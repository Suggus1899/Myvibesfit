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

  Future<AuthResult> joinOrg(String joinCode) async {
    final res = await _dio.post('/v1/orgs/join', data: {'join_code': joinCode});
    return AuthResult.fromJson(res.data);
  }

  Future<List<ExerciseSummary>> exercises({String? pattern}) async {
    final res = await _dio.get('/v1/exercises', queryParameters: {'pattern': ?pattern, 'limit': 100});
    return (res.data as List).map((e) => ExerciseSummary.fromJson(e)).toList();
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

  Future<Map<String, dynamic>> syncSessions(List<Map<String, dynamic>> sessions) async {
    final res = await _dio.post('/v1/sync/sessions', data: {'sessions': sessions});
    return res.data as Map<String, dynamic>;
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

  Future<void> logHabit({required String clientHabitId, required String clientLocalId, required String logDate}) async {
    await _dio.post('/v1/habits/$clientHabitId/log', data: {
      'client_local_id': clientLocalId,
      'log_date': logDate,
      'value': 1,
      'is_completed': true,
    });
  }

  Future<List<PersonalRecordInfo>> records() async {
    final res = await _dio.get('/v1/progress/records');
    return (res.data as List).map((r) => PersonalRecordInfo.fromJson(r)).toList();
  }

  Future<List<SetLogPoint>> exerciseHistory(String exerciseId) async {
    final res = await _dio.get('/v1/progress/exercises/$exerciseId', queryParameters: {'days': 365});
    return (res.data as List).map((s) => SetLogPoint.fromJson(s)).toList();
  }
}
