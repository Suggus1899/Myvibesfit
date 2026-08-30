import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/network/models.dart';
import '../../core/providers.dart';
import '../../core/theme/app_colors.dart';

final profileProvider = FutureProvider.autoDispose<ClientProfile>((ref) => ref.watch(apiRepositoryProvider).profile());

const _goals = {
  'general_health': 'Salud general',
  'hypertrophy': 'Ganar músculo',
  'strength': 'Ganar fuerza',
  'fat_loss': 'Bajar grasa',
  'endurance': 'Resistencia',
};

const _levels = {
  'beginner': 'Principiante',
  'intermediate': 'Intermedio',
  'advanced': 'Avanzado',
};

const _equipment = ['barbell', 'dumbbell', 'machine', 'cable', 'kettlebell', 'bodyweight', 'bands'];

class ProfileScreen extends ConsumerWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profileAsync = ref.watch(profileProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Mi perfil')),
      body: profileAsync.when(
        data: (p) => _ProfileForm(initial: p),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, _) => const Center(child: Text('No se pudo cargar tu perfil')),
      ),
    );
  }
}

class _ProfileForm extends ConsumerStatefulWidget {
  const _ProfileForm({required this.initial});
  final ClientProfile initial;

  @override
  ConsumerState<_ProfileForm> createState() => _ProfileFormState();
}

class _ProfileFormState extends ConsumerState<_ProfileForm> {
  late String _goal = widget.initial.primaryGoal;
  late String _level = widget.initial.experience;
  late int _days = widget.initial.daysPerWeek;
  late int _minutes = widget.initial.sessionMinutes;
  late String _units = widget.initial.unitSystem;
  late final Set<String> _equip = {...widget.initial.availableEquipment};
  late final _heightCtrl = TextEditingController(text: widget.initial.heightCm?.toStringAsFixed(0) ?? '');
  late final _limitCtrl = TextEditingController(text: widget.initial.limitations);

  bool _saving = false;
  String? _message;

  @override
  void dispose() {
    _heightCtrl.dispose();
    _limitCtrl.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    setState(() {
      _saving = true;
      _message = null;
    });
    try {
      await ref.read(apiRepositoryProvider).saveProfile(
            sex: widget.initial.sex,
            experience: _level,
            primaryGoal: _goal,
            daysPerWeek: _days,
            sessionMinutes: _minutes,
            unitSystem: _units,
            heightCm: double.tryParse(_heightCtrl.text.replaceAll(',', '.')),
            availableEquipment: _equip.toList(),
            limitations: _limitCtrl.text.trim(),
          );
      ref.invalidate(profileProvider);
      if (mounted) setState(() => _message = 'Perfil guardado.');
    } catch (_) {
      if (mounted) setState(() => _message = 'No se pudo guardar. Probá de nuevo.');
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        if (!widget.initial.isOnboarded)
          Card(
            color: colors.brand.withValues(alpha: 0.15),
            child: const Padding(
              padding: EdgeInsets.all(12),
              child: Text('Completá tu perfil para que tu coach y las sugerencias se ajusten a vos.'),
            ),
          ),
        const SizedBox(height: 12),
        Text('Objetivo', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        DropdownButtonFormField<String>(
          initialValue: _goal,
          items: [for (final e in _goals.entries) DropdownMenuItem(value: e.key, child: Text(e.value))],
          onChanged: (v) => setState(() => _goal = v ?? _goal),
        ),
        const SizedBox(height: 16),
        Text('Experiencia', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        DropdownButtonFormField<String>(
          initialValue: _level,
          items: [for (final e in _levels.entries) DropdownMenuItem(value: e.key, child: Text(e.value))],
          onChanged: (v) => setState(() => _level = v ?? _level),
        ),
        const SizedBox(height: 20),
        Text('Días por semana: $_days', style: Theme.of(context).textTheme.titleMedium),
        Slider(
          value: _days.toDouble(),
          min: 1,
          max: 7,
          divisions: 6,
          label: '$_days',
          onChanged: (v) => setState(() => _days = v.round()),
        ),
        Text('Minutos por sesión: $_minutes', style: Theme.of(context).textTheme.titleMedium),
        Slider(
          value: _minutes.toDouble(),
          min: 20,
          max: 120,
          divisions: 10,
          label: '$_minutes',
          onChanged: (v) => setState(() => _minutes = v.round()),
        ),
        const SizedBox(height: 12),
        TextField(
          controller: _heightCtrl,
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          decoration: const InputDecoration(labelText: 'Altura (cm)'),
        ),
        const SizedBox(height: 20),
        Text('Equipamiento disponible', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          children: [
            for (final eq in _equipment)
              FilterChip(
                label: Text(eq),
                selected: _equip.contains(eq),
                onSelected: (on) => setState(() => on ? _equip.add(eq) : _equip.remove(eq)),
              ),
          ],
        ),
        const SizedBox(height: 20),
        TextField(
          controller: _limitCtrl,
          maxLines: 2,
          decoration: const InputDecoration(labelText: 'Lesiones o limitaciones (opcional)'),
        ),
        const SizedBox(height: 20),
        Text('Unidades', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        SegmentedButton<String>(
          segments: const [
            ButtonSegment(value: 'metric', label: Text('kg / cm')),
            ButtonSegment(value: 'imperial', label: Text('lb / in')),
          ],
          selected: {_units},
          onSelectionChanged: (s) => setState(() => _units = s.first),
        ),
        if (_message != null) ...[
          const SizedBox(height: 12),
          Text(_message!, style: TextStyle(color: colors.textMuted)),
        ],
        const SizedBox(height: 24),
        FilledButton(
          onPressed: _saving ? null : _save,
          child: _saving ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)) : const Text('Guardar perfil'),
        ),
      ],
    );
  }
}
