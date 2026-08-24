-- Reglas de progresion globales (org_id NULL), listas para usar sin que
-- el coach tenga que crear las suyas desde cero.

INSERT INTO progression_rule (name, type, params, is_system) VALUES
('Doble progresion estandar', 'double_progression', '{"increment_kg":2.5,"round_to_kg":2.5}', true),
('Progresion lineal', 'linear_load', '{"increment_kg":2.5,"round_to_kg":2.5}', true),
('Autorregulado por RPE', 'rpe_autoregulated', '{"adjustment_pct_per_rpe":0.025,"round_to_kg":2.5}', true),
('Porcentaje de 1RM (75%)', 'percentage_1rm', '{"percent":0.75,"round_to_kg":2.5}', true);
