package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionStatus string

const (
	SessionStatusInProgress SessionStatus = "in_progress"
	SessionStatusCompleted  SessionStatus = "completed"
	SessionStatusAbandoned  SessionStatus = "abandoned"
)

type SetType string

const (
	SetTypeWarmup  SetType = "warmup"
	SetTypeWorking SetType = "working"
	SetTypeDrop    SetType = "drop"
	SetTypeFailure SetType = "failure"
	SetTypeBackoff SetType = "backoff"
)

type RecordType string

const (
	RecordTypeMaxWeight    RecordType = "max_weight"
	RecordTypeMaxReps      RecordType = "max_reps"
	RecordTypeEstimated1RM RecordType = "estimated_1rm"
	RecordTypeMaxVolumeSet RecordType = "max_volume_set"
)

type WorkoutSession struct {
	ID                uuid.UUID
	ClientLocalID     uuid.UUID
	UserID            uuid.UUID
	OrgID             *uuid.UUID
	AssignedWorkoutID *uuid.UUID
	Name              string
	Status            SessionStatus
	StartedAt         time.Time
	EndedAt           *time.Time
	DurationSeconds   *int
	TotalVolumeKg     float64
	PerceivedEffort   *int
	Mood              *int
	Notes             string
	SyncedAt          time.Time
	CreatedAt         time.Time
}

type SessionExercise struct {
	ID                 uuid.UUID
	SessionID          uuid.UUID
	ExerciseID         uuid.UUID
	AssignedExerciseID *uuid.UUID
	OrderIndex         int16
	SupersetGroup      *int
	Note               string
}

type SetLog struct {
	ID                int64
	ClientLocalID     uuid.UUID
	SessionExerciseID uuid.UUID
	UserID            uuid.UUID
	ExerciseID        uuid.UUID
	SetNumber         int16
	Type              SetType
	WeightKg          *float64
	Reps              *int
	RPE               *float64
	RIR               *int
	DurationSeconds   *int
	DistanceM         *float64
	RestTakenSeconds  *int
	IsCompleted       bool
	PerformedAt       time.Time
}

type PersonalRecord struct {
	ID         int64
	UserID     uuid.UUID
	ExerciseID uuid.UUID
	Type       RecordType
	Value      float64
	SetLogID   *int64
	AchievedAt time.Time
}

// SessionRepository espeja sync.sql.go (6 metodos: registro de
// entrenamiento/series y records personales).
type SessionRepository interface {
	UpsertWorkoutSession(ctx context.Context, s WorkoutSession) (WorkoutSession, error)
	UpsertSessionExercise(ctx context.Context, e SessionExercise) (SessionExercise, error)
	UpsertSetLog(ctx context.Context, s SetLog) (SetLog, error)
	GetPersonalRecord(ctx context.Context, userID, exerciseID uuid.UUID, recordType RecordType) (PersonalRecord, error)
	UpsertPersonalRecord(ctx context.Context, r PersonalRecord) (PersonalRecord, error)
	CountCompletedSessionsOnDate(ctx context.Context, userID uuid.UUID, date time.Time) (int64, error)
}
