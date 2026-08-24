-- name: GetOrgMembership :one
SELECT * FROM membership WHERE org_id = $1 AND user_id = $2 AND status = 'active' LIMIT 1;

-- name: GetActiveAssignmentByClient :one
SELECT * FROM assignment WHERE client_user_id = $1 AND status = 'active';

-- name: CreateAssignment :one
INSERT INTO assignment (org_id, program_id, client_user_id, coach_user_id, name, start_date, status)
VALUES ($1, $2, $3, $4, $5, $6, 'active')
RETURNING *;

-- name: CancelAssignment :one
UPDATE assignment SET status = 'cancelled', end_date = CURRENT_DATE
WHERE id = $1 AND org_id = $2 AND status = 'active'
RETURNING *;

-- name: GetAssignmentByID :one
SELECT * FROM assignment WHERE id = $1;

-- name: CreateAssignedWorkout :one
INSERT INTO assigned_workout (assignment_id, source_workout_id, week_number, day_index, name, note, is_deload, scheduled_on, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending')
RETURNING *;

-- name: ListAssignedWorkouts :many
SELECT * FROM assigned_workout WHERE assignment_id = $1 ORDER BY week_number, day_index;

-- name: CreateAssignedExercise :one
INSERT INTO assigned_exercise (
  assigned_workout_id, exercise_id, order_index, superset_group, target_sets,
  target_reps_min, target_reps_max, target_rpe, target_weight_kg, rest_seconds,
  tempo, note, progression_rule_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: ListAssignedExercises :many
SELECT * FROM assigned_exercise WHERE assigned_workout_id = $1 ORDER BY order_index;
