-- name: CreateUser :one
INSERT INTO app_user (email, password_hash, full_name)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM app_user WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT * FROM app_user WHERE id = $1 AND deleted_at IS NULL;

-- name: TouchUserLogin :exec
UPDATE app_user SET last_login_at = now() WHERE id = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_token (user_id, token_hash, device_label, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_token WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RevokeRefreshTokenByHash :exec
UPDATE refresh_token SET revoked_at = now() WHERE token_hash = $1;

-- name: GetOrganizationByJoinCode :one
SELECT * FROM organization WHERE join_code = $1 AND status = 'active';

-- name: GetOrganizationByID :one
SELECT * FROM organization WHERE id = $1;

-- name: CreateOrganization :one
INSERT INTO organization (name, slug, join_code, brand_color)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateMembership :one
INSERT INTO membership (org_id, user_id, role, status, joined_at)
VALUES ($1, $2, $3, 'active', now())
ON CONFLICT (org_id, user_id, role)
DO UPDATE SET status = 'active', joined_at = now()
RETURNING *;

-- name: GetActiveMembershipByUser :one
SELECT * FROM membership
WHERE user_id = $1 AND status = 'active'
ORDER BY joined_at DESC
LIMIT 1;

-- name: ListOrgMembers :many
SELECT m.id AS membership_id, m.user_id, m.role, m.status, m.joined_at,
       u.full_name, u.email, u.avatar_url
FROM membership m
JOIN app_user u ON u.id = m.user_id
WHERE m.org_id = $1 AND m.status = 'active'
ORDER BY u.full_name;

-- name: UpdateMemberRole :one
UPDATE membership SET role = $3
WHERE id = $1 AND org_id = $2 AND status = 'active'
RETURNING *;
