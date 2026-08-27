package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type ExerciseRepository struct{ q db.Querier }

func NewExerciseRepository(q db.Querier) *ExerciseRepository { return &ExerciseRepository{q: q} }

var _ domain.ExerciseRepository = (*ExerciseRepository)(nil)

func (r *ExerciseRepository) Create(ctx context.Context, e domain.Exercise) (domain.Exercise, error) {
	row, err := r.q.CreateExercise(ctx, db.CreateExerciseParams{
		OrgID: uuidPtrToPgUUID(e.OrgID), Slug: e.Slug, Name: e.Name, Description: stringToPgText(e.Description),
		Instructions: e.Instructions, Pattern: db.MovementPattern(e.Pattern), Mechanic: db.ExerciseMechanic(e.Mechanic),
		PrimaryMuscle: e.PrimaryMuscle, SecondaryMuscles: e.SecondaryMuscles, Equipment: e.Equipment,
		Difficulty: db.ExperienceLevel(e.Difficulty), Tracking: db.TrackingMode(e.Tracking), IsUnilateral: e.IsUnilateral,
		VideoUrl: stringToPgText(e.VideoURL), ThumbnailUrl: stringToPgText(e.ThumbnailURL), CreatedBy: uuidPtrToPgUUID(e.CreatedBy),
	})
	if err != nil {
		return domain.Exercise{}, err
	}
	return toDomainExercise(row), nil
}

func (r *ExerciseRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Exercise, error) {
	row, err := r.q.GetExerciseByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Exercise{}, domain.ErrNotFound
		}
		return domain.Exercise{}, err
	}
	return toDomainExercise(row), nil
}

func (r *ExerciseRepository) List(ctx context.Context, f domain.ExerciseFilter) ([]domain.Exercise, error) {
	params := db.ListExercisesParams{
		OrgID: uuidPtrToPgUUID(f.OrgID), Muscle: stringToPgText(f.Muscle), LimitCount: f.Limit, OffsetCount: f.Offset,
	}
	if f.Pattern != "" {
		params.Pattern = db.NullMovementPattern{MovementPattern: db.MovementPattern(f.Pattern), Valid: true}
	}
	rows, err := r.q.ListExercises(ctx, params)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Exercise, len(rows))
	for i, row := range rows {
		out[i] = toDomainExercise(row)
	}
	return out, nil
}

func (r *ExerciseRepository) Update(ctx context.Context, orgID uuid.UUID, e domain.Exercise) (domain.Exercise, error) {
	row, err := r.q.UpdateExercise(ctx, db.UpdateExerciseParams{
		ID: e.ID, OrgID: uuidPtrToPgUUID(&orgID), Name: e.Name, Description: stringToPgText(e.Description), Instructions: e.Instructions,
		Pattern: db.MovementPattern(e.Pattern), Mechanic: db.ExerciseMechanic(e.Mechanic), PrimaryMuscle: e.PrimaryMuscle,
		SecondaryMuscles: e.SecondaryMuscles, Equipment: e.Equipment, Difficulty: db.ExperienceLevel(e.Difficulty),
		Tracking: db.TrackingMode(e.Tracking), IsUnilateral: e.IsUnilateral, VideoUrl: stringToPgText(e.VideoURL), ThumbnailUrl: stringToPgText(e.ThumbnailURL),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Exercise{}, domain.ErrNotFound
		}
		return domain.Exercise{}, err
	}
	return toDomainExercise(row), nil
}

func (r *ExerciseRepository) Deactivate(ctx context.Context, id, orgID uuid.UUID) error {
	rows, err := r.q.DeactivateExercise(ctx, db.DeactivateExerciseParams{ID: id, OrgID: uuidPtrToPgUUID(&orgID)})
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func toDomainExercise(e db.Exercise) domain.Exercise {
	return domain.Exercise{
		ID: e.ID, OrgID: pgUUIDToPtr(e.OrgID), Slug: e.Slug, Name: e.Name, Description: pgTextToString(e.Description),
		Instructions: e.Instructions, Pattern: string(e.Pattern), Mechanic: string(e.Mechanic), PrimaryMuscle: e.PrimaryMuscle,
		SecondaryMuscles: e.SecondaryMuscles, Equipment: e.Equipment, Difficulty: string(e.Difficulty), Tracking: string(e.Tracking),
		IsUnilateral: e.IsUnilateral, VideoURL: pgTextToString(e.VideoUrl), ThumbnailURL: pgTextToString(e.ThumbnailUrl),
		CreatedBy: pgUUIDToPtr(e.CreatedBy), IsActive: e.IsActive, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
	}
}
