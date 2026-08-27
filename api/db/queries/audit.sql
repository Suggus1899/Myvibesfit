-- name: InsertAuditLog :exec
INSERT INTO audit_log (org_id, actor_user_id, action, entity_type, entity_id, metadata)
VALUES ($1, $2, $3, $4, $5, $6);
