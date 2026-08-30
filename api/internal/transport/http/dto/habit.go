package dto

import "github.com/google/uuid"

type HabitDTO struct {
	ID            uuid.UUID `json:"id"`
	Slug          string    `json:"slug"`
	Name          string    `json:"name"`
	Icon          string    `json:"icon"`
	Unit          string    `json:"unit"`
	DefaultTarget *float64  `json:"default_target,omitempty"`
	IsSystem      bool      `json:"is_system"`
}

type MyHabitDTO struct {
	ID          uuid.UUID `json:"id"`
	HabitID     uuid.UUID `json:"habit_id"`
	Name        string    `json:"name"`
	Icon        string    `json:"icon"`
	Unit        string    `json:"unit"`
	TargetValue *float64  `json:"target_value,omitempty"`
	Frequency   string    `json:"frequency"`
	DaysOfWeek  []int     `json:"days_of_week"`
}

type SubscribeHabitRequest struct {
	HabitID     uuid.UUID `json:"habit_id"`
	TargetValue *float64  `json:"target_value"`
	Frequency   string    `json:"frequency"`
	DaysOfWeek  []int     `json:"days_of_week"`
}

type LogHabitRequest struct {
	ClientLocalID uuid.UUID `json:"client_local_id"`
	LogDate       string    `json:"log_date"`
	Value         float64   `json:"value"`
	IsCompleted   bool      `json:"is_completed"`
}

type HabitLogDTO struct {
	ID            int64     `json:"id"`
	ClientHabitID uuid.UUID `json:"client_habit_id,omitempty"`
	LogDate       string    `json:"log_date"`
	Value         float64   `json:"value"`
	IsCompleted   bool      `json:"is_completed"`
}

type HabitLogResponse struct {
	Log                  HabitLogDTO      `json:"log"`
	UnlockedAchievements []AchievementDTO `json:"unlocked_achievements"`
}

type StreakDTO struct {
	Kind         string `json:"kind"`
	CurrentCount int    `json:"current_count"`
	LongestCount int    `json:"longest_count"`
	LastActiveOn string `json:"last_active_on,omitempty"`
}

type UserStatsDTO struct {
	TotalXP       int         `json:"total_xp"`
	Level         int         `json:"level"`
	TotalSessions int         `json:"total_sessions"`
	TotalVolumeKg float64     `json:"total_volume_kg"`
	Streaks       []StreakDTO `json:"streaks"`
}

type UserAchievementDTO struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        string  `json:"icon"`
	Tier        int     `json:"tier"`
	Progress    float64 `json:"progress"`
	EarnedAt    string  `json:"earned_at,omitempty"`
}
