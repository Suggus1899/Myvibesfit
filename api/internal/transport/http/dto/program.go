package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ProgressionRuleDTO struct {
	ID       uuid.UUID       `json:"id"`
	OrgID    *uuid.UUID      `json:"org_id,omitempty"`
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Params   json.RawMessage `json:"params"`
	IsSystem bool            `json:"is_system"`
}

type CreateProgressionRuleRequest struct {
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	Params json.RawMessage `json:"params"`
}

type ProgramDTO struct {
	ID          uuid.UUID `json:"id"`
	OrgID       uuid.UUID `json:"org_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Goal        string    `json:"goal"`
	Level       string    `json:"level"`
	TotalWeeks  int       `json:"total_weeks"`
	DaysPerWeek int       `json:"days_per_week"`
	Status      string    `json:"status"`
}

type CreateProgramRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Goal        string `json:"goal"`
	Level       string `json:"level"`
	TotalWeeks  int    `json:"total_weeks"`
	DaysPerWeek int    `json:"days_per_week"`
}

type UpdateProgramRequest = CreateProgramRequest

type ProgramWorkoutDTO struct {
	ID               uuid.UUID `json:"id"`
	ProgramID        uuid.UUID `json:"program_id"`
	WeekNumber       int       `json:"week_number"`
	DayIndex         int       `json:"day_index"`
	Name             string    `json:"name"`
	Note             string    `json:"note,omitempty"`
	IsDeload         bool      `json:"is_deload"`
	EstimatedMinutes *int      `json:"estimated_minutes,omitempty"`
}

type CreateProgramWorkoutRequest struct {
	WeekNumber       int    `json:"week_number"`
	DayIndex         int    `json:"day_index"`
	Name             string `json:"name"`
	Note             string `json:"note"`
	IsDeload         bool   `json:"is_deload"`
	EstimatedMinutes *int   `json:"estimated_minutes"`
}

type UpdateProgramWorkoutRequest struct {
	Name             string `json:"name"`
	Note             string `json:"note"`
	IsDeload         bool   `json:"is_deload"`
	EstimatedMinutes *int   `json:"estimated_minutes"`
}

type ProgramExerciseDTO struct {
	ID                uuid.UUID  `json:"id"`
	ProgramWorkoutID  uuid.UUID  `json:"program_workout_id"`
	ExerciseID        uuid.UUID  `json:"exercise_id"`
	OrderIndex        int        `json:"order_index"`
	SupersetGroup     *int       `json:"superset_group,omitempty"`
	TargetSets        int        `json:"target_sets"`
	TargetRepsMin     *int       `json:"target_reps_min,omitempty"`
	TargetRepsMax     *int       `json:"target_reps_max,omitempty"`
	TargetRPE         *float64   `json:"target_rpe,omitempty"`
	TargetPct1RM      *float64   `json:"target_pct_1rm,omitempty"`
	RestSeconds       int        `json:"rest_seconds"`
	Tempo             string     `json:"tempo,omitempty"`
	Note              string     `json:"note,omitempty"`
	ProgressionRuleID *uuid.UUID `json:"progression_rule_id,omitempty"`
}

type CreateProgramExerciseRequest struct {
	ExerciseID        uuid.UUID  `json:"exercise_id"`
	OrderIndex        int        `json:"order_index"`
	SupersetGroup     *int       `json:"superset_group"`
	TargetSets        int        `json:"target_sets"`
	TargetRepsMin     *int       `json:"target_reps_min"`
	TargetRepsMax     *int       `json:"target_reps_max"`
	TargetRPE         *float64   `json:"target_rpe"`
	TargetPct1RM      *float64   `json:"target_pct_1rm"`
	RestSeconds       int        `json:"rest_seconds"`
	Tempo             string     `json:"tempo"`
	Note              string     `json:"note"`
	ProgressionRuleID *uuid.UUID `json:"progression_rule_id"`
}

type UpdateProgramExerciseRequest struct {
	OrderIndex        int        `json:"order_index"`
	SupersetGroup     *int       `json:"superset_group"`
	TargetSets        int        `json:"target_sets"`
	TargetRepsMin     *int       `json:"target_reps_min"`
	TargetRepsMax     *int       `json:"target_reps_max"`
	TargetRPE         *float64   `json:"target_rpe"`
	TargetPct1RM      *float64   `json:"target_pct_1rm"`
	RestSeconds       int        `json:"rest_seconds"`
	Tempo             string     `json:"tempo"`
	Note              string     `json:"note"`
	ProgressionRuleID *uuid.UUID `json:"progression_rule_id"`
}
