import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/network/models.dart';
import '../../core/theme/app_colors.dart';
import '../home/home_screen.dart';
import 'workout_providers.dart';

class WeeklyPlanScreen extends ConsumerWidget {
  const WeeklyPlanScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final assignmentAsync = ref.watch(currentAssignmentProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Mi plan')),
      body: assignmentAsync.when(
        data: (assignment) {
          if (assignment == null || assignment.workouts.isEmpty) {
            return const Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Text('Todavía no tenés un plan asignado.', textAlign: TextAlign.center),
              ),
            );
          }

          // El backend devuelve el mesociclo completo; agrupamos por semana
          // para que se vea el bloque entero, no solo el proximo dia.
          final byWeek = <int, List<AssignedWorkoutInfo>>{};
          for (final w in assignment.workouts) {
            byWeek.putIfAbsent(w.weekNumber, () => []).add(w);
          }
          final weeks = byWeek.keys.toList()..sort();

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(assignment.name, style: Theme.of(context).textTheme.headlineMedium),
              const SizedBox(height: 16),
              for (final week in weeks) ...[
                Padding(
                  padding: const EdgeInsets.only(top: 8, bottom: 8),
                  child: Text('Semana $week', style: Theme.of(context).textTheme.titleMedium),
                ),
                ...(byWeek[week]!..sort((a, b) => a.dayIndex.compareTo(b.dayIndex))).map((w) => _WorkoutTile(workout: w)),
              ],
            ],
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, _) => const Center(child: Text('No se pudo cargar tu plan')),
      ),
    );
  }
}

class _WorkoutTile extends ConsumerWidget {
  const _WorkoutTile({required this.workout});
  final AssignedWorkoutInfo workout;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colors = context.colors;
    final done = workout.status == 'completed';

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: done ? colors.success.withValues(alpha: 0.2) : colors.surfaceRaised,
          child: done
              ? Icon(Icons.check, color: colors.success)
              : Text('${workout.dayIndex}', style: TextStyle(color: colors.textMuted, fontWeight: FontWeight.w700)),
        ),
        title: Text(workout.name),
        subtitle: Text([
          '${workout.exercises.length} ejercicios',
          if (workout.scheduledOn.isNotEmpty) workout.scheduledOn,
        ].join(' · ')),
        trailing: done
            ? null
            : TextButton(
                onPressed: () {
                  ref.read(activeWorkoutProvider.notifier).start(fromAssigned: workout);
                  context.go('/workout');
                },
                child: const Text('Empezar'),
              ),
      ),
    );
  }
}
