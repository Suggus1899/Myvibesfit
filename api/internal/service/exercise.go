package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type ExerciseService struct {
	q db.Querier
}

func NewExerciseService(q db.Querier) *ExerciseService {
	return &ExerciseService{q: q}
}

type ListExercisesFilter struct {
	OrgID   *uuid.UUID
	Pattern string
	Muscle  string
	Limit   int32
	Offset  int32
}

func (s *ExerciseService) List(ctx context.Context, f ListExercisesFilter) ([]db.Exercise, error) {
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	params := db.ListExercisesParams{
		OrgID:       uuidToPg(f.OrgID),
		Muscle:      textToPg(f.Muscle),
		LimitCount:  limit,
		OffsetCount: f.Offset,
	}
	if f.Pattern != "" {
		params.Pattern = db.NullMovementPattern{MovementPattern: db.MovementPattern(f.Pattern), Valid: true}
	}

	return s.q.ListExercises(ctx, params)
}

func (s *ExerciseService) Get(ctx context.Context, id uuid.UUID) (db.Exercise, error) {
	ex, err := s.q.GetExerciseByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Exercise{}, domain.ErrNotFound
		}
		return db.Exercise{}, err
	}
	return ex, nil
}

type CreateExerciseInput struct {
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
	CreatedBy        uuid.UUID
}

func (s *ExerciseService) Create(ctx context.Context, in CreateExerciseInput) (db.Exercise, error) {
	if in.Slug == "" || in.Name == "" || in.PrimaryMuscle == "" || in.Pattern == "" {
		return db.Exercise{}, domain.ErrInvalidInput
	}

	mechanic := db.ExerciseMechanic(in.Mechanic)
	if mechanic == "" {
		mechanic = db.ExerciseMechanicCompound
	}
	difficulty := db.ExperienceLevel(in.Difficulty)
	if difficulty == "" {
		difficulty = db.ExperienceLevelBeginner
	}
	tracking := db.TrackingMode(in.Tracking)
	if tracking == "" {
		tracking = db.TrackingModeWeightReps
	}

	return s.q.CreateExercise(ctx, db.CreateExerciseParams{
		OrgID:            uuidToPg(in.OrgID),
		Slug:             in.Slug,
		Name:             in.Name,
		Description:      textToPg(in.Description),
		Instructions:     nonNilSlice(in.Instructions),
		Pattern:          db.MovementPattern(in.Pattern),
		Mechanic:         mechanic,
		PrimaryMuscle:    in.PrimaryMuscle,
		SecondaryMuscles: nonNilSlice(in.SecondaryMuscles),
		Equipment:        nonNilSlice(in.Equipment),
		Difficulty:       difficulty,
		Tracking:         tracking,
		IsUnilateral:     in.IsUnilateral,
		VideoUrl:         textToPg(in.VideoURL),
		ThumbnailUrl:     textToPg(in.ThumbnailURL),
		CreatedBy:        uuidToPg(&in.CreatedBy),
	})
}

type UpdateExerciseInput struct {
	ID               uuid.UUID
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
}

func (s *ExerciseService) Update(ctx context.Context, in UpdateExerciseInput) (db.Exercise, error) {
	if in.Name == "" || in.PrimaryMuscle == "" || in.Pattern == "" {
		return db.Exercise{}, domain.ErrInvalidInput
	}

	mechanic := db.ExerciseMechanic(in.Mechanic)
	if mechanic == "" {
		mechanic = db.ExerciseMechanicCompound
	}
	difficulty := db.ExperienceLevel(in.Difficulty)
	if difficulty == "" {
		difficulty = db.ExperienceLevelBeginner
	}
	tracking := db.TrackingMode(in.Tracking)
	if tracking == "" {
		tracking = db.TrackingModeWeightReps
	}

	ex, err := s.q.UpdateExercise(ctx, db.UpdateExerciseParams{
		ID:               in.ID,
		Name:             in.Name,
		Description:      textToPg(in.Description),
		Instructions:     nonNilSlice(in.Instructions),
		Pattern:          db.MovementPattern(in.Pattern),
		Mechanic:         mechanic,
		PrimaryMuscle:    in.PrimaryMuscle,
		SecondaryMuscles: nonNilSlice(in.SecondaryMuscles),
		Equipment:        nonNilSlice(in.Equipment),
		Difficulty:       difficulty,
		Tracking:         tracking,
		IsUnilateral:     in.IsUnilateral,
		VideoUrl:         textToPg(in.VideoURL),
		ThumbnailUrl:     textToPg(in.ThumbnailURL),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Exercise{}, domain.ErrNotFound
		}
		return db.Exercise{}, err
	}
	return ex, nil
}

func (s *ExerciseService) Deactivate(ctx context.Context, id uuid.UUID) error {
	return s.q.DeactivateExercise(ctx, id)
}

func uuidToPg(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

// nonNilSlice evita mandar NULL a columnas text[] NOT NULL cuando el
// request JSON simplemente omite el campo (decodifica a nil, no a []).
func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func textToPg(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}
