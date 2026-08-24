-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- EJECUCION DE ENTRENAMIENTOS
-- client_local_id: UUID que genera el movil. Da idempotencia a la cola
-- de sincronizacion: reenviar la misma sesion no la duplica.
-- ============================================================

CREATE TYPE session_status AS ENUM ('in_progress','completed','abandoned');

CREATE TABLE workout_session (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  client_local_id     uuid NOT NULL,
  user_id             uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  org_id              uuid REFERENCES organization(id) ON DELETE SET NULL,
  assigned_workout_id uuid REFERENCES assigned_workout(id) ON DELETE SET NULL,
  name                text NOT NULL,
  status              session_status NOT NULL DEFAULT 'in_progress',
  started_at          timestamptz NOT NULL,
  ended_at            timestamptz,
  duration_seconds    integer,
  total_volume_kg     numeric(10,2) NOT NULL DEFAULT 0,
  perceived_effort    smallint CHECK (perceived_effort BETWEEN 1 AND 10),
  mood                smallint CHECK (mood BETWEEN 1 AND 5),
  notes               text,
  synced_at           timestamptz NOT NULL DEFAULT now(),
  created_at          timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, client_local_id)
);
CREATE INDEX ON workout_session (user_id, started_at DESC);
CREATE INDEX ON workout_session (org_id, started_at DESC) WHERE status = 'completed';

CREATE TABLE session_exercise (
  id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id           uuid NOT NULL REFERENCES workout_session(id) ON DELETE CASCADE,
  exercise_id          uuid NOT NULL REFERENCES exercise(id),
  assigned_exercise_id uuid REFERENCES assigned_exercise(id) ON DELETE SET NULL,
  order_index          smallint NOT NULL,
  superset_group       smallint,
  note                 text,
  UNIQUE (session_id, order_index)
);

CREATE TYPE set_type AS ENUM ('warmup','working','drop','failure','backoff');

-- Tabla de mayor volumen del sistema.
-- user_id / exercise_id / performed_at estan desnormalizados a proposito:
-- la consulta "progreso de este ejercicio en 6 meses" se resuelve con un
-- solo indice en lugar de tres joins sobre millones de filas.
CREATE TABLE set_log (
  id                  bigserial PRIMARY KEY,
  client_local_id     uuid NOT NULL,
  session_exercise_id uuid NOT NULL REFERENCES session_exercise(id) ON DELETE CASCADE,
  user_id             uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  exercise_id         uuid NOT NULL REFERENCES exercise(id),
  set_number          smallint NOT NULL,
  type                set_type NOT NULL DEFAULT 'working',
  weight_kg           numeric(6,2),
  reps                smallint,
  rpe                 numeric(3,1) CHECK (rpe BETWEEN 1 AND 10),
  rir                 smallint CHECK (rir BETWEEN 0 AND 10),
  duration_seconds    integer,
  distance_m          numeric(8,2),
  rest_taken_seconds  integer,
  is_completed        boolean NOT NULL DEFAULT true,
  performed_at        timestamptz NOT NULL DEFAULT now(),
  UNIQUE (session_exercise_id, set_number),
  UNIQUE (user_id, client_local_id)
);
CREATE INDEX set_log_progress_idx
  ON set_log (user_id, exercise_id, performed_at DESC)
  WHERE type = 'working' AND is_completed;

CREATE TYPE record_type AS ENUM ('max_weight','max_reps','estimated_1rm','max_volume_set');

CREATE TABLE personal_record (
  id            bigserial PRIMARY KEY,
  user_id       uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  exercise_id   uuid NOT NULL REFERENCES exercise(id) ON DELETE CASCADE,
  type          record_type NOT NULL,
  value         numeric(8,2) NOT NULL,
  set_log_id    bigint REFERENCES set_log(id) ON DELETE SET NULL,
  achieved_at   timestamptz NOT NULL,
  UNIQUE (user_id, exercise_id, type)
);
CREATE INDEX ON personal_record (user_id, achieved_at DESC);

-- ============================================================
-- HABITOS
-- ============================================================

CREATE TYPE habit_unit AS ENUM ('boolean','count','minutes','milliliters','steps');
CREATE TYPE habit_frequency AS ENUM ('daily','weekly','specific_days');

CREATE TABLE habit (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id         uuid REFERENCES organization(id) ON DELETE CASCADE,
  slug           citext NOT NULL,
  name           text NOT NULL,
  icon           text NOT NULL DEFAULT 'check',
  unit           habit_unit NOT NULL DEFAULT 'boolean',
  default_target numeric(8,2),
  is_system      boolean NOT NULL DEFAULT false,
  created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX habit_global_slug_uq ON habit (slug) WHERE org_id IS NULL;
CREATE UNIQUE INDEX habit_org_slug_uq    ON habit (org_id, slug) WHERE org_id IS NOT NULL;

CREATE TABLE client_habit (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  habit_id      uuid NOT NULL REFERENCES habit(id) ON DELETE CASCADE,
  assigned_by   uuid REFERENCES app_user(id),
  target_value  numeric(8,2),
  frequency     habit_frequency NOT NULL DEFAULT 'daily',
  days_of_week  smallint[] NOT NULL DEFAULT '{}',
  reminder_time time,
  started_on    date NOT NULL DEFAULT CURRENT_DATE,
  ended_on      date,
  UNIQUE (user_id, habit_id, started_on)
);
CREATE INDEX ON client_habit (user_id) WHERE ended_on IS NULL;

CREATE TABLE habit_log (
  id              bigserial PRIMARY KEY,
  client_local_id uuid NOT NULL,
  client_habit_id uuid NOT NULL REFERENCES client_habit(id) ON DELETE CASCADE,
  user_id         uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  log_date        date NOT NULL,
  value           numeric(8,2) NOT NULL DEFAULT 1,
  is_completed    boolean NOT NULL DEFAULT true,
  logged_at       timestamptz NOT NULL DEFAULT now(),
  UNIQUE (client_habit_id, log_date),
  UNIQUE (user_id, client_local_id)
);
CREATE INDEX ON habit_log (user_id, log_date DESC);

-- ============================================================
-- GAMIFICACION
-- ============================================================

CREATE TYPE streak_kind AS ENUM ('workout','habit','overall');

CREATE TABLE user_streak (
  user_id         uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  kind            streak_kind NOT NULL,
  current_count   integer NOT NULL DEFAULT 0,
  longest_count   integer NOT NULL DEFAULT 0,
  last_active_on  date,
  freezes_left    smallint NOT NULL DEFAULT 0,
  PRIMARY KEY (user_id, kind)
);

CREATE TABLE user_stats (
  user_id           uuid PRIMARY KEY REFERENCES app_user(id) ON DELETE CASCADE,
  total_xp          integer NOT NULL DEFAULT 0,
  level             smallint NOT NULL DEFAULT 1,
  total_sessions    integer NOT NULL DEFAULT 0,
  total_volume_kg   numeric(12,2) NOT NULL DEFAULT 0,
  updated_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TYPE xp_source AS ENUM ('session','habit','personal_record','streak','achievement');

CREATE TABLE xp_event (
  id            bigserial PRIMARY KEY,
  user_id       uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  source        xp_source NOT NULL,
  points        integer NOT NULL,
  reference_id  text,
  created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON xp_event (user_id, created_at DESC);

CREATE TABLE achievement (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code          citext NOT NULL UNIQUE,
  name          text NOT NULL,
  description   text NOT NULL,
  icon          text NOT NULL,
  tier          smallint NOT NULL DEFAULT 1,
  xp_reward     integer NOT NULL DEFAULT 0,
  criteria      jsonb NOT NULL
);

CREATE TABLE user_achievement (
  user_id         uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  achievement_id  uuid NOT NULL REFERENCES achievement(id) ON DELETE CASCADE,
  progress        numeric(5,2) NOT NULL DEFAULT 0,
  earned_at       timestamptz,
  PRIMARY KEY (user_id, achievement_id)
);

-- ============================================================
-- SUGERENCIAS DE IA
-- La IA nunca escribe en assigned_exercise: propone, el coach aprueba.
-- input_snapshot guarda los datos exactos con los que se genero; sin eso
-- una sugerencia no es auditable.
-- ============================================================

CREATE TYPE suggestion_kind AS ENUM ('volume_adjust','load_adjust','exercise_swap','deload','rest_day','habit_nudge');
CREATE TYPE suggestion_status AS ENUM ('pending','approved','rejected','expired','auto_applied');

CREATE TABLE ai_suggestion (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id          uuid NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
  client_user_id  uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  coach_user_id   uuid REFERENCES app_user(id) ON DELETE SET NULL,
  assignment_id   uuid REFERENCES assignment(id) ON DELETE CASCADE,
  kind            suggestion_kind NOT NULL,
  payload         jsonb NOT NULL,
  rationale       text NOT NULL,
  confidence      numeric(3,2),
  model           text NOT NULL,
  input_snapshot  jsonb NOT NULL,
  status          suggestion_status NOT NULL DEFAULT 'pending',
  reviewed_by     uuid REFERENCES app_user(id),
  reviewed_at     timestamptz,
  applied_at      timestamptz,
  expires_at      timestamptz NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON ai_suggestion (coach_user_id, status) WHERE status = 'pending';
CREATE INDEX ON ai_suggestion (client_user_id, created_at DESC);

-- ============================================================
-- NOTIFICACIONES Y AUDITORIA
-- ============================================================

CREATE TABLE device_token (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  token         text NOT NULL UNIQUE,
  platform      text NOT NULL CHECK (platform IN ('ios','android')),
  last_seen_at  timestamptz NOT NULL DEFAULT now(),
  created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON device_token (user_id);

CREATE TABLE audit_log (
  id            bigserial PRIMARY KEY,
  org_id        uuid REFERENCES organization(id) ON DELETE SET NULL,
  actor_user_id uuid REFERENCES app_user(id) ON DELETE SET NULL,
  action        text NOT NULL,
  entity_type   text NOT NULL,
  entity_id     text NOT NULL,
  metadata      jsonb NOT NULL DEFAULT '{}',
  ip_address    inet,
  created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON audit_log (org_id, created_at DESC);
CREATE INDEX ON audit_log (entity_type, entity_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_log, device_token, ai_suggestion, user_achievement, achievement,
  xp_event, user_stats, user_streak, habit_log, client_habit, habit,
  personal_record, set_log, session_exercise, workout_session CASCADE;
DROP TYPE IF EXISTS suggestion_status, suggestion_kind, xp_source, streak_kind,
  habit_frequency, habit_unit, record_type, set_type, session_status;
-- +goose StatementEnd
