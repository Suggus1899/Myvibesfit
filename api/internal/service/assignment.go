package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type AssignmentService struct {
	repo        domain.AssignmentRepository
	programRepo domain.ProgramRepository
	uow         domain.UnitOfWork
	notifier    userNotifier
	audit       *AuditLogger
}

// userNotifier es el subconjunto de NotificationService que este caso de uso
// necesita. Con una interfaz el test no tiene que armar toda la cadena de push.
type userNotifier interface {
	NotifyUser(ctx context.Context, userID uuid.UUID, n domain.Notification)
}

func NewAssignmentService(repo domain.AssignmentRepository, programRepo domain.ProgramRepository, uow domain.UnitOfWork, audit *AuditLogger, notifier userNotifier) *AssignmentService {
	return &AssignmentService{repo: repo, programRepo: programRepo, uow: uow, audit: audit, notifier: notifier}
}

type AssignInput struct {
	OrgID        uuid.UUID
	ProgramID    uuid.UUID
	ClientUserID uuid.UUID
	CoachUserID  uuid.UUID
	StartDate    time.Time
}

// Assign copia un programa publicado a una asignacion propia del cliente.
// La lectura de la plantilla (programa/dias/ejercicios) es de solo lectura y
// corre antes de abrir transaccion; la escritura de assignment/
// assigned_workout/assigned_exercise corre entera dentro de uow.Execute: si
// algo falla a mitad de copia, no queda una asignacion a medias.
func (s *AssignmentService) Assign(ctx context.Context, in AssignInput) (domain.Assignment, error) {
	program, err := s.programRepo.GetByID(ctx, in.ProgramID, in.OrgID)
	if err != nil {
		return domain.Assignment{}, err
	}
	if program.Status != domain.ProgramStatusPublished {
		return domain.Assignment{}, fmt.Errorf("%w: el programa debe estar publicado antes de asignarlo", domain.ErrInvalidInput)
	}

	if _, err := s.repo.GetOrgMembership(ctx, in.OrgID, in.ClientUserID); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Assignment{}, fmt.Errorf("%w: el cliente no pertenece a este gimnasio", domain.ErrInvalidInput)
		}
		return domain.Assignment{}, err
	}

	if _, err := s.repo.GetActiveByClient(ctx, in.ClientUserID); err == nil {
		return domain.Assignment{}, fmt.Errorf("%w: el cliente ya tiene una asignacion activa", domain.ErrAlreadyExists)
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.Assignment{}, err
	}

	workouts, err := s.programRepo.ListWorkouts(ctx, in.ProgramID)
	if err != nil {
		return domain.Assignment{}, err
	}
	exercisesByWorkout := make(map[uuid.UUID][]domain.ProgramExercise, len(workouts))
	for _, w := range workouts {
		exercises, err := s.programRepo.ListExercises(ctx, w.ID)
		if err != nil {
			return domain.Assignment{}, err
		}
		exercisesByWorkout[w.ID] = exercises
	}

	var assignment domain.Assignment
	err = s.uow.Execute(ctx, func(repos domain.TxRepos) error {
		programID, coachID := in.ProgramID, in.CoachUserID
		a, err := repos.Assignments.Create(ctx, domain.Assignment{
			OrgID: in.OrgID, ProgramID: &programID, ClientUserID: in.ClientUserID,
			CoachUserID: &coachID, Name: program.Name, StartDate: in.StartDate,
		})
		if err != nil {
			return err
		}
		assignment = a

		if err := repos.Coaches.LinkClient(ctx, in.OrgID, in.CoachUserID, in.ClientUserID); err != nil {
			return err
		}

		for _, w := range workouts {
			// day_index es un offset dentro de la semana (1-7), no un dia de
			// calendario fijo: la fecha real se ancla a start_date.
			scheduledOn := in.StartDate.AddDate(0, 0, (int(w.WeekNumber)-1)*7+(int(w.DayIndex)-1))
			sourceWorkoutID := w.ID
			aw, err := repos.Assignments.CreateWorkout(ctx, domain.AssignedWorkout{
				AssignmentID: a.ID, SourceWorkoutID: &sourceWorkoutID, WeekNumber: w.WeekNumber, DayIndex: w.DayIndex,
				Name: w.Name, Note: w.Note, IsDeload: w.IsDeload, ScheduledOn: &scheduledOn,
			})
			if err != nil {
				return err
			}

			for _, pe := range exercisesByWorkout[w.ID] {
				if _, err := repos.Assignments.CreateExercise(ctx, domain.AssignedExercise{
					AssignedWorkoutID: aw.ID, ExerciseID: pe.ExerciseID, OrderIndex: pe.OrderIndex,
					SupersetGroup: pe.SupersetGroup, TargetSets: pe.TargetSets, TargetRepsMin: pe.TargetRepsMin,
					TargetRepsMax: pe.TargetRepsMax, TargetRPE: pe.TargetRPE,
					// target_pct_1rm de la plantilla se resuelve a un peso
					// concreto cuando exista 1RM del cliente (fase 6+); por
					// ahora queda sin objetivo de peso fijo.
					TargetWeightKg: nil, RestSeconds: pe.RestSeconds, Tempo: pe.Tempo, Note: pe.Note,
					ProgressionRuleID: pe.ProgressionRuleID,
				}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return domain.Assignment{}, err
	}

	s.audit.Log(ctx, in.OrgID, in.CoachUserID, "assignment.assign", "assignment", assignment.ID.String(), map[string]any{
		"program_id": in.ProgramID, "client_user_id": in.ClientUserID,
	})
	// Despues del commit, nunca dentro de uow.Execute: una llamada de red al
	// proveedor de push mantendria la transaccion abierta a merced de su latencia.
	s.notifier.NotifyUser(ctx, in.ClientUserID, domain.PlanAssignedNotification(program.Name))
	return assignment, nil
}

func (s *AssignmentService) Cancel(ctx context.Context, id, orgID, actorID uuid.UUID) (domain.Assignment, error) {
	a, err := s.repo.Cancel(ctx, id, orgID)
	if err != nil {
		return domain.Assignment{}, err
	}
	s.audit.Log(ctx, orgID, actorID, "assignment.cancel", "assignment", id.String(), nil)
	return a, nil
}

type AssignedWorkoutDetail struct {
	Workout   domain.AssignedWorkout
	Exercises []domain.AssignedExercise
}

type AssignmentDetail struct {
	Assignment domain.Assignment
	Workouts   []AssignedWorkoutDetail
}

func (s *AssignmentService) GetCurrentForClient(ctx context.Context, clientID uuid.UUID) (AssignmentDetail, error) {
	a, err := s.repo.GetActiveByClient(ctx, clientID)
	if err != nil {
		return AssignmentDetail{}, err
	}

	workouts, err := s.repo.ListWorkouts(ctx, a.ID)
	if err != nil {
		return AssignmentDetail{}, err
	}

	detail := AssignmentDetail{Assignment: a, Workouts: make([]AssignedWorkoutDetail, 0, len(workouts))}
	for _, w := range workouts {
		exercises, err := s.repo.ListExercises(ctx, w.ID)
		if err != nil {
			return AssignmentDetail{}, err
		}
		detail.Workouts = append(detail.Workouts, AssignedWorkoutDetail{Workout: w, Exercises: exercises})
	}
	return detail, nil
}
