import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/network/models.dart';
import '../../core/providers.dart';
import '../../core/theme/app_colors.dart';

final exerciseDetailProvider =
    FutureProvider.autoDispose.family<ExerciseDetail, String>((ref, id) => ref.watch(apiRepositoryProvider).exercise(id));

class ExerciseDetailScreen extends ConsumerWidget {
  const ExerciseDetailScreen({super.key, required this.exerciseId});
  final String exerciseId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detailAsync = ref.watch(exerciseDetailProvider(exerciseId));

    return Scaffold(
      appBar: AppBar(title: const Text('Ejercicio')),
      body: detailAsync.when(
        data: (e) => _Detail(exercise: e),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, _) => const Center(child: Padding(padding: EdgeInsets.all(24), child: Text('No se pudo cargar el ejercicio.'))),
      ),
    );
  }
}

class _Detail extends StatelessWidget {
  const _Detail({required this.exercise});
  final ExerciseDetail exercise;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final tags = [
      if (exercise.primaryMuscle.isNotEmpty) exercise.primaryMuscle,
      ...exercise.secondaryMuscles,
      ...exercise.equipment,
      if (exercise.difficulty.isNotEmpty) exercise.difficulty,
      if (exercise.isUnilateral) 'unilateral',
    ];

    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Text(exercise.name, style: Theme.of(context).textTheme.headlineMedium),
        if (exercise.pattern.isNotEmpty) ...[
          const SizedBox(height: 4),
          Text(exercise.pattern, style: TextStyle(color: colors.textMuted)),
        ],
        if (tags.isNotEmpty) ...[
          const SizedBox(height: 16),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: tags
                .map((t) => Chip(
                      label: Text(t),
                      backgroundColor: colors.surfaceRaised,
                      side: BorderSide(color: colors.border),
                    ))
                .toList(),
          ),
        ],
        if (exercise.description.isNotEmpty) ...[
          const SizedBox(height: 24),
          Text(exercise.description, style: Theme.of(context).textTheme.bodyMedium),
        ],
        if (exercise.instructions.isNotEmpty) ...[
          const SizedBox(height: 24),
          Text('Cómo se hace', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 12),
          ...exercise.instructions.asMap().entries.map(
                (entry) => Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      CircleAvatar(
                        radius: 12,
                        backgroundColor: colors.brand,
                        child: Text('${entry.key + 1}',
                            style: TextStyle(color: colors.brandInk, fontSize: 12, fontWeight: FontWeight.w700)),
                      ),
                      const SizedBox(width: 12),
                      Expanded(child: Text(entry.value, style: Theme.of(context).textTheme.bodyMedium)),
                    ],
                  ),
                ),
              ),
        ],
        if (exercise.videoUrl.isNotEmpty) ...[
          const SizedBox(height: 16),
          // Sin reproductor embebido todavia: mostramos la URL para no
          // prometer un player que no existe.
          Text('Video de referencia', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 4),
          SelectableText(exercise.videoUrl, style: TextStyle(color: colors.info)),
        ],
      ],
    );
  }
}
