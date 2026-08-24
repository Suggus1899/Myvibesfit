-- name: ListHabits :many
SELECT * FROM habit
WHERE org_id IS NULL OR org_id = sqlc.narg('org_id')
ORDER BY is_system DESC, name;

-- name: GetHabitByID :one
SELECT * FROM habit WHERE id = $1;

-- name: GetActiveClientHabit :one
SELECT * FROM client_habit WHERE user_id = $1 AND habit_id = $2 AND ended_on IS NULL;

-- name: SubscribeHabit :one
INSERT INTO client_habit (user_id, habit_id, assigned_by, target_value, frequency, days_of_week)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListMyHabits :many
SELECT ch.id, ch.user_id, ch.habit_id, ch.target_value, ch.frequency, ch.days_of_week,
       ch.started_on, ch.ended_on, h.name AS habit_name, h.icon AS habit_icon, h.unit AS habit_unit
FROM client_habit ch
JOIN habit h ON h.id = ch.habit_id
WHERE ch.user_id = $1 AND ch.ended_on IS NULL
ORDER BY h.name;

-- name: UnsubscribeHabit :exec
UPDATE client_habit SET ended_on = CURRENT_DATE WHERE id = $1 AND user_id = $2;

-- name: UpsertHabitLog :one
INSERT INTO habit_log (client_local_id, client_habit_id, user_id, log_date, value, is_completed)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, client_local_id) DO UPDATE SET
  value = EXCLUDED.value, is_completed = EXCLUDED.is_completed
RETURNING *;

-- name: ListHabitLogsForDate :many
SELECT * FROM habit_log WHERE user_id = $1 AND log_date = $2;

-- name: CountActiveHabitsForUser :one
SELECT count(*) FROM client_habit WHERE user_id = $1 AND ended_on IS NULL;

-- name: CountCompletedHabitLogsForDate :one
SELECT count(*) FROM habit_log hl
JOIN client_habit ch ON ch.id = hl.client_habit_id
WHERE hl.user_id = $1 AND hl.log_date = $2 AND hl.is_completed = true AND ch.ended_on IS NULL;
