-- name: ListExercises :many
SELECT * FROM exercise
WHERE is_active = true
  AND (org_id IS NULL OR org_id = sqlc.narg('org_id'))
  AND (sqlc.narg('pattern')::movement_pattern IS NULL OR pattern = sqlc.narg('pattern'))
  AND (sqlc.narg('muscle')::text IS NULL OR primary_muscle = sqlc.narg('muscle'))
ORDER BY name
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: GetExerciseByID :one
SELECT * FROM exercise WHERE id = $1 AND is_active = true;

-- name: CreateExercise :one
INSERT INTO exercise (
  org_id, slug, name, description, instructions, pattern, mechanic,
  primary_muscle, secondary_muscles, equipment, difficulty, tracking,
  is_unilateral, video_url, thumbnail_url, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING *;

-- name: UpdateExercise :one
UPDATE exercise SET
  name = $2, description = $3, instructions = $4, pattern = $5,
  mechanic = $6, primary_muscle = $7, secondary_muscles = $8,
  equipment = $9, difficulty = $10, tracking = $11, is_unilateral = $12,
  video_url = $13, thumbnail_url = $14
WHERE id = $1
RETURNING *;

-- name: DeactivateExercise :exec
UPDATE exercise SET is_active = false WHERE id = $1;
