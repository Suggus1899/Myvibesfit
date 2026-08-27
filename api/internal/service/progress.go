package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

type ProgressService struct {
	repo domain.ProgressRepository
}

func NewProgressService(repo domain.ProgressRepository) *ProgressService {
	return &ProgressService{repo: repo}
}

func (s *ProgressService) ExerciseHistory(ctx context.Context, userID, exerciseID uuid.UUID, since time.Time, limit int32) ([]domain.SetLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	return s.repo.ListSetLogsForExercise(ctx, userID, exerciseID, since, limit)
}

func (s *ProgressService) Records(ctx context.Context, userID uuid.UUID, exerciseID *uuid.UUID) ([]domain.PersonalRecord, error) {
	return s.repo.ListPersonalRecords(ctx, userID, exerciseID)
}

func (s *ProgressService) Volume(ctx context.Context, userID uuid.UUID, since time.Time) ([]domain.SessionVolumePoint, error) {
	return s.repo.ListSessionsForVolume(ctx, userID, since)
}

type BodyMetricInput struct {
	UserID     uuid.UUID
	MeasuredOn time.Time
	WeightKg   *float64
	BodyFatPct *float64
	Note       string
}

func (s *ProgressService) LogBodyMetric(ctx context.Context, in BodyMetricInput) (domain.BodyMetric, error) {
	if in.WeightKg == nil && in.BodyFatPct == nil {
		return domain.BodyMetric{}, domain.ErrInvalidInput
	}
	return s.repo.UpsertBodyMetric(ctx, domain.BodyMetric{
		UserID: in.UserID, MeasuredOn: in.MeasuredOn, WeightKg: in.WeightKg, BodyFatPct: in.BodyFatPct, Note: in.Note,
	})
}

func (s *ProgressService) BodyMetrics(ctx context.Context, userID uuid.UUID, limit int32) ([]domain.BodyMetric, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return s.repo.ListBodyMetrics(ctx, userID, limit)
}
