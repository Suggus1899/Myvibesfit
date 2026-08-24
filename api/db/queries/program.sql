-- name: CreateProgressionRule :one
INSERT INTO progression_rule (org_id, name, type, params, is_system)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListProgressionRules :many
SELECT * FROM progression_rule
WHERE org_id IS NULL OR org_id = sqlc.narg('org_id')
ORDER BY is_system DESC, name;

-- name: GetProgressionRuleByID :one
SELECT * FROM progression_rule WHERE id = $1;

-- name: CreateProgram :one
INSERT INTO program (org_id, created_by, name, description, goal, level, total_weeks, days_per_week)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetProgramByID :one
SELECT * FROM program WHERE id = $1 AND org_id = $2;

-- name: ListProgramsByOrg :many
SELECT * FROM program
WHERE org_id = $1
  AND (sqlc.narg('status')::program_status IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: UpdateProgram :one
UPDATE program SET
  name = $3, description = $4, goal = $5, level = $6, total_weeks = $7, days_per_week = $8
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: SetProgramStatus :one
UPDATE program SET status = $3 WHERE id = $1 AND org_id = $2 RETURNING *;

-- name: CreateProgramWorkout :one
INSERT INTO program_workout (program_id, week_number, day_index, name, note, is_deload, estimated_minutes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListProgramWorkouts :many
SELECT * FROM program_workout WHERE program_id = $1 ORDER BY week_number, day_index;

-- name: GetProgramWorkoutWithOrg :one
SELECT pw.*, p.org_id AS program_org_id
FROM program_workout pw
JOIN program p ON p.id = pw.program_id
WHERE pw.id = $1;

-- name: UpdateProgramWorkout :one
UPDATE program_workout SET
  name = $2, note = $3, is_deload = $4, estimated_minutes = $5
WHERE id = $1
RETURNING *;

-- name: DeleteProgramWorkout :exec
DELETE FROM program_workout WHERE id = $1;

-- name: CreateProgramExercise :one
INSERT INTO program_exercise (
  program_workout_id, exercise_id, order_index, superset_group, target_sets,
  target_reps_min, target_reps_max, target_rpe, target_pct_1rm, rest_seconds,
  tempo, note, progression_rule_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: ListProgramExercises :many
SELECT * FROM program_exercise WHERE program_workout_id = $1 ORDER BY order_index;

-- name: GetProgramExerciseWithOrg :one
SELECT pe.*, p.org_id AS program_org_id
FROM program_exercise pe
JOIN program_workout pw ON pw.id = pe.program_workout_id
JOIN program p ON p.id = pw.program_id
WHERE pe.id = $1;

-- name: UpdateProgramExercise :one
UPDATE program_exercise SET
  order_index = $2, superset_group = $3, target_sets = $4, target_reps_min = $5,
  target_reps_max = $6, target_rpe = $7, target_pct_1rm = $8, rest_seconds = $9,
  tempo = $10, note = $11, progression_rule_id = $12
WHERE id = $1
RETURNING *;

-- name: DeleteProgramExercise :exec
DELETE FROM program_exercise WHERE id = $1;
