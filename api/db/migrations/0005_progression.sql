-- +goose Up
-- +goose StatementBegin

-- override_source marca que el target de este ejercicio asignado lo escribio
-- algo que no es el motor de progresion (hoy: una sugerencia de IA aprobada
-- por el coach). El motor lo respeta una sola vez: cuando le tocaria
-- sobreescribir ese target, no lo hace y limpia el flag, asi la proxima vez
-- vuelve al calculo automatico. Un flag que se autoconsume alcanza — no hace
-- falta una tabla de prioridades ni auditoria aparte.
ALTER TABLE assigned_exercise
  ADD COLUMN override_source text
  CHECK (override_source IS NULL OR override_source IN ('ai_suggestion'));

-- La progresion busca "la proxima ocurrencia de este ejercicio en dias sin
-- completar", que entra por assigned_workout + exercise.
CREATE INDEX assigned_exercise_workout_exercise_idx
  ON assigned_exercise (assigned_workout_id, exercise_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS assigned_exercise_workout_exercise_idx;
ALTER TABLE assigned_exercise DROP COLUMN override_source;

-- +goose StatementEnd
