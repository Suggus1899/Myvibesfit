import 'workout_store_web.dart' if (dart.library.io) 'workout_store_io.dart' as impl;

/// Set logueado localmente. client_local_id es la clave de idempotencia
/// hacia el servidor: reenviarlo nunca duplica filas alla.
class LocalSet {
  final String localId;
  final String exerciseId;
  final int setNumber;
  final String type;
  final double? weightKg;
  final int? reps;
  final double? rpe;
  final bool isCompleted;
  final DateTime performedAt;

  const LocalSet({
    required this.localId,
    required this.exerciseId,
    required this.setNumber,
    required this.type,
    required this.performedAt,
    this.weightKg,
    this.reps,
    this.rpe,
    this.isCompleted = true,
  });
}

class LocalSessionExercise {
  final String exerciseId;
  final int orderIndex;
  final List<LocalSet> sets;

  const LocalSessionExercise({required this.exerciseId, required this.orderIndex, required this.sets});
}

class LocalSession {
  final String localId;
  final String name;
  final String status;
  final String? assignedWorkoutId;
  final DateTime startedAt;
  final DateTime? endedAt;
  final int? perceivedEffort;
  final int? mood;
  final String? notes;
  final bool synced;
  final List<LocalSessionExercise> exercises;

  const LocalSession({
    required this.localId,
    required this.name,
    required this.status,
    required this.startedAt,
    required this.exercises,
    this.assignedWorkoutId,
    this.endedAt,
    this.perceivedEffort,
    this.mood,
    this.notes,
    this.synced = false,
  });

  LocalSession copyWith({bool? synced}) => LocalSession(
        localId: localId,
        name: name,
        status: status,
        assignedWorkoutId: assignedWorkoutId,
        startedAt: startedAt,
        endedAt: endedAt,
        perceivedEffort: perceivedEffort,
        mood: mood,
        notes: notes,
        synced: synced ?? this.synced,
        exercises: exercises,
      );
}

/// Persistencia local de entrenamientos. La implementacion real (Drift +
/// SQLite) vive en workout_store_io.dart para plataformas nativas; la web
/// usa un store en memoria (workout_store_web.dart) porque SQLite-WASM
/// requiere assets adicionales que no aportan nada a esta vista previa.
/// ponytail: si el panel/demo web necesita persistir de verdad, migrar a
/// drift_flutter con sqlite3.wasm.
abstract class WorkoutStore {
  Future<void> saveSession(LocalSession session);
  Future<List<LocalSession>> unsyncedSessions();
  Future<void> markSynced(String sessionLocalId);
  Future<List<LocalSession>> recentSessions({int limit = 20});
}

WorkoutStore createWorkoutStore() => impl.createWorkoutStore();
