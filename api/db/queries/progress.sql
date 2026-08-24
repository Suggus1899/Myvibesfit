-- name: ListSetLogsForExercise :many
SELECT * FROM set_log
WHERE user_id = $1 AND exercise_id = $2 AND type = 'working' AND is_completed = true
  AND performed_at >= $3
ORDER BY performed_at DESC
LIMIT $4;

-- name: ListPersonalRecords :many
SELECT * FROM personal_record
WHERE user_id = $1 AND (sqlc.narg('exercise_id')::uuid IS NULL OR exercise_id = sqlc.narg('exercise_id'))
ORDER BY achieved_at DESC;

-- name: ListSessionsForVolume :many
SELECT id, started_at, total_volume_kg, duration_seconds
FROM workout_session
WHERE user_id = $1 AND status = 'completed' AND started_at >= $2
ORDER BY started_at DESC;

-- name: UpsertBodyMetric :one
INSERT INTO body_metric (user_id, measured_on, weight_kg, body_fat_pct, note)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, measured_on) DO UPDATE SET
  weight_kg = EXCLUDED.weight_kg, body_fat_pct = EXCLUDED.body_fat_pct, note = EXCLUDED.note
RETURNING *;

-- name: ListBodyMetrics :many
SELECT * FROM body_metric WHERE user_id = $1 ORDER BY measured_on DESC LIMIT $2;
