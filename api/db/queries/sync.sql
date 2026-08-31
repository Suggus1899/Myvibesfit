-- name: UpsertWorkoutSession :one
INSERT INTO workout_session (
  client_local_id, user_id, org_id, assigned_workout_id, name, status,
  started_at, ended_at, duration_seconds, total_volume_kg, perceived_effort, mood, notes
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (user_id, client_local_id) DO UPDATE SET
  status = EXCLUDED.status, ended_at = EXCLUDED.ended_at, duration_seconds = EXCLUDED.duration_seconds,
  total_volume_kg = EXCLUDED.total_volume_kg, perceived_effort = EXCLUDED.perceived_effort,
  mood = EXCLUDED.mood, notes = EXCLUDED.notes, synced_at = now()
-- xmax = 0 distingue una fila recien insertada de una que el upsert
-- actualizo: es como Postgres deja ver si el ON CONFLICT se disparo. Sin
-- esto, reenviar el mismo lote vuelve a otorgar XP, racha y logros.
RETURNING *, (xmax = 0) AS inserted;

-- name: UpsertSessionExercise :one
INSERT INTO session_exercise (session_id, exercise_id, assigned_exercise_id, order_index, superset_group, note)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (session_id, order_index) DO UPDATE SET
  exercise_id = EXCLUDED.exercise_id, assigned_exercise_id = EXCLUDED.assigned_exercise_id,
  superset_group = EXCLUDED.superset_group, note = EXCLUDED.note
RETURNING *;

-- name: UpsertSetLog :one
INSERT INTO set_log (
  client_local_id, session_exercise_id, user_id, exercise_id, set_number, type,
  weight_kg, reps, rpe, rir, duration_seconds, distance_m, rest_taken_seconds,
  is_completed, performed_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
)
ON CONFLICT (user_id, client_local_id) DO UPDATE SET
  weight_kg = EXCLUDED.weight_kg, reps = EXCLUDED.reps, rpe = EXCLUDED.rpe, rir = EXCLUDED.rir,
  duration_seconds = EXCLUDED.duration_seconds, distance_m = EXCLUDED.distance_m,
  rest_taken_seconds = EXCLUDED.rest_taken_seconds, is_completed = EXCLUDED.is_completed
RETURNING *;

-- name: GetPersonalRecord :one
SELECT * FROM personal_record WHERE user_id = $1 AND exercise_id = $2 AND type = $3;

-- name: UpsertPersonalRecord :one
INSERT INTO personal_record (user_id, exercise_id, type, value, set_log_id, achieved_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, exercise_id, type) DO UPDATE SET
  value = EXCLUDED.value, set_log_id = EXCLUDED.set_log_id, achieved_at = EXCLUDED.achieved_at
RETURNING *;

-- name: CountCompletedSessionsOnDate :one
SELECT count(*) FROM workout_session
WHERE user_id = $1 AND status = 'completed' AND started_at::date = $2::date;

-- name: ListPersonalRecordsForExercise :many
-- Los 4 tipos de record de un ejercicio en una sola ida a la base: antes se
-- consultaba uno por uno por cada serie de trabajo.
SELECT * FROM personal_record WHERE user_id = $1 AND exercise_id = $2;
