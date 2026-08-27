import 'workout_store.dart';

/// Store en memoria para web (ver nota en workout_store.dart). Se pierde al
/// recargar la pagina: suficiente para previsualizar la UI, no para uso real.
class WebWorkoutStore implements WorkoutStore {
  final List<LocalSession> _sessions = [];

  @override
  Future<void> saveSession(LocalSession session) async {
    _sessions.removeWhere((s) => s.localId == session.localId);
    _sessions.add(session);
  }

  @override
  Future<List<LocalSession>> unsyncedSessions() async => _sessions.where((s) => !s.synced).toList();

  @override
  Future<void> markSynced(String sessionLocalId) async {
    final index = _sessions.indexWhere((s) => s.localId == sessionLocalId);
    if (index != -1) {
      _sessions[index] = _sessions[index].copyWith(synced: true);
    }
  }

  @override
  Future<List<LocalSession>> recentSessions({int limit = 20}) async {
    final sorted = [..._sessions]..sort((a, b) => b.startedAt.compareTo(a.startedAt));
    return sorted.take(limit).toList();
  }
}

WorkoutStore createWorkoutStore() => WebWorkoutStore();
