-- +goose Up
-- +goose StatementBegin

-- Decision de producto: un usuario tiene un unico rol activo por
-- organizacion (para escalar permisos se sube el rol existente en vez de
-- agregar uno nuevo). Sin este indice, GetOrgMembership/
-- GetActiveMembershipByUser (LIMIT 1 sin ORDER BY) resuelven el rol
-- efectivo de forma no deterministica si llegaran a coexistir dos filas
-- activas para el mismo usuario+org.
CREATE UNIQUE INDEX membership_one_active_role_per_org_uq
  ON membership (org_id, user_id) WHERE status = 'active';

-- exercise/assignment solo se desactivan o cancelan (soft-delete); un
-- ON DELETE CASCADE aca seria perdida silenciosa de historial de PRs o del
-- rastro de auditoria de sugerencias de IA si alguna vez se agrega un
-- hard-delete real.
ALTER TABLE personal_record DROP CONSTRAINT personal_record_exercise_id_fkey;
ALTER TABLE personal_record ADD CONSTRAINT personal_record_exercise_id_fkey
  FOREIGN KEY (exercise_id) REFERENCES exercise(id) ON DELETE RESTRICT;

ALTER TABLE ai_suggestion DROP CONSTRAINT ai_suggestion_assignment_id_fkey;
ALTER TABLE ai_suggestion ADD CONSTRAINT ai_suggestion_assignment_id_fkey
  FOREIGN KEY (assignment_id) REFERENCES assignment(id) ON DELETE RESTRICT;

-- Los CHECK de rango de la plantilla (program_workout/program_exercise) no
-- se habian copiado a su instancia asignada (assigned_workout/
-- assigned_exercise), que se llena por codigo de aplicacion en
-- AssignmentService.Assign, no por input directo del usuario, pero conviene
-- que la base los garantice igual en las dos copias.
ALTER TABLE assigned_workout ADD CONSTRAINT assigned_workout_week_number_check
  CHECK (week_number >= 1);
ALTER TABLE assigned_workout ADD CONSTRAINT assigned_workout_day_index_check
  CHECK (day_index BETWEEN 1 AND 7);
ALTER TABLE assigned_exercise ADD CONSTRAINT assigned_exercise_target_rpe_check
  CHECK (target_rpe IS NULL OR target_rpe BETWEEN 1 AND 10);
ALTER TABLE program_exercise ADD CONSTRAINT program_exercise_reps_range_check
  CHECK (target_reps_min IS NULL OR target_reps_max IS NULL OR target_reps_min <= target_reps_max);
ALTER TABLE assigned_exercise ADD CONSTRAINT assigned_exercise_reps_range_check
  CHECK (target_reps_min IS NULL OR target_reps_max IS NULL OR target_reps_min <= target_reps_max);

-- Indices faltantes para patrones de acceso reales: listados admin por org,
-- analisis de impacto al desactivar un ejercicio, lookup inverso de logros
-- y auditoria por actor.
CREATE INDEX assignment_org_status_idx ON assignment (org_id, status);
CREATE INDEX program_exercise_exercise_id_idx ON program_exercise (exercise_id);
CREATE INDEX assigned_exercise_exercise_id_idx ON assigned_exercise (exercise_id);
CREATE INDEX workout_session_assigned_workout_id_idx ON workout_session (assigned_workout_id) WHERE assigned_workout_id IS NOT NULL;
CREATE INDEX user_achievement_achievement_id_idx ON user_achievement (achievement_id);
CREATE INDEX audit_log_actor_user_id_idx ON audit_log (actor_user_id);
CREATE INDEX ai_suggestion_assignment_id_idx ON ai_suggestion (assignment_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS ai_suggestion_assignment_id_idx;
DROP INDEX IF EXISTS audit_log_actor_user_id_idx;
DROP INDEX IF EXISTS user_achievement_achievement_id_idx;
DROP INDEX IF EXISTS workout_session_assigned_workout_id_idx;
DROP INDEX IF EXISTS assigned_exercise_exercise_id_idx;
DROP INDEX IF EXISTS program_exercise_exercise_id_idx;
DROP INDEX IF EXISTS assignment_org_status_idx;

ALTER TABLE assigned_exercise DROP CONSTRAINT assigned_exercise_reps_range_check;
ALTER TABLE program_exercise DROP CONSTRAINT program_exercise_reps_range_check;
ALTER TABLE assigned_exercise DROP CONSTRAINT assigned_exercise_target_rpe_check;
ALTER TABLE assigned_workout DROP CONSTRAINT assigned_workout_day_index_check;
ALTER TABLE assigned_workout DROP CONSTRAINT assigned_workout_week_number_check;

ALTER TABLE ai_suggestion DROP CONSTRAINT ai_suggestion_assignment_id_fkey;
ALTER TABLE ai_suggestion ADD CONSTRAINT ai_suggestion_assignment_id_fkey
  FOREIGN KEY (assignment_id) REFERENCES assignment(id) ON DELETE CASCADE;

ALTER TABLE personal_record DROP CONSTRAINT personal_record_exercise_id_fkey;
ALTER TABLE personal_record ADD CONSTRAINT personal_record_exercise_id_fkey
  FOREIGN KEY (exercise_id) REFERENCES exercise(id) ON DELETE CASCADE;

DROP INDEX IF EXISTS membership_one_active_role_per_org_uq;

-- +goose StatementEnd
