package domain

import (
	"context"
	"fmt"
	"math"
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

	// MarkApplied sella que la sugerencia efectivamente muto el plan.
	// Aprobar y aplicar tienen que ser atomicos: marcarla aprobada sin
	// aplicarla le hace creer al coach que el ajuste esta puesto.
	MarkApplied(ctx context.Context, id, orgID uuid.UUID) error
}

// AppliableKinds son los unicos kinds que mutan assigned_exercise al
// aprobarse, porque son los que traen un delta numerico ya validado por
// Suggestion.Validate. exercise_swap viaja con un nombre de ejercicio en
// texto libre sin id resoluble contra el catalogo — resolverlo es una
// decision de producto pendiente, no un bug. deload, rest_day y habit_nudge
// no proponen un target: sugieren saltear o descansar.
var AppliableKinds = map[string]bool{"load_adjust": true, "volume_adjust": true}

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

	// Contexto del onboarding (client_profile). Sin esto la IA propone a
	// ciegas: no sabe si el cliente busca fuerza o bajar grasa, ni con que
	// equipamiento cuenta, ni que lesiones tiene que esquivar.
	PrimaryGoal        string   `json:"primary_goal,omitempty"`
	Experience         string   `json:"experience,omitempty"`
	DaysPerWeekTarget  int      `json:"days_per_week_target,omitempty"`
	SessionMinutes     int      `json:"session_minutes_target,omitempty"`
	AvailableEquipment []string `json:"available_equipment,omitempty"`
	Limitations        string   `json:"limitations,omitempty"`
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

// Topes de magnitud (docs/ARCHITECTURE.md §4): la IA propone ajustes
// incrementales, no saltos. Un salto grande casi siempre es una alucinacion
// o un error de unidades, y el coach no deberia siquiera tener que
// rechazarlo a mano.
const (
	MaxLoadDeltaPercent   = 10.0
	MaxVolumeDeltaPercent = 30.0
)

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

	delta := math.Abs(s.Payload.DeltaPercent)
	switch s.Kind {
	case "load_adjust":
		if delta > MaxLoadDeltaPercent {
			return fmt.Errorf("salto de carga %.1f%% supera el tope de %.0f%%", delta, MaxLoadDeltaPercent)
		}
	case "volume_adjust":
		if delta > MaxVolumeDeltaPercent {
			return fmt.Errorf("salto de volumen %.1f%% supera el tope de %.0f%%", delta, MaxVolumeDeltaPercent)
		}
	}
	return nil
}

// SuggestionProposer es el port hacia el proveedor de IA (adapter/anthropic).
type SuggestionProposer interface {
	Propose(ctx context.Context, summary AssignmentSummary) (Suggestion, error)
}
