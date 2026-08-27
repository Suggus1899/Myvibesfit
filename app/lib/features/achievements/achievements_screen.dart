import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/network/models.dart';
import '../../core/providers.dart';
import '../../core/theme/app_colors.dart';

final achievementsProvider = FutureProvider.autoDispose<List<UserAchievementInfo>>((ref) => ref.watch(apiRepositoryProvider).achievements());

class AchievementsScreen extends ConsumerWidget {
  const AchievementsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final achievementsAsync = ref.watch(achievementsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Logros')),
      body: achievementsAsync.when(
        data: (achievements) {
          if (achievements.isEmpty) {
            return const Center(child: Padding(padding: EdgeInsets.all(24), child: Text('Entrena y cumple hábitos para desbloquear logros.')));
          }
          return GridView.builder(
            padding: const EdgeInsets.all(16),
            gridDelegate:
                const SliverGridDelegateWithFixedCrossAxisCount(crossAxisCount: 2, mainAxisSpacing: 12, crossAxisSpacing: 12, childAspectRatio: 0.9),
            itemCount: achievements.length,
            itemBuilder: (context, i) => _AchievementCard(achievement: achievements[i]),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => const Center(child: Text('No se pudieron cargar los logros')),
      ),
    );
  }
}

class _AchievementCard extends StatelessWidget {
  const _AchievementCard({required this.achievement});
  final UserAchievementInfo achievement;

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final earned = achievement.isEarned;
    return Card(
      color: earned ? colors.surfaceRaised : colors.surface,
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.emoji_events, size: 40, color: earned ? colors.brand : colors.border),
            const SizedBox(height: 8),
            Text(achievement.name,
                textAlign: TextAlign.center, style: Theme.of(context).textTheme.titleLarge?.copyWith(color: earned ? colors.text : colors.textMuted)),
            const SizedBox(height: 4),
            Text(achievement.description, textAlign: TextAlign.center, style: Theme.of(context).textTheme.bodySmall),
          ],
        ),
      ),
    );
  }
}
