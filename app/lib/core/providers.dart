import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'network/api_client.dart';
import 'network/api_repository.dart';
import 'network/models.dart';
import 'network/token_storage.dart';
import 'storage/workout_store.dart';

final tokenStorageProvider = Provider<TokenStorage>((ref) => TokenStorage());

final apiClientProvider = Provider<ApiClient>((ref) => ApiClient(ref.watch(tokenStorageProvider)));

final apiRepositoryProvider = Provider<ApiRepository>(
  (ref) => ApiRepository(ref.watch(apiClientProvider), ref.watch(tokenStorageProvider)),
);

final workoutStoreProvider = Provider<WorkoutStore>((ref) => createWorkoutStore());

/// Catalogo de ejercicios compartido: los planes asignados solo traen
/// exercise_id, este provider es la unica fuente para resolver el nombre.
final exerciseCatalogProvider = FutureProvider.autoDispose<List<ExerciseSummary>>((ref) => ref.watch(apiRepositoryProvider).exercises());

String resolveExerciseName(List<ExerciseSummary> catalog, String exerciseId, {String? fallback}) {
  for (final e in catalog) {
    if (e.id == exerciseId) return e.name;
  }
  return fallback ?? exerciseId;
}
