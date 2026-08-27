-- name: ListActiveAssignmentsForWorker :many
SELECT a.id AS assignment_id, a.org_id, a.client_user_id, a.coach_user_id,
       u.full_name AS client_name
FROM assignment a
JOIN app_user u ON u.id = a.client_user_id
WHERE a.status = 'active';

-- name: CountCompletedSessionsSince :one
SELECT count(*) FROM workout_session
WHERE user_id = $1 AND status = 'completed' AND started_at >= $2;

-- name: CountAssignedWorkoutsInRange :one
SELECT count(*) FROM assigned_workout
WHERE assignment_id = $1 AND scheduled_on IS NOT NULL
  AND scheduled_on BETWEEN $2 AND $3;

-- name: CountCompletedHabitLogsSince :one
SELECT count(*) FROM habit_log hl
JOIN client_habit ch ON ch.id = hl.client_habit_id
WHERE hl.user_id = $1 AND hl.log_date >= $2 AND hl.is_completed = true AND ch.ended_on IS NULL;

-- name: ListRecentWorkingSets :many
SELECT sl.exercise_id, e.name AS exercise_name, sl.weight_kg, sl.reps, sl.rpe, sl.performed_at
FROM set_log sl
JOIN exercise e ON e.id = sl.exercise_id
WHERE sl.user_id = $1 AND sl.type = 'working' AND sl.is_completed = true AND sl.performed_at >= $2
ORDER BY sl.performed_at ASC
LIMIT $3;

-- name: CreateAISuggestion :one
INSERT INTO ai_suggestion (
  org_id, client_user_id, coach_user_id, assignment_id, kind, payload,
  rationale, confidence, model, input_snapshot, expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: ListPendingSuggestionsForCoach :many
SELECT s.*, u.full_name AS client_name
FROM ai_suggestion s
JOIN app_user u ON u.id = s.client_user_id
WHERE s.coach_user_id = $1 AND s.org_id = $2 AND s.status = 'pending'
ORDER BY s.created_at DESC;

-- name: GetAISuggestionByID :one
SELECT * FROM ai_suggestion WHERE id = $1 AND org_id = $2;

-- name: ReviewAISuggestion :one
UPDATE ai_suggestion
SET status = $3, reviewed_by = $4, reviewed_at = now()
WHERE id = $1 AND org_id = $2 AND status = 'pending'
RETURNING *;
