-- Catalogo global inicial (org_id NULL). Cubre los 11 patrones de movimiento
-- con ejercicios estandar de gimnasio. Sin video/thumbnail: la produccion de
-- media es un proyecto aparte (banco propio o licenciado), fuera de este seed.
-- ponytail: ampliar a 200-300 con media cuando haya contenido real que subir.

INSERT INTO exercise (slug, name, pattern, mechanic, primary_muscle, secondary_muscles, equipment, difficulty, tracking, is_unilateral, instructions) VALUES
-- squat
('back-squat', 'Back Squat', 'squat', 'compound', 'quadriceps', ARRAY['glutes','hamstrings','core'], ARRAY['barbell','rack'], 'intermediate', 'weight_reps', false, ARRAY['Barra apoyada en trapecio, pies al ancho de hombros','Baja controlando cadera y rodilla a la vez','Sube empujando el piso']),
('front-squat', 'Front Squat', 'squat', 'compound', 'quadriceps', ARRAY['core','glutes'], ARRAY['barbell','rack'], 'advanced', 'weight_reps', false, ARRAY['Barra en rack frontal, codos altos','Torso lo mas vertical posible durante el descenso']),
('goblet-squat', 'Goblet Squat', 'squat', 'compound', 'quadriceps', ARRAY['glutes','core'], ARRAY['dumbbell'], 'beginner', 'weight_reps', false, ARRAY['Sostener la mancuerna contra el pecho','Sentadilla profunda manteniendo el torso vertical']),
('leg-press', 'Leg Press', 'squat', 'compound', 'quadriceps', ARRAY['glutes','hamstrings'], ARRAY['machine'], 'beginner', 'weight_reps', false, ARRAY['Pies al ancho de hombros en la plataforma','No bloquear rodillas al extender']),

-- hinge
('conventional-deadlift', 'Conventional Deadlift', 'hinge', 'compound', 'hamstrings', ARRAY['glutes','back','core'], ARRAY['barbell'], 'advanced', 'weight_reps', false, ARRAY['Barra pegada a las espinillas','Espalda neutra, empuja el piso con las piernas']),
('romanian-deadlift', 'Romanian Deadlift', 'hinge', 'compound', 'hamstrings', ARRAY['glutes','back'], ARRAY['barbell'], 'intermediate', 'weight_reps', false, ARRAY['Rodillas con flexion minima','Cadera hacia atras manteniendo la barra pegada a las piernas']),
('hip-thrust', 'Hip Thrust', 'hinge', 'compound', 'glutes', ARRAY['hamstrings'], ARRAY['barbell','bench'], 'intermediate', 'weight_reps', false, ARRAY['Espalda alta apoyada en el banco','Extiende cadera hasta linea recta rodilla-cadera-hombro']),
('kettlebell-swing', 'Kettlebell Swing', 'hinge', 'compound', 'glutes', ARRAY['hamstrings','core'], ARRAY['kettlebell'], 'beginner', 'weight_reps', false, ARRAY['Impulso desde la cadera, no desde los brazos','Kettlebell sube por inercia hasta la altura del pecho']),

-- horizontal_push
('barbell-bench-press', 'Barbell Bench Press', 'horizontal_push', 'compound', 'chest', ARRAY['triceps','shoulders'], ARRAY['barbell','bench'], 'intermediate', 'weight_reps', false, ARRAY['Escapulas retraidas contra el banco','Baja la barra al pecho con control']),
('incline-dumbbell-press', 'Incline Dumbbell Press', 'horizontal_push', 'compound', 'chest', ARRAY['shoulders','triceps'], ARRAY['dumbbell','bench'], 'intermediate', 'weight_reps', false, ARRAY['Banco a 30-45 grados','Baja las mancuernas a la altura del pecho superior']),
('push-up', 'Push-Up', 'horizontal_push', 'compound', 'chest', ARRAY['triceps','core'], ARRAY['bodyweight'], 'beginner', 'bodyweight_reps', false, ARRAY['Cuerpo en linea recta de cabeza a talones','Baja hasta que el pecho casi toque el piso']),
('cable-chest-fly', 'Cable Chest Fly', 'horizontal_push', 'isolation', 'chest', ARRAY['shoulders'], ARRAY['cable'], 'beginner', 'weight_reps', false, ARRAY['Ligera flexion de codo constante','Junta las manos al frente del pecho']),

-- vertical_push
('overhead-press', 'Overhead Press', 'vertical_push', 'compound', 'shoulders', ARRAY['triceps','core'], ARRAY['barbell'], 'intermediate', 'weight_reps', false, ARRAY['Core apretado, sin arquear la espalda baja','Empuja la barra en linea recta sobre la cabeza']),
('dumbbell-shoulder-press', 'Dumbbell Shoulder Press', 'vertical_push', 'compound', 'shoulders', ARRAY['triceps'], ARRAY['dumbbell'], 'beginner', 'weight_reps', false, ARRAY['Mancuernas a la altura de los hombros','Empuja hacia arriba sin bloquear violentamente el codo']),
('pike-push-up', 'Pike Push-Up', 'vertical_push', 'compound', 'shoulders', ARRAY['triceps'], ARRAY['bodyweight'], 'intermediate', 'bodyweight_reps', false, ARRAY['Cadera elevada formando una V invertida','Baja la cabeza hacia el piso entre las manos']),

-- horizontal_pull
('barbell-row', 'Barbell Row', 'horizontal_pull', 'compound', 'back', ARRAY['biceps','shoulders'], ARRAY['barbell'], 'intermediate', 'weight_reps', false, ARRAY['Torso inclinado, espalda neutra','Lleva la barra hacia el abdomen apretando escapulas']),
('seated-cable-row', 'Seated Cable Row', 'horizontal_pull', 'compound', 'back', ARRAY['biceps'], ARRAY['cable'], 'beginner', 'weight_reps', false, ARRAY['Espalda recta durante todo el movimiento','Tira hasta el abdomen sin balancear el torso']),
('inverted-row', 'Inverted Row', 'horizontal_pull', 'compound', 'back', ARRAY['biceps','core'], ARRAY['bodyweight','bar'], 'beginner', 'bodyweight_reps', false, ARRAY['Cuerpo recto colgando bajo la barra','Tira el pecho hacia la barra']),

-- vertical_pull
('pull-up', 'Pull-Up', 'vertical_pull', 'compound', 'back', ARRAY['biceps'], ARRAY['bodyweight','bar'], 'advanced', 'bodyweight_reps', false, ARRAY['Agarre prono al ancho de hombros','Sube hasta que el menton pase la barra']),
('lat-pulldown', 'Lat Pulldown', 'vertical_pull', 'compound', 'back', ARRAY['biceps'], ARRAY['cable','machine'], 'beginner', 'weight_reps', false, ARRAY['Tira la barra hacia la parte alta del pecho','Evita usar impulso del torso']),
('chin-up', 'Chin-Up', 'vertical_pull', 'compound', 'back', ARRAY['biceps'], ARRAY['bodyweight','bar'], 'advanced', 'bodyweight_reps', false, ARRAY['Agarre supino al ancho de hombros','Sube controlando todo el recorrido']),

-- lunge
('walking-lunge', 'Walking Lunge', 'lunge', 'compound', 'quadriceps', ARRAY['glutes','hamstrings'], ARRAY['dumbbell'], 'beginner', 'weight_reps', true, ARRAY['Paso largo, rodilla trasera casi toca el piso','Alterna pierna en cada paso']),
('bulgarian-split-squat', 'Bulgarian Split Squat', 'lunge', 'compound', 'quadriceps', ARRAY['glutes'], ARRAY['dumbbell','bench'], 'intermediate', 'weight_reps', true, ARRAY['Pie trasero elevado en el banco','Desciende hasta que el muslo delantero quede paralelo']),
('reverse-lunge', 'Reverse Lunge', 'lunge', 'compound', 'quadriceps', ARRAY['glutes','hamstrings'], ARRAY['bodyweight'], 'beginner', 'bodyweight_reps', true, ARRAY['Paso hacia atras controlado','Rodilla delantera se mantiene sobre el tobillo']),

-- carry
('farmers-carry', 'Farmers Carry', 'carry', 'compound', 'forearms', ARRAY['core','shoulders'], ARRAY['dumbbell'], 'beginner', 'duration', false, ARRAY['Pecho arriba, hombros hacia atras','Camina con pasos cortos y controlados']),
('suitcase-carry', 'Suitcase Carry', 'carry', 'compound', 'core', ARRAY['forearms'], ARRAY['dumbbell'], 'intermediate', 'duration', true, ARRAY['Peso en un solo lado','Evita inclinar el torso hacia el peso']),

-- rotation
('cable-woodchop', 'Cable Woodchop', 'rotation', 'isolation', 'core', ARRAY['shoulders'], ARRAY['cable'], 'intermediate', 'weight_reps', true, ARRAY['Rotacion desde la cadera, no solo los brazos','Movimiento diagonal controlado de alto a bajo']),
('russian-twist', 'Russian Twist', 'rotation', 'isolation', 'core', ARRAY[]::text[], ARRAY['bodyweight'], 'beginner', 'bodyweight_reps', false, ARRAY['Pies elevados, torso inclinado hacia atras','Rota tocando el piso a cada lado']),
('pallof-press', 'Pallof Press', 'rotation', 'isolation', 'core', ARRAY['shoulders'], ARRAY['cable'], 'beginner', 'weight_reps', true, ARRAY['Resiste la rotacion que impone el cable','Empuja al frente y regresa sin dejar que el torso gire']),

-- isolation
('bicep-curl', 'Bicep Curl', 'isolation', 'isolation', 'biceps', ARRAY['forearms'], ARRAY['dumbbell'], 'beginner', 'weight_reps', false, ARRAY['Codos pegados al torso','Sube controlando, evita balanceo']),
('triceps-pushdown', 'Triceps Pushdown', 'isolation', 'isolation', 'triceps', ARRAY[]::text[], ARRAY['cable'], 'beginner', 'weight_reps', false, ARRAY['Codos fijos junto al torso','Extiende completamente sin bloquear con fuerza']),
('lateral-raise', 'Lateral Raise', 'isolation', 'isolation', 'shoulders', ARRAY[]::text[], ARRAY['dumbbell'], 'beginner', 'weight_reps', false, ARRAY['Ligera flexion de codo','Eleva hasta la altura del hombro, sin usar impulso']),
('leg-curl', 'Leg Curl', 'isolation', 'isolation', 'hamstrings', ARRAY[]::text[], ARRAY['machine'], 'beginner', 'weight_reps', false, ARRAY['Cadera fija contra el respaldo','Flexiona la rodilla llevando el talon a los gluteos']),
('leg-extension', 'Leg Extension', 'isolation', 'isolation', 'quadriceps', ARRAY[]::text[], ARRAY['machine'], 'beginner', 'weight_reps', false, ARRAY['Espalda contra el respaldo','Extiende sin bloquear violentamente la rodilla']),
('standing-calf-raise', 'Standing Calf Raise', 'isolation', 'isolation', 'calves', ARRAY[]::text[], ARRAY['machine'], 'beginner', 'weight_reps', false, ARRAY['Rango completo, baja hasta sentir estiramiento','Sube hasta la punta de los pies']),
('face-pull', 'Face Pull', 'isolation', 'isolation', 'shoulders', ARRAY['back'], ARRAY['cable'], 'beginner', 'weight_reps', false, ARRAY['Tira hacia la cara separando las manos','Codos altos durante todo el recorrido']),

-- cardio
('running', 'Running', 'cardio', 'compound', 'legs', ARRAY['core'], ARRAY['bodyweight'], 'beginner', 'distance_duration', false, ARRAY['Cadencia constante','Ajusta ritmo segun el objetivo de la sesion']),
('rowing-machine', 'Rowing Machine', 'cardio', 'compound', 'back', ARRAY['legs','core'], ARRAY['machine'], 'beginner', 'distance_duration', false, ARRAY['Empuje con piernas antes de tirar con los brazos','Secuencia: piernas, espalda, brazos - y a la inversa al volver']),
('assault-bike', 'Assault Bike', 'cardio', 'compound', 'legs', ARRAY['shoulders','core'], ARRAY['machine'], 'beginner', 'distance_duration', false, ARRAY['Brazos y piernas trabajan a la vez','Regula intensidad segun el intervalo objetivo']),
('jump-rope', 'Jump Rope', 'cardio', 'compound', 'calves', ARRAY['core'], ARRAY['jump-rope'], 'beginner', 'duration', false, ARRAY['Saltos bajos y rapidos','Muneca gira la cuerda, no el brazo completo'])

ON CONFLICT (slug) WHERE org_id IS NULL DO NOTHING;
