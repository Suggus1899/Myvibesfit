package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CoachClientRow es el read-model desnormalizado de coach.sql
// (ListCoachClients): no es un agregado de dominio, es un reporte.
type CoachClientRow struct {
	ClientUserID     uuid.UUID
	FullName         string
	AvatarURL        string
	AssignmentID     *uuid.UUID
	AssignmentName   string
	AssignmentStatus string
	LastSessionAt    time.Time
	StreakDays       int32
	RecentPRs        int32
}

// CoachRepository espeja coach.sql.go (1 metodo).
type CoachRepository interface {
	ListClients(ctx context.Context, orgID, coachUserID uuid.UUID) ([]CoachClientRow, error)
}
