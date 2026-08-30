import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/network/models.dart';
import '../../core/providers.dart';
import '../../core/theme/app_colors.dart';
import '../auth/auth_providers.dart';

final meProvider = FutureProvider.autoDispose<AuthUser>((ref) => ref.watch(apiRepositoryProvider).me());

class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  final _joinCodeCtrl = TextEditingController();
  bool _joining = false;
  String? _message;

  @override
  void dispose() {
    _joinCodeCtrl.dispose();
    super.dispose();
  }

  Future<void> _join() async {
    final code = _joinCodeCtrl.text.trim();
    if (code.isEmpty) return;
    setState(() {
      _joining = true;
      _message = null;
    });
    try {
      await ref.read(apiRepositoryProvider).joinOrg(code);
      _joinCodeCtrl.clear();
      ref.invalidate(meProvider);
      if (mounted) setState(() => _message = 'Te uniste al gimnasio.');
    } catch (_) {
      if (mounted) setState(() => _message = 'Código inválido.');
    } finally {
      if (mounted) setState(() => _joining = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    final meAsync = ref.watch(meProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Configuración')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: ListView(
          children: [
            meAsync.when(
              data: (me) => Card(
                child: ListTile(
                  title: Text(me.fullName),
                  subtitle: Text(me.orgId == null ? '${me.email} · sin gimnasio' : '${me.email} · ${me.role ?? "miembro"}'),
                ),
              ),
              loading: () => const Center(child: Padding(padding: EdgeInsets.all(16), child: CircularProgressIndicator())),
              error: (e, _) => const SizedBox.shrink(),
            ),
            const SizedBox(height: 12),
            Card(
              child: ListTile(
                leading: const Icon(Icons.person_outline),
                title: const Text('Mi perfil'),
                subtitle: const Text('Objetivo, experiencia, equipamiento y unidades'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push('/profile'),
              ),
            ),
            const SizedBox(height: 24),
            Text('Unirse a un gimnasio', style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 8),
            Text('Pedile el código a tu coach.', style: TextStyle(color: colors.textMuted)),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _joinCodeCtrl,
                    textCapitalization: TextCapitalization.characters,
                    decoration: const InputDecoration(labelText: 'Código de gimnasio'),
                  ),
                ),
                const SizedBox(width: 12),
                FilledButton(
                  onPressed: _joining ? null : _join,
                  child: _joining ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2)) : const Text('Unirse'),
                ),
              ],
            ),
            if (_message != null) ...[
              const SizedBox(height: 8),
              Text(_message!, style: TextStyle(color: colors.textMuted)),
            ],
            const SizedBox(height: 32),
            OutlinedButton.icon(
              icon: const Icon(Icons.logout),
              label: const Text('Cerrar sesión'),
              onPressed: () => ref.read(authControllerProvider.notifier).logout(),
            ),
          ],
        ),
      ),
    );
  }
}
