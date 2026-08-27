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
  name = $3, description = $4, instructions = $5, pattern = $6,
  mechanic = $7, primary_muscle = $8, secondary_muscles = $9,
  equipment = $10, difficulty = $11, tracking = $12, is_unilateral = $13,
  video_url = $14, thumbnail_url = $15
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeactivateExercise :execrows
UPDATE exercise SET is_active = false WHERE id = $1 AND org_id = $2;
