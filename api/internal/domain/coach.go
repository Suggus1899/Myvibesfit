package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AttentionThreshold: sin entrenar en mas de 3 dias con asignacion activa
// es la definicion de "necesita atencion" en el overview del coach. Es una
// regla de negocio, por eso vive en domain y no en el service.
const AttentionThreshold = 3 * 24 * time.Hour

// NeedsAttention decide si un cliente aparece marcado en el panel: solo
// aplica a clientes con plan activo — sin plan no hay nada que reclamar.
func NeedsAttention(hasAssignment bool, lastSession *time.Time, now time.Time) bool {
	if !hasAssignment {
		return false
	}
	return lastSession == nil || now.Sub(*lastSession) > AttentionThreshold
}

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

// CoachRepository espeja coach.sql.go (2 metodos).
type CoachRepository interface {
	ListClients(ctx context.Context, orgID, coachUserID uuid.UUID) ([]CoachClientRow, error)

	// LinkClient vincula coach-cliente (upsert sobre coach_client_active_uq):
	// un cliente tiene un solo coach activo por org a la vez, y reasignarlo
	// a otro programa/coach mueve el vinculo, no lo duplica.
	LinkClient(ctx context.Context, orgID, coachUserID, clientUserID uuid.UUID) error
}
