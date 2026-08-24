-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

-- Trigger compartido para updated_at
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- TENENCIA
-- ============================================================

CREATE TYPE org_status AS ENUM ('active', 'suspended', 'cancelled');

CREATE TABLE organization (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name            text NOT NULL,
  slug            citext NOT NULL UNIQUE,
  join_code       text NOT NULL UNIQUE,
  logo_url        text,
  brand_color     text NOT NULL DEFAULT '#C6FF4F',
  timezone        text NOT NULL DEFAULT 'UTC',
  status          org_status NOT NULL DEFAULT 'active',
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE TRIGGER organization_updated_at BEFORE UPDATE ON organization
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ============================================================
-- IDENTIDAD (registro abierto: el usuario existe sin organizacion)
-- ============================================================

CREATE TABLE app_user (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email           citext NOT NULL UNIQUE,
  password_hash   text,
  full_name       text NOT NULL,
  avatar_url      text,
  locale          text NOT NULL DEFAULT 'es',
  timezone        text NOT NULL DEFAULT 'UTC',
  email_verified_at timestamptz,
  is_platform_admin boolean NOT NULL DEFAULT false,
  last_login_at   timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  deleted_at      timestamptz
);
CREATE TRIGGER app_user_updated_at BEFORE UPDATE ON app_user
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TYPE auth_provider AS ENUM ('password', 'google', 'apple');

CREATE TABLE user_identity (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  provider        auth_provider NOT NULL,
  provider_uid    text NOT NULL,
  created_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (provider, provider_uid)
);

CREATE TABLE refresh_token (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  token_hash      text NOT NULL UNIQUE,
  device_label    text,
  expires_at      timestamptz NOT NULL,
  revoked_at      timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON refresh_token (user_id) WHERE revoked_at IS NULL;

-- Membresia N:M -- un usuario puede ser coach en un gym y cliente en otro
CREATE TYPE member_role AS ENUM ('owner', 'admin', 'coach', 'client');
CREATE TYPE member_status AS ENUM ('pending', 'active', 'revoked');

CREATE TABLE membership (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id          uuid NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
  user_id         uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  role            member_role NOT NULL,
  status          member_status NOT NULL DEFAULT 'pending',
  invited_by      uuid REFERENCES app_user(id),
  joined_at       timestamptz,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (org_id, user_id, role)
);
CREATE INDEX ON membership (user_id) WHERE status = 'active';
CREATE INDEX ON membership (org_id, role) WHERE status = 'active';
CREATE TRIGGER membership_updated_at BEFORE UPDATE ON membership
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Asignacion coach -> cliente dentro de una organizacion
CREATE TABLE coach_client (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id          uuid NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
  coach_user_id   uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  client_user_id  uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  started_at      timestamptz NOT NULL DEFAULT now(),
  ended_at        timestamptz,
  CHECK (coach_user_id <> client_user_id)
);
CREATE UNIQUE INDEX coach_client_active_uq
  ON coach_client (org_id, client_user_id) WHERE ended_at IS NULL;
CREATE INDEX ON coach_client (coach_user_id) WHERE ended_at IS NULL;

-- ============================================================
-- PERFIL Y METRICAS DEL CLIENTE
-- ============================================================

CREATE TYPE biological_sex AS ENUM ('male', 'female', 'unspecified');
CREATE TYPE experience_level AS ENUM ('beginner', 'intermediate', 'advanced');
CREATE TYPE training_goal AS ENUM ('strength', 'hypertrophy', 'fat_loss', 'endurance', 'general_health');

CREATE TABLE client_profile (
  user_id             uuid PRIMARY KEY REFERENCES app_user(id) ON DELETE CASCADE,
  birth_date          date,
  sex                 biological_sex NOT NULL DEFAULT 'unspecified',
  height_cm           numeric(5,1),
  experience          experience_level NOT NULL DEFAULT 'beginner',
  primary_goal        training_goal NOT NULL DEFAULT 'general_health',
  days_per_week       smallint NOT NULL DEFAULT 3 CHECK (days_per_week BETWEEN 1 AND 7),
  session_minutes     smallint NOT NULL DEFAULT 60,
  available_equipment text[] NOT NULL DEFAULT '{}',
  limitations         text,
  unit_system         text NOT NULL DEFAULT 'metric' CHECK (unit_system IN ('metric','imperial')),
  onboarded_at        timestamptz,
  updated_at          timestamptz NOT NULL DEFAULT now()
);
CREATE TRIGGER client_profile_updated_at BEFORE UPDATE ON client_profile
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE body_metric (
  id              bigserial PRIMARY KEY,
  user_id         uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  measured_on     date NOT NULL,
  weight_kg       numeric(5,2),
  body_fat_pct    numeric(4,1),
  measurements    jsonb NOT NULL DEFAULT '{}',
  note            text,
  created_at      timestamptz NOT NULL DEFAULT now(),
  UNIQUE (user_id, measured_on)
);
CREATE INDEX ON body_metric (user_id, measured_on DESC);

-- Fotos de progreso: dato sensible, storage privado y URL firmada bajo demanda
CREATE TABLE progress_photo (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         uuid NOT NULL REFERENCES app_user(id) ON DELETE CASCADE,
  storage_key     text NOT NULL,
  pose            text,
  taken_on        date NOT NULL,
  shared_with_coach boolean NOT NULL DEFAULT false,
  created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON progress_photo (user_id, taken_on DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS progress_photo, body_metric, client_profile, coach_client, membership,
  refresh_token, user_identity, app_user, organization CASCADE;
DROP TYPE IF EXISTS training_goal, experience_level, biological_sex, member_status, member_role,
  auth_provider, org_status;
DROP FUNCTION IF EXISTS set_updated_at();
-- +goose StatementEnd
