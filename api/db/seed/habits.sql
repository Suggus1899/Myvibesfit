-- Habitos globales (org_id NULL), listos para que cualquier cliente se
-- suscriba sin que el gimnasio tenga que crear los suyos.

INSERT INTO habit (slug, name, icon, unit, default_target, is_system) VALUES
('drink-water', 'Tomar agua', 'droplet', 'milliliters', 2000, true),
('sleep-8h', 'Dormir 8 horas', 'moon', 'minutes', 480, true),
('walk-steps', 'Caminar', 'footprints', 'steps', 8000, true),
('stretch', 'Estirar', 'check', 'boolean', 1, true),
('protein-target', 'Cumplir proteina', 'check', 'boolean', 1, true)
ON CONFLICT (slug) WHERE org_id IS NULL DO NOTHING;
