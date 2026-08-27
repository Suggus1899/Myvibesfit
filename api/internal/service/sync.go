package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type SyncService struct {
	uow domain.UnitOfWork
	gam *domain.GamificationService
}

func NewSyncService(uow domain.UnitOfWork, gam *domain.GamificationService) *SyncService {
	return &SyncService{uow: uow, gam: gam}
}

type SyncSetInput struct {
	ClientLocalID    uuid.UUID
	SetNumber        int
	Type             string
	WeightKg         *float64
	Reps             *int
	RPE              *float64
	RIR              *int
	DurationSeconds  *int
	DistanceM        *float64
	RestTakenSeconds *int
	IsCompleted      bool
	PerformedAt      time.Time
}

type SyncExerciseInput struct {
	ExerciseID         uuid.UUID
	AssignedExerciseID *uuid.UUID
	OrderIndex         int
	SupersetGroup      *int
	Note               string
	Sets               []SyncSetInput
}

type SyncSessionInput struct {
	ClientLocalID     uuid.UUID
	Name              string
	Status            string
	AssignedWorkoutID *uuid.UUID
	StartedAt         time.Time
	EndedAt           *time.Time
	DurationSeconds   *int
	PerceivedEffort   *int
	Mood              *int
	Notes             string
	Exercises         []SyncExerciseInput
}

type SyncResult struct {
	Sessions             []domain.WorkoutSession
	NewPersonalRecords   int
	UnlockedAchievements []domain.Achievement
}

// SyncSessions aplica un lote de sesiones de entrenamiento generadas en el
// movil (posiblemente offline). Cada sesion/serie trae su client_local_id:
// reenviar el mismo lote nunca duplica filas. Todo corre dentro de
// uow.Execute (XP/racha/logros incluidos), para no dejar estadisticas a
// medias si algo falla.
func (s *SyncService) SyncSessions(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID, sessions []SyncSessionInput) (SyncResult, error) {
	if len(sessions) == 0 {
		return SyncResult{}, domain.ErrInvalidInput
	}

	result := SyncResult{}

	err := s.uow.Execute(ctx, func(repos domain.TxRepos) error {
		var completedCount int32
		var completedVolume float64
		var lastActiveDate time.Time
		var newPRs bool

		for _, sess := range sessions {
			volume := sessionVolume(sess)

			row, err := repos.Sessions.UpsertWorkoutSession(ctx, domain.WorkoutSession{
				ClientLocalID: sess.ClientLocalID, UserID: userID, OrgID: orgID,
				AssignedWorkoutID: sess.AssignedWorkoutID, Name: sess.Name, Status: domain.SessionStatus(sess.Status),
				StartedAt: sess.StartedAt, EndedAt: sess.EndedAt, DurationSeconds: sess.DurationSeconds,
				TotalVolumeKg: volume, PerceivedEffort: sess.PerceivedEffort, Mood: sess.Mood, Notes: sess.Notes,
			})
			if err != nil {
				return err
			}
			result.Sessions = append(result.Sessions, row)

			for _, ex := range sess.Exercises {
				se, err := repos.Sessions.UpsertSessionExercise(ctx, domain.SessionExercise{
					SessionID: row.ID, ExerciseID: ex.ExerciseID, AssignedExerciseID: ex.AssignedExerciseID,
					OrderIndex: int16(ex.OrderIndex), SupersetGroup: ex.SupersetGroup, Note: ex.Note,
				})
				if err != nil {
					return err
				}

				for _, set := range ex.Sets {
					setType := domain.SetType(set.Type)
					if setType == "" {
						setType = domain.SetTypeWorking
					}
					setRow, err := repos.Sessions.UpsertSetLog(ctx, domain.SetLog{
						ClientLocalID: set.ClientLocalID, SessionExerciseID: se.ID, UserID: userID, ExerciseID: ex.ExerciseID,
						SetNumber: int16(set.SetNumber), Type: setType, WeightKg: set.WeightKg, Reps: set.Reps,
						RPE: set.RPE, RIR: set.RIR, DurationSeconds: set.DurationSeconds, DistanceM: set.DistanceM,
						RestTakenSeconds: set.RestTakenSeconds, IsCompleted: set.IsCompleted, PerformedAt: set.PerformedAt,
					})
					if err != nil {
						return err
					}

					if setType == domain.SetTypeWorking && set.IsCompleted && set.WeightKg != nil && set.Reps != nil {
						beat, err := s.checkPersonalRecords(ctx, repos, userID, ex.ExerciseID, setRow.ID, *set.WeightKg, *set.Reps, set.PerformedAt)
						if err != nil {
							return err
						}
						if beat {
							newPRs = true
							result.NewPersonalRecords++
						}
					}
				}
			}

			if sess.Status == string(domain.SessionStatusCompleted) {
				completedCount++
				completedVolume += volume
				if sess.StartedAt.After(lastActiveDate) {
					lastActiveDate = sess.StartedAt
				}
			}
		}

		xpEarned := completedCount * domain.XPPerSession
		if newPRs {
			xpEarned += int32(result.NewPersonalRecords) * domain.XPPerPersonalRecord
		}

		var stats domain.UserStat
		var err error
		if completedCount > 0 || newPRs {
			if completedCount > 0 {
				if err := s.gam.RecordXPEvent(ctx, repos.Gamification, userID, domain.XpSourceSession, completedCount*domain.XPPerSession, ""); err != nil {
					return err
				}
			}
			if newPRs {
				if err := s.gam.RecordXPEvent(ctx, repos.Gamification, userID, domain.XpSourcePersonalRecord, int32(result.NewPersonalRecords)*domain.XPPerPersonalRecord, ""); err != nil {
					return err
				}
			}
			stats, err = s.gam.ApplyStatsDelta(ctx, repos.Gamification, userID, domain.StatsDelta{XP: xpEarned, Sessions: completedCount, VolumeKg: completedVolume})
			if err != nil {
				return err
			}
		} else {
			stats, err = repos.Gamification.GetUserStats(ctx, userID)
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return err
			}
		}

		var workoutStreak domain.UserStreak
		if completedCount > 0 {
			workoutStreak, err = s.gam.BumpStreak(ctx, repos.Gamification, userID, domain.StreakKindWorkout, lastActiveDate)
			if err != nil {
				return err
			}
			if _, err := s.gam.BumpStreak(ctx, repos.Gamification, userID, domain.StreakKindOverall, lastActiveDate); err != nil {
				return err
			}
		}

		if completedCount > 0 || newPRs {
			unlocked, err := s.gam.CheckAchievements(ctx, repos.Gamification, userID, domain.AchievementSnapshot{
				TotalSessions: int(stats.TotalSessions), WorkoutStreak: int(workoutStreak.CurrentCount), HasNewPR: newPRs,
			})
			if err != nil {
				return err
			}
			if err := s.gam.GrantAchievementXP(ctx, repos.Gamification, userID, unlocked); err != nil {
				return err
			}
			result.UnlockedAchievements = unlocked
		}
		return nil
	})
	if err != nil {
		return SyncResult{}, err
	}
	return result, nil
}

func (s *SyncService) checkPersonalRecords(ctx context.Context, repos domain.TxRepos, userID, exerciseID uuid.UUID, setLogID int64, weightKg float64, reps int, performedAt time.Time) (bool, error) {
	beat := false
	candidates := map[domain.RecordType]float64{
		domain.RecordTypeMaxWeight:    weightKg,
		domain.RecordTypeMaxReps:      float64(reps),
		domain.RecordTypeEstimated1RM: weightKg * (1 + float64(reps)/30.0),
		domain.RecordTypeMaxVolumeSet: weightKg * float64(reps),
	}
	for recordType, value := range candidates {
		current, err := repos.Sessions.GetPersonalRecord(ctx, userID, exerciseID, recordType)
		if err != nil && !errors.Is(err, domain.ErrNotFound) {
			return false, err
		}
		if err == nil && current.Value >= value {
			continue
		}
		logID := setLogID
		if _, err := repos.Sessions.UpsertPersonalRecord(ctx, domain.PersonalRecord{
			UserID: userID, ExerciseID: exerciseID, Type: recordType, Value: value, SetLogID: &logID, AchievedAt: performedAt,
		}); err != nil {
			return false, err
		}
		beat = true
	}
	return beat, nil
}

func sessionVolume(sess SyncSessionInput) float64 {
	var total float64
	for _, ex := range sess.Exercises {
		for _, set := range ex.Sets {
			if set.Type != "working" && set.Type != "" {
				continue
			}
			if !set.IsCompleted || set.WeightKg == nil || set.Reps == nil {
				continue
			}
			total += *set.WeightKg * float64(*set.Reps)
		}
	}
	return total
}
