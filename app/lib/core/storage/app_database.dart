import 'dart:io';

import 'package:drift/drift.dart';
import 'package:drift/native.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';

part 'app_database.g.dart';

@DataClassName('LocalSessionRow')
class LocalSessions extends Table {
  TextColumn get localId => text()();
  TextColumn get name => text()();
  TextColumn get status => text()();
  TextColumn get assignedWorkoutId => text().nullable()();
  DateTimeColumn get startedAt => dateTime()();
  DateTimeColumn get endedAt => dateTime().nullable()();
  IntColumn get perceivedEffort => integer().nullable()();
  IntColumn get mood => integer().nullable()();
  TextColumn get notes => text().nullable()();
  BoolColumn get synced => boolean().withDefault(const Constant(false))();

  @override
  Set<Column> get primaryKey => {localId};
}

@DataClassName('LocalSessionExerciseRow')
class LocalSessionExercises extends Table {
  IntColumn get id => integer().autoIncrement()();
  TextColumn get sessionLocalId => text().references(LocalSessions, #localId)();
  TextColumn get exerciseId => text()();
  IntColumn get orderIndex => integer()();
}

@DataClassName('LocalSetRow')
class LocalSets extends Table {
  TextColumn get localId => text()();
  IntColumn get sessionExerciseId => integer().references(LocalSessionExercises, #id)();
  TextColumn get exerciseId => text()();
  IntColumn get setNumber => integer()();
  TextColumn get type => text().withDefault(const Constant('working'))();
  RealColumn get weightKg => real().nullable()();
  IntColumn get reps => integer().nullable()();
  RealColumn get rpe => real().nullable()();
  BoolColumn get isCompleted => boolean().withDefault(const Constant(true))();
  DateTimeColumn get performedAt => dateTime()();

  @override
  Set<Column> get primaryKey => {localId};
}

@DriftDatabase(tables: [LocalSessions, LocalSessionExercises, LocalSets])
class AppDatabase extends _$AppDatabase {
  AppDatabase() : super(_openConnection());
  AppDatabase.forTesting(super.executor);

  @override
  int get schemaVersion => 1;

  static QueryExecutor _openConnection() {
    return LazyDatabase(() async {
      final dir = await getApplicationDocumentsDirectory();
      final file = File(p.join(dir.path, 'myvibesfit.sqlite'));
      return NativeDatabase.createInBackground(file);
    });
  }
}
