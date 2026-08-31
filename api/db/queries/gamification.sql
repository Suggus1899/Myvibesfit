-- name: GetUserStats :one
SELECT * FROM user_stats WHERE user_id = $1;

-- name: UpsertUserStats :one
INSERT INTO user_stats (user_id, total_xp, level, total_sessions, total_volume_kg)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id) DO UPDATE SET
  total_xp = EXCLUDED.total_xp, level = EXCLUDED.level, total_sessions = EXCLUDED.total_sessions,
  total_volume_kg = EXCLUDED.total_volume_kg, updated_at = now()
RETURNING *;

-- name: GetUserStreak :one
SELECT * FROM user_streak WHERE user_id = $1 AND kind = $2;

-- name: UpsertUserStreak :one
INSERT INTO user_streak (user_id, kind, current_count, longest_count, last_active_on)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, kind) DO UPDATE SET
  current_count = EXCLUDED.current_count, longest_count = EXCLUDED.longest_count,
  last_active_on = EXCLUDED.last_active_on
RETURNING *;

-- name: ListUserStreaks :many
SELECT * FROM user_streak WHERE user_id = $1;

-- name: CreateXPEvent :one
INSERT INTO xp_event (user_id, source, points, reference_id) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: ListAchievements :many
SELECT * FROM achievement ORDER BY tier, name;

-- name: GetUserAchievement :one
SELECT * FROM user_achievement WHERE user_id = $1 AND achievement_id = $2;

-- name: UpsertUserAchievement :one
INSERT INTO user_achievement (user_id, achievement_id, progress, earned_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, achievement_id) DO UPDATE SET
  progress = EXCLUDED.progress,
  earned_at = COALESCE(user_achievement.earned_at, EXCLUDED.earned_at)
RETURNING *;

-- name: ListUserAchievements :many
SELECT ua.user_id, ua.achievement_id, ua.progress, ua.earned_at,
       a.code, a.name, a.description, a.icon, a.tier, a.xp_reward
FROM user_achievement ua
JOIN achievement a ON a.id = ua.achievement_id
WHERE ua.user_id = $1
ORDER BY a.tier, a.name;

-- name: GetUserStatsForUpdate :one
-- Variante con lock de fila para el read-modify-write de ApplyStatsDelta.
-- La version sin lock se usa en las lecturas de solo lectura (/me/stats),
-- que no deberian bloquear a nadie.
SELECT * FROM user_stats WHERE user_id = $1 FOR UPDATE;

-- name: GetUserStreakForUpdate :one
SELECT * FROM user_streak WHERE user_id = $1 AND kind = $2 FOR UPDATE;
