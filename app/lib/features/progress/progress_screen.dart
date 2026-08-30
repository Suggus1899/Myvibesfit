import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/network/models.dart';
import '../../core/providers.dart';
import '../../core/theme/app_colors.dart';

final recordsProvider = FutureProvider.autoDispose<List<PersonalRecordInfo>>((ref) => ref.watch(apiRepositoryProvider).records());
final selectedExerciseProvider = StateProvider<String?>((ref) => null);
final exerciseHistoryProvider = FutureProvider.autoDispose<List<SetLogPoint>>((ref) async {
  final exerciseId = ref.watch(selectedExerciseProvider);
  if (exerciseId == null) return [];
  return ref.watch(apiRepositoryProvider).exerciseHistory(exerciseId);
});

class ProgressScreen extends ConsumerWidget {
  const ProgressScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final recordsAsync = ref.watch(recordsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Progreso'),
        actions: [
          IconButton(
            icon: const Icon(Icons.monitor_weight_outlined),
            tooltip: 'Peso y medidas',
            onPressed: () => context.push('/body-metrics'),
          ),
          IconButton(
            icon: const Icon(Icons.history),
            tooltip: 'Historial',
            onPressed: () => context.push('/history'),
          ),
        ],
      ),
      body: recordsAsync.when(
        data: (records) {
          if (records.isEmpty) {
            return const Center(child: Padding(padding: EdgeInsets.all(24), child: Text('Todavía no hay records. Completa un entrenamiento primero.')));
          }
          final byExercise = <String, List<PersonalRecordInfo>>{};
          for (final r in records) {
            byExercise.putIfAbsent(r.exerciseId, () => []).add(r);
          }
          return ListView(
            padding: const EdgeInsets.all(16),
            children: byExercise.entries.map((entry) => _ExerciseRecordsCard(exerciseId: entry.key, records: entry.value)).toList(),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => const Center(child: Text('No se pudo cargar el progreso')),
      ),
    );
  }
}

class _ExerciseRecordsCard extends ConsumerWidget {
  const _ExerciseRecordsCard({required this.exerciseId, required this.records});
  final String exerciseId;
  final List<PersonalRecordInfo> records;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colors = context.colors;
    final selected = ref.watch(selectedExerciseProvider) == exerciseId;
    final maxWeightRecords = records.where((r) => r.type == 'max_weight');
    final maxWeight = maxWeightRecords.isEmpty ? null : maxWeightRecords.first.value;
    final catalog = ref.watch(exerciseCatalogProvider).valueOrNull ?? const <ExerciseSummary>[];
    final name = resolveExerciseName(catalog, exerciseId);

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Expanded(
                  child: InkWell(
                    onTap: () => context.push('/exercises/$exerciseId'),
                    child: Text(name, style: Theme.of(context).textTheme.titleLarge),
                  ),
                ),
                if (maxWeight != null) Text('${maxWeight.toStringAsFixed(1)} kg', style: TextStyle(color: colors.brand, fontWeight: FontWeight.w800)),
              ],
            ),
            TextButton(
              onPressed: () => ref.read(selectedExerciseProvider.notifier).state = selected ? null : exerciseId,
              child: Text(selected ? 'Ocultar gráfica' : 'Ver gráfica'),
            ),
            if (selected) const _ExerciseChart(),
          ],
        ),
      ),
    );
  }
}

class _ExerciseChart extends ConsumerWidget {
  const _ExerciseChart();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colors = context.colors;
    final historyAsync = ref.watch(exerciseHistoryProvider);

    return historyAsync.when(
      data: (points) {
        if (points.isEmpty) return const SizedBox.shrink();
        final sorted = points.reversed.toList();
        final spots = <FlSpot>[
          for (var i = 0; i < sorted.length; i++)
            if (sorted[i].weightKg != null) FlSpot(i.toDouble(), sorted[i].weightKg!),
        ];
        return SizedBox(
          height: 160,
          child: Padding(
            padding: const EdgeInsets.only(top: 12),
            child: LineChart(LineChartData(
              gridData: const FlGridData(show: false),
              titlesData: const FlTitlesData(show: false),
              borderData: FlBorderData(show: false),
              lineBarsData: [
                LineChartBarData(spots: spots, isCurved: true, color: colors.brand, barWidth: 3, dotData: const FlDotData(show: false)),
              ],
            )),
          ),
        );
      },
      loading: () => const SizedBox(height: 40, child: Center(child: CircularProgressIndicator(strokeWidth: 2))),
      error: (e, _) => const SizedBox.shrink(),
    );
  }
}
