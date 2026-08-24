-- +goose Up
-- +goose StatementBegin

-- ============================================================
-- CATALOGO DE EJERCICIOS
-- org_id NULL = ejercicio global del sistema; con valor = propio del gimnasio
-- ============================================================

CREATE TYPE movement_pattern AS ENUM (
  'squat','hinge','horizontal_push','vertical_push','horizontal_pull','vertical_pull',
  'lunge','carry','rotation','isolation','cardio'
);
CREATE TYPE exercise_mechanic AS ENUM ('compound', 'isolation');
CREATE TYPE tracking_mode AS ENUM ('weight_reps','bodyweight_reps','weighted_bodyweight','duration','distance_duration');

CREATE TABLE exercise (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id            uuid REFERENCES organization(id) ON DELETE CASCADE,
  slug              citext NOT NULL,
  name              text NOT NULL,
  description       text,
  instructions      text[] NOT NULL DEFAULT '{}',
  pattern           movement_pattern NOT NULL,
  mechanic          exercise_mechanic NOT NULL DEFAULT 'compound',
  primary_muscle    text NOT NULL,
  secondary_muscles text[] NOT NULL DEFAULT '{}',
  equipment         text[] NOT NULL DEFAULT '{}',
  difficulty        experience_level NOT NULL DEFAULT 'beginner',
  tracking          tracking_mode NOT NULL DEFAULT 'weight_reps',
  is_unilateral     boolean NOT NULL DEFAULT false,
  video_url         text,
  thumbnail_url     text,
  created_by        uuid REFERENCES app_user(id),
  is_active         boolean NOT NULL DEFAULT true,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now()
);
-- NULL no colisiona en UNIQUE: hacen falta dos indices parciales
CREATE UNIQUE INDEX exercise_global_slug_uq ON exercise (slug) WHERE org_id IS NULL;
CREATE UNIQUE INDEX exercise_org_slug_uq    ON exercise (org_id, slug) WHERE org_id IS NOT NULL;
CREATE INDEX exercise_lookup_idx ON exercise (org_id, pattern, primary_muscle) WHERE is_active;
CREATE TRIGGER exercise_updated_at BEFORE UPDATE ON exercise
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Sustituciones validas: alimenta el swap manual y el que sugiere la IA
CREATE TABLE exercise_alternative (
  exercise_id     uuid NOT NULL REFERENCES exercise(id) ON DELETE CASCADE,
  alternative_id  uuid NOT NULL REFERENCES exercise(id) ON DELETE CASCADE,
  reason          text,
  PRIMARY KEY (exercise_id, alternative_id),
  CHECK (exercise_id <> alternative_id)
);

-- ============================================================
-- REGLAS DE PROGRESION (deterministas; la IA nunca las reemplaza)
-- ============================================================

CREATE TYPE progression_type AS ENUM ('double_progression','linear_load','rpe_autoregulated','percentage_1rm','none');

CREATE TABLE progression_rule (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id        uuid REFERENCES organization(id) ON DELETE CASCADE,
  name          text NOT NULL,
  type          progression_type NOT NULL,
  params        jsonb NOT NULL DEFAULT '{}',
  is_system     boolean NOT NULL DEFAULT false,
  created_at    timestamptz NOT NULL DEFAULT now()
);

-- ============================================================
-- PROGRAMAS (plantilla que disena el coach en el panel web)
-- ============================================================

CREATE TYPE program_status AS ENUM ('draft','published','archived');

CREATE TABLE program (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id          uuid NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
  created_by      uuid NOT NULL REFERENCES app_user(id),
  name            text NOT NULL,
  description     text,
  goal            training_goal NOT NULL DEFAULT 'general_health',
  level           experience_level NOT NULL DEFAULT 'beginner',
  total_weeks     smallint NOT NULL DEFAULT 4 CHECK (total_weeks BETWEEN 1 AND 52),
  days_per_week   smallint NOT NULL DEFAULT 3,
  status          program_status NOT NULL DEFAULT 'draft',
  is_shared       boolean NOT NULL DEFAULT true,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON program (org_id, status);
CREATE TRIGGER program_updated_at BEFORE UPDATE ON program
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE program_workout (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  program_id        uuid NOT NULL REFERENCES program(id) ON DELETE CASCADE,
  week_number       smallint NOT NULL CHECK (week_number >= 1),
  day_index         smallint NOT NULL CHECK (day_index BETWEEN 1 AND 7),
  name              text NOT NULL,
  note              text,
  is_deload         boolean NOT NULL DEFAULT false,
  estimated_minutes smallint,
  UNIQUE (program_id, week_number, day_index)
);

CREATE TABLE program_exercise (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  program_workout_id  uuid NOT NULL REFERENCES program_workout(id) ON DELETE CASCADE,
  exercise_id         uuid NOT NULL REFERENCES exercise(id),
  order_index         smallint NOT NULL,
  superset_group      smallint,
  target_sets         smallint NOT NULL DEFAULT 3,
  target_reps_min     smallint,
  target_reps_max     smallint,
  target_rpe          numeric(3,1) CHECK (target_rpe BETWEEN 1 AND 10),
  target_pct_1rm      numeric(4,1),
  rest_seconds        smallint NOT NULL DEFAULT 90,
  tempo               text,
  note                text,
  progression_rule_id uuid REFERENCES progression_rule(id),
  UNIQUE (program_workout_id, order_index)
);

-- ============================================================
-- ASIGNACION: instancia del programa para UN cliente.
-- Se copia al asignar. El coach edita la copia sin romper la plantilla
-- ni el historial de los demas clientes.
-- ============================================================

CREATE TYPE assignment_status AS ENUM ('active','paused','completed','cancelled');
CREATE TYPE assigned_workout_status AS ENUM ('pending','completed','skipped');

CREATE TABLE assignment (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id          uuid NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
  program_id      uuid REFERENCES program(id) ON DELETE SET NULL,
  client_user_id  uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  coach_user_id   uuid REFERENCES app_user(id),
  name            text NOT NULL,
  start_date      date NOT NULL,
  end_date        date,
  status          assignment_status NOT NULL DEFAULT 'active',
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX assignment_one_active_uq
  ON assignment (client_user_id) WHERE status = 'active';
CREATE INDEX ON assignment (coach_user_id, status);
CREATE TRIGGER assignment_updated_at BEFORE UPDATE ON assignment
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE assigned_workout (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  assignment_id     uuid NOT NULL REFERENCES assignment(id) ON DELETE CASCADE,
  source_workout_id uuid REFERENCES program_workout(id) ON DELETE SET NULL,
  week_number       smallint NOT NULL,
  day_index         smallint NOT NULL,
  name              text NOT NULL,
  note              text,
  is_deload         boolean NOT NULL DEFAULT false,
  scheduled_on      date,
  status            assigned_workout_status NOT NULL DEFAULT 'pending',
  UNIQUE (assignment_id, week_number, day_index)
);
CREATE INDEX ON assigned_workout (assignment_id, scheduled_on);

CREATE TABLE assigned_exercise (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  assigned_workout_id uuid NOT NULL REFERENCES assigned_workout(id) ON DELETE CASCADE,
  exercise_id         uuid NOT NULL REFERENCES exercise(id),
  order_index         smallint NOT NULL,
  superset_group      smallint,
  target_sets         smallint NOT NULL DEFAULT 3,
  target_reps_min     smallint,
  target_reps_max     smallint,
  target_rpe          numeric(3,1),
  target_weight_kg    numeric(6,2),
  rest_seconds        smallint NOT NULL DEFAULT 90,
  tempo               text,
  note                text,
  progression_rule_id uuid REFERENCES progression_rule(id),
  UNIQUE (assigned_workout_id, order_index)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS assigned_exercise, assigned_workout, assignment,
  program_exercise, program_workout, program, progression_rule,
  exercise_alternative, exercise CASCADE;
DROP TYPE IF EXISTS assigned_workout_status, assignment_status, program_status,
  progression_type, tracking_mode, exercise_mechanic, movement_pattern;
-- +goose StatementEnd
