import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'network/api_client.dart';
import 'network/api_repository.dart';
import 'network/token_storage.dart';
import 'storage/workout_store.dart';

final tokenStorageProvider = Provider<TokenStorage>((ref) => TokenStorage());

final apiClientProvider = Provider<ApiClient>((ref) => ApiClient(ref.watch(tokenStorageProvider)));

final apiRepositoryProvider = Provider<ApiRepository>(
  (ref) => ApiRepository(ref.watch(apiClientProvider), ref.watch(tokenStorageProvider)),
);

final workoutStoreProvider = Provider<WorkoutStore>((ref) => createWorkoutStore());
