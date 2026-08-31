package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const DefaultHabitFrequency = "daily"

type Habit struct {
	ID            uuid.UUID
	OrgID         *uuid.UUID
	Slug          string
	Name          string
	Icon          string
	Unit          string
	DefaultTarget *float64
	IsSystem      bool
	CreatedAt     time.Time
}

type ClientHabit struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	HabitID     uuid.UUID
	AssignedBy  *uuid.UUID
	TargetValue *float64
	Frequency   string
	DaysOfWeek  []int16
	StartedOn   time.Time
	EndedOn     *time.Time
}

// MyHabit es el join client_habit+habit que consume la pantalla "mis
// habitos" del cliente.
type MyHabit struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	HabitID     uuid.UUID
	TargetValue *float64
	Frequency   string
	DaysOfWeek  []int16
	StartedOn   time.Time
	EndedOn     *time.Time
	HabitName   string
	HabitIcon   string
	HabitUnit   string
}

type HabitLog struct {
	ID            int64
	ClientLocalID uuid.UUID
	ClientHabitID uuid.UUID
	UserID        uuid.UUID
	LogDate       time.Time
	Value         float64
	IsCompleted   bool
	LoggedAt      time.Time

	// Inserted distingue el primer registro del dia de un reenvio. La fila es
	// idempotente por (client_habit_id, log_date); el XP y la racha tambien
	// tienen que serlo.
	Inserted bool
}

// HabitRepository espeja habit.sql.go (10 metodos).
type HabitRepository interface {
	ListHabits(ctx context.Context, orgID *uuid.UUID) ([]Habit, error)
	GetHabitByID(ctx context.Context, id uuid.UUID) (Habit, error)
	ListMyHabits(ctx context.Context, userID uuid.UUID) ([]MyHabit, error)
	GetActiveClientHabit(ctx context.Context, userID, habitID uuid.UUID) (ClientHabit, error)
	GetClientHabitOwner(ctx context.Context, clientHabitID uuid.UUID) (uuid.UUID, error)
	SubscribeHabit(ctx context.Context, c ClientHabit) (ClientHabit, error)
	UnsubscribeHabit(ctx context.Context, id, userID uuid.UUID) error
	UpsertLog(ctx context.Context, l HabitLog) (HabitLog, error)
	ListLogsForDate(ctx context.Context, userID uuid.UUID, date time.Time) ([]HabitLog, error)
	CountActiveForUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountCompletedLogsForDate(ctx context.Context, userID uuid.UUID, date time.Time) (int64, error)
}
