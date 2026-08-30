package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/progression"
)

type SyncService struct {
	uow domain.UnitOfWork
	gam *domain.GamificationService

	// Lecturas del plan y del catalogo de reglas, de solo lectura y previas a
	// la transaccion (mismo patron que AssignmentService.Assign).
	assignments domain.AssignmentRepository
	programs    domain.ProgramRepository
}

func NewSyncService(uow domain.UnitOfWork, gam *domain.GamificationService, assignments domain.AssignmentRepository, programs domain.ProgramRepository) *SyncService {
	return &SyncService{uow: uow, gam: gam, assignments: assignments, programs: programs}
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

	// El plan y las reglas de progresion son datos de solo lectura: se leen
	// antes de abrir la transaccion, igual que la plantilla en Assign.
	plans, err := s.loadPlanContext(ctx, sessions)
	if err != nil {
		return SyncResult{}, err
	}

	result := SyncResult{}

	err = s.uow.Execute(ctx, func(repos domain.TxRepos) error {
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
				if err := s.progressPlan(ctx, repos, sess, plans); err != nil {
					return err
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

// planContext es lo que hace falta saber del plan para progresar un dia
// entrenado, leido antes de la transaccion.
type planContext struct {
	assignmentID uuid.UUID
	// exercises indexa los ejercicios de ese dia por exercise_id, para poder
	// cruzar lo que el movil reporto (que solo trae exercise_id) con el
	// ejercicio asignado que lo origino.
	exercises map[uuid.UUID]domain.AssignedExercise
	rules     map[uuid.UUID]domain.ProgressionRule
}

func (s *SyncService) loadPlanContext(ctx context.Context, sessions []SyncSessionInput) (map[uuid.UUID]planContext, error) {
	out := map[uuid.UUID]planContext{}

	for _, sess := range sessions {
		if sess.AssignedWorkoutID == nil || sess.Status != string(domain.SessionStatusCompleted) {
			continue
		}
		if _, done := out[*sess.AssignedWorkoutID]; done {
			continue
		}

		assignmentID, err := s.assignments.GetWorkoutAssignmentID(ctx, *sess.AssignedWorkoutID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue // entrenamiento libre o plan ya cancelado: nada que progresar
			}
			return nil, err
		}

		exercises, err := s.assignments.ListExercises(ctx, *sess.AssignedWorkoutID)
		if err != nil {
			return nil, err
		}

		pc := planContext{
			assignmentID: assignmentID,
			exercises:    make(map[uuid.UUID]domain.AssignedExercise, len(exercises)),
			rules:        map[uuid.UUID]domain.ProgressionRule{},
		}
		for _, ex := range exercises {
			pc.exercises[ex.ExerciseID] = ex
			if ex.ProgressionRuleID == nil {
				continue
			}
			if _, cached := pc.rules[*ex.ProgressionRuleID]; cached {
				continue
			}
			rule, err := s.programs.GetProgressionRuleByID(ctx, *ex.ProgressionRuleID)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					continue // regla borrada: el ejercicio simplemente no progresa
				}
				return nil, err
			}
			pc.rules[*ex.ProgressionRuleID] = rule
		}
		out[*sess.AssignedWorkoutID] = pc
	}
	return out, nil
}

// progressPlan cierra el dia entrenado y escribe el objetivo de la proxima
// ocurrencia de cada ejercicio que tenga regla de progresion.
func (s *SyncService) progressPlan(ctx context.Context, repos domain.TxRepos, sess SyncSessionInput, plans map[uuid.UUID]planContext) error {
	if sess.AssignedWorkoutID == nil {
		return nil // entrenamiento libre
	}
	pc, ok := plans[*sess.AssignedWorkoutID]
	if !ok {
		return nil
	}

	if err := repos.Assignments.MarkWorkoutCompleted(ctx, *sess.AssignedWorkoutID); err != nil {
		return err
	}

	for _, ex := range sess.Exercises {
		assigned, ok := pc.exercises[ex.ExerciseID]
		if !ok || assigned.ProgressionRuleID == nil {
			continue // ejercicio agregado a mano, o sin regla: no progresa
		}
		rule, ok := pc.rules[*assigned.ProgressionRuleID]
		if !ok {
			continue
		}

		next, found, err := repos.Assignments.FindNextExerciseOccurrence(ctx, pc.assignmentID, ex.ExerciseID, *sess.AssignedWorkoutID)
		if err != nil {
			return err
		}
		if !found {
			continue // ultima vez que aparece en el plan
		}

		// Una sugerencia de IA aprobada gana esta vez. Se consume el flag para
		// que a partir de la proxima sesion vuelva el calculo automatico.
		if next.OverrideSource != nil {
			if _, err := repos.Assignments.UpdateExerciseTargets(ctx, next.ID, next.TargetWeightKg, next.TargetRepsMin, next.TargetRepsMax, nil); err != nil {
				return err
			}
			continue
		}

		out, err := progression.Apply(progression.Type(rule.Type), rule.Params, progressionInput(assigned, ex))
		if err != nil {
			// Una regla con type invalido es dato malo, no una falla del sync:
			// el resto de la sesion ya se guardo bien.
			continue
		}

		weight := out.NextWeightKg
		repsMin, repsMax := out.NextRepsMin, out.NextRepsMax
		if _, err := repos.Assignments.UpdateExerciseTargets(ctx, next.ID, &weight, &repsMin, &repsMax, nil); err != nil {
			return err
		}
	}
	return nil
}

// progressionInput traduce lo que el cliente reporto a la entrada del motor.
// Success = todas las series de trabajo alcanzaron el piso del rango.
func progressionInput(assigned domain.AssignedExercise, ex SyncExerciseInput) progression.Input {
	in := progression.Input{
		TargetRepsMin: derefInt(assigned.TargetRepsMin),
		TargetRepsMax: derefInt(assigned.TargetRepsMax),
	}
	if assigned.TargetWeightKg != nil {
		in.LastWeightKg = *assigned.TargetWeightKg
	}
	if assigned.TargetRPE != nil {
		in.TargetRPE = *assigned.TargetRPE
	}

	var rpeSum float64
	var rpeCount int
	for _, set := range ex.Sets {
		if (set.Type != "working" && set.Type != "") || !set.IsCompleted {
			continue
		}
		if set.Reps != nil {
			in.RepsAchieved = append(in.RepsAchieved, *set.Reps)
		}
		// El peso real levantado manda sobre el prescrito: si el cliente
		// cambio la carga, la progresion parte de lo que efectivamente hizo.
		if set.WeightKg != nil {
			in.LastWeightKg = *set.WeightKg
		}
		if set.RPE != nil {
			rpeSum += *set.RPE
			rpeCount++
		}
	}
	if rpeCount > 0 {
		in.ActualRPE = rpeSum / float64(rpeCount)
	}

	in.Success = len(in.RepsAchieved) > 0
	for _, reps := range in.RepsAchieved {
		if reps < in.TargetRepsMin {
			in.Success = false
			break
		}
	}
	return in
}

func derefInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
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
