import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/network/models.dart';
import '../../core/providers.dart';
import '../../core/theme/app_colors.dart';
import '../home/home_screen.dart';
import 'workout_providers.dart';

class ActiveWorkoutScreen extends ConsumerWidget {
  const ActiveWorkoutScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final session = ref.watch(activeWorkoutProvider);
    if (session == null) {
      return const _StartWorkoutView();
    }
    return _InProgressView(session: session);
  }
}

class _StartWorkoutView extends ConsumerWidget {
  const _StartWorkoutView();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final assignmentAsync = ref.watch(currentAssignmentProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Entrenar')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            assignmentAsync.when(
              data: (a) {
                if (a == null || a.workouts.isEmpty) return const SizedBox.shrink();
                final next = a.workouts.firstWhere((w) => w.status == 'pending', orElse: () => a.workouts.first);
                return ElevatedButton(
                  onPressed: () => ref.read(activeWorkoutProvider.notifier).start(fromAssigned: next),
                  child: Text('Empezar "${next.name}" del plan'),
                );
              },
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => const SizedBox.shrink(),
            ),
            const SizedBox(height: 12),
            OutlinedButton(
              onPressed: () => ref.read(activeWorkoutProvider.notifier).start(),
              child: const Text('Entrenamiento libre'),
            ),
          ],
        ),
      ),
    );
  }
}

class _InProgressView extends ConsumerWidget {
  const _InProgressView({required this.session});
  final WorkoutSessionState session;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colors = context.colors;
    final rest = ref.watch(restTimerProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(session.name),
        actions: [
          IconButton(
            icon: const Icon(Icons.close),
            tooltip: 'Descartar entrenamiento',
            onPressed: () => ref.read(activeWorkoutProvider.notifier).discard(),
          ),
        ],
      ),
      body: Column(
        children: [
          if (rest.running)
            Container(
              width: double.infinity,
              color: colors.info.withValues(alpha: 0.15),
              padding: const EdgeInsets.all(12),
              child: Text(
                'Descanso: ${rest.secondsLeft}s',
                textAlign: TextAlign.center,
                style: TextStyle(color: colors.info, fontWeight: FontWeight.w700),
              ),
            ),
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: session.exercises.length + 1,
              itemBuilder: (context, i) {
                if (i == session.exercises.length) {
                  return const _AddExerciseButton();
                }
                return _ExerciseCard(exerciseIndex: i, exercise: session.exercises[i]);
              },
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: ElevatedButton(
              onPressed: () async {
                final result = await ref.read(activeWorkoutProvider.notifier).finish();
                if (!context.mounted) return;
                final rewardParts = <String>[];
                if (result != null && result.newPersonalRecords > 0) {
                  rewardParts.add(
                    '${result.newPersonalRecords} PR${result.newPersonalRecords > 1 ? "s" : ""} nuevo${result.newPersonalRecords > 1 ? "s" : ""} 🎉',
                  );
                }
                if (result != null && result.unlockedAchievements.isNotEmpty) {
                  rewardParts.add('🏆 ${result.unlockedAchievements.map((a) => a.name).join(", ")}');
                }
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(
                      rewardParts.isEmpty
                          ? 'Entrenamiento guardado'
                          : 'Entrenamiento guardado — ${rewardParts.join(" · ")}',
                    ),
                  ),
                );
              },
              child: const Text('Terminar entrenamiento'),
            ),
          ),
        ],
      ),
    );
  }
}

class _ExerciseCard extends ConsumerWidget {
  const _ExerciseCard({required this.exerciseIndex, required this.exercise});
  final int exerciseIndex;
  final ExerciseEntry exercise;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final catalog = ref.watch(exerciseCatalogProvider).valueOrNull ?? const <ExerciseSummary>[];
    final displayName = resolveExerciseName(catalog, exercise.exerciseId, fallback: exercise.name);
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            InkWell(
              onTap: () => context.push('/exercises/${exercise.exerciseId}'),
              child: Row(
                children: [
                  Expanded(child: Text(displayName, style: Theme.of(context).textTheme.titleLarge)),
                  const Icon(Icons.info_outline, size: 18),
                ],
              ),
            ),
            const SizedBox(height: 8),
            ...exercise.sets.asMap().entries.map(
              (entry) => _SetRow(exerciseIndex: exerciseIndex, setIndex: entry.key, set: entry.value),
            ),
            TextButton.icon(
              onPressed: () => ref.read(activeWorkoutProvider.notifier).addSet(exerciseIndex),
              icon: const Icon(Icons.add),
              label: const Text('Agregar serie'),
            ),
          ],
        ),
      ),
    );
  }
}

class _SetRow extends ConsumerStatefulWidget {
  const _SetRow({required this.exerciseIndex, required this.setIndex, required this.set});
  final int exerciseIndex;
  final int setIndex;
  final SetEntry set;

  @override
  ConsumerState<_SetRow> createState() => _SetRowState();
}

class _SetRowState extends ConsumerState<_SetRow> {
  late double _weight = widget.set.weightKg ?? 0;
  late int _reps = widget.set.reps ?? 0;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final completed = widget.set.isCompleted;

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          SizedBox(width: 28, child: Text('${widget.set.setNumber}', style: Theme.of(context).textTheme.bodyMedium)),
          Expanded(
            child: _Stepper(label: 'kg', value: _weight, step: 2.5, onChanged: (v) => setState(() => _weight = v)),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: _Stepper(
              label: 'reps',
              value: _reps.toDouble(),
              step: 1,
              onChanged: (v) => setState(() => _reps = v.toInt()),
            ),
          ),
          const SizedBox(width: 12),
          IconButton(
            // Peso 0 es valido (ejercicios con peso corporal); 0 reps no lo es.
            icon: Icon(
              completed ? Icons.check_circle : Icons.check_circle_outline,
              color: completed ? colors.success : colors.textMuted,
            ),
            tooltip: completed ? 'Serie completada' : 'Completar serie',
            onPressed: completed || _reps <= 0
                ? null
                : () => ref
                      .read(activeWorkoutProvider.notifier)
                      .completeSet(widget.exerciseIndex, widget.setIndex, weightKg: _weight, reps: _reps),
          ),
        ],
      ),
    );
  }
}

class _Stepper extends StatelessWidget {
  const _Stepper({required this.label, required this.value, required this.step, required this.onChanged});
  final String label;
  final double value;
  final double step;
  final ValueChanged<double> onChanged;

  String get _formatted => value == value.roundToDouble() ? value.toInt().toString() : value.toStringAsFixed(1);

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        // Sin tooltip estos iconos no anuncian nada al lector de pantalla, y
        // visualDensity.compact dejaba el area tactil por debajo de los 44x44
        // que exige docs/DESIGN.md §8.
        IconButton(
          icon: const Icon(Icons.remove, size: 18),
          tooltip: 'Reducir $label',
          onPressed: () => onChanged((value - step).clamp(0, 999)),
          constraints: const BoxConstraints(minWidth: 44, minHeight: 44),
        ),
        Semantics(
          label: label,
          value: _formatted,
          child: ExcludeSemantics(
            child: Column(
              children: [
                Text(_formatted, style: Theme.of(context).textTheme.titleLarge),
                Text(label, style: Theme.of(context).textTheme.bodySmall),
              ],
            ),
          ),
        ),
        IconButton(
          icon: const Icon(Icons.add, size: 18),
          tooltip: 'Aumentar $label',
          onPressed: () => onChanged(value + step),
          constraints: const BoxConstraints(minWidth: 44, minHeight: 44),
        ),
      ],
    );
  }
}

class _AddExerciseButton extends ConsumerWidget {
  const _AddExerciseButton();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return OutlinedButton.icon(
      icon: const Icon(Icons.add),
      label: const Text('Agregar ejercicio'),
      onPressed: () async {
        final api = ref.read(apiRepositoryProvider);
        final exercises = await api.exercises();
        if (!context.mounted) return;
        final selected = await showModalBottomSheet<ExerciseSummary>(
          context: context,
          builder: (context) => ListView(
            children: exercises
                .map(
                  (e) => ListTile(
                    title: Text(e.name),
                    subtitle: Text(e.primaryMuscle),
                    onTap: () => Navigator.pop(context, e),
                  ),
                )
                .toList(),
          ),
        );
        if (selected != null) {
          ref.read(activeWorkoutProvider.notifier).addExercise(selected.id, selected.name);
        }
      },
    );
  }
}
