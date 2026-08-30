import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../core/providers.dart';
import '../../core/storage/workout_store.dart';
import '../../core/theme/app_colors.dart';

final sessionHistoryProvider =
    FutureProvider.autoDispose<List<LocalSession>>((ref) => ref.watch(workoutStoreProvider).recentSessions());

class SessionHistoryScreen extends ConsumerWidget {
  const SessionHistoryScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final historyAsync = ref.watch(sessionHistoryProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Historial')),
      body: historyAsync.when(
        data: (sessions) {
          if (sessions.isEmpty) {
            return const Center(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Text('Todavía no completaste ningún entrenamiento.', textAlign: TextAlign.center),
              ),
            );
          }
          return ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: sessions.length,
            itemBuilder: (context, i) => _SessionTile(session: sessions[i]),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, _) => const Center(child: Text('No se pudo cargar el historial')),
      ),
    );
  }
}

class _SessionTile extends StatelessWidget {
  const _SessionTile({required this.session});
  final LocalSession session;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final totalSets = session.exercises.fold<int>(0, (sum, e) => sum + e.sets.length);
    final volume = session.exercises.fold<double>(
      0,
      (sum, e) => sum + e.sets.fold<double>(0, (s, set) => s + ((set.weightKg ?? 0) * (set.reps ?? 0))),
    );
    final duration = session.endedAt?.difference(session.startedAt);

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        title: Row(
          children: [
            Expanded(child: Text(session.name)),
            if (!session.synced)
              Tooltip(
                message: 'Sin sincronizar',
                child: Icon(Icons.cloud_off, size: 16, color: colors.warning),
              ),
          ],
        ),
        subtitle: Text([
          DateFormat('d MMM').format(session.startedAt),
          '${session.exercises.length} ejercicios',
          '$totalSets series',
          if (volume > 0) '${volume.toStringAsFixed(0)} kg',
          if (duration != null && duration.inMinutes > 0) '${duration.inMinutes} min',
        ].join(' · ')),
      ),
    );
  }
}
