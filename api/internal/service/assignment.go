package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type AssignmentService struct {
	q    db.Querier
	pool *pgxpool.Pool
}

func NewAssignmentService(q db.Querier, pool *pgxpool.Pool) *AssignmentService {
	return &AssignmentService{q: q, pool: pool}
}

type AssignInput struct {
	OrgID        uuid.UUID
	ProgramID    uuid.UUID
	ClientUserID uuid.UUID
	CoachUserID  uuid.UUID
	StartDate    time.Time
}

// Assign copia un programa publicado a una asignacion propia del cliente.
// Todo corre en una transaccion: si algo falla a mitad de copia, no queda
// una asignacion a medias. Editar la copia despues nunca toca la plantilla
// (program/program_workout/program_exercise), porque assigned_* son filas
// independientes.
func (s *AssignmentService) Assign(ctx context.Context, in AssignInput) (db.Assignment, error) {
	program, err := s.q.GetProgramByID(ctx, db.GetProgramByIDParams{ID: in.ProgramID, OrgID: in.OrgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Assignment{}, domain.ErrNotFound
		}
		return db.Assignment{}, err
	}
	if program.Status != db.ProgramStatusPublished {
		return db.Assignment{}, fmt.Errorf("%w: el programa debe estar publicado antes de asignarlo", domain.ErrInvalidInput)
	}

	if _, err := s.q.GetOrgMembership(ctx, db.GetOrgMembershipParams{OrgID: in.OrgID, UserID: in.ClientUserID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Assignment{}, fmt.Errorf("%w: el cliente no pertenece a este gimnasio", domain.ErrInvalidInput)
		}
		return db.Assignment{}, err
	}

	if _, err := s.q.GetActiveAssignmentByClient(ctx, in.ClientUserID); err == nil {
		return db.Assignment{}, fmt.Errorf("%w: el cliente ya tiene una asignacion activa", domain.ErrAlreadyExists)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return db.Assignment{}, err
	}

	workouts, err := s.q.ListProgramWorkouts(ctx, in.ProgramID)
	if err != nil {
		return db.Assignment{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return db.Assignment{}, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op si ya hizo commit
	qtx := db.New(tx)

	assignment, err := qtx.CreateAssignment(ctx, db.CreateAssignmentParams{
		OrgID:        in.OrgID,
		ProgramID:    uuidToPg(&in.ProgramID),
		ClientUserID: in.ClientUserID,
		CoachUserID:  uuidToPg(&in.CoachUserID),
		Name:         program.Name,
		StartDate:    in.StartDate,
	})
	if err != nil {
		return db.Assignment{}, err
	}

	for _, w := range workouts {
		exercises, err := qtx.ListProgramExercises(ctx, w.ID)
		if err != nil {
			return db.Assignment{}, err
		}

		// day_index es un offset dentro de la semana (1-7), no un dia de
		// calendario fijo: la fecha real se ancla a start_date.
		scheduledOn := in.StartDate.AddDate(0, 0, (int(w.WeekNumber)-1)*7+(int(w.DayIndex)-1))
		aw, err := qtx.CreateAssignedWorkout(ctx, db.CreateAssignedWorkoutParams{
			AssignmentID:    assignment.ID,
			SourceWorkoutID: uuidToPg(&w.ID),
			WeekNumber:      w.WeekNumber,
			DayIndex:        w.DayIndex,
			Name:            w.Name,
			Note:            w.Note,
			IsDeload:        w.IsDeload,
			ScheduledOn:     &scheduledOn,
		})
		if err != nil {
			return db.Assignment{}, err
		}

		for _, pe := range exercises {
			if _, err := qtx.CreateAssignedExercise(ctx, db.CreateAssignedExerciseParams{
				AssignedWorkoutID: aw.ID,
				ExerciseID:        pe.ExerciseID,
				OrderIndex:        pe.OrderIndex,
				SupersetGroup:     pe.SupersetGroup,
				TargetSets:        pe.TargetSets,
				TargetRepsMin:     pe.TargetRepsMin,
				TargetRepsMax:     pe.TargetRepsMax,
				TargetRpe:         pe.TargetRpe,
				// target_pct_1rm de la plantilla se resuelve a un peso
				// concreto cuando exista 1RM del cliente (fase 6+); por
				// ahora queda sin objetivo de peso fijo.
				TargetWeightKg:    nil,
				RestSeconds:       pe.RestSeconds,
				Tempo:             pe.Tempo,
				Note:              pe.Note,
				ProgressionRuleID: pe.ProgressionRuleID,
			}); err != nil {
				return db.Assignment{}, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return db.Assignment{}, err
	}
	return assignment, nil
}

func (s *AssignmentService) Cancel(ctx context.Context, id, orgID uuid.UUID) (db.Assignment, error) {
	a, err := s.q.CancelAssignment(ctx, db.CancelAssignmentParams{ID: id, OrgID: orgID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Assignment{}, domain.ErrNotFound
		}
		return db.Assignment{}, err
	}
	return a, nil
}

type AssignedWorkoutDetail struct {
	Workout   db.AssignedWorkout
	Exercises []db.AssignedExercise
}

type AssignmentDetail struct {
	Assignment db.Assignment
	Workouts   []AssignedWorkoutDetail
}

func (s *AssignmentService) GetCurrentForClient(ctx context.Context, clientID uuid.UUID) (AssignmentDetail, error) {
	a, err := s.q.GetActiveAssignmentByClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AssignmentDetail{}, domain.ErrNotFound
		}
		return AssignmentDetail{}, err
	}

	workouts, err := s.q.ListAssignedWorkouts(ctx, a.ID)
	if err != nil {
		return AssignmentDetail{}, err
	}

	detail := AssignmentDetail{Assignment: a, Workouts: make([]AssignedWorkoutDetail, 0, len(workouts))}
	for _, w := range workouts {
		exercises, err := s.q.ListAssignedExercises(ctx, w.ID)
		if err != nil {
			return AssignmentDetail{}, err
		}
		detail.Workouts = append(detail.Workouts, AssignedWorkoutDetail{Workout: w, Exercises: exercises})
	}
	return detail, nil
}
