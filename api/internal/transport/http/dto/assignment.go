package dto

import "github.com/google/uuid"

type AssignRequest struct {
	ClientUserID uuid.UUID `json:"client_user_id"`
	StartDate    string    `json:"start_date"` // YYYY-MM-DD
}

type AssignmentDTO struct {
	ID           uuid.UUID  `json:"id"`
	ProgramID    *uuid.UUID `json:"program_id,omitempty"`
	ClientUserID uuid.UUID  `json:"client_user_id"`
	CoachUserID  *uuid.UUID `json:"coach_user_id,omitempty"`
	Name         string     `json:"name"`
	StartDate    string     `json:"start_date"`
	EndDate      string     `json:"end_date,omitempty"`
	Status       string     `json:"status"`
}

type AssignedExerciseDTO struct {
	ID             uuid.UUID `json:"id"`
	ExerciseID     uuid.UUID `json:"exercise_id"`
	OrderIndex     int       `json:"order_index"`
	SupersetGroup  *int      `json:"superset_group,omitempty"`
	TargetSets     int       `json:"target_sets"`
	TargetRepsMin  *int      `json:"target_reps_min,omitempty"`
	TargetRepsMax  *int      `json:"target_reps_max,omitempty"`
	TargetRPE      *float64  `json:"target_rpe,omitempty"`
	TargetWeightKg *float64  `json:"target_weight_kg,omitempty"`
	RestSeconds    int       `json:"rest_seconds"`
	Tempo          string    `json:"tempo,omitempty"`
	Note           string    `json:"note,omitempty"`
}

type AssignedWorkoutDTO struct {
	ID          uuid.UUID             `json:"id"`
	WeekNumber  int                   `json:"week_number"`
	DayIndex    int                   `json:"day_index"`
	Name        string                `json:"name"`
	Note        string                `json:"note,omitempty"`
	IsDeload    bool                  `json:"is_deload"`
	ScheduledOn string                `json:"scheduled_on,omitempty"`
	Status      string                `json:"status"`
	Exercises   []AssignedExerciseDTO `json:"exercises"`
}

type AssignmentDetailDTO struct {
	Assignment AssignmentDTO        `json:"assignment"`
	Workouts   []AssignedWorkoutDTO `json:"workouts"`
}
