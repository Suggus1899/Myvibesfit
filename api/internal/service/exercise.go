package service

import (
	"context"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type ExerciseService struct {
	repo domain.ExerciseRepository
}

func NewExerciseService(repo domain.ExerciseRepository) *ExerciseService {
	return &ExerciseService{repo: repo}
}

func (s *ExerciseService) List(ctx context.Context, f domain.ExerciseFilter) ([]domain.Exercise, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 50
	}
	return s.repo.List(ctx, f)
}

func (s *ExerciseService) Get(ctx context.Context, id uuid.UUID) (domain.Exercise, error) {
	return s.repo.GetByID(ctx, id)
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

func (s *ExerciseService) Create(ctx context.Context, in CreateExerciseInput) (domain.Exercise, error) {
	if in.Slug == "" || in.Name == "" || in.PrimaryMuscle == "" || in.Pattern == "" {
		return domain.Exercise{}, domain.ErrInvalidInput
	}

	createdBy := in.CreatedBy
	return s.repo.Create(ctx, domain.Exercise{
		OrgID: in.OrgID, Slug: in.Slug, Name: in.Name, Description: in.Description, Instructions: nonNilSlice(in.Instructions),
		Pattern: in.Pattern, Mechanic: stringOrDefault(in.Mechanic, domain.DefaultExerciseMechanic),
		PrimaryMuscle: in.PrimaryMuscle, SecondaryMuscles: nonNilSlice(in.SecondaryMuscles), Equipment: nonNilSlice(in.Equipment),
		Difficulty: stringOrDefault(in.Difficulty, domain.DefaultExperienceLevel), Tracking: stringOrDefault(in.Tracking, domain.DefaultTrackingMode),
		IsUnilateral: in.IsUnilateral, VideoURL: in.VideoURL, ThumbnailURL: in.ThumbnailURL, CreatedBy: &createdBy,
	})
}

type UpdateExerciseInput struct {
	ID               uuid.UUID
	OrgID            uuid.UUID
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

func (s *ExerciseService) Update(ctx context.Context, in UpdateExerciseInput) (domain.Exercise, error) {
	if in.Name == "" || in.PrimaryMuscle == "" || in.Pattern == "" {
		return domain.Exercise{}, domain.ErrInvalidInput
	}

	return s.repo.Update(ctx, in.OrgID, domain.Exercise{
		ID: in.ID, Name: in.Name, Description: in.Description, Instructions: nonNilSlice(in.Instructions),
		Pattern: in.Pattern, Mechanic: stringOrDefault(in.Mechanic, domain.DefaultExerciseMechanic),
		PrimaryMuscle: in.PrimaryMuscle, SecondaryMuscles: nonNilSlice(in.SecondaryMuscles), Equipment: nonNilSlice(in.Equipment),
		Difficulty: stringOrDefault(in.Difficulty, domain.DefaultExperienceLevel), Tracking: stringOrDefault(in.Tracking, domain.DefaultTrackingMode),
		IsUnilateral: in.IsUnilateral, VideoURL: in.VideoURL, ThumbnailURL: in.ThumbnailURL,
	})
}

func (s *ExerciseService) Deactivate(ctx context.Context, id, orgID uuid.UUID) error {
	return s.repo.Deactivate(ctx, id, orgID)
}

// nonNilSlice evita mandar NULL a columnas text[] NOT NULL cuando el
// request JSON simplemente omite el campo (decodifica a nil, no a []).
func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func stringOrDefault(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
