import 'package:dio/dio.dart';

import 'api_client.dart';
import 'models.dart';
import 'token_storage.dart';

/// Punto unico de acceso a la API. Las pantallas no tocan dio directamente.
class ApiRepository {
  ApiRepository(this._client, this._tokenStorage);

  final ApiClient _client;
  final TokenStorage _tokenStorage;

  Dio get _dio => _client.dio;

  Future<AuthResult> register({required String email, required String password, required String fullName}) async {
    final res = await _dio.post('/v1/auth/register', data: {'email': email, 'password': password, 'full_name': fullName});
    final result = AuthResult.fromJson(res.data);
    await _tokenStorage.save(accessToken: result.accessToken, refreshToken: result.refreshToken);
    return result;
  }

  Future<AuthResult> login({required String email, required String password}) async {
    final res = await _dio.post('/v1/auth/login', data: {'email': email, 'password': password});
    final result = AuthResult.fromJson(res.data);
    await _tokenStorage.save(accessToken: result.accessToken, refreshToken: result.refreshToken);
    return result;
  }

  Future<void> logout() => _tokenStorage.clear();

  Future<bool> hasSession() async => (await _tokenStorage.readAccessToken()) != null;

  Future<AuthUser> me() async {
    final res = await _dio.get('/v1/me');
    return AuthUser.fromJson(res.data);
  }

  Future<ClientProfile> profile() async {
    final res = await _dio.get('/v1/me/profile');
    return ClientProfile.fromJson(res.data);
  }

  Future<ClientProfile> saveProfile({
    required String sex,
    required String experience,
    required String primaryGoal,
    required int daysPerWeek,
    required int sessionMinutes,
    required String unitSystem,
    double? heightCm,
    List<String>? availableEquipment,
    String? limitations,
  }) async {
    final res = await _dio.put('/v1/me/profile', data: {
      'sex': sex,
      'experience': experience,
      'primary_goal': primaryGoal,
      'days_per_week': daysPerWeek,
      'session_minutes': sessionMinutes,
      'unit_system': unitSystem,
      'height_cm': heightCm,
      'available_equipment': availableEquipment ?? [],
      'limitations': limitations ?? '',
    });
    return ClientProfile.fromJson(res.data);
  }

  Future<AuthResult> joinOrg(String joinCode) async {
    final res = await _dio.post('/v1/orgs/join', data: {'join_code': joinCode});
    final result = AuthResult.fromJson(res.data);
    // A diferencia de login/register, esto se olvidaba de guardar los tokens
    // nuevos: el JWT en disco quedaba sin org_id hasta el proximo refresh.
    await _tokenStorage.save(accessToken: result.accessToken, refreshToken: result.refreshToken);
    return result;
  }

  Future<List<ExerciseSummary>> exercises({String? pattern}) async {
    final res = await _dio.get('/v1/exercises', queryParameters: {'pattern': ?pattern, 'limit': 100});
    return (res.data as List).map((e) => ExerciseSummary.fromJson(e)).toList();
  }

  Future<ExerciseDetail> exercise(String id) async {
    final res = await _dio.get('/v1/exercises/$id');
    return ExerciseDetail.fromJson(res.data);
  }

  Future<CurrentAssignment?> currentAssignment() async {
    try {
      final res = await _dio.get('/v1/me/assignment/current');
      return CurrentAssignment.fromJson(res.data);
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) return null;
      rethrow;
    }
  }

  Future<SyncResult> syncSessions(List<Map<String, dynamic>> sessions) async {
    final res = await _dio.post('/v1/sync/sessions', data: {'sessions': sessions});
    return SyncResult.fromJson(res.data as Map<String, dynamic>);
  }

  Future<UserStats> stats() async {
    final res = await _dio.get('/v1/me/stats');
    return UserStats.fromJson(res.data);
  }

  Future<List<UserAchievementInfo>> achievements() async {
    final res = await _dio.get('/v1/me/achievements');
    return (res.data as List).map((a) => UserAchievementInfo.fromJson(a)).toList();
  }

  Future<List<Map<String, dynamic>>> allHabits() async {
    final res = await _dio.get('/v1/habits');
    return (res.data as List).cast<Map<String, dynamic>>();
  }

  Future<List<MyHabit>> myHabits() async {
    final res = await _dio.get('/v1/me/habits');
    return (res.data as List).map((h) => MyHabit.fromJson(h)).toList();
  }

  Future<void> subscribeHabit(String habitId) async {
    await _dio.post('/v1/habits/subscribe', data: {'habit_id': habitId, 'frequency': 'daily'});
  }

  Future<void> unsubscribeHabit(String clientHabitId) async {
    await _dio.delete('/v1/habits/$clientHabitId');
  }

  /// Ids de client_habit ya marcados en la fecha dada — permite que la UI
  /// arranque con el estado real del servidor en vez de asumir "sin marcar".
  Future<Set<String>> habitLogsForDate(String date) async {
    final res = await _dio.get('/v1/habits/logs', queryParameters: {'date': date});
    return (res.data as List).map((l) => l['client_habit_id'] as String).toSet();
  }

  Future<List<UserAchievementInfo>> logHabit({required String clientHabitId, required String clientLocalId, required String logDate}) async {
    final res = await _dio.post('/v1/habits/$clientHabitId/log', data: {
      'client_local_id': clientLocalId,
      'log_date': logDate,
      'value': 1,
      'is_completed': true,
    });
    final unlocked = (res.data as Map<String, dynamic>)['unlocked_achievements'] as List? ?? [];
    return unlocked.map((a) => UserAchievementInfo.fromJson(a)).toList();
  }

  Future<List<PersonalRecordInfo>> records() async {
    final res = await _dio.get('/v1/progress/records');
    return (res.data as List).map((r) => PersonalRecordInfo.fromJson(r)).toList();
  }

  Future<List<SetLogPoint>> exerciseHistory(String exerciseId) async {
    final res = await _dio.get('/v1/progress/exercises/$exerciseId', queryParameters: {'days': 365});
    return (res.data as List).map((s) => SetLogPoint.fromJson(s)).toList();
  }

  Future<void> logBodyMetric({required String measuredOn, double? weightKg, double? bodyFatPct, String? note}) async {
    await _dio.post('/v1/body-metrics', data: {
      'measured_on': measuredOn,
      'weight_kg': weightKg,
      'body_fat_pct': bodyFatPct,
      'note': note ?? '',
    });
  }

  Future<List<BodyMetric>> bodyMetrics() async {
    final res = await _dio.get('/v1/body-metrics');
    return (res.data as List).map((m) => BodyMetric.fromJson(m)).toList();
  }
}
