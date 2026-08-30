import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../workout/workout_providers.dart';

class HomeShell extends ConsumerStatefulWidget {
  const HomeShell({super.key, required this.child, required this.location});

  final Widget child;
  final String location;

  static const _tabs = ['/', '/progress', '/habits', '/achievements'];

  @override
  ConsumerState<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends ConsumerState<HomeShell> with WidgetsBindingObserver {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    // Reintento al abrir la app: una sesion sincronizada sin red durante el
    // ultimo entrenamiento no deberia esperar a que se termine otro.
    Future.microtask(() => ref.read(syncQueueProvider.notifier).syncNow());
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      ref.read(syncQueueProvider.notifier).syncNow();
    }
  }

  int get _currentIndex {
    final index = HomeShell._tabs.indexOf(widget.location);
    return index == -1 ? 0 : index;
  }

  @override
  Widget build(BuildContext context) {
    // /workout no es una tab: es una sesion en curso, no tiene sentido dejar
    // la barra inferior visible (ni resaltando "Inicio" por descarte).
    final isWorkout = widget.location == '/workout';
    return Scaffold(
      body: widget.child,
      floatingActionButton: widget.location == '/'
          ? FloatingActionButton.extended(
              onPressed: () => context.go('/workout'),
              icon: const Icon(Icons.fitness_center),
              label: const Text('Entrenar'),
            )
          : null,
      bottomNavigationBar: isWorkout
          ? null
          : NavigationBar(
              selectedIndex: _currentIndex,
              onDestinationSelected: (i) => context.go(HomeShell._tabs[i]),
              destinations: const [
                NavigationDestination(icon: Icon(Icons.home_outlined), selectedIcon: Icon(Icons.home), label: 'Inicio'),
                NavigationDestination(icon: Icon(Icons.show_chart_outlined), selectedIcon: Icon(Icons.show_chart), label: 'Progreso'),
                NavigationDestination(icon: Icon(Icons.check_circle_outline), selectedIcon: Icon(Icons.check_circle), label: 'Hábitos'),
                NavigationDestination(icon: Icon(Icons.emoji_events_outlined), selectedIcon: Icon(Icons.emoji_events), label: 'Logros'),
              ],
            ),
    );
  }
}
