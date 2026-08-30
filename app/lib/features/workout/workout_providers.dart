import 'dart:async';

import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../core/network/models.dart';
import '../../core/providers.dart';
import '../../core/storage/workout_store.dart';

const _uuid = Uuid();

class SetEntry {
  final String localId;
  final int setNumber;
  double? weightKg;
  int? reps;
  double? rpe;
  bool isCompleted;
  DateTime? performedAt;

  SetEntry({required this.localId, required this.setNumber, this.weightKg, this.reps, this.rpe, this.isCompleted = false, this.performedAt});
}

class ExerciseEntry {
  final String exerciseId;
  final String name;
  final int orderIndex;
  final int restSeconds;
  final List<SetEntry> sets;

  ExerciseEntry({required this.exerciseId, required this.name, required this.orderIndex, required this.restSeconds, required this.sets});
}

class WorkoutSessionState {
  final String localId;
  final String name;
  final DateTime startedAt;
  final String? assignedWorkoutId;
  final List<ExerciseEntry> exercises;

  WorkoutSessionState({
    required this.localId,
    required this.name,
    required this.startedAt,
    required this.exercises,
    this.assignedWorkoutId,
  });
}

class ActiveWorkoutController extends StateNotifier<WorkoutSessionState?> {
  ActiveWorkoutController(this._ref) : super(null);

  final Ref _ref;

  void start({String name = 'Entrenamiento libre', AssignedWorkoutInfo? fromAssigned}) {
    final exercises = <ExerciseEntry>[];
    if (fromAssigned != null) {
      for (final e in fromAssigned.exercises) {
        exercises.add(ExerciseEntry(
          exerciseId: e.exerciseId,
          name: e.exerciseId, // se resuelve a nombre real via catalogo en la UI si hace falta
          orderIndex: e.orderIndex,
          restSeconds: e.restSeconds,
          sets: List.generate(
            e.targetSets,
            (i) => SetEntry(localId: _uuid.v4(), setNumber: i + 1),
          ),
        ));
      }
    }
    state = WorkoutSessionState(
      localId: _uuid.v4(),
      name: fromAssigned?.name ?? name,
      startedAt: DateTime.now(),
      assignedWorkoutId: fromAssigned?.id,
      exercises: exercises,
    );
  }

  void addExercise(String exerciseId, String name) {
    final current = state;
    if (current == null) return;
    current.exercises.add(ExerciseEntry(
      exerciseId: exerciseId,
      name: name,
      orderIndex: current.exercises.length + 1,
      restSeconds: 90,
      sets: [SetEntry(localId: _uuid.v4(), setNumber: 1)],
    ));
    state = WorkoutSessionState(
      localId: current.localId,
      name: current.name,
      startedAt: current.startedAt,
      assignedWorkoutId: current.assignedWorkoutId,
      exercises: [...current.exercises],
    );
  }

  void addSet(int exerciseIndex) {
    final current = state;
    if (current == null) return;
    final ex = current.exercises[exerciseIndex];
    ex.sets.add(SetEntry(localId: _uuid.v4(), setNumber: ex.sets.length + 1));
    state = WorkoutSessionState(
      localId: current.localId,
      name: current.name,
      startedAt: current.startedAt,
      assignedWorkoutId: current.assignedWorkoutId,
      exercises: [...current.exercises],
    );
  }

  void completeSet(int exerciseIndex, int setIndex, {double? weightKg, int? reps, double? rpe}) {
    final current = state;
    if (current == null) return;
    final set = current.exercises[exerciseIndex].sets[setIndex];
    set.weightKg = weightKg ?? set.weightKg;
    set.reps = reps ?? set.reps;
    set.rpe = rpe ?? set.rpe;
    set.isCompleted = true;
    set.performedAt = DateTime.now();
    state = WorkoutSessionState(
      localId: current.localId,
      name: current.name,
      startedAt: current.startedAt,
      assignedWorkoutId: current.assignedWorkoutId,
      exercises: [...current.exercises],
    );
    _ref.read(restTimerProvider.notifier).start(current.exercises[exerciseIndex].restSeconds);
  }

  Future<SyncResult?> finish() async {
    final current = state;
    if (current == null) return null;

    final localSession = LocalSession(
      localId: current.localId,
      name: current.name,
      status: 'completed',
      assignedWorkoutId: current.assignedWorkoutId,
      startedAt: current.startedAt,
      endedAt: DateTime.now(),
      exercises: current.exercises
          .map((e) => LocalSessionExercise(
                exerciseId: e.exerciseId,
                orderIndex: e.orderIndex,
                sets: e.sets
                    .where((s) => s.isCompleted)
                    .map((s) => LocalSet(
                          localId: s.localId,
                          exerciseId: e.exerciseId,
                          setNumber: s.setNumber,
                          type: 'working',
                          weightKg: s.weightKg,
                          reps: s.reps,
                          rpe: s.rpe,
                          isCompleted: true,
                          performedAt: s.performedAt ?? DateTime.now(),
                        ))
                    .toList(),
              ))
          .toList(),
    );

    await _ref.read(workoutStoreProvider).saveSession(localSession);
    state = null;
    _ref.read(restTimerProvider.notifier).cancel();
    return _ref.read(syncQueueProvider.notifier).syncNow();
  }

  void discard() {
    state = null;
    _ref.read(restTimerProvider.notifier).cancel();
  }
}

final activeWorkoutProvider = StateNotifierProvider<ActiveWorkoutController, WorkoutSessionState?>((ref) => ActiveWorkoutController(ref));

/// Temporizador de descanso: cuenta regresiva simple con vibracion (haptics
/// del SDK, sin dependencia extra) al terminar.
class RestTimerState {
  final int secondsLeft;
  final bool running;
  const RestTimerState({this.secondsLeft = 0, this.running = false});
}

class RestTimerController extends StateNotifier<RestTimerState> {
  RestTimerController() : super(const RestTimerState());
  Timer? _timer;

  void start(int seconds) {
    _timer?.cancel();
    state = RestTimerState(secondsLeft: seconds, running: true);
    _timer = Timer.periodic(const Duration(seconds: 1), (t) {
      final left = state.secondsLeft - 1;
      if (left <= 0) {
        t.cancel();
        state = const RestTimerState(secondsLeft: 0, running: false);
        HapticFeedback.mediumImpact();
      } else {
        state = RestTimerState(secondsLeft: left, running: true);
      }
    });
  }

  void cancel() {
    _timer?.cancel();
    state = const RestTimerState();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }
}

final restTimerProvider = StateNotifierProvider<RestTimerController, RestTimerState>((ref) => RestTimerController());

/// Sincroniza las sesiones locales pendientes con el servidor. Idempotente
/// por client_local_id: se puede llamar tantas veces como se quiera.
class SyncQueueController extends StateNotifier<bool> {
  SyncQueueController(this._ref) : super(false);
  final Ref _ref;

  Future<SyncResult?> syncNow() async {
    if (state) return null; // ya sincronizando
    state = true;
    try {
      final store = _ref.read(workoutStoreProvider);
      final pending = await store.unsyncedSessions();
      if (pending.isEmpty) return null;

      final api = _ref.read(apiRepositoryProvider);
      final payload = pending
          .map((s) => {
                'client_local_id': s.localId,
                'name': s.name,
                'status': s.status,
                'assigned_workout_id': s.assignedWorkoutId,
                'started_at': s.startedAt.toIso8601String(),
                'ended_at': s.endedAt?.toIso8601String(),
                'notes': s.notes,
                'exercises': s.exercises
                    .map((e) => {
                          'exercise_id': e.exerciseId,
                          'order_index': e.orderIndex,
                          'sets': e.sets
                              .map((set) => {
                                    'client_local_id': set.localId,
                                    'set_number': set.setNumber,
                                    'type': set.type,
                                    'weight_kg': set.weightKg,
                                    'reps': set.reps,
                                    'rpe': set.rpe,
                                    'is_completed': set.isCompleted,
                                    'performed_at': set.performedAt.toIso8601String(),
                                  })
                              .toList(),
                        })
                    .toList(),
              })
          .toList();

      final result = await api.syncSessions(payload);
      for (final s in pending) {
        await store.markSynced(s.localId);
      }
      return result;
    } catch (_) {
      // sin red: se reintenta al reabrir la app, al volver a foreground, o
      // cuando la UI lo dispare manualmente (ver pendingSyncCountProvider).
      return null;
    } finally {
      state = false;
    }
  }
}

final syncQueueProvider = StateNotifierProvider<SyncQueueController, bool>((ref) => SyncQueueController(ref));

/// Cuantas sesiones locales todavia no llegaron al servidor. La UI lo usa
/// para mostrar un aviso + boton de reintento manual.
final pendingSyncCountProvider = FutureProvider.autoDispose<int>((ref) async {
  final pending = await ref.watch(workoutStoreProvider).unsyncedSessions();
  return pending.length;
});
