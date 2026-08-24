package dto

import (
	"time"

	"github.com/google/uuid"
)

type SyncSetRequest struct {
	ClientLocalID    uuid.UUID `json:"client_local_id"`
	SetNumber        int       `json:"set_number"`
	Type             string    `json:"type"`
	WeightKg         *float64  `json:"weight_kg"`
	Reps             *int      `json:"reps"`
	RPE              *float64  `json:"rpe"`
	RIR              *int      `json:"rir"`
	DurationSeconds  *int      `json:"duration_seconds"`
	DistanceM        *float64  `json:"distance_m"`
	RestTakenSeconds *int      `json:"rest_taken_seconds"`
	IsCompleted      bool      `json:"is_completed"`
	PerformedAt      time.Time `json:"performed_at"`
}

type SyncExerciseRequest struct {
	ExerciseID         uuid.UUID        `json:"exercise_id"`
	AssignedExerciseID *uuid.UUID       `json:"assigned_exercise_id"`
	OrderIndex         int              `json:"order_index"`
	SupersetGroup      *int             `json:"superset_group"`
	Note               string           `json:"note"`
	Sets               []SyncSetRequest `json:"sets"`
}

type SyncSessionRequest struct {
	ClientLocalID     uuid.UUID             `json:"client_local_id"`
	Name              string                `json:"name"`
	Status            string                `json:"status"`
	AssignedWorkoutID *uuid.UUID            `json:"assigned_workout_id"`
	StartedAt         time.Time             `json:"started_at"`
	EndedAt           *time.Time            `json:"ended_at"`
	DurationSeconds   *int                  `json:"duration_seconds"`
	PerceivedEffort   *int                  `json:"perceived_effort"`
	Mood              *int                  `json:"mood"`
	Notes             string                `json:"notes"`
	Exercises         []SyncExerciseRequest `json:"exercises"`
}

type SyncRequest struct {
	Sessions []SyncSessionRequest `json:"sessions"`
}

type SyncedSessionDTO struct {
	ID            uuid.UUID `json:"id"`
	ClientLocalID uuid.UUID `json:"client_local_id"`
	Status        string    `json:"status"`
	TotalVolumeKg float64   `json:"total_volume_kg"`
}

type AchievementDTO struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	Tier        int       `json:"tier"`
	XPReward    int       `json:"xp_reward"`
}

type SyncResponse struct {
	Sessions             []SyncedSessionDTO `json:"sessions"`
	NewPersonalRecords   int                `json:"new_personal_records"`
	UnlockedAchievements []AchievementDTO   `json:"unlocked_achievements"`
}
