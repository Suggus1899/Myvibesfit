-- name: ListCoachClients :many
SELECT
  cc.client_user_id,
  u.full_name,
  u.avatar_url,
  a.id AS assignment_id,
  a.name AS assignment_name,
  COALESCE((a.status)::text, '')::text AS assignment_status,
  COALESCE(
    (SELECT started_at FROM workout_session
      WHERE user_id = cc.client_user_id AND status = 'completed'
      ORDER BY started_at DESC LIMIT 1),
    'epoch'::timestamptz
  )::timestamptz AS last_session_at,
  COALESCE(st.current_count, 0)::int AS streak_days,
  COALESCE(pr.recent_prs, 0)::int AS recent_prs
FROM coach_client cc
JOIN app_user u ON u.id = cc.client_user_id
LEFT JOIN assignment a ON a.client_user_id = cc.client_user_id AND a.status = 'active'
LEFT JOIN user_streak st ON st.user_id = cc.client_user_id AND st.kind = 'workout'
LEFT JOIN LATERAL (
  SELECT count(*)::int AS recent_prs FROM personal_record
  WHERE user_id = cc.client_user_id AND achieved_at >= now() - interval '14 days'
) pr ON true
WHERE cc.org_id = $1 AND cc.coach_user_id = $2 AND cc.ended_at IS NULL
ORDER BY u.full_name;

-- name: LinkCoachClient :exec
INSERT INTO coach_client (org_id, coach_user_id, client_user_id)
VALUES ($1, $2, $3)
ON CONFLICT (org_id, client_user_id) WHERE ended_at IS NULL
DO UPDATE SET coach_user_id = EXCLUDED.coach_user_id;
