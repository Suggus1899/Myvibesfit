package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Assignment struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	ProgramID    *uuid.UUID
	ClientUserID uuid.UUID
	CoachUserID  *uuid.UUID
	Name         string
	StartDate    time.Time
	EndDate      *time.Time
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AssignedWorkout struct {
	ID              uuid.UUID
	AssignmentID    uuid.UUID
	SourceWorkoutID *uuid.UUID
	WeekNumber      int16
	DayIndex        int16
	Name            string
	Note            string
	IsDeload        bool
	ScheduledOn     *time.Time
	Status          string
}

type AssignedExercise struct {
	ID                uuid.UUID
	AssignedWorkoutID uuid.UUID
	ExerciseID        uuid.UUID
	OrderIndex        int16
	SupersetGroup     *int
	TargetSets        int16
	TargetRepsMin     *int
	TargetRepsMax     *int
	TargetRPE         *float64
	TargetWeightKg    *float64
	RestSeconds       int16
	Tempo             string
	Note              string
	ProgressionRuleID *uuid.UUID
}

// AssignmentRepository espeja assignment.sql.go (9 metodos). Incluye
// GetOrgMembership porque solo lo usa AssignmentService.Assign para validar
// que el cliente pertenece al gimnasio antes de asignarle un programa.
type AssignmentRepository interface {
	Create(ctx context.Context, a Assignment) (Assignment, error)
	GetByID(ctx context.Context, id uuid.UUID) (Assignment, error)
	GetActiveByClient(ctx context.Context, clientUserID uuid.UUID) (Assignment, error)
	Cancel(ctx context.Context, id, orgID uuid.UUID) (Assignment, error)

	CreateWorkout(ctx context.Context, w AssignedWorkout) (AssignedWorkout, error)
	ListWorkouts(ctx context.Context, assignmentID uuid.UUID) ([]AssignedWorkout, error)
	CreateExercise(ctx context.Context, e AssignedExercise) (AssignedExercise, error)
	ListExercises(ctx context.Context, assignedWorkoutID uuid.UUID) ([]AssignedExercise, error)

	GetOrgMembership(ctx context.Context, orgID, userID uuid.UUID) (Membership, error)
}
