package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const DefaultSuggestionModel = "claude-opus-5"

type SuggestionStatus string

const (
	SuggestionStatusPending     SuggestionStatus = "pending"
	SuggestionStatusApproved    SuggestionStatus = "approved"
	SuggestionStatusRejected    SuggestionStatus = "rejected"
	SuggestionStatusExpired     SuggestionStatus = "expired"
	SuggestionStatusAutoApplied SuggestionStatus = "auto_applied"
)

var validSuggestionKinds = map[string]bool{
	"volume_adjust": true, "load_adjust": true, "exercise_swap": true,
	"deload": true, "rest_day": true, "habit_nudge": true,
}

type AISuggestion struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	ClientUserID  uuid.UUID
	CoachUserID   *uuid.UUID
	AssignmentID  *uuid.UUID
	Kind          string
	Payload       []byte
	Rationale     string
	Confidence    *float64
	Model         string
	InputSnapshot []byte
	Status        SuggestionStatus
	ReviewedBy    *uuid.UUID
	ReviewedAt    *time.Time
	AppliedAt     *time.Time
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

// PendingSuggestion agrega el nombre del cliente (join con app_user) que
// consume la tarjeta de aprobacion del coach.
type PendingSuggestion struct {
	AISuggestion
	ClientName string
}

type ActiveAssignmentForWorker struct {
	AssignmentID uuid.UUID
	OrgID        uuid.UUID
	ClientUserID uuid.UUID
	CoachUserID  *uuid.UUID
	ClientName   string
}

type RecentWorkingSet struct {
	ExerciseID   uuid.UUID
	ExerciseName string
	WeightKg     *float64
	Reps         *int
	RPE          *float64
	PerformedAt  time.Time
}

// AISuggestionRepository espeja ai_suggestion.sql.go (9 metodos): las
// queries de agregacion que arman el resumen del worker, mas el CRUD de
// ai_suggestion que consume el flujo de aprobacion del coach.
type AISuggestionRepository interface {
	ListActiveAssignmentsForWorker(ctx context.Context) ([]ActiveAssignmentForWorker, error)
	CountAssignedWorkoutsInRange(ctx context.Context, assignmentID uuid.UUID, from, to time.Time) (int64, error)
	CountCompletedSessionsSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error)
	CountCompletedHabitLogsSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error)
	ListRecentWorkingSets(ctx context.Context, userID uuid.UUID, since time.Time, limit int32) ([]RecentWorkingSet, error)
	Create(ctx context.Context, s AISuggestion) (AISuggestion, error)
	GetByID(ctx context.Context, id, orgID uuid.UUID) (AISuggestion, error)
	ListPendingForCoach(ctx context.Context, coachID, orgID uuid.UUID) ([]PendingSuggestion, error)
	Review(ctx context.Context, id, orgID uuid.UUID, status SuggestionStatus, reviewedBy uuid.UUID) (AISuggestion, error)
}

// --- lo que se manda a/recibe de Claude (movido de internal/ai) ---

type AssignmentSummary struct {
	ClientName        string          `json:"client_name"`
	PeriodDays        int             `json:"period_days"`
	AssignedWorkouts  int             `json:"assigned_workouts_in_period"`
	CompletedSessions int             `json:"completed_sessions_in_period"`
	CurrentStreakDays int             `json:"current_streak_days"`
	HabitAdherencePct float64         `json:"habit_adherence_pct"`
	RecentPRs         []PRSummary     `json:"recent_prs"`
	ExerciseTrends    []ExerciseTrend `json:"exercise_trends"`
}

type PRSummary struct {
	ExerciseName string  `json:"exercise_name"`
	Type         string  `json:"type"`
	Value        float64 `json:"value"`
	AchievedAt   string  `json:"achieved_at"`
}

type ExerciseTrend struct {
	ExerciseName  string   `json:"exercise_name"`
	SetCount      int      `json:"set_count"`
	FirstWeightKg *float64 `json:"first_weight_kg,omitempty"`
	LastWeightKg  *float64 `json:"last_weight_kg,omitempty"`
	AvgRPE        *float64 `json:"avg_rpe,omitempty"`
}

type SuggestionPayload struct {
	TargetExerciseID      string  `json:"target_exercise_id,omitempty"`
	DeltaPercent          float64 `json:"delta_percent,omitempty"`
	SuggestedExerciseName string  `json:"suggested_exercise_name,omitempty"`
	Note                  string  `json:"note,omitempty"`
}

type Suggestion struct {
	Kind       string            `json:"kind"`
	Rationale  string            `json:"rationale"`
	Confidence float64           `json:"confidence"`
	Payload    SuggestionPayload `json:"payload"`
}

func (s Suggestion) Validate() error {
	if !validSuggestionKinds[s.Kind] {
		return fmt.Errorf("kind invalido: %q", s.Kind)
	}
	if s.Rationale == "" {
		return fmt.Errorf("rationale vacio")
	}
	if s.Confidence < 0 || s.Confidence > 1 {
		return fmt.Errorf("confidence fuera de rango: %v", s.Confidence)
	}
	return nil
}

// SuggestionProposer es el port hacia el proveedor de IA (adapter/anthropic).
type SuggestionProposer interface {
	Propose(ctx context.Context, summary AssignmentSummary) (Suggestion, error)
}
