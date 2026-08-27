package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type BodyMetric struct {
	ID           int64
	UserID       uuid.UUID
	MeasuredOn   time.Time
	WeightKg     *float64
	BodyFatPct   *float64
	Measurements []byte
	Note         string
	CreatedAt    time.Time
}

// SessionVolumePoint es la proyeccion de workout_session que consume el
// grafico de volumen (progress.sql: ListSessionsForVolume).
type SessionVolumePoint struct {
	ID              uuid.UUID
	StartedAt       time.Time
	TotalVolumeKg   float64
	DurationSeconds *int
}

// ProgressRepository espeja progress.sql.go (5 metodos: lecturas de
// progreso + registro de metricas corporales).
type ProgressRepository interface {
	ListSetLogsForExercise(ctx context.Context, userID, exerciseID uuid.UUID, since time.Time, limit int32) ([]SetLog, error)
	ListPersonalRecords(ctx context.Context, userID uuid.UUID, exerciseID *uuid.UUID) ([]PersonalRecord, error)
	ListSessionsForVolume(ctx context.Context, userID uuid.UUID, since time.Time) ([]SessionVolumePoint, error)
	UpsertBodyMetric(ctx context.Context, m BodyMetric) (BodyMetric, error)
	ListBodyMetrics(ctx context.Context, userID uuid.UUID, limit int32) ([]BodyMetric, error)
}
