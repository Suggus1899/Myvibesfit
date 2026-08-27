import 'package:drift/drift.dart';

import 'app_database.dart';
import 'workout_store.dart';

class IoWorkoutStore implements WorkoutStore {
  IoWorkoutStore() : _db = AppDatabase();

  final AppDatabase _db;

  @override
  Future<void> saveSession(LocalSession session) async {
    await _db.transaction(() async {
      await _db.into(_db.localSessions).insertOnConflictUpdate(LocalSessionsCompanion(
            localId: Value(session.localId),
            name: Value(session.name),
            status: Value(session.status),
            assignedWorkoutId: Value(session.assignedWorkoutId),
            startedAt: Value(session.startedAt),
            endedAt: Value(session.endedAt),
            perceivedEffort: Value(session.perceivedEffort),
            mood: Value(session.mood),
            notes: Value(session.notes),
            synced: Value(session.synced),
          ));

      for (final ex in session.exercises) {
        final exerciseRowId = await _db.into(_db.localSessionExercises).insert(LocalSessionExercisesCompanion.insert(
              sessionLocalId: session.localId,
              exerciseId: ex.exerciseId,
              orderIndex: ex.orderIndex,
            ));

        for (final set in ex.sets) {
          await _db.into(_db.localSets).insertOnConflictUpdate(LocalSetsCompanion(
                localId: Value(set.localId),
                sessionExerciseId: Value(exerciseRowId),
                exerciseId: Value(set.exerciseId),
                setNumber: Value(set.setNumber),
                type: Value(set.type),
                weightKg: Value(set.weightKg),
                reps: Value(set.reps),
                rpe: Value(set.rpe),
                isCompleted: Value(set.isCompleted),
                performedAt: Value(set.performedAt),
              ));
        }
      }
    });
  }

  @override
  Future<List<LocalSession>> unsyncedSessions() async {
    final rows = await (_db.select(_db.localSessions)..where((s) => s.synced.equals(false))).get();
    return Future.wait(rows.map(_hydrate));
  }

  @override
  Future<void> markSynced(String sessionLocalId) async {
    await (_db.update(_db.localSessions)..where((s) => s.localId.equals(sessionLocalId)))
        .write(const LocalSessionsCompanion(synced: Value(true)));
  }

  @override
  Future<List<LocalSession>> recentSessions({int limit = 20}) async {
    final rows = await (_db.select(_db.localSessions)
          ..orderBy([(s) => OrderingTerm.desc(s.startedAt)])
          ..limit(limit))
        .get();
    return Future.wait(rows.map(_hydrate));
  }

  Future<LocalSession> _hydrate(LocalSessionRow row) async {
    final exerciseRows = await (_db.select(_db.localSessionExercises)..where((e) => e.sessionLocalId.equals(row.localId))).get();
    final exercises = <LocalSessionExercise>[];
    for (final exRow in exerciseRows) {
      final setRows = await (_db.select(_db.localSets)..where((s) => s.sessionExerciseId.equals(exRow.id))).get();
      exercises.add(LocalSessionExercise(
        exerciseId: exRow.exerciseId,
        orderIndex: exRow.orderIndex,
        sets: setRows
            .map((s) => LocalSet(
                  localId: s.localId,
                  exerciseId: s.exerciseId,
                  setNumber: s.setNumber,
                  type: s.type,
                  weightKg: s.weightKg,
                  reps: s.reps,
                  rpe: s.rpe,
                  isCompleted: s.isCompleted,
                  performedAt: s.performedAt,
                ))
            .toList(),
      ));
    }
    return LocalSession(
      localId: row.localId,
      name: row.name,
      status: row.status,
      assignedWorkoutId: row.assignedWorkoutId,
      startedAt: row.startedAt,
      endedAt: row.endedAt,
      perceivedEffort: row.perceivedEffort,
      mood: row.mood,
      notes: row.notes,
      synced: row.synced,
      exercises: exercises,
    );
  }
}

WorkoutStore createWorkoutStore() => IoWorkoutStore();
