import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/achievements/achievements_screen.dart';
import '../../features/auth/auth_providers.dart';
import '../../features/auth/login_screen.dart';
import '../../features/auth/register_screen.dart';
import '../../features/auth/splash_screen.dart';
import '../../features/habits/habits_screen.dart';
import '../../features/home/home_screen.dart';
import '../../features/home/home_shell.dart';
import '../../features/progress/progress_screen.dart';
import '../../features/workout/active_workout_screen.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authStatus = ref.watch(authControllerProvider).status;

  return GoRouter(
    initialLocation: '/',
    redirect: (context, state) {
      // Mientras no sabemos si hay sesion, no se monta ninguna pantalla que
      // dispare llamadas a la API (evita 401 de arranque).
      if (authStatus == AuthStatus.unknown) {
        return state.matchedLocation == '/splash' ? null : '/splash';
      }
      if (state.matchedLocation == '/splash') {
        return authStatus == AuthStatus.authenticated ? '/' : '/login';
      }
      final loggingIn = state.matchedLocation == '/login' || state.matchedLocation == '/register';
      if (authStatus == AuthStatus.unauthenticated && !loggingIn) return '/login';
      if (authStatus == AuthStatus.authenticated && loggingIn) return '/';
      return null;
    },
    routes: [
      GoRoute(path: '/splash', builder: (context, state) => const SplashScreen()),
      GoRoute(path: '/login', builder: (context, state) => const LoginScreen()),
      GoRoute(path: '/register', builder: (context, state) => const RegisterScreen()),
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
});
