-- Logros del sistema. El code es la clave que interpreta
-- internal/service/gamification.go (achievementCriteria); el resto es
-- contenido puro (nombre, descripcion, icono, recompensa).

INSERT INTO achievement (code, name, description, icon, tier, xp_reward, criteria) VALUES
('first-workout', 'Primer entrenamiento', 'Completaste tu primera sesion', 'flag', 1, 50, '{"total_sessions":1}'),
('ten-workouts', 'Diez entrenamientos', 'Completaste 10 sesiones', 'trophy', 2, 100, '{"total_sessions":10}'),
('fifty-workouts', 'Cincuenta entrenamientos', 'Completaste 50 sesiones', 'trophy', 3, 300, '{"total_sessions":50}'),
('streak-7', 'Una semana seguida', 'Racha de 7 dias entrenando', 'fire', 2, 150, '{"streak_days":7}'),
('streak-30', 'Un mes seguido', 'Racha de 30 dias entrenando', 'fire', 3, 500, '{"streak_days":30}'),
('first-pr', 'Primer record', 'Superaste tu primer record personal', 'medal', 1, 100, '{"has_pr":true}')
ON CONFLICT (code) DO NOTHING;
