import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/providers.dart';
import '../../core/theme/app_colors.dart';
import 'auth_providers.dart';

class RegisterScreen extends ConsumerStatefulWidget {
  const RegisterScreen({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  final _nameCtrl = TextEditingController();
  final _emailCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  final _joinCodeCtrl = TextEditingController();
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _nameCtrl.dispose();
    _emailCtrl.dispose();
    _passwordCtrl.dispose();
    _joinCodeCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ref.read(authControllerProvider.notifier).register(_emailCtrl.text.trim(), _passwordCtrl.text, _nameCtrl.text.trim());
    } catch (e) {
      // La cuenta no se creo: este es el unico caso donde el error es sobre
      // el registro en si.
      if (mounted) setState(() => _error = 'No se pudo crear la cuenta. ¿El correo ya existe?');
      if (mounted) setState(() => _loading = false);
      return;
    }

    final joinCode = _joinCodeCtrl.text.trim();
    if (joinCode.isNotEmpty) {
      try {
        await ref.read(apiRepositoryProvider).joinOrg(joinCode);
      } catch (e) {
        // La cuenta ya existe y ya esta autenticada (register() la dejo asi);
        // un codigo invalido no es un fallo de registro, se puede unir despues.
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Cuenta creada. El código de gimnasio no era válido — podés unirte después.')),
          );
        }
      }
    }
    if (mounted) setState(() => _loading = false);
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.colors;
    return Scaffold(
      appBar: AppBar(leading: BackButton(onPressed: () => context.go('/login'))),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 400),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text('Crear cuenta', style: Theme.of(context).textTheme.headlineMedium),
                  const SizedBox(height: 24),
                  TextField(controller: _nameCtrl, decoration: const InputDecoration(labelText: 'Nombre completo')),
                  const SizedBox(height: 12),
                  TextField(controller: _emailCtrl, decoration: const InputDecoration(labelText: 'Correo'), keyboardType: TextInputType.emailAddress),
                  const SizedBox(height: 12),
                  TextField(controller: _passwordCtrl, decoration: const InputDecoration(labelText: 'Contraseña'), obscureText: true),
                  const SizedBox(height: 12),
                  TextField(controller: _joinCodeCtrl, decoration: const InputDecoration(labelText: 'Código de gimnasio (opcional)')),
                  if (_error != null) ...[
                    const SizedBox(height: 12),
                    Text(_error!, style: TextStyle(color: colors.danger)),
                  ],
                  const SizedBox(height: 24),
                  ElevatedButton(
                    onPressed: _loading ? null : _submit,
                    child: _loading ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)) : const Text('Crear cuenta'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
