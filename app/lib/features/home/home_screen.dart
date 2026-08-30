import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/network/models.dart';
import '../../core/providers.dart';
import '../../core/theme/app_colors.dart';
import '../auth/auth_providers.dart';
import '../workout/workout_providers.dart';

final statsProvider = FutureProvider.autoDispose<UserStats>((ref) => ref.watch(apiRepositoryProvider).stats());
final currentAssignmentProvider = FutureProvider.autoDispose<CurrentAssignment?>((ref) => ref.watch(apiRepositoryProvider).currentAssignment());

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colors = context.colors;
    final statsAsync = ref.watch(statsProvider);
    final assignmentAsync = ref.watch(currentAssignmentProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Myvibesfit'),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings_outlined),
            onPressed: () => context.push('/settings'),
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () => ref.read(authControllerProvider.notifier).logout(),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(statsProvider);
          ref.invalidate(currentAssignmentProvider);
          ref.invalidate(pendingSyncCountProvider);
        },
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            const _PendingSyncBanner(),
            statsAsync.when(
              data: (stats) => _StreakAndXpRow(stats: stats),
              loading: () => const SizedBox(height: 100, child: Center(child: CircularProgressIndicator())),
              error: (e, _) => Text('No se pudieron cargar tus stats', style: TextStyle(color: colors.danger)),
            ),
            const SizedBox(height: 16),
            assignmentAsync.when(
              data: (assignment) => _TodayWorkoutCard(assignment: assignment),
              loading: () => const SizedBox(height: 80, child: Center(child: CircularProgressIndicator())),
              error: (e, _) => const SizedBox.shrink(),
            ),
          ],
        ),
      ),
    );
  }
}

class _PendingSyncBanner extends ConsumerWidget {
  const _PendingSyncBanner();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colors = context.colors;
    final pending = ref.watch(pendingSyncCountProvider).valueOrNull ?? 0;
    if (pending == 0) return const SizedBox.shrink();

    final syncing = ref.watch(syncQueueProvider);
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Card(
        color: colors.warning.withValues(alpha: 0.12),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              Icon(Icons.cloud_off, color: colors.warning),
              const SizedBox(width: 10),
              Expanded(
                child: Text('$pending entrenamiento${pending > 1 ? "s" : ""} sin sincronizar', style: TextStyle(color: colors.text)),
              ),
              TextButton(
                onPressed: syncing
                    ? null
                    : () async {
                        await ref.read(syncQueueProvider.notifier).syncNow();
                        ref.invalidate(pendingSyncCountProvider);
                      },
                child: syncing ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2)) : const Text('Reintentar'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StreakAndXpRow extends StatelessWidget {
  const _StreakAndXpRow({required this.stats});
  final UserStats stats;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final workoutStreak = stats.streakFor('workout');
    final progress = (stats.totalXP % 500) / 500;

    return Row(
      children: [
        Expanded(
          child: Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(children: [
                    Icon(Icons.local_fire_department, color: colors.streakFire, size: 28),
                    const SizedBox(width: 4),
                    Text('$workoutStreak', style: Theme.of(context).textTheme.displayMedium),
                  ]),
                  const SizedBox(height: 4),
                  Text('días seguidos', style: Theme.of(context).textTheme.bodySmall),
                ],
              ),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Nivel ${stats.level}', style: Theme.of(context).textTheme.titleLarge),
                  const SizedBox(height: 8),
                  ClipRRect(
                    borderRadius: BorderRadius.circular(8),
                    child: LinearProgressIndicator(value: progress, minHeight: 8, backgroundColor: colors.border, color: colors.brand),
                  ),
                  const SizedBox(height: 4),
                  Text('${stats.totalXP} XP', style: Theme.of(context).textTheme.bodySmall),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _TodayWorkoutCard extends StatelessWidget {
  const _TodayWorkoutCard({required this.assignment});
  final CurrentAssignment? assignment;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    if (assignment == null || assignment!.workouts.isEmpty) {
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Sin plan asignado', style: Theme.of(context).textTheme.titleLarge),
              const SizedBox(height: 8),
              Text('Pídele a tu coach que te asigne un programa, o entrena libre.', style: Theme.of(context).textTheme.bodyMedium),
              const SizedBox(height: 16),
              ElevatedButton(onPressed: () => context.go('/workout'), child: const Text('Entrenar de todos modos')),
            ],
          ),
        ),
      );
    }

    final next = assignment!.workouts.firstWhere((w) => w.status == 'pending', orElse: () => assignment!.workouts.first);
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('PLAN ACTUAL', style: Theme.of(context).textTheme.labelLarge),
            const SizedBox(height: 4),
            Text(next.name, style: Theme.of(context).textTheme.headlineMedium),
            const SizedBox(height: 4),
            Text('Semana ${next.weekNumber} · Día ${next.dayIndex} · ${next.exercises.length} ejercicios',
                style: Theme.of(context).textTheme.bodySmall),
            const SizedBox(height: 16),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: colors.brand),
              onPressed: () => context.go('/workout'),
              child: const Text('Empezar entrenamiento'),
            ),
            TextButton(onPressed: () => context.push('/plan'), child: const Text('Ver plan completo')),
          ],
        ),
      ),
    );
  }
}
