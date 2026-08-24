package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type ProgressService struct {
	q db.Querier
}

func NewProgressService(q db.Querier) *ProgressService {
	return &ProgressService{q: q}
}

func (s *ProgressService) ExerciseHistory(ctx context.Context, userID, exerciseID uuid.UUID, since time.Time, limit int32) ([]db.SetLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	return s.q.ListSetLogsForExercise(ctx, db.ListSetLogsForExerciseParams{
		UserID: userID, ExerciseID: exerciseID, PerformedAt: since, Limit: limit,
	})
}

func (s *ProgressService) Records(ctx context.Context, userID uuid.UUID, exerciseID *uuid.UUID) ([]db.PersonalRecord, error) {
	return s.q.ListPersonalRecords(ctx, db.ListPersonalRecordsParams{UserID: userID, ExerciseID: uuidToPg(exerciseID)})
}

func (s *ProgressService) Volume(ctx context.Context, userID uuid.UUID, since time.Time) ([]db.ListSessionsForVolumeRow, error) {
	return s.q.ListSessionsForVolume(ctx, db.ListSessionsForVolumeParams{UserID: userID, StartedAt: since})
}

type BodyMetricInput struct {
	UserID     uuid.UUID
	MeasuredOn time.Time
	WeightKg   *float64
	BodyFatPct *float64
	Note       string
}

func (s *ProgressService) LogBodyMetric(ctx context.Context, in BodyMetricInput) (db.BodyMetric, error) {
	if in.WeightKg == nil && in.BodyFatPct == nil {
		return db.BodyMetric{}, domain.ErrInvalidInput
	}
	return s.q.UpsertBodyMetric(ctx, db.UpsertBodyMetricParams{
		UserID: in.UserID, MeasuredOn: in.MeasuredOn, WeightKg: in.WeightKg,
		BodyFatPct: in.BodyFatPct, Note: textToPg(in.Note),
	})
}

func (s *ProgressService) BodyMetrics(ctx context.Context, userID uuid.UUID, limit int32) ([]db.BodyMetric, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.q.ListBodyMetrics(ctx, db.ListBodyMetricsParams{UserID: userID, Limit: limit})
}
