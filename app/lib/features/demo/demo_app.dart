import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/network/mock_api_repository.dart';
import '../../core/providers.dart';
import '../../core/theme/app_theme.dart';
import '../achievements/achievements_screen.dart';
import '../habits/habits_screen.dart';
import '../home/home_screen.dart';
import '../home/home_shell.dart';
import '../progress/progress_screen.dart';
import '../workout/active_workout_screen.dart';

final _demoRouter = GoRouter(
  initialLocation: '/',
  routes: [
    ShellRoute(
      builder: (context, state, child) => HomeShell(location: state.matchedLocation, child: child),
      routes: [
        GoRoute(path: '/', builder: (context, state) => const HomeScreen()),
        GoRoute(path: '/workout', builder: (context, state) => const ActiveWorkoutScreen()),
        GoRoute(path: '/progress', builder: (context, state) => const ProgressScreen()),
        GoRoute(path: '/habits', builder: (context, state) => const HabitsScreen()),
        GoRoute(path: '/achievements', builder: (context, state) => const AchievementsScreen()),
      ],
    ),
  ],
);

/// Navega las pantallas reales de la app con datos falsos, sin backend ni
/// login. Se abre como una pantalla separada (su propio MaterialApp/
/// ProviderScope) para no tocar el flujo de auth real; el boton flotante
/// cierra el demo y vuelve a login.
class DemoApp extends StatelessWidget {
  const DemoApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ProviderScope(
      overrides: [apiRepositoryProvider.overrideWithValue(MockApiRepository())],
      child: MaterialApp.router(
        title: 'Myvibesfit (demo)',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.from(brightness: Brightness.light),
        darkTheme: AppTheme.from(brightness: Brightness.dark),
        themeMode: ThemeMode.system,
        routerConfig: _demoRouter,
        builder: (context, child) => Stack(
          children: [
            ?child,
            Positioned(
              top: 12,
              right: 12,
              child: SafeArea(
                child: FloatingActionButton.small(
                  heroTag: 'exit-demo',
                  backgroundColor: Colors.black87,
                  onPressed: () => Navigator.of(context, rootNavigator: true).pop(),
                  child: const Icon(Icons.close, color: Colors.white),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
