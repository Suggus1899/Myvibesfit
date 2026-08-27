package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type ProgressRepository struct{ q db.Querier }

func NewProgressRepository(q db.Querier) *ProgressRepository { return &ProgressRepository{q: q} }

var _ domain.ProgressRepository = (*ProgressRepository)(nil)

func (r *ProgressRepository) ListSetLogsForExercise(ctx context.Context, userID, exerciseID uuid.UUID, since time.Time, limit int32) ([]domain.SetLog, error) {
	rows, err := r.q.ListSetLogsForExercise(ctx, db.ListSetLogsForExerciseParams{
		UserID: userID, ExerciseID: exerciseID, PerformedAt: since, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.SetLog, len(rows))
	for i, row := range rows {
		out[i] = toDomainSetLog(row)
	}
	return out, nil
}

func (r *ProgressRepository) ListPersonalRecords(ctx context.Context, userID uuid.UUID, exerciseID *uuid.UUID) ([]domain.PersonalRecord, error) {
	rows, err := r.q.ListPersonalRecords(ctx, db.ListPersonalRecordsParams{UserID: userID, ExerciseID: uuidPtrToPgUUID(exerciseID)})
	if err != nil {
		return nil, err
	}
	out := make([]domain.PersonalRecord, len(rows))
	for i, row := range rows {
		out[i] = toDomainPersonalRecord(row)
	}
	return out, nil
}

func (r *ProgressRepository) ListSessionsForVolume(ctx context.Context, userID uuid.UUID, since time.Time) ([]domain.SessionVolumePoint, error) {
	rows, err := r.q.ListSessionsForVolume(ctx, db.ListSessionsForVolumeParams{UserID: userID, StartedAt: since})
	if err != nil {
		return nil, err
	}
	out := make([]domain.SessionVolumePoint, len(rows))
	for i, row := range rows {
		out[i] = domain.SessionVolumePoint{
			ID: row.ID, StartedAt: row.StartedAt, TotalVolumeKg: row.TotalVolumeKg, DurationSeconds: pgInt4ToPtr(row.DurationSeconds),
		}
	}
	return out, nil
}

func (r *ProgressRepository) UpsertBodyMetric(ctx context.Context, m domain.BodyMetric) (domain.BodyMetric, error) {
	row, err := r.q.UpsertBodyMetric(ctx, db.UpsertBodyMetricParams{
		UserID: m.UserID, MeasuredOn: m.MeasuredOn, WeightKg: m.WeightKg, BodyFatPct: m.BodyFatPct, Note: stringToPgText(m.Note),
	})
	if err != nil {
		return domain.BodyMetric{}, err
	}
	return toDomainBodyMetric(row), nil
}

func (r *ProgressRepository) ListBodyMetrics(ctx context.Context, userID uuid.UUID, limit int32) ([]domain.BodyMetric, error) {
	rows, err := r.q.ListBodyMetrics(ctx, db.ListBodyMetricsParams{UserID: userID, Limit: limit})
	if err != nil {
		return nil, err
	}
	out := make([]domain.BodyMetric, len(rows))
	for i, row := range rows {
		out[i] = toDomainBodyMetric(row)
	}
	return out, nil
}

func toDomainBodyMetric(m db.BodyMetric) domain.BodyMetric {
	return domain.BodyMetric{
		ID: m.ID, UserID: m.UserID, MeasuredOn: m.MeasuredOn, WeightKg: m.WeightKg, BodyFatPct: m.BodyFatPct,
		Measurements: m.Measurements, Note: pgTextToString(m.Note), CreatedAt: m.CreatedAt,
	}
}
