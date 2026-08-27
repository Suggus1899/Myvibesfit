package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultExerciseMechanic = "compound"
	DefaultExperienceLevel  = "beginner"
	DefaultTrackingMode     = "weight_reps"
)

type Exercise struct {
	ID               uuid.UUID
	OrgID            *uuid.UUID
	Slug             string
	Name             string
	Description      string
	Instructions     []string
	Pattern          string
	Mechanic         string
	PrimaryMuscle    string
	SecondaryMuscles []string
	Equipment        []string
	Difficulty       string
	Tracking         string
	IsUnilateral     bool
	VideoURL         string
	ThumbnailURL     string
	CreatedBy        *uuid.UUID
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ExerciseFilter struct {
	OrgID   *uuid.UUID
	Pattern string
	Muscle  string
	Limit   int32
	Offset  int32
}

// ExerciseRepository espeja exercise.sql.go (5 metodos).
type ExerciseRepository interface {
	Create(ctx context.Context, e Exercise) (Exercise, error)
	GetByID(ctx context.Context, id uuid.UUID) (Exercise, error)
	List(ctx context.Context, f ExerciseFilter) ([]Exercise, error)
	Update(ctx context.Context, orgID uuid.UUID, e Exercise) (Exercise, error)
	Deactivate(ctx context.Context, id, orgID uuid.UUID) error
}
