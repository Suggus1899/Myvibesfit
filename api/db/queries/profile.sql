-- name: GetClientProfile :one
SELECT * FROM client_profile WHERE user_id = $1;

-- name: UpsertClientProfile :one
INSERT INTO client_profile (
  user_id, birth_date, sex, height_cm, experience, primary_goal,
  days_per_week, session_minutes, available_equipment, limitations,
  unit_system, onboarded_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now()
)
ON CONFLICT (user_id) DO UPDATE SET
  birth_date = EXCLUDED.birth_date,
  sex = EXCLUDED.sex,
  height_cm = EXCLUDED.height_cm,
  experience = EXCLUDED.experience,
  primary_goal = EXCLUDED.primary_goal,
  days_per_week = EXCLUDED.days_per_week,
  session_minutes = EXCLUDED.session_minutes,
  available_equipment = EXCLUDED.available_equipment,
  limitations = EXCLUDED.limitations,
  unit_system = EXCLUDED.unit_system,
  -- onboarded_at marca la primera vez: no se pisa en ediciones posteriores.
  onboarded_at = COALESCE(client_profile.onboarded_at, now())
RETURNING *;
