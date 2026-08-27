package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ProgramStatus string

const (
	ProgramStatusDraft     ProgramStatus = "draft"
	ProgramStatusPublished ProgramStatus = "published"
	ProgramStatusArchived  ProgramStatus = "archived"
)

const DefaultTrainingGoal = "general_health"

type Program struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	CreatedBy   uuid.UUID
	Name        string
	Description string
	Goal        string
	Level       string
	TotalWeeks  int16
	DaysPerWeek int16
	Status      ProgramStatus
	IsShared    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ListProgramsFilter struct {
	OrgID  uuid.UUID
	Status string
	Limit  int32
	Offset int32
}

type ProgramWorkout struct {
	ID               uuid.UUID
	ProgramID        uuid.UUID
	WeekNumber       int16
	DayIndex         int16
	Name             string
	Note             string
	IsDeload         bool
	EstimatedMinutes *int
}

type ProgramExercise struct {
	ID                uuid.UUID
	ProgramWorkoutID  uuid.UUID
	ExerciseID        uuid.UUID
	OrderIndex        int16
	SupersetGroup     *int
	TargetSets        int16
	TargetRepsMin     *int
	TargetRepsMax     *int
	TargetRPE         *float64
	TargetPct1RM      *float64
	RestSeconds       int16
	Tempo             string
	Note              string
	ProgressionRuleID *uuid.UUID
}

type ProgressionRule struct {
	ID        uuid.UUID
	OrgID     *uuid.UUID
	Name      string
	Type      string
	Params    []byte
	IsSystem  bool
	CreatedAt time.Time
}

// ProgramRepository espeja program.sql.go (18 metodos: Program, ProgramWorkout,
// ProgramExercise, ProgressionRule). GetWorkoutOrgID/GetExerciseOrgID
// reemplazan a GetProgramWorkoutWithOrg/GetProgramExerciseWithOrg: en el
// codigo original esas queries devuelven la fila completa con join pero
// solo se usa program_org_id (chequeo de pertenencia), asi que el port solo
// expone lo que de verdad se consume.
type ProgramRepository interface {
	Create(ctx context.Context, p Program) (Program, error)
	GetByID(ctx context.Context, id, orgID uuid.UUID) (Program, error)
	ListByOrg(ctx context.Context, f ListProgramsFilter) ([]Program, error)
	Update(ctx context.Context, p Program) (Program, error)
	SetStatus(ctx context.Context, id, orgID uuid.UUID, status ProgramStatus) (Program, error)

	CreateWorkout(ctx context.Context, w ProgramWorkout) (ProgramWorkout, error)
	ListWorkouts(ctx context.Context, programID uuid.UUID) ([]ProgramWorkout, error)
	UpdateWorkout(ctx context.Context, w ProgramWorkout) (ProgramWorkout, error)
	DeleteWorkout(ctx context.Context, id uuid.UUID) error
	GetWorkoutOrgID(ctx context.Context, workoutID uuid.UUID) (uuid.UUID, error)

	CreateExercise(ctx context.Context, e ProgramExercise) (ProgramExercise, error)
	ListExercises(ctx context.Context, workoutID uuid.UUID) ([]ProgramExercise, error)
	UpdateExercise(ctx context.Context, e ProgramExercise) (ProgramExercise, error)
	DeleteExercise(ctx context.Context, id uuid.UUID) error
	GetExerciseOrgID(ctx context.Context, exerciseID uuid.UUID) (uuid.UUID, error)

	ListProgressionRules(ctx context.Context, orgID *uuid.UUID) ([]ProgressionRule, error)
	CreateProgressionRule(ctx context.Context, r ProgressionRule) (ProgressionRule, error)
	GetProgressionRuleByID(ctx context.Context, id uuid.UUID) (ProgressionRule, error)
}
