package postgres

import (
	"context"

	"myvibesfit/api/internal/domain"
	"myvibesfit/api/internal/repository/db"
)

type AuditRepository struct{ q db.Querier }

func NewAuditRepository(q db.Querier) *AuditRepository { return &AuditRepository{q: q} }

var _ domain.AuditRepository = (*AuditRepository)(nil)

func (r *AuditRepository) Insert(ctx context.Context, e domain.AuditEntry) error {
	return r.q.InsertAuditLog(ctx, db.InsertAuditLogParams{
		OrgID: uuidPtrToPgUUID(&e.OrgID), ActorUserID: uuidPtrToPgUUID(&e.ActorID),
		Action: e.Action, EntityType: e.EntityType, EntityID: e.EntityID, Metadata: e.Metadata,
	})
}
