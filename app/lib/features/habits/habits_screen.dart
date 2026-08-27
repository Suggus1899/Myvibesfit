import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';
import 'package:uuid/uuid.dart';

import '../../core/network/models.dart';
import '../../core/providers.dart';
import '../../core/theme/app_colors.dart';
import '../home/home_screen.dart';

const _uuid = Uuid();

final myHabitsProvider = FutureProvider.autoDispose<List<MyHabit>>((ref) => ref.watch(apiRepositoryProvider).myHabits());
final allHabitsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) => ref.watch(apiRepositoryProvider).allHabits());

class HabitsScreen extends ConsumerWidget {
  const HabitsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final myHabitsAsync = ref.watch(myHabitsProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Hábitos')),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showHabitPicker(context, ref),
        child: const Icon(Icons.add),
      ),
      body: myHabitsAsync.when(
        data: (habits) {
          if (habits.isEmpty) {
            return const Center(child: Padding(padding: EdgeInsets.all(24), child: Text('Suscríbete a un hábito con el botón +')));
          }
          return ListView.builder(
            padding: const EdgeInsets.all(16),
            itemCount: habits.length,
            itemBuilder: (context, i) => _HabitCheckRow(habit: habits[i]),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => const Center(child: Text('No se pudieron cargar tus hábitos')),
      ),
    );
  }

  void _showHabitPicker(BuildContext context, WidgetRef ref) async {
    final all = await ref.read(allHabitsProvider.future);
    if (!context.mounted) return;
    await showModalBottomSheet(
      context: context,
      builder: (context) => ListView(
        children: all
            .map((h) => ListTile(
                  leading: const Icon(Icons.check_circle_outline),
                  title: Text(h['name']),
                  onTap: () async {
                    await ref.read(apiRepositoryProvider).subscribeHabit(h['id']);
                    ref.invalidate(myHabitsProvider);
                    if (context.mounted) Navigator.pop(context);
                  },
                ))
            .toList(),
      ),
    );
  }
}

class _HabitCheckRow extends ConsumerStatefulWidget {
  const _HabitCheckRow({required this.habit});
  final MyHabit habit;

  @override
  ConsumerState<_HabitCheckRow> createState() => _HabitCheckRowState();
}

class _HabitCheckRowState extends ConsumerState<_HabitCheckRow> {
  bool _checkedToday = false;
  bool _loading = false;

  Future<void> _check() async {
    setState(() => _loading = true);
    try {
      final today = DateFormat('yyyy-MM-dd').format(DateTime.now());
      await ref.read(apiRepositoryProvider).logHabit(clientHabitId: widget.habit.id, clientLocalId: _uuid.v4(), logDate: today);
      setState(() => _checkedToday = true);
      ref.invalidate(statsProvider);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: ListTile(
        leading: CircleAvatar(backgroundColor: colors.surfaceRaised, child: Icon(Icons.check, color: _checkedToday ? colors.success : colors.textMuted)),
        title: Text(widget.habit.name),
        trailing: _loading
            ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2))
            : IconButton(
                icon: Icon(_checkedToday ? Icons.check_circle : Icons.check_circle_outline, color: _checkedToday ? colors.success : colors.textMuted),
                onPressed: _checkedToday ? null : _check,
              ),
      ),
    );
  }
}
