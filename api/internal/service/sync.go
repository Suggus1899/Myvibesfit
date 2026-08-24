package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type SyncService struct {
	q    db.Querier
	pool *pgxpool.Pool
	gam  *GamificationService
}

func NewSyncService(q db.Querier, pool *pgxpool.Pool, gam *GamificationService) *SyncService {
	return &SyncService{q: q, pool: pool, gam: gam}
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
	Sessions             []db.WorkoutSession
	NewPersonalRecords   int
	UnlockedAchievements []db.Achievement
}

// SyncSessions aplica un lote de sesiones de entrenamiento generadas en el
// movil (posiblemente offline). Cada sesion/serie trae su client_local_id:
// reenviar el mismo lote nunca duplica filas. Todo corre en una sola
// transaccion, incluyendo XP/racha/logros, para no dejar estadisticas a
// medias si algo falla.
func (s *SyncService) SyncSessions(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID, sessions []SyncSessionInput) (SyncResult, error) {
	if len(sessions) == 0 {
		return SyncResult{}, domain.ErrInvalidInput
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return SyncResult{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	qtx := db.New(tx)

	result := SyncResult{}
	var completedCount int32
	var completedVolume float64
	var lastActiveDate time.Time
	var newPRs bool

	for _, sess := range sessions {
		volume := sessionVolume(sess)

		row, err := qtx.UpsertWorkoutSession(ctx, db.UpsertWorkoutSessionParams{
			ClientLocalID:     sess.ClientLocalID,
			UserID:            userID,
			OrgID:             uuidToPg(orgID),
			AssignedWorkoutID: uuidToPg(sess.AssignedWorkoutID),
			Name:              sess.Name,
			Status:            db.SessionStatus(sess.Status),
			StartedAt:         sess.StartedAt,
			EndedAt:           sess.EndedAt,
			DurationSeconds:   intToPgInt4(sess.DurationSeconds),
			TotalVolumeKg:     volume,
			PerceivedEffort:   intToPgInt2(sess.PerceivedEffort),
			Mood:              intToPgInt2(sess.Mood),
			Notes:             textToPg(sess.Notes),
		})
		if err != nil {
			return SyncResult{}, err
		}
		result.Sessions = append(result.Sessions, row)

		for _, ex := range sess.Exercises {
			se, err := qtx.UpsertSessionExercise(ctx, db.UpsertSessionExerciseParams{
				SessionID:          row.ID,
				ExerciseID:         ex.ExerciseID,
				AssignedExerciseID: uuidToPg(ex.AssignedExerciseID),
				OrderIndex:         int16(ex.OrderIndex),
				SupersetGroup:      intToPgInt2(ex.SupersetGroup),
				Note:               textToPg(ex.Note),
			})
			if err != nil {
				return SyncResult{}, err
			}

			for _, set := range ex.Sets {
				setType := db.SetType(set.Type)
				if setType == "" {
					setType = db.SetTypeWorking
				}
				setRow, err := qtx.UpsertSetLog(ctx, db.UpsertSetLogParams{
					ClientLocalID:     set.ClientLocalID,
					SessionExerciseID: se.ID,
					UserID:            userID,
					ExerciseID:        ex.ExerciseID,
					SetNumber:         int16(set.SetNumber),
					Type:              setType,
					WeightKg:          set.WeightKg,
					Reps:              intToPgInt2(set.Reps),
					Rpe:               set.RPE,
					Rir:               intToPgInt2(set.RIR),
					DurationSeconds:   intToPgInt4(set.DurationSeconds),
					DistanceM:         set.DistanceM,
					RestTakenSeconds:  intToPgInt4(set.RestTakenSeconds),
					IsCompleted:       set.IsCompleted,
					PerformedAt:       set.PerformedAt,
				})
				if err != nil {
					return SyncResult{}, err
				}

				if setType == db.SetTypeWorking && set.IsCompleted && set.WeightKg != nil && set.Reps != nil {
					beat, err := s.checkPersonalRecords(ctx, qtx, userID, ex.ExerciseID, setRow.ID, *set.WeightKg, *set.Reps, set.PerformedAt)
					if err != nil {
						return SyncResult{}, err
					}
					if beat {
						newPRs = true
						result.NewPersonalRecords++
					}
				}
			}
		}

		if sess.Status == string(db.SessionStatusCompleted) {
			completedCount++
			completedVolume += volume
			if sess.StartedAt.After(lastActiveDate) {
				lastActiveDate = sess.StartedAt
			}
		}
	}

	xpEarned := completedCount * xpPerSession
	if newPRs {
		xpEarned += int32(result.NewPersonalRecords) * xpPerPersonalRecord
	}

	var stats db.UserStat
	if completedCount > 0 || newPRs {
		if completedCount > 0 {
			if err := s.gam.RecordXPEvent(ctx, qtx, userID, db.XpSourceSession, completedCount*xpPerSession, ""); err != nil {
				return SyncResult{}, err
			}
		}
		if newPRs {
			if err := s.gam.RecordXPEvent(ctx, qtx, userID, db.XpSourcePersonalRecord, int32(result.NewPersonalRecords)*xpPerPersonalRecord, ""); err != nil {
				return SyncResult{}, err
			}
		}
		stats, err = s.gam.ApplyStatsDelta(ctx, qtx, userID, StatsDelta{XP: xpEarned, Sessions: completedCount, VolumeKg: completedVolume})
		if err != nil {
			return SyncResult{}, err
		}
	} else {
		stats, err = qtx.GetUserStats(ctx, userID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return SyncResult{}, err
		}
	}

	workoutStreak := db.UserStreak{}
	if completedCount > 0 {
		workoutStreak, err = s.gam.BumpStreak(ctx, qtx, userID, db.StreakKindWorkout, lastActiveDate)
		if err != nil {
			return SyncResult{}, err
		}
		if _, err := s.gam.BumpStreak(ctx, qtx, userID, db.StreakKindOverall, lastActiveDate); err != nil {
			return SyncResult{}, err
		}
	}

	if completedCount > 0 || newPRs {
		unlocked, err := s.gam.CheckAchievements(ctx, qtx, userID, AchievementSnapshot{
			TotalSessions: int(stats.TotalSessions), WorkoutStreak: int(workoutStreak.CurrentCount), HasNewPR: newPRs,
		})
		if err != nil {
			return SyncResult{}, err
		}
		if err := s.gam.GrantAchievementXP(ctx, qtx, userID, unlocked); err != nil {
			return SyncResult{}, err
		}
		result.UnlockedAchievements = unlocked
	}

	if err := tx.Commit(ctx); err != nil {
		return SyncResult{}, err
	}
	return result, nil
}

func (s *SyncService) checkPersonalRecords(ctx context.Context, qtx db.Querier, userID, exerciseID uuid.UUID, setLogID int64, weightKg float64, reps int, performedAt time.Time) (bool, error) {
	beat := false
	candidates := map[db.RecordType]float64{
		db.RecordTypeMaxWeight:    weightKg,
		db.RecordTypeMaxReps:      float64(reps),
		db.RecordTypeEstimated1rm: weightKg * (1 + float64(reps)/30.0),
		db.RecordTypeMaxVolumeSet: weightKg * float64(reps),
	}
	for recordType, value := range candidates {
		current, err := qtx.GetPersonalRecord(ctx, db.GetPersonalRecordParams{UserID: userID, ExerciseID: exerciseID, Type: recordType})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}
		if err == nil && current.Value >= value {
			continue
		}
		if _, err := qtx.UpsertPersonalRecord(ctx, db.UpsertPersonalRecordParams{
			UserID: userID, ExerciseID: exerciseID, Type: recordType, Value: value,
			SetLogID: pgtype.Int8{Int64: setLogID, Valid: true}, AchievedAt: performedAt,
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

func intToPgInt4(v *int) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*v), Valid: true}
}
