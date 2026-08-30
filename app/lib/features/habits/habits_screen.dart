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
final todayHabitLogsProvider = FutureProvider.autoDispose<Set<String>>((ref) {
  final today = DateFormat('yyyy-MM-dd').format(DateTime.now());
  return ref.watch(apiRepositoryProvider).habitLogsForDate(today);
});

class HabitsScreen extends ConsumerWidget {
  const HabitsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final myHabitsAsync = ref.watch(myHabitsProvider);
    // Si esto todavia esta cargando o fallo, arrancamos asumiendo nada
    // marcado en vez de bloquear toda la lista por un fetch secundario.
    final checkedIds = ref.watch(todayHabitLogsProvider).valueOrNull ?? const <String>{};

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
            itemBuilder: (context, i) => _HabitCheckRow(habit: habits[i], initiallyChecked: checkedIds.contains(habits[i].id)),
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
  const _HabitCheckRow({required this.habit, required this.initiallyChecked});
  final MyHabit habit;
  final bool initiallyChecked;

  @override
  ConsumerState<_HabitCheckRow> createState() => _HabitCheckRowState();
}

class _HabitCheckRowState extends ConsumerState<_HabitCheckRow> {
  late bool _checkedToday = widget.initiallyChecked;
  bool _loading = false;

  Future<void> _check() async {
    setState(() => _loading = true);
    try {
      final today = DateFormat('yyyy-MM-dd').format(DateTime.now());
      final unlocked =
          await ref.read(apiRepositoryProvider).logHabit(clientHabitId: widget.habit.id, clientLocalId: _uuid.v4(), logDate: today);
      setState(() => _checkedToday = true);
      ref.invalidate(statsProvider);
      if (mounted && unlocked.isNotEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('🏆 Logro desbloqueado: ${unlocked.map((a) => a.name).join(", ")}')),
        );
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('No se pudo registrar el hábito. Probá de nuevo.')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _unsubscribe() async {
    try {
      await ref.read(apiRepositoryProvider).unsubscribeHabit(widget.habit.id);
    } finally {
      ref.invalidate(myHabitsProvider);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    return Dismissible(
      key: ValueKey(widget.habit.id),
      direction: DismissDirection.endToStart,
      confirmDismiss: (_) => showDialog<bool>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Dejar de seguir este hábito'),
          content: Text('¿Ya no querés seguir "${widget.habit.name}"?'),
          actions: [
            TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Cancelar')),
            TextButton(onPressed: () => Navigator.pop(context, true), child: const Text('Dejar de seguir')),
          ],
        ),
      ),
      onDismissed: (_) => _unsubscribe(),
      background: Container(
        alignment: Alignment.centerRight,
        padding: const EdgeInsets.only(right: 20),
        color: colors.danger,
        child: const Icon(Icons.delete_outline, color: Colors.white),
      ),
      child: Card(
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
      ),
    );
  }
}
