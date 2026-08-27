package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"

	"myvibesfit/api/internal/domain"
)

// AuditLogger registra mutaciones sensibles en audit_log. Un fallo al
// escribir la auditoria nunca debe tumbar la mutacion que la origino, asi
// que solo se loguea el error localmente en vez de propagarlo.
type AuditLogger struct {
	repo domain.AuditRepository
}

func NewAuditLogger(repo domain.AuditRepository) *AuditLogger {
	return &AuditLogger{repo: repo}
}

func (a *AuditLogger) Log(ctx context.Context, orgID, actorID uuid.UUID, action, entityType, entityID string, metadata any) {
	meta, err := json.Marshal(metadata)
	if err != nil {
		meta = []byte("{}")
	}
	if err := a.repo.Insert(ctx, domain.AuditEntry{
		OrgID: orgID, ActorID: actorID, Action: action, EntityType: entityType, EntityID: entityID, Metadata: meta,
	}); err != nil {
		log.Printf("audit log write failed: action=%s entity=%s/%s: %v", action, entityType, entityID, err)
	}
}
