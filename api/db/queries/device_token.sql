-- name: UpsertDeviceToken :one
-- El conflicto se resuelve por token, no por (user_id, token): el token es
-- UNIQUE global y FCM devuelve el mismo cuando otro usuario entra en el
-- mismo telefono, asi que la fila tiene que cambiar de dueno.
INSERT INTO device_token (user_id, token, platform)
VALUES ($1, $2, $3)
ON CONFLICT (token) DO UPDATE
  SET user_id = EXCLUDED.user_id,
      platform = EXCLUDED.platform,
      last_seen_at = now()
RETURNING *;

-- name: ListDeviceTokensByUser :many
SELECT * FROM device_token
WHERE user_id = $1
ORDER BY last_seen_at DESC;

-- name: DeleteDeviceToken :execrows
DELETE FROM device_token
WHERE token = $1 AND user_id = $2;
