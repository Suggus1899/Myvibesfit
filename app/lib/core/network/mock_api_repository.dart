import 'models.dart';
import 'api_repository.dart';

const _exerciseIds = ['Sentadilla', 'Press banca', 'Peso muerto', 'Remo con barra', 'Press militar', 'Curl de biceps'];

/// Repositorio con datos falsos: implementa el mismo contrato que
/// [ApiRepository] pero sin tocar la red, para poder navegar la app entera
/// sin backend (ver DemoApp). Los ids de ejercicio son nombres legibles a
/// proposito: la UI no resuelve exercise_id a nombre (limitacion existente,
/// ver comentario en workout_providers.dart), asi que esto lo simula gratis.
class MockApiRepository implements ApiRepository {
  DateTime _daysAgo(int d) => DateTime.now().subtract(Duration(days: d));

  @override
  Future<AuthResult> register({required String email, required String password, required String fullName}) async =>
      AuthResult.fromJson({
        'access_token': 'demo', 'refresh_token': 'demo',
        'user': {'id': 'demo-user', 'email': email, 'full_name': fullName, 'org_id': null, 'role': 'client'},
      });

  @override
  Future<AuthResult> login({required String email, required String password}) =>
      register(email: email, password: password, fullName: 'Cliente Demo');

  @override
  Future<void> logout() async {}

  @override
  Future<bool> hasSession() async => true;

  @override
  Future<AuthUser> me() async => AuthUser.fromJson(
      {'id': 'demo-user', 'email': 'demo@myvibesfit.app', 'full_name': 'Cliente Demo', 'org_id': 'demo-org', 'role': 'client'});

  @override
  Future<AuthResult> joinOrg(String joinCode) => register(email: 'demo@myvibesfit.app', password: '', fullName: 'Cliente Demo');

  @override
  Future<List<ExerciseSummary>> exercises({String? pattern}) async => _exerciseIds
      .map((name) => ExerciseSummary.fromJson({'id': name, 'name': name, 'pattern': 'push', 'primary_muscle': 'general'}))
      .toList();

  @override
  Future<CurrentAssignment?> currentAssignment() async => CurrentAssignment.fromJson({
        'assignment': {'id': 'demo-assignment', 'name': 'Plan Fuerza 4 dias'},
        'workouts': [
          {
            'id': 'demo-workout-1',
            'week_number': 2,
            'day_index': 3,
            'name': 'Empuje',
            'scheduled_on': DateTime.now().toIso8601String().substring(0, 10),
            'status': 'pending',
            'exercises': [
              {'id': 'e1', 'exercise_id': 'Press banca', 'order_index': 1, 'target_sets': 4, 'target_reps_min': 6, 'target_reps_max': 8, 'target_rpe': 8.0, 'rest_seconds': 120},
              {'id': 'e2', 'exercise_id': 'Press militar', 'order_index': 2, 'target_sets': 3, 'target_reps_min': 8, 'target_reps_max': 10, 'target_rpe': 7.5, 'rest_seconds': 90},
            ],
          },
          {
            'id': 'demo-workout-0',
            'week_number': 2,
            'day_index': 2,
            'name': 'Tirón',
            'scheduled_on': _daysAgo(1).toIso8601String().substring(0, 10),
            'status': 'completed',
            'exercises': [
              {'id': 'e3', 'exercise_id': 'Remo con barra', 'order_index': 1, 'target_sets': 4, 'target_reps_min': 8, 'target_reps_max': 10, 'target_rpe': 8.0, 'rest_seconds': 90},
            ],
          },
        ],
      });

  @override
  Future<Map<String, dynamic>> syncSessions(List<Map<String, dynamic>> sessions) async => {'synced': sessions.length};

  @override
  Future<UserStats> stats() async => UserStats.fromJson({
        'total_xp': 1240,
        'level': 4,
        'total_sessions': 23,
        'streaks': [
          {'kind': 'workout', 'current_count': 7, 'last_active_on': null},
        ],
      });

  @override
  Future<List<UserAchievementInfo>> achievements() async => [
        {'code': 'first_workout', 'name': 'Primer entrenamiento', 'description': 'Completaste tu primera sesión', 'icon': '', 'earned_at': _daysAgo(20).toIso8601String()},
        {'code': 'streak_7', 'name': 'Racha de 7', 'description': '7 días seguidos entrenando', 'icon': '', 'earned_at': _daysAgo(1).toIso8601String()},
        {'code': 'pr_master', 'name': 'Rompe récords', 'description': 'Logra 5 records personales', 'icon': '', 'earned_at': ''},
        {'code': 'habit_10', 'name': 'Constancia', 'description': 'Cumple un hábito 10 veces', 'icon': '', 'earned_at': ''},
      ].map((j) => UserAchievementInfo.fromJson(j)).toList();

  @override
  Future<List<Map<String, dynamic>>> allHabits() async => [
        {'id': 'habit-water', 'name': 'Tomar 2L de agua'},
        {'id': 'habit-sleep', 'name': 'Dormir 8 horas'},
        {'id': 'habit-steps', 'name': 'Caminar 8000 pasos'},
      ];

  @override
  Future<List<MyHabit>> myHabits() async => [
        MyHabit.fromJson({'id': 'my-habit-1', 'habit_id': 'habit-water', 'name': 'Tomar 2L de agua', 'icon': ''}),
        MyHabit.fromJson({'id': 'my-habit-2', 'habit_id': 'habit-sleep', 'name': 'Dormir 8 horas', 'icon': ''}),
      ];

  @override
  Future<void> subscribeHabit(String habitId) async {}

  @override
  Future<void> logHabit({required String clientHabitId, required String clientLocalId, required String logDate}) async {}

  @override
  Future<List<PersonalRecordInfo>> records() async => [
        {'exercise_id': 'Press banca', 'type': 'max_weight', 'value': 82.5},
        {'exercise_id': 'Sentadilla', 'type': 'max_weight', 'value': 110.0},
        {'exercise_id': 'Peso muerto', 'type': 'max_weight', 'value': 140.0},
      ].map((j) => PersonalRecordInfo.fromJson(j)).toList();

  @override
  Future<List<SetLogPoint>> exerciseHistory(String exerciseId) async {
    final base = 60 + exerciseId.hashCode % 20;
    return List.generate(8, (i) {
      final weight = base + i * 2.5;
      return SetLogPoint.fromJson({'weight_kg': weight, 'reps': 6, 'performed_at': _daysAgo((7 - i) * 5).toIso8601String()});
    });
  }
}
