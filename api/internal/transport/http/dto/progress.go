package dto

import (
	"time"

	"github.com/google/uuid"
)

type SetLogDTO struct {
	ID          int64     `json:"id"`
	ExerciseID  uuid.UUID `json:"exercise_id"`
	SetNumber   int       `json:"set_number"`
	Type        string    `json:"type"`
	WeightKg    *float64  `json:"weight_kg,omitempty"`
	Reps        *int      `json:"reps,omitempty"`
	RPE         *float64  `json:"rpe,omitempty"`
	PerformedAt time.Time `json:"performed_at"`
}

type PersonalRecordDTO struct {
	ID         int64     `json:"id"`
	ExerciseID uuid.UUID `json:"exercise_id"`
	Type       string    `json:"type"`
	Value      float64   `json:"value"`
	AchievedAt time.Time `json:"achieved_at"`
}

type VolumePointDTO struct {
	SessionID       uuid.UUID `json:"session_id"`
	StartedAt       time.Time `json:"started_at"`
	TotalVolumeKg   float64   `json:"total_volume_kg"`
	DurationSeconds *int      `json:"duration_seconds,omitempty"`
}

type BodyMetricDTO struct {
	MeasuredOn string   `json:"measured_on"`
	WeightKg   *float64 `json:"weight_kg,omitempty"`
	BodyFatPct *float64 `json:"body_fat_pct,omitempty"`
	Note       string   `json:"note,omitempty"`
}

type LogBodyMetricRequest struct {
	MeasuredOn string   `json:"measured_on"`
	WeightKg   *float64 `json:"weight_kg"`
	BodyFatPct *float64 `json:"body_fat_pct"`
	Note       string   `json:"note"`
}
