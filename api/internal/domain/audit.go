package domain

import (
	"context"

	"github.com/google/uuid"
)

type AuditEntry struct {
	OrgID      uuid.UUID
	ActorID    uuid.UUID
	Action     string
	EntityType string
	EntityID   string
	Metadata   []byte
}

// AuditRepository espeja audit.sql.go (1 metodo).
type AuditRepository interface {
	Insert(ctx context.Context, e AuditEntry) error
}
