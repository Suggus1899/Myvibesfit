// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'app_database.dart';

// ignore_for_file: type=lint
class $LocalSessionsTable extends LocalSessions
    with TableInfo<$LocalSessionsTable, LocalSessionRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $LocalSessionsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _localIdMeta = const VerificationMeta(
    'localId',
  );
  @override
  late final GeneratedColumn<String> localId = GeneratedColumn<String>(
    'local_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _nameMeta = const VerificationMeta('name');
  @override
  late final GeneratedColumn<String> name = GeneratedColumn<String>(
    'name',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _statusMeta = const VerificationMeta('status');
  @override
  late final GeneratedColumn<String> status = GeneratedColumn<String>(
    'status',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _assignedWorkoutIdMeta = const VerificationMeta(
    'assignedWorkoutId',
  );
  @override
  late final GeneratedColumn<String> assignedWorkoutId =
      GeneratedColumn<String>(
        'assigned_workout_id',
        aliasedName,
        true,
        type: DriftSqlType.string,
        requiredDuringInsert: false,
      );
  static const VerificationMeta _startedAtMeta = const VerificationMeta(
    'startedAt',
  );
  @override
  late final GeneratedColumn<DateTime> startedAt = GeneratedColumn<DateTime>(
    'started_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _endedAtMeta = const VerificationMeta(
    'endedAt',
  );
  @override
  late final GeneratedColumn<DateTime> endedAt = GeneratedColumn<DateTime>(
    'ended_at',
    aliasedName,
    true,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _perceivedEffortMeta = const VerificationMeta(
    'perceivedEffort',
  );
  @override
  late final GeneratedColumn<int> perceivedEffort = GeneratedColumn<int>(
    'perceived_effort',
    aliasedName,
    true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _moodMeta = const VerificationMeta('mood');
  @override
  late final GeneratedColumn<int> mood = GeneratedColumn<int>(
    'mood',
    aliasedName,
    true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _notesMeta = const VerificationMeta('notes');
  @override
  late final GeneratedColumn<String> notes = GeneratedColumn<String>(
    'notes',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _syncedMeta = const VerificationMeta('synced');
  @override
  late final GeneratedColumn<bool> synced = GeneratedColumn<bool>(
    'synced',
    aliasedName,
    false,
    type: DriftSqlType.bool,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'CHECK ("synced" IN (0, 1))',
    ),
    defaultValue: const Constant(false),
  );
  @override
  List<GeneratedColumn> get $columns => [
    localId,
    name,
    status,
    assignedWorkoutId,
    startedAt,
    endedAt,
    perceivedEffort,
    mood,
    notes,
    synced,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'local_sessions';
  @override
  VerificationContext validateIntegrity(
    Insertable<LocalSessionRow> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('local_id')) {
      context.handle(
        _localIdMeta,
        localId.isAcceptableOrUnknown(data['local_id']!, _localIdMeta),
      );
    } else if (isInserting) {
      context.missing(_localIdMeta);
    }
    if (data.containsKey('name')) {
      context.handle(
        _nameMeta,
        name.isAcceptableOrUnknown(data['name']!, _nameMeta),
      );
    } else if (isInserting) {
      context.missing(_nameMeta);
    }
    if (data.containsKey('status')) {
      context.handle(
        _statusMeta,
        status.isAcceptableOrUnknown(data['status']!, _statusMeta),
      );
    } else if (isInserting) {
      context.missing(_statusMeta);
    }
    if (data.containsKey('assigned_workout_id')) {
      context.handle(
        _assignedWorkoutIdMeta,
        assignedWorkoutId.isAcceptableOrUnknown(
          data['assigned_workout_id']!,
          _assignedWorkoutIdMeta,
        ),
      );
    }
    if (data.containsKey('started_at')) {
      context.handle(
        _startedAtMeta,
        startedAt.isAcceptableOrUnknown(data['started_at']!, _startedAtMeta),
      );
    } else if (isInserting) {
      context.missing(_startedAtMeta);
    }
    if (data.containsKey('ended_at')) {
      context.handle(
        _endedAtMeta,
        endedAt.isAcceptableOrUnknown(data['ended_at']!, _endedAtMeta),
      );
    }
    if (data.containsKey('perceived_effort')) {
      context.handle(
        _perceivedEffortMeta,
        perceivedEffort.isAcceptableOrUnknown(
          data['perceived_effort']!,
          _perceivedEffortMeta,
        ),
      );
    }
    if (data.containsKey('mood')) {
      context.handle(
        _moodMeta,
        mood.isAcceptableOrUnknown(data['mood']!, _moodMeta),
      );
    }
    if (data.containsKey('notes')) {
      context.handle(
        _notesMeta,
        notes.isAcceptableOrUnknown(data['notes']!, _notesMeta),
      );
    }
    if (data.containsKey('synced')) {
      context.handle(
        _syncedMeta,
        synced.isAcceptableOrUnknown(data['synced']!, _syncedMeta),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {localId};
  @override
  LocalSessionRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return LocalSessionRow(
      localId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}local_id'],
      )!,
      name: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}name'],
      )!,
      status: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}status'],
      )!,
      assignedWorkoutId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}assigned_workout_id'],
      ),
      startedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}started_at'],
      )!,
      endedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}ended_at'],
      ),
      perceivedEffort: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}perceived_effort'],
      ),
      mood: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}mood'],
      ),
      notes: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}notes'],
      ),
      synced: attachedDatabase.typeMapping.read(
        DriftSqlType.bool,
        data['${effectivePrefix}synced'],
      )!,
    );
  }

  @override
  $LocalSessionsTable createAlias(String alias) {
    return $LocalSessionsTable(attachedDatabase, alias);
  }
}

class LocalSessionRow extends DataClass implements Insertable<LocalSessionRow> {
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
  const LocalSessionRow({
    required this.localId,
    required this.name,
    required this.status,
    this.assignedWorkoutId,
    required this.startedAt,
    this.endedAt,
    this.perceivedEffort,
    this.mood,
    this.notes,
    required this.synced,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['local_id'] = Variable<String>(localId);
    map['name'] = Variable<String>(name);
    map['status'] = Variable<String>(status);
    if (!nullToAbsent || assignedWorkoutId != null) {
      map['assigned_workout_id'] = Variable<String>(assignedWorkoutId);
    }
    map['started_at'] = Variable<DateTime>(startedAt);
    if (!nullToAbsent || endedAt != null) {
      map['ended_at'] = Variable<DateTime>(endedAt);
    }
    if (!nullToAbsent || perceivedEffort != null) {
      map['perceived_effort'] = Variable<int>(perceivedEffort);
    }
    if (!nullToAbsent || mood != null) {
      map['mood'] = Variable<int>(mood);
    }
    if (!nullToAbsent || notes != null) {
      map['notes'] = Variable<String>(notes);
    }
    map['synced'] = Variable<bool>(synced);
    return map;
  }

  LocalSessionsCompanion toCompanion(bool nullToAbsent) {
    return LocalSessionsCompanion(
      localId: Value(localId),
      name: Value(name),
      status: Value(status),
      assignedWorkoutId: assignedWorkoutId == null && nullToAbsent
          ? const Value.absent()
          : Value(assignedWorkoutId),
      startedAt: Value(startedAt),
      endedAt: endedAt == null && nullToAbsent
          ? const Value.absent()
          : Value(endedAt),
      perceivedEffort: perceivedEffort == null && nullToAbsent
          ? const Value.absent()
          : Value(perceivedEffort),
      mood: mood == null && nullToAbsent ? const Value.absent() : Value(mood),
      notes: notes == null && nullToAbsent
          ? const Value.absent()
          : Value(notes),
      synced: Value(synced),
    );
  }

  factory LocalSessionRow.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return LocalSessionRow(
      localId: serializer.fromJson<String>(json['localId']),
      name: serializer.fromJson<String>(json['name']),
      status: serializer.fromJson<String>(json['status']),
      assignedWorkoutId: serializer.fromJson<String?>(
        json['assignedWorkoutId'],
      ),
      startedAt: serializer.fromJson<DateTime>(json['startedAt']),
      endedAt: serializer.fromJson<DateTime?>(json['endedAt']),
      perceivedEffort: serializer.fromJson<int?>(json['perceivedEffort']),
      mood: serializer.fromJson<int?>(json['mood']),
      notes: serializer.fromJson<String?>(json['notes']),
      synced: serializer.fromJson<bool>(json['synced']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'localId': serializer.toJson<String>(localId),
      'name': serializer.toJson<String>(name),
      'status': serializer.toJson<String>(status),
      'assignedWorkoutId': serializer.toJson<String?>(assignedWorkoutId),
      'startedAt': serializer.toJson<DateTime>(startedAt),
      'endedAt': serializer.toJson<DateTime?>(endedAt),
      'perceivedEffort': serializer.toJson<int?>(perceivedEffort),
      'mood': serializer.toJson<int?>(mood),
      'notes': serializer.toJson<String?>(notes),
      'synced': serializer.toJson<bool>(synced),
    };
  }

  LocalSessionRow copyWith({
    String? localId,
    String? name,
    String? status,
    Value<String?> assignedWorkoutId = const Value.absent(),
    DateTime? startedAt,
    Value<DateTime?> endedAt = const Value.absent(),
    Value<int?> perceivedEffort = const Value.absent(),
    Value<int?> mood = const Value.absent(),
    Value<String?> notes = const Value.absent(),
    bool? synced,
  }) => LocalSessionRow(
    localId: localId ?? this.localId,
    name: name ?? this.name,
    status: status ?? this.status,
    assignedWorkoutId: assignedWorkoutId.present
        ? assignedWorkoutId.value
        : this.assignedWorkoutId,
    startedAt: startedAt ?? this.startedAt,
    endedAt: endedAt.present ? endedAt.value : this.endedAt,
    perceivedEffort: perceivedEffort.present
        ? perceivedEffort.value
        : this.perceivedEffort,
    mood: mood.present ? mood.value : this.mood,
    notes: notes.present ? notes.value : this.notes,
    synced: synced ?? this.synced,
  );
  LocalSessionRow copyWithCompanion(LocalSessionsCompanion data) {
    return LocalSessionRow(
      localId: data.localId.present ? data.localId.value : this.localId,
      name: data.name.present ? data.name.value : this.name,
      status: data.status.present ? data.status.value : this.status,
      assignedWorkoutId: data.assignedWorkoutId.present
          ? data.assignedWorkoutId.value
          : this.assignedWorkoutId,
      startedAt: data.startedAt.present ? data.startedAt.value : this.startedAt,
      endedAt: data.endedAt.present ? data.endedAt.value : this.endedAt,
      perceivedEffort: data.perceivedEffort.present
          ? data.perceivedEffort.value
          : this.perceivedEffort,
      mood: data.mood.present ? data.mood.value : this.mood,
      notes: data.notes.present ? data.notes.value : this.notes,
      synced: data.synced.present ? data.synced.value : this.synced,
    );
  }

  @override
  String toString() {
    return (StringBuffer('LocalSessionRow(')
          ..write('localId: $localId, ')
          ..write('name: $name, ')
          ..write('status: $status, ')
          ..write('assignedWorkoutId: $assignedWorkoutId, ')
          ..write('startedAt: $startedAt, ')
          ..write('endedAt: $endedAt, ')
          ..write('perceivedEffort: $perceivedEffort, ')
          ..write('mood: $mood, ')
          ..write('notes: $notes, ')
          ..write('synced: $synced')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
    localId,
    name,
    status,
    assignedWorkoutId,
    startedAt,
    endedAt,
    perceivedEffort,
    mood,
    notes,
    synced,
  );
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is LocalSessionRow &&
          other.localId == this.localId &&
          other.name == this.name &&
          other.status == this.status &&
          other.assignedWorkoutId == this.assignedWorkoutId &&
          other.startedAt == this.startedAt &&
          other.endedAt == this.endedAt &&
          other.perceivedEffort == this.perceivedEffort &&
          other.mood == this.mood &&
          other.notes == this.notes &&
          other.synced == this.synced);
}

class LocalSessionsCompanion extends UpdateCompanion<LocalSessionRow> {
  final Value<String> localId;
  final Value<String> name;
  final Value<String> status;
  final Value<String?> assignedWorkoutId;
  final Value<DateTime> startedAt;
  final Value<DateTime?> endedAt;
  final Value<int?> perceivedEffort;
  final Value<int?> mood;
  final Value<String?> notes;
  final Value<bool> synced;
  final Value<int> rowid;
  const LocalSessionsCompanion({
    this.localId = const Value.absent(),
    this.name = const Value.absent(),
    this.status = const Value.absent(),
    this.assignedWorkoutId = const Value.absent(),
    this.startedAt = const Value.absent(),
    this.endedAt = const Value.absent(),
    this.perceivedEffort = const Value.absent(),
    this.mood = const Value.absent(),
    this.notes = const Value.absent(),
    this.synced = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  LocalSessionsCompanion.insert({
    required String localId,
    required String name,
    required String status,
    this.assignedWorkoutId = const Value.absent(),
    required DateTime startedAt,
    this.endedAt = const Value.absent(),
    this.perceivedEffort = const Value.absent(),
    this.mood = const Value.absent(),
    this.notes = const Value.absent(),
    this.synced = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : localId = Value(localId),
       name = Value(name),
       status = Value(status),
       startedAt = Value(startedAt);
  static Insertable<LocalSessionRow> custom({
    Expression<String>? localId,
    Expression<String>? name,
    Expression<String>? status,
    Expression<String>? assignedWorkoutId,
    Expression<DateTime>? startedAt,
    Expression<DateTime>? endedAt,
    Expression<int>? perceivedEffort,
    Expression<int>? mood,
    Expression<String>? notes,
    Expression<bool>? synced,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (localId != null) 'local_id': localId,
      if (name != null) 'name': name,
      if (status != null) 'status': status,
      if (assignedWorkoutId != null) 'assigned_workout_id': assignedWorkoutId,
      if (startedAt != null) 'started_at': startedAt,
      if (endedAt != null) 'ended_at': endedAt,
      if (perceivedEffort != null) 'perceived_effort': perceivedEffort,
      if (mood != null) 'mood': mood,
      if (notes != null) 'notes': notes,
      if (synced != null) 'synced': synced,
      if (rowid != null) 'rowid': rowid,
    });
  }

  LocalSessionsCompanion copyWith({
    Value<String>? localId,
    Value<String>? name,
    Value<String>? status,
    Value<String?>? assignedWorkoutId,
    Value<DateTime>? startedAt,
    Value<DateTime?>? endedAt,
    Value<int?>? perceivedEffort,
    Value<int?>? mood,
    Value<String?>? notes,
    Value<bool>? synced,
    Value<int>? rowid,
  }) {
    return LocalSessionsCompanion(
      localId: localId ?? this.localId,
      name: name ?? this.name,
      status: status ?? this.status,
      assignedWorkoutId: assignedWorkoutId ?? this.assignedWorkoutId,
      startedAt: startedAt ?? this.startedAt,
      endedAt: endedAt ?? this.endedAt,
      perceivedEffort: perceivedEffort ?? this.perceivedEffort,
      mood: mood ?? this.mood,
      notes: notes ?? this.notes,
      synced: synced ?? this.synced,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (localId.present) {
      map['local_id'] = Variable<String>(localId.value);
    }
    if (name.present) {
      map['name'] = Variable<String>(name.value);
    }
    if (status.present) {
      map['status'] = Variable<String>(status.value);
    }
    if (assignedWorkoutId.present) {
      map['assigned_workout_id'] = Variable<String>(assignedWorkoutId.value);
    }
    if (startedAt.present) {
      map['started_at'] = Variable<DateTime>(startedAt.value);
    }
    if (endedAt.present) {
      map['ended_at'] = Variable<DateTime>(endedAt.value);
    }
    if (perceivedEffort.present) {
      map['perceived_effort'] = Variable<int>(perceivedEffort.value);
    }
    if (mood.present) {
      map['mood'] = Variable<int>(mood.value);
    }
    if (notes.present) {
      map['notes'] = Variable<String>(notes.value);
    }
    if (synced.present) {
      map['synced'] = Variable<bool>(synced.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('LocalSessionsCompanion(')
          ..write('localId: $localId, ')
          ..write('name: $name, ')
          ..write('status: $status, ')
          ..write('assignedWorkoutId: $assignedWorkoutId, ')
          ..write('startedAt: $startedAt, ')
          ..write('endedAt: $endedAt, ')
          ..write('perceivedEffort: $perceivedEffort, ')
          ..write('mood: $mood, ')
          ..write('notes: $notes, ')
          ..write('synced: $synced, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $LocalSessionExercisesTable extends LocalSessionExercises
    with TableInfo<$LocalSessionExercisesTable, LocalSessionExerciseRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $LocalSessionExercisesTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<int> id = GeneratedColumn<int>(
    'id',
    aliasedName,
    false,
    hasAutoIncrement: true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'PRIMARY KEY AUTOINCREMENT',
    ),
  );
  static const VerificationMeta _sessionLocalIdMeta = const VerificationMeta(
    'sessionLocalId',
  );
  @override
  late final GeneratedColumn<String> sessionLocalId = GeneratedColumn<String>(
    'session_local_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'REFERENCES local_sessions (local_id)',
    ),
  );
  static const VerificationMeta _exerciseIdMeta = const VerificationMeta(
    'exerciseId',
  );
  @override
  late final GeneratedColumn<String> exerciseId = GeneratedColumn<String>(
    'exercise_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _orderIndexMeta = const VerificationMeta(
    'orderIndex',
  );
  @override
  late final GeneratedColumn<int> orderIndex = GeneratedColumn<int>(
    'order_index',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  @override
  List<GeneratedColumn> get $columns => [
    id,
    sessionLocalId,
    exerciseId,
    orderIndex,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'local_session_exercises';
  @override
  VerificationContext validateIntegrity(
    Insertable<LocalSessionExerciseRow> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    }
    if (data.containsKey('session_local_id')) {
      context.handle(
        _sessionLocalIdMeta,
        sessionLocalId.isAcceptableOrUnknown(
          data['session_local_id']!,
          _sessionLocalIdMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_sessionLocalIdMeta);
    }
    if (data.containsKey('exercise_id')) {
      context.handle(
        _exerciseIdMeta,
        exerciseId.isAcceptableOrUnknown(data['exercise_id']!, _exerciseIdMeta),
      );
    } else if (isInserting) {
      context.missing(_exerciseIdMeta);
    }
    if (data.containsKey('order_index')) {
      context.handle(
        _orderIndexMeta,
        orderIndex.isAcceptableOrUnknown(data['order_index']!, _orderIndexMeta),
      );
    } else if (isInserting) {
      context.missing(_orderIndexMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  LocalSessionExerciseRow map(
    Map<String, dynamic> data, {
    String? tablePrefix,
  }) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return LocalSessionExerciseRow(
      id: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}id'],
      )!,
      sessionLocalId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}session_local_id'],
      )!,
      exerciseId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}exercise_id'],
      )!,
      orderIndex: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}order_index'],
      )!,
    );
  }

  @override
  $LocalSessionExercisesTable createAlias(String alias) {
    return $LocalSessionExercisesTable(attachedDatabase, alias);
  }
}

class LocalSessionExerciseRow extends DataClass
    implements Insertable<LocalSessionExerciseRow> {
  final int id;
  final String sessionLocalId;
  final String exerciseId;
  final int orderIndex;
  const LocalSessionExerciseRow({
    required this.id,
    required this.sessionLocalId,
    required this.exerciseId,
    required this.orderIndex,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<int>(id);
    map['session_local_id'] = Variable<String>(sessionLocalId);
    map['exercise_id'] = Variable<String>(exerciseId);
    map['order_index'] = Variable<int>(orderIndex);
    return map;
  }

  LocalSessionExercisesCompanion toCompanion(bool nullToAbsent) {
    return LocalSessionExercisesCompanion(
      id: Value(id),
      sessionLocalId: Value(sessionLocalId),
      exerciseId: Value(exerciseId),
      orderIndex: Value(orderIndex),
    );
  }

  factory LocalSessionExerciseRow.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return LocalSessionExerciseRow(
      id: serializer.fromJson<int>(json['id']),
      sessionLocalId: serializer.fromJson<String>(json['sessionLocalId']),
      exerciseId: serializer.fromJson<String>(json['exerciseId']),
      orderIndex: serializer.fromJson<int>(json['orderIndex']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<int>(id),
      'sessionLocalId': serializer.toJson<String>(sessionLocalId),
      'exerciseId': serializer.toJson<String>(exerciseId),
      'orderIndex': serializer.toJson<int>(orderIndex),
    };
  }

  LocalSessionExerciseRow copyWith({
    int? id,
    String? sessionLocalId,
    String? exerciseId,
    int? orderIndex,
  }) => LocalSessionExerciseRow(
    id: id ?? this.id,
    sessionLocalId: sessionLocalId ?? this.sessionLocalId,
    exerciseId: exerciseId ?? this.exerciseId,
    orderIndex: orderIndex ?? this.orderIndex,
  );
  LocalSessionExerciseRow copyWithCompanion(
    LocalSessionExercisesCompanion data,
  ) {
    return LocalSessionExerciseRow(
      id: data.id.present ? data.id.value : this.id,
      sessionLocalId: data.sessionLocalId.present
          ? data.sessionLocalId.value
          : this.sessionLocalId,
      exerciseId: data.exerciseId.present
          ? data.exerciseId.value
          : this.exerciseId,
      orderIndex: data.orderIndex.present
          ? data.orderIndex.value
          : this.orderIndex,
    );
  }

  @override
  String toString() {
    return (StringBuffer('LocalSessionExerciseRow(')
          ..write('id: $id, ')
          ..write('sessionLocalId: $sessionLocalId, ')
          ..write('exerciseId: $exerciseId, ')
          ..write('orderIndex: $orderIndex')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(id, sessionLocalId, exerciseId, orderIndex);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is LocalSessionExerciseRow &&
          other.id == this.id &&
          other.sessionLocalId == this.sessionLocalId &&
          other.exerciseId == this.exerciseId &&
          other.orderIndex == this.orderIndex);
}

class LocalSessionExercisesCompanion
    extends UpdateCompanion<LocalSessionExerciseRow> {
  final Value<int> id;
  final Value<String> sessionLocalId;
  final Value<String> exerciseId;
  final Value<int> orderIndex;
  const LocalSessionExercisesCompanion({
    this.id = const Value.absent(),
    this.sessionLocalId = const Value.absent(),
    this.exerciseId = const Value.absent(),
    this.orderIndex = const Value.absent(),
  });
  LocalSessionExercisesCompanion.insert({
    this.id = const Value.absent(),
    required String sessionLocalId,
    required String exerciseId,
    required int orderIndex,
  }) : sessionLocalId = Value(sessionLocalId),
       exerciseId = Value(exerciseId),
       orderIndex = Value(orderIndex);
  static Insertable<LocalSessionExerciseRow> custom({
    Expression<int>? id,
    Expression<String>? sessionLocalId,
    Expression<String>? exerciseId,
    Expression<int>? orderIndex,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (sessionLocalId != null) 'session_local_id': sessionLocalId,
      if (exerciseId != null) 'exercise_id': exerciseId,
      if (orderIndex != null) 'order_index': orderIndex,
    });
  }

  LocalSessionExercisesCompanion copyWith({
    Value<int>? id,
    Value<String>? sessionLocalId,
    Value<String>? exerciseId,
    Value<int>? orderIndex,
  }) {
    return LocalSessionExercisesCompanion(
      id: id ?? this.id,
      sessionLocalId: sessionLocalId ?? this.sessionLocalId,
      exerciseId: exerciseId ?? this.exerciseId,
      orderIndex: orderIndex ?? this.orderIndex,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<int>(id.value);
    }
    if (sessionLocalId.present) {
      map['session_local_id'] = Variable<String>(sessionLocalId.value);
    }
    if (exerciseId.present) {
      map['exercise_id'] = Variable<String>(exerciseId.value);
    }
    if (orderIndex.present) {
      map['order_index'] = Variable<int>(orderIndex.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('LocalSessionExercisesCompanion(')
          ..write('id: $id, ')
          ..write('sessionLocalId: $sessionLocalId, ')
          ..write('exerciseId: $exerciseId, ')
          ..write('orderIndex: $orderIndex')
          ..write(')'))
        .toString();
  }
}

class $LocalSetsTable extends LocalSets
    with TableInfo<$LocalSetsTable, LocalSetRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $LocalSetsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _localIdMeta = const VerificationMeta(
    'localId',
  );
  @override
  late final GeneratedColumn<String> localId = GeneratedColumn<String>(
    'local_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _sessionExerciseIdMeta = const VerificationMeta(
    'sessionExerciseId',
  );
  @override
  late final GeneratedColumn<int> sessionExerciseId = GeneratedColumn<int>(
    'session_exercise_id',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'REFERENCES local_session_exercises (id)',
    ),
  );
  static const VerificationMeta _exerciseIdMeta = const VerificationMeta(
    'exerciseId',
  );
  @override
  late final GeneratedColumn<String> exerciseId = GeneratedColumn<String>(
    'exercise_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _setNumberMeta = const VerificationMeta(
    'setNumber',
  );
  @override
  late final GeneratedColumn<int> setNumber = GeneratedColumn<int>(
    'set_number',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _typeMeta = const VerificationMeta('type');
  @override
  late final GeneratedColumn<String> type = GeneratedColumn<String>(
    'type',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
    defaultValue: const Constant('working'),
  );
  static const VerificationMeta _weightKgMeta = const VerificationMeta(
    'weightKg',
  );
  @override
  late final GeneratedColumn<double> weightKg = GeneratedColumn<double>(
    'weight_kg',
    aliasedName,
    true,
    type: DriftSqlType.double,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _repsMeta = const VerificationMeta('reps');
  @override
  late final GeneratedColumn<int> reps = GeneratedColumn<int>(
    'reps',
    aliasedName,
    true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _rpeMeta = const VerificationMeta('rpe');
  @override
  late final GeneratedColumn<double> rpe = GeneratedColumn<double>(
    'rpe',
    aliasedName,
    true,
    type: DriftSqlType.double,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _isCompletedMeta = const VerificationMeta(
    'isCompleted',
  );
  @override
  late final GeneratedColumn<bool> isCompleted = GeneratedColumn<bool>(
    'is_completed',
    aliasedName,
    false,
    type: DriftSqlType.bool,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways(
      'CHECK ("is_completed" IN (0, 1))',
    ),
    defaultValue: const Constant(true),
  );
  static const VerificationMeta _performedAtMeta = const VerificationMeta(
    'performedAt',
  );
  @override
  late final GeneratedColumn<DateTime> performedAt = GeneratedColumn<DateTime>(
    'performed_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: true,
  );
  @override
  List<GeneratedColumn> get $columns => [
    localId,
    sessionExerciseId,
    exerciseId,
    setNumber,
    type,
    weightKg,
    reps,
    rpe,
    isCompleted,
    performedAt,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'local_sets';
  @override
  VerificationContext validateIntegrity(
    Insertable<LocalSetRow> instance, {
    bool isInserting = false,
  }) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('local_id')) {
      context.handle(
        _localIdMeta,
        localId.isAcceptableOrUnknown(data['local_id']!, _localIdMeta),
      );
    } else if (isInserting) {
      context.missing(_localIdMeta);
    }
    if (data.containsKey('session_exercise_id')) {
      context.handle(
        _sessionExerciseIdMeta,
        sessionExerciseId.isAcceptableOrUnknown(
          data['session_exercise_id']!,
          _sessionExerciseIdMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_sessionExerciseIdMeta);
    }
    if (data.containsKey('exercise_id')) {
      context.handle(
        _exerciseIdMeta,
        exerciseId.isAcceptableOrUnknown(data['exercise_id']!, _exerciseIdMeta),
      );
    } else if (isInserting) {
      context.missing(_exerciseIdMeta);
    }
    if (data.containsKey('set_number')) {
      context.handle(
        _setNumberMeta,
        setNumber.isAcceptableOrUnknown(data['set_number']!, _setNumberMeta),
      );
    } else if (isInserting) {
      context.missing(_setNumberMeta);
    }
    if (data.containsKey('type')) {
      context.handle(
        _typeMeta,
        type.isAcceptableOrUnknown(data['type']!, _typeMeta),
      );
    }
    if (data.containsKey('weight_kg')) {
      context.handle(
        _weightKgMeta,
        weightKg.isAcceptableOrUnknown(data['weight_kg']!, _weightKgMeta),
      );
    }
    if (data.containsKey('reps')) {
      context.handle(
        _repsMeta,
        reps.isAcceptableOrUnknown(data['reps']!, _repsMeta),
      );
    }
    if (data.containsKey('rpe')) {
      context.handle(
        _rpeMeta,
        rpe.isAcceptableOrUnknown(data['rpe']!, _rpeMeta),
      );
    }
    if (data.containsKey('is_completed')) {
      context.handle(
        _isCompletedMeta,
        isCompleted.isAcceptableOrUnknown(
          data['is_completed']!,
          _isCompletedMeta,
        ),
      );
    }
    if (data.containsKey('performed_at')) {
      context.handle(
        _performedAtMeta,
        performedAt.isAcceptableOrUnknown(
          data['performed_at']!,
          _performedAtMeta,
        ),
      );
    } else if (isInserting) {
      context.missing(_performedAtMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {localId};
  @override
  LocalSetRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return LocalSetRow(
      localId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}local_id'],
      )!,
      sessionExerciseId: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}session_exercise_id'],
      )!,
      exerciseId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}exercise_id'],
      )!,
      setNumber: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}set_number'],
      )!,
      type: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}type'],
      )!,
      weightKg: attachedDatabase.typeMapping.read(
        DriftSqlType.double,
        data['${effectivePrefix}weight_kg'],
      ),
      reps: attachedDatabase.typeMapping.read(
        DriftSqlType.int,
        data['${effectivePrefix}reps'],
      ),
      rpe: attachedDatabase.typeMapping.read(
        DriftSqlType.double,
        data['${effectivePrefix}rpe'],
      ),
      isCompleted: attachedDatabase.typeMapping.read(
        DriftSqlType.bool,
        data['${effectivePrefix}is_completed'],
      )!,
      performedAt: attachedDatabase.typeMapping.read(
        DriftSqlType.dateTime,
        data['${effectivePrefix}performed_at'],
      )!,
    );
  }

  @override
  $LocalSetsTable createAlias(String alias) {
    return $LocalSetsTable(attachedDatabase, alias);
  }
}

class LocalSetRow extends DataClass implements Insertable<LocalSetRow> {
  final String localId;
  final int sessionExerciseId;
  final String exerciseId;
  final int setNumber;
  final String type;
  final double? weightKg;
  final int? reps;
  final double? rpe;
  final bool isCompleted;
  final DateTime performedAt;
  const LocalSetRow({
    required this.localId,
    required this.sessionExerciseId,
    required this.exerciseId,
    required this.setNumber,
    required this.type,
    this.weightKg,
    this.reps,
    this.rpe,
    required this.isCompleted,
    required this.performedAt,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['local_id'] = Variable<String>(localId);
    map['session_exercise_id'] = Variable<int>(sessionExerciseId);
    map['exercise_id'] = Variable<String>(exerciseId);
    map['set_number'] = Variable<int>(setNumber);
    map['type'] = Variable<String>(type);
    if (!nullToAbsent || weightKg != null) {
      map['weight_kg'] = Variable<double>(weightKg);
    }
    if (!nullToAbsent || reps != null) {
      map['reps'] = Variable<int>(reps);
    }
    if (!nullToAbsent || rpe != null) {
      map['rpe'] = Variable<double>(rpe);
    }
    map['is_completed'] = Variable<bool>(isCompleted);
    map['performed_at'] = Variable<DateTime>(performedAt);
    return map;
  }

  LocalSetsCompanion toCompanion(bool nullToAbsent) {
    return LocalSetsCompanion(
      localId: Value(localId),
      sessionExerciseId: Value(sessionExerciseId),
      exerciseId: Value(exerciseId),
      setNumber: Value(setNumber),
      type: Value(type),
      weightKg: weightKg == null && nullToAbsent
          ? const Value.absent()
          : Value(weightKg),
      reps: reps == null && nullToAbsent ? const Value.absent() : Value(reps),
      rpe: rpe == null && nullToAbsent ? const Value.absent() : Value(rpe),
      isCompleted: Value(isCompleted),
      performedAt: Value(performedAt),
    );
  }

  factory LocalSetRow.fromJson(
    Map<String, dynamic> json, {
    ValueSerializer? serializer,
  }) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return LocalSetRow(
      localId: serializer.fromJson<String>(json['localId']),
      sessionExerciseId: serializer.fromJson<int>(json['sessionExerciseId']),
      exerciseId: serializer.fromJson<String>(json['exerciseId']),
      setNumber: serializer.fromJson<int>(json['setNumber']),
      type: serializer.fromJson<String>(json['type']),
      weightKg: serializer.fromJson<double?>(json['weightKg']),
      reps: serializer.fromJson<int?>(json['reps']),
      rpe: serializer.fromJson<double?>(json['rpe']),
      isCompleted: serializer.fromJson<bool>(json['isCompleted']),
      performedAt: serializer.fromJson<DateTime>(json['performedAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'localId': serializer.toJson<String>(localId),
      'sessionExerciseId': serializer.toJson<int>(sessionExerciseId),
      'exerciseId': serializer.toJson<String>(exerciseId),
      'setNumber': serializer.toJson<int>(setNumber),
      'type': serializer.toJson<String>(type),
      'weightKg': serializer.toJson<double?>(weightKg),
      'reps': serializer.toJson<int?>(reps),
      'rpe': serializer.toJson<double?>(rpe),
      'isCompleted': serializer.toJson<bool>(isCompleted),
      'performedAt': serializer.toJson<DateTime>(performedAt),
    };
  }

  LocalSetRow copyWith({
    String? localId,
    int? sessionExerciseId,
    String? exerciseId,
    int? setNumber,
    String? type,
    Value<double?> weightKg = const Value.absent(),
    Value<int?> reps = const Value.absent(),
    Value<double?> rpe = const Value.absent(),
    bool? isCompleted,
    DateTime? performedAt,
  }) => LocalSetRow(
    localId: localId ?? this.localId,
    sessionExerciseId: sessionExerciseId ?? this.sessionExerciseId,
    exerciseId: exerciseId ?? this.exerciseId,
    setNumber: setNumber ?? this.setNumber,
    type: type ?? this.type,
    weightKg: weightKg.present ? weightKg.value : this.weightKg,
    reps: reps.present ? reps.value : this.reps,
    rpe: rpe.present ? rpe.value : this.rpe,
    isCompleted: isCompleted ?? this.isCompleted,
    performedAt: performedAt ?? this.performedAt,
  );
  LocalSetRow copyWithCompanion(LocalSetsCompanion data) {
    return LocalSetRow(
      localId: data.localId.present ? data.localId.value : this.localId,
      sessionExerciseId: data.sessionExerciseId.present
          ? data.sessionExerciseId.value
          : this.sessionExerciseId,
      exerciseId: data.exerciseId.present
          ? data.exerciseId.value
          : this.exerciseId,
      setNumber: data.setNumber.present ? data.setNumber.value : this.setNumber,
      type: data.type.present ? data.type.value : this.type,
      weightKg: data.weightKg.present ? data.weightKg.value : this.weightKg,
      reps: data.reps.present ? data.reps.value : this.reps,
      rpe: data.rpe.present ? data.rpe.value : this.rpe,
      isCompleted: data.isCompleted.present
          ? data.isCompleted.value
          : this.isCompleted,
      performedAt: data.performedAt.present
          ? data.performedAt.value
          : this.performedAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('LocalSetRow(')
          ..write('localId: $localId, ')
          ..write('sessionExerciseId: $sessionExerciseId, ')
          ..write('exerciseId: $exerciseId, ')
          ..write('setNumber: $setNumber, ')
          ..write('type: $type, ')
          ..write('weightKg: $weightKg, ')
          ..write('reps: $reps, ')
          ..write('rpe: $rpe, ')
          ..write('isCompleted: $isCompleted, ')
          ..write('performedAt: $performedAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
    localId,
    sessionExerciseId,
    exerciseId,
    setNumber,
    type,
    weightKg,
    reps,
    rpe,
    isCompleted,
    performedAt,
  );
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is LocalSetRow &&
          other.localId == this.localId &&
          other.sessionExerciseId == this.sessionExerciseId &&
          other.exerciseId == this.exerciseId &&
          other.setNumber == this.setNumber &&
          other.type == this.type &&
          other.weightKg == this.weightKg &&
          other.reps == this.reps &&
          other.rpe == this.rpe &&
          other.isCompleted == this.isCompleted &&
          other.performedAt == this.performedAt);
}

class LocalSetsCompanion extends UpdateCompanion<LocalSetRow> {
  final Value<String> localId;
  final Value<int> sessionExerciseId;
  final Value<String> exerciseId;
  final Value<int> setNumber;
  final Value<String> type;
  final Value<double?> weightKg;
  final Value<int?> reps;
  final Value<double?> rpe;
  final Value<bool> isCompleted;
  final Value<DateTime> performedAt;
  final Value<int> rowid;
  const LocalSetsCompanion({
    this.localId = const Value.absent(),
    this.sessionExerciseId = const Value.absent(),
    this.exerciseId = const Value.absent(),
    this.setNumber = const Value.absent(),
    this.type = const Value.absent(),
    this.weightKg = const Value.absent(),
    this.reps = const Value.absent(),
    this.rpe = const Value.absent(),
    this.isCompleted = const Value.absent(),
    this.performedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  LocalSetsCompanion.insert({
    required String localId,
    required int sessionExerciseId,
    required String exerciseId,
    required int setNumber,
    this.type = const Value.absent(),
    this.weightKg = const Value.absent(),
    this.reps = const Value.absent(),
    this.rpe = const Value.absent(),
    this.isCompleted = const Value.absent(),
    required DateTime performedAt,
    this.rowid = const Value.absent(),
  }) : localId = Value(localId),
       sessionExerciseId = Value(sessionExerciseId),
       exerciseId = Value(exerciseId),
       setNumber = Value(setNumber),
       performedAt = Value(performedAt);
  static Insertable<LocalSetRow> custom({
    Expression<String>? localId,
    Expression<int>? sessionExerciseId,
    Expression<String>? exerciseId,
    Expression<int>? setNumber,
    Expression<String>? type,
    Expression<double>? weightKg,
    Expression<int>? reps,
    Expression<double>? rpe,
    Expression<bool>? isCompleted,
    Expression<DateTime>? performedAt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (localId != null) 'local_id': localId,
      if (sessionExerciseId != null) 'session_exercise_id': sessionExerciseId,
      if (exerciseId != null) 'exercise_id': exerciseId,
      if (setNumber != null) 'set_number': setNumber,
      if (type != null) 'type': type,
      if (weightKg != null) 'weight_kg': weightKg,
      if (reps != null) 'reps': reps,
      if (rpe != null) 'rpe': rpe,
      if (isCompleted != null) 'is_completed': isCompleted,
      if (performedAt != null) 'performed_at': performedAt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  LocalSetsCompanion copyWith({
    Value<String>? localId,
    Value<int>? sessionExerciseId,
    Value<String>? exerciseId,
    Value<int>? setNumber,
    Value<String>? type,
    Value<double?>? weightKg,
    Value<int?>? reps,
    Value<double?>? rpe,
    Value<bool>? isCompleted,
    Value<DateTime>? performedAt,
    Value<int>? rowid,
  }) {
    return LocalSetsCompanion(
      localId: localId ?? this.localId,
      sessionExerciseId: sessionExerciseId ?? this.sessionExerciseId,
      exerciseId: exerciseId ?? this.exerciseId,
      setNumber: setNumber ?? this.setNumber,
      type: type ?? this.type,
      weightKg: weightKg ?? this.weightKg,
      reps: reps ?? this.reps,
      rpe: rpe ?? this.rpe,
      isCompleted: isCompleted ?? this.isCompleted,
      performedAt: performedAt ?? this.performedAt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (localId.present) {
      map['local_id'] = Variable<String>(localId.value);
    }
    if (sessionExerciseId.present) {
      map['session_exercise_id'] = Variable<int>(sessionExerciseId.value);
    }
    if (exerciseId.present) {
      map['exercise_id'] = Variable<String>(exerciseId.value);
    }
    if (setNumber.present) {
      map['set_number'] = Variable<int>(setNumber.value);
    }
    if (type.present) {
      map['type'] = Variable<String>(type.value);
    }
    if (weightKg.present) {
      map['weight_kg'] = Variable<double>(weightKg.value);
    }
    if (reps.present) {
      map['reps'] = Variable<int>(reps.value);
    }
    if (rpe.present) {
      map['rpe'] = Variable<double>(rpe.value);
    }
    if (isCompleted.present) {
      map['is_completed'] = Variable<bool>(isCompleted.value);
    }
    if (performedAt.present) {
      map['performed_at'] = Variable<DateTime>(performedAt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('LocalSetsCompanion(')
          ..write('localId: $localId, ')
          ..write('sessionExerciseId: $sessionExerciseId, ')
          ..write('exerciseId: $exerciseId, ')
          ..write('setNumber: $setNumber, ')
          ..write('type: $type, ')
          ..write('weightKg: $weightKg, ')
          ..write('reps: $reps, ')
          ..write('rpe: $rpe, ')
          ..write('isCompleted: $isCompleted, ')
          ..write('performedAt: $performedAt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

abstract class _$AppDatabase extends GeneratedDatabase {
  _$AppDatabase(QueryExecutor e) : super(e);
  $AppDatabaseManager get managers => $AppDatabaseManager(this);
  late final $LocalSessionsTable localSessions = $LocalSessionsTable(this);
  late final $LocalSessionExercisesTable localSessionExercises =
      $LocalSessionExercisesTable(this);
  late final $LocalSetsTable localSets = $LocalSetsTable(this);
  @override
  Iterable<TableInfo<Table, Object?>> get allTables =>
      allSchemaEntities.whereType<TableInfo<Table, Object?>>();
  @override
  List<DatabaseSchemaEntity> get allSchemaEntities => [
    localSessions,
    localSessionExercises,
    localSets,
  ];
}

typedef $$LocalSessionsTableCreateCompanionBuilder =
    LocalSessionsCompanion Function({
      required String localId,
      required String name,
      required String status,
      Value<String?> assignedWorkoutId,
      required DateTime startedAt,
      Value<DateTime?> endedAt,
      Value<int?> perceivedEffort,
      Value<int?> mood,
      Value<String?> notes,
      Value<bool> synced,
      Value<int> rowid,
    });
typedef $$LocalSessionsTableUpdateCompanionBuilder =
    LocalSessionsCompanion Function({
      Value<String> localId,
      Value<String> name,
      Value<String> status,
      Value<String?> assignedWorkoutId,
      Value<DateTime> startedAt,
      Value<DateTime?> endedAt,
      Value<int?> perceivedEffort,
      Value<int?> mood,
      Value<String?> notes,
      Value<bool> synced,
      Value<int> rowid,
    });

final class $$LocalSessionsTableReferences
    extends
        BaseReferences<_$AppDatabase, $LocalSessionsTable, LocalSessionRow> {
  $$LocalSessionsTableReferences(
    super.$_db,
    super.$_table,
    super.$_typedResult,
  );

  static MultiTypedResultKey<
    $LocalSessionExercisesTable,
    List<LocalSessionExerciseRow>
  >
  _localSessionExercisesRefsTable(
    _$AppDatabase db,
  ) => MultiTypedResultKey.fromTable(
    db.localSessionExercises,
    aliasName:
        'local_sessions__local_id__local_session_exercises__session_local_id',
  );

  $$LocalSessionExercisesTableProcessedTableManager
  get localSessionExercisesRefs {
    final manager =
        $$LocalSessionExercisesTableTableManager(
          $_db,
          $_db.localSessionExercises,
        ).filter(
          (f) => f.sessionLocalId.localId.sqlEquals(
            $_itemColumn<String>('local_id')!,
          ),
        );

    final cache = $_typedResult.readTableOrNull(
      _localSessionExercisesRefsTable($_db),
    );
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: cache),
    );
  }
}

class $$LocalSessionsTableFilterComposer
    extends Composer<_$AppDatabase, $LocalSessionsTable> {
  $$LocalSessionsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get localId => $composableBuilder(
    column: $table.localId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get name => $composableBuilder(
    column: $table.name,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get status => $composableBuilder(
    column: $table.status,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get assignedWorkoutId => $composableBuilder(
    column: $table.assignedWorkoutId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get startedAt => $composableBuilder(
    column: $table.startedAt,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get endedAt => $composableBuilder(
    column: $table.endedAt,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get perceivedEffort => $composableBuilder(
    column: $table.perceivedEffort,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get mood => $composableBuilder(
    column: $table.mood,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get notes => $composableBuilder(
    column: $table.notes,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<bool> get synced => $composableBuilder(
    column: $table.synced,
    builder: (column) => ColumnFilters(column),
  );

  Expression<bool> localSessionExercisesRefs(
    Expression<bool> Function($$LocalSessionExercisesTableFilterComposer f) f,
  ) {
    final $$LocalSessionExercisesTableFilterComposer composer =
        $composerBuilder(
          composer: this,
          getCurrentColumn: (t) => t.localId,
          referencedTable: $db.localSessionExercises,
          getReferencedColumn: (t) => t.sessionLocalId,
          builder:
              (
                joinBuilder, {
                $addJoinBuilderToRootComposer,
                $removeJoinBuilderFromRootComposer,
              }) => $$LocalSessionExercisesTableFilterComposer(
                $db: $db,
                $table: $db.localSessionExercises,
                $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
                joinBuilder: joinBuilder,
                $removeJoinBuilderFromRootComposer:
                    $removeJoinBuilderFromRootComposer,
              ),
        );
    return f(composer);
  }
}

class $$LocalSessionsTableOrderingComposer
    extends Composer<_$AppDatabase, $LocalSessionsTable> {
  $$LocalSessionsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get localId => $composableBuilder(
    column: $table.localId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get name => $composableBuilder(
    column: $table.name,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get status => $composableBuilder(
    column: $table.status,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get assignedWorkoutId => $composableBuilder(
    column: $table.assignedWorkoutId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get startedAt => $composableBuilder(
    column: $table.startedAt,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get endedAt => $composableBuilder(
    column: $table.endedAt,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get perceivedEffort => $composableBuilder(
    column: $table.perceivedEffort,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get mood => $composableBuilder(
    column: $table.mood,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get notes => $composableBuilder(
    column: $table.notes,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<bool> get synced => $composableBuilder(
    column: $table.synced,
    builder: (column) => ColumnOrderings(column),
  );
}

class $$LocalSessionsTableAnnotationComposer
    extends Composer<_$AppDatabase, $LocalSessionsTable> {
  $$LocalSessionsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get localId =>
      $composableBuilder(column: $table.localId, builder: (column) => column);

  GeneratedColumn<String> get name =>
      $composableBuilder(column: $table.name, builder: (column) => column);

  GeneratedColumn<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => column);

  GeneratedColumn<String> get assignedWorkoutId => $composableBuilder(
    column: $table.assignedWorkoutId,
    builder: (column) => column,
  );

  GeneratedColumn<DateTime> get startedAt =>
      $composableBuilder(column: $table.startedAt, builder: (column) => column);

  GeneratedColumn<DateTime> get endedAt =>
      $composableBuilder(column: $table.endedAt, builder: (column) => column);

  GeneratedColumn<int> get perceivedEffort => $composableBuilder(
    column: $table.perceivedEffort,
    builder: (column) => column,
  );

  GeneratedColumn<int> get mood =>
      $composableBuilder(column: $table.mood, builder: (column) => column);

  GeneratedColumn<String> get notes =>
      $composableBuilder(column: $table.notes, builder: (column) => column);

  GeneratedColumn<bool> get synced =>
      $composableBuilder(column: $table.synced, builder: (column) => column);

  Expression<T> localSessionExercisesRefs<T extends Object>(
    Expression<T> Function($$LocalSessionExercisesTableAnnotationComposer a) f,
  ) {
    final $$LocalSessionExercisesTableAnnotationComposer composer =
        $composerBuilder(
          composer: this,
          getCurrentColumn: (t) => t.localId,
          referencedTable: $db.localSessionExercises,
          getReferencedColumn: (t) => t.sessionLocalId,
          builder:
              (
                joinBuilder, {
                $addJoinBuilderToRootComposer,
                $removeJoinBuilderFromRootComposer,
              }) => $$LocalSessionExercisesTableAnnotationComposer(
                $db: $db,
                $table: $db.localSessionExercises,
                $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
                joinBuilder: joinBuilder,
                $removeJoinBuilderFromRootComposer:
                    $removeJoinBuilderFromRootComposer,
              ),
        );
    return f(composer);
  }
}

class $$LocalSessionsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $LocalSessionsTable,
          LocalSessionRow,
          $$LocalSessionsTableFilterComposer,
          $$LocalSessionsTableOrderingComposer,
          $$LocalSessionsTableAnnotationComposer,
          $$LocalSessionsTableCreateCompanionBuilder,
          $$LocalSessionsTableUpdateCompanionBuilder,
          (LocalSessionRow, $$LocalSessionsTableReferences),
          LocalSessionRow,
          PrefetchHooks Function({bool localSessionExercisesRefs})
        > {
  $$LocalSessionsTableTableManager(_$AppDatabase db, $LocalSessionsTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$LocalSessionsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$LocalSessionsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$LocalSessionsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> localId = const Value.absent(),
                Value<String> name = const Value.absent(),
                Value<String> status = const Value.absent(),
                Value<String?> assignedWorkoutId = const Value.absent(),
                Value<DateTime> startedAt = const Value.absent(),
                Value<DateTime?> endedAt = const Value.absent(),
                Value<int?> perceivedEffort = const Value.absent(),
                Value<int?> mood = const Value.absent(),
                Value<String?> notes = const Value.absent(),
                Value<bool> synced = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => LocalSessionsCompanion(
                localId: localId,
                name: name,
                status: status,
                assignedWorkoutId: assignedWorkoutId,
                startedAt: startedAt,
                endedAt: endedAt,
                perceivedEffort: perceivedEffort,
                mood: mood,
                notes: notes,
                synced: synced,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String localId,
                required String name,
                required String status,
                Value<String?> assignedWorkoutId = const Value.absent(),
                required DateTime startedAt,
                Value<DateTime?> endedAt = const Value.absent(),
                Value<int?> perceivedEffort = const Value.absent(),
                Value<int?> mood = const Value.absent(),
                Value<String?> notes = const Value.absent(),
                Value<bool> synced = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => LocalSessionsCompanion.insert(
                localId: localId,
                name: name,
                status: status,
                assignedWorkoutId: assignedWorkoutId,
                startedAt: startedAt,
                endedAt: endedAt,
                perceivedEffort: perceivedEffort,
                mood: mood,
                notes: notes,
                synced: synced,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable(table),
                  $$LocalSessionsTableReferences(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: ({localSessionExercisesRefs = false}) {
            return PrefetchHooks(
              db: db,
              explicitlyWatchedTables: [
                if (localSessionExercisesRefs) db.localSessionExercises,
              ],
              addJoins: null,
              getPrefetchedDataCallback: (items) async {
                return [
                  if (localSessionExercisesRefs)
                    await $_getPrefetchedData<
                      LocalSessionRow,
                      $LocalSessionsTable,
                      LocalSessionExerciseRow
                    >(
                      currentTable: table,
                      referencedTable: $$LocalSessionsTableReferences
                          ._localSessionExercisesRefsTable(db),
                      managerFromTypedResult: (p0) =>
                          $$LocalSessionsTableReferences(
                            db,
                            table,
                            p0,
                          ).localSessionExercisesRefs,
                      referencedItemsForCurrentItem: (item, referencedItems) =>
                          referencedItems.where(
                            (e) => e.sessionLocalId == item.localId,
                          ),
                      typedResults: items,
                    ),
                ];
              },
            );
          },
        ),
      );
}

typedef $$LocalSessionsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $LocalSessionsTable,
      LocalSessionRow,
      $$LocalSessionsTableFilterComposer,
      $$LocalSessionsTableOrderingComposer,
      $$LocalSessionsTableAnnotationComposer,
      $$LocalSessionsTableCreateCompanionBuilder,
      $$LocalSessionsTableUpdateCompanionBuilder,
      (LocalSessionRow, $$LocalSessionsTableReferences),
      LocalSessionRow,
      PrefetchHooks Function({bool localSessionExercisesRefs})
    >;
typedef $$LocalSessionExercisesTableCreateCompanionBuilder =
    LocalSessionExercisesCompanion Function({
      Value<int> id,
      required String sessionLocalId,
      required String exerciseId,
      required int orderIndex,
    });
typedef $$LocalSessionExercisesTableUpdateCompanionBuilder =
    LocalSessionExercisesCompanion Function({
      Value<int> id,
      Value<String> sessionLocalId,
      Value<String> exerciseId,
      Value<int> orderIndex,
    });

final class $$LocalSessionExercisesTableReferences
    extends
        BaseReferences<
          _$AppDatabase,
          $LocalSessionExercisesTable,
          LocalSessionExerciseRow
        > {
  $$LocalSessionExercisesTableReferences(
    super.$_db,
    super.$_table,
    super.$_typedResult,
  );

  static $LocalSessionsTable _sessionLocalIdTable(_$AppDatabase db) =>
      db.localSessions.createAlias(
        'local_session_exercises__session_local_id__local_sessions__local_id',
      );

  $$LocalSessionsTableProcessedTableManager get sessionLocalId {
    final $_column = $_itemColumn<String>('session_local_id')!;

    final manager = $$LocalSessionsTableTableManager(
      $_db,
      $_db.localSessions,
    ).filter((f) => f.localId.sqlEquals($_column));
    final item = $_typedResult.readTableOrNull(_sessionLocalIdTable($_db));
    if (item == null) return manager;
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: [item]),
    );
  }

  static MultiTypedResultKey<$LocalSetsTable, List<LocalSetRow>>
  _localSetsRefsTable(_$AppDatabase db) => MultiTypedResultKey.fromTable(
    db.localSets,
    aliasName: 'local_session_exercises__id__local_sets__session_exercise_id',
  );

  $$LocalSetsTableProcessedTableManager get localSetsRefs {
    final manager = $$LocalSetsTableTableManager(
      $_db,
      $_db.localSets,
    ).filter((f) => f.sessionExerciseId.id.sqlEquals($_itemColumn<int>('id')!));

    final cache = $_typedResult.readTableOrNull(_localSetsRefsTable($_db));
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: cache),
    );
  }
}

class $$LocalSessionExercisesTableFilterComposer
    extends Composer<_$AppDatabase, $LocalSessionExercisesTable> {
  $$LocalSessionExercisesTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<int> get id => $composableBuilder(
    column: $table.id,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get exerciseId => $composableBuilder(
    column: $table.exerciseId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get orderIndex => $composableBuilder(
    column: $table.orderIndex,
    builder: (column) => ColumnFilters(column),
  );

  $$LocalSessionsTableFilterComposer get sessionLocalId {
    final $$LocalSessionsTableFilterComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.sessionLocalId,
      referencedTable: $db.localSessions,
      getReferencedColumn: (t) => t.localId,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$LocalSessionsTableFilterComposer(
            $db: $db,
            $table: $db.localSessions,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }

  Expression<bool> localSetsRefs(
    Expression<bool> Function($$LocalSetsTableFilterComposer f) f,
  ) {
    final $$LocalSetsTableFilterComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.id,
      referencedTable: $db.localSets,
      getReferencedColumn: (t) => t.sessionExerciseId,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$LocalSetsTableFilterComposer(
            $db: $db,
            $table: $db.localSets,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return f(composer);
  }
}

class $$LocalSessionExercisesTableOrderingComposer
    extends Composer<_$AppDatabase, $LocalSessionExercisesTable> {
  $$LocalSessionExercisesTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<int> get id => $composableBuilder(
    column: $table.id,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get exerciseId => $composableBuilder(
    column: $table.exerciseId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get orderIndex => $composableBuilder(
    column: $table.orderIndex,
    builder: (column) => ColumnOrderings(column),
  );

  $$LocalSessionsTableOrderingComposer get sessionLocalId {
    final $$LocalSessionsTableOrderingComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.sessionLocalId,
      referencedTable: $db.localSessions,
      getReferencedColumn: (t) => t.localId,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$LocalSessionsTableOrderingComposer(
            $db: $db,
            $table: $db.localSessions,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }
}

class $$LocalSessionExercisesTableAnnotationComposer
    extends Composer<_$AppDatabase, $LocalSessionExercisesTable> {
  $$LocalSessionExercisesTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<int> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get exerciseId => $composableBuilder(
    column: $table.exerciseId,
    builder: (column) => column,
  );

  GeneratedColumn<int> get orderIndex => $composableBuilder(
    column: $table.orderIndex,
    builder: (column) => column,
  );

  $$LocalSessionsTableAnnotationComposer get sessionLocalId {
    final $$LocalSessionsTableAnnotationComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.sessionLocalId,
      referencedTable: $db.localSessions,
      getReferencedColumn: (t) => t.localId,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$LocalSessionsTableAnnotationComposer(
            $db: $db,
            $table: $db.localSessions,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return composer;
  }

  Expression<T> localSetsRefs<T extends Object>(
    Expression<T> Function($$LocalSetsTableAnnotationComposer a) f,
  ) {
    final $$LocalSetsTableAnnotationComposer composer = $composerBuilder(
      composer: this,
      getCurrentColumn: (t) => t.id,
      referencedTable: $db.localSets,
      getReferencedColumn: (t) => t.sessionExerciseId,
      builder:
          (
            joinBuilder, {
            $addJoinBuilderToRootComposer,
            $removeJoinBuilderFromRootComposer,
          }) => $$LocalSetsTableAnnotationComposer(
            $db: $db,
            $table: $db.localSets,
            $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
            joinBuilder: joinBuilder,
            $removeJoinBuilderFromRootComposer:
                $removeJoinBuilderFromRootComposer,
          ),
    );
    return f(composer);
  }
}

class $$LocalSessionExercisesTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $LocalSessionExercisesTable,
          LocalSessionExerciseRow,
          $$LocalSessionExercisesTableFilterComposer,
          $$LocalSessionExercisesTableOrderingComposer,
          $$LocalSessionExercisesTableAnnotationComposer,
          $$LocalSessionExercisesTableCreateCompanionBuilder,
          $$LocalSessionExercisesTableUpdateCompanionBuilder,
          (LocalSessionExerciseRow, $$LocalSessionExercisesTableReferences),
          LocalSessionExerciseRow,
          PrefetchHooks Function({bool sessionLocalId, bool localSetsRefs})
        > {
  $$LocalSessionExercisesTableTableManager(
    _$AppDatabase db,
    $LocalSessionExercisesTable table,
  ) : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$LocalSessionExercisesTableFilterComposer(
                $db: db,
                $table: table,
              ),
          createOrderingComposer: () =>
              $$LocalSessionExercisesTableOrderingComposer(
                $db: db,
                $table: table,
              ),
          createComputedFieldComposer: () =>
              $$LocalSessionExercisesTableAnnotationComposer(
                $db: db,
                $table: table,
              ),
          updateCompanionCallback:
              ({
                Value<int> id = const Value.absent(),
                Value<String> sessionLocalId = const Value.absent(),
                Value<String> exerciseId = const Value.absent(),
                Value<int> orderIndex = const Value.absent(),
              }) => LocalSessionExercisesCompanion(
                id: id,
                sessionLocalId: sessionLocalId,
                exerciseId: exerciseId,
                orderIndex: orderIndex,
              ),
          createCompanionCallback:
              ({
                Value<int> id = const Value.absent(),
                required String sessionLocalId,
                required String exerciseId,
                required int orderIndex,
              }) => LocalSessionExercisesCompanion.insert(
                id: id,
                sessionLocalId: sessionLocalId,
                exerciseId: exerciseId,
                orderIndex: orderIndex,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable(table),
                  $$LocalSessionExercisesTableReferences(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback:
              ({sessionLocalId = false, localSetsRefs = false}) {
                return PrefetchHooks(
                  db: db,
                  explicitlyWatchedTables: [if (localSetsRefs) db.localSets],
                  addJoins:
                      <
                        T extends TableManagerState<
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic,
                          dynamic
                        >
                      >(state) {
                        if (sessionLocalId) {
                          state =
                              state.withJoin(
                                    currentTable: table,
                                    currentColumn: table.sessionLocalId,
                                    referencedTable:
                                        $$LocalSessionExercisesTableReferences
                                            ._sessionLocalIdTable(db),
                                    referencedColumn:
                                        $$LocalSessionExercisesTableReferences
                                            ._sessionLocalIdTable(db)
                                            .localId,
                                  )
                                  as T;
                        }

                        return state;
                      },
                  getPrefetchedDataCallback: (items) async {
                    return [
                      if (localSetsRefs)
                        await $_getPrefetchedData<
                          LocalSessionExerciseRow,
                          $LocalSessionExercisesTable,
                          LocalSetRow
                        >(
                          currentTable: table,
                          referencedTable:
                              $$LocalSessionExercisesTableReferences
                                  ._localSetsRefsTable(db),
                          managerFromTypedResult: (p0) =>
                              $$LocalSessionExercisesTableReferences(
                                db,
                                table,
                                p0,
                              ).localSetsRefs,
                          referencedItemsForCurrentItem:
                              (item, referencedItems) => referencedItems.where(
                                (e) => e.sessionExerciseId == item.id,
                              ),
                          typedResults: items,
                        ),
                    ];
                  },
                );
              },
        ),
      );
}

typedef $$LocalSessionExercisesTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $LocalSessionExercisesTable,
      LocalSessionExerciseRow,
      $$LocalSessionExercisesTableFilterComposer,
      $$LocalSessionExercisesTableOrderingComposer,
      $$LocalSessionExercisesTableAnnotationComposer,
      $$LocalSessionExercisesTableCreateCompanionBuilder,
      $$LocalSessionExercisesTableUpdateCompanionBuilder,
      (LocalSessionExerciseRow, $$LocalSessionExercisesTableReferences),
      LocalSessionExerciseRow,
      PrefetchHooks Function({bool sessionLocalId, bool localSetsRefs})
    >;
typedef $$LocalSetsTableCreateCompanionBuilder =
    LocalSetsCompanion Function({
      required String localId,
      required int sessionExerciseId,
      required String exerciseId,
      required int setNumber,
      Value<String> type,
      Value<double?> weightKg,
      Value<int?> reps,
      Value<double?> rpe,
      Value<bool> isCompleted,
      required DateTime performedAt,
      Value<int> rowid,
    });
typedef $$LocalSetsTableUpdateCompanionBuilder =
    LocalSetsCompanion Function({
      Value<String> localId,
      Value<int> sessionExerciseId,
      Value<String> exerciseId,
      Value<int> setNumber,
      Value<String> type,
      Value<double?> weightKg,
      Value<int?> reps,
      Value<double?> rpe,
      Value<bool> isCompleted,
      Value<DateTime> performedAt,
      Value<int> rowid,
    });

final class $$LocalSetsTableReferences
    extends BaseReferences<_$AppDatabase, $LocalSetsTable, LocalSetRow> {
  $$LocalSetsTableReferences(super.$_db, super.$_table, super.$_typedResult);

  static $LocalSessionExercisesTable _sessionExerciseIdTable(
    _$AppDatabase db,
  ) => db.localSessionExercises.createAlias(
    'local_sets__session_exercise_id__local_session_exercises__id',
  );

  $$LocalSessionExercisesTableProcessedTableManager get sessionExerciseId {
    final $_column = $_itemColumn<int>('session_exercise_id')!;

    final manager = $$LocalSessionExercisesTableTableManager(
      $_db,
      $_db.localSessionExercises,
    ).filter((f) => f.id.sqlEquals($_column));
    final item = $_typedResult.readTableOrNull(_sessionExerciseIdTable($_db));
    if (item == null) return manager;
    return ProcessedTableManager(
      manager.$state.copyWith(prefetchedData: [item]),
    );
  }
}

class $$LocalSetsTableFilterComposer
    extends Composer<_$AppDatabase, $LocalSetsTable> {
  $$LocalSetsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get localId => $composableBuilder(
    column: $table.localId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get exerciseId => $composableBuilder(
    column: $table.exerciseId,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get setNumber => $composableBuilder(
    column: $table.setNumber,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<String> get type => $composableBuilder(
    column: $table.type,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<double> get weightKg => $composableBuilder(
    column: $table.weightKg,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<int> get reps => $composableBuilder(
    column: $table.reps,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<double> get rpe => $composableBuilder(
    column: $table.rpe,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<bool> get isCompleted => $composableBuilder(
    column: $table.isCompleted,
    builder: (column) => ColumnFilters(column),
  );

  ColumnFilters<DateTime> get performedAt => $composableBuilder(
    column: $table.performedAt,
    builder: (column) => ColumnFilters(column),
  );

  $$LocalSessionExercisesTableFilterComposer get sessionExerciseId {
    final $$LocalSessionExercisesTableFilterComposer composer =
        $composerBuilder(
          composer: this,
          getCurrentColumn: (t) => t.sessionExerciseId,
          referencedTable: $db.localSessionExercises,
          getReferencedColumn: (t) => t.id,
          builder:
              (
                joinBuilder, {
                $addJoinBuilderToRootComposer,
                $removeJoinBuilderFromRootComposer,
              }) => $$LocalSessionExercisesTableFilterComposer(
                $db: $db,
                $table: $db.localSessionExercises,
                $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
                joinBuilder: joinBuilder,
                $removeJoinBuilderFromRootComposer:
                    $removeJoinBuilderFromRootComposer,
              ),
        );
    return composer;
  }
}

class $$LocalSetsTableOrderingComposer
    extends Composer<_$AppDatabase, $LocalSetsTable> {
  $$LocalSetsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get localId => $composableBuilder(
    column: $table.localId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get exerciseId => $composableBuilder(
    column: $table.exerciseId,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get setNumber => $composableBuilder(
    column: $table.setNumber,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<String> get type => $composableBuilder(
    column: $table.type,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<double> get weightKg => $composableBuilder(
    column: $table.weightKg,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<int> get reps => $composableBuilder(
    column: $table.reps,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<double> get rpe => $composableBuilder(
    column: $table.rpe,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<bool> get isCompleted => $composableBuilder(
    column: $table.isCompleted,
    builder: (column) => ColumnOrderings(column),
  );

  ColumnOrderings<DateTime> get performedAt => $composableBuilder(
    column: $table.performedAt,
    builder: (column) => ColumnOrderings(column),
  );

  $$LocalSessionExercisesTableOrderingComposer get sessionExerciseId {
    final $$LocalSessionExercisesTableOrderingComposer composer =
        $composerBuilder(
          composer: this,
          getCurrentColumn: (t) => t.sessionExerciseId,
          referencedTable: $db.localSessionExercises,
          getReferencedColumn: (t) => t.id,
          builder:
              (
                joinBuilder, {
                $addJoinBuilderToRootComposer,
                $removeJoinBuilderFromRootComposer,
              }) => $$LocalSessionExercisesTableOrderingComposer(
                $db: $db,
                $table: $db.localSessionExercises,
                $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
                joinBuilder: joinBuilder,
                $removeJoinBuilderFromRootComposer:
                    $removeJoinBuilderFromRootComposer,
              ),
        );
    return composer;
  }
}

class $$LocalSetsTableAnnotationComposer
    extends Composer<_$AppDatabase, $LocalSetsTable> {
  $$LocalSetsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get localId =>
      $composableBuilder(column: $table.localId, builder: (column) => column);

  GeneratedColumn<String> get exerciseId => $composableBuilder(
    column: $table.exerciseId,
    builder: (column) => column,
  );

  GeneratedColumn<int> get setNumber =>
      $composableBuilder(column: $table.setNumber, builder: (column) => column);

  GeneratedColumn<String> get type =>
      $composableBuilder(column: $table.type, builder: (column) => column);

  GeneratedColumn<double> get weightKg =>
      $composableBuilder(column: $table.weightKg, builder: (column) => column);

  GeneratedColumn<int> get reps =>
      $composableBuilder(column: $table.reps, builder: (column) => column);

  GeneratedColumn<double> get rpe =>
      $composableBuilder(column: $table.rpe, builder: (column) => column);

  GeneratedColumn<bool> get isCompleted => $composableBuilder(
    column: $table.isCompleted,
    builder: (column) => column,
  );

  GeneratedColumn<DateTime> get performedAt => $composableBuilder(
    column: $table.performedAt,
    builder: (column) => column,
  );

  $$LocalSessionExercisesTableAnnotationComposer get sessionExerciseId {
    final $$LocalSessionExercisesTableAnnotationComposer composer =
        $composerBuilder(
          composer: this,
          getCurrentColumn: (t) => t.sessionExerciseId,
          referencedTable: $db.localSessionExercises,
          getReferencedColumn: (t) => t.id,
          builder:
              (
                joinBuilder, {
                $addJoinBuilderToRootComposer,
                $removeJoinBuilderFromRootComposer,
              }) => $$LocalSessionExercisesTableAnnotationComposer(
                $db: $db,
                $table: $db.localSessionExercises,
                $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
                joinBuilder: joinBuilder,
                $removeJoinBuilderFromRootComposer:
                    $removeJoinBuilderFromRootComposer,
              ),
        );
    return composer;
  }
}

class $$LocalSetsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $LocalSetsTable,
          LocalSetRow,
          $$LocalSetsTableFilterComposer,
          $$LocalSetsTableOrderingComposer,
          $$LocalSetsTableAnnotationComposer,
          $$LocalSetsTableCreateCompanionBuilder,
          $$LocalSetsTableUpdateCompanionBuilder,
          (LocalSetRow, $$LocalSetsTableReferences),
          LocalSetRow,
          PrefetchHooks Function({bool sessionExerciseId})
        > {
  $$LocalSetsTableTableManager(_$AppDatabase db, $LocalSetsTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$LocalSetsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$LocalSetsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$LocalSetsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> localId = const Value.absent(),
                Value<int> sessionExerciseId = const Value.absent(),
                Value<String> exerciseId = const Value.absent(),
                Value<int> setNumber = const Value.absent(),
                Value<String> type = const Value.absent(),
                Value<double?> weightKg = const Value.absent(),
                Value<int?> reps = const Value.absent(),
                Value<double?> rpe = const Value.absent(),
                Value<bool> isCompleted = const Value.absent(),
                Value<DateTime> performedAt = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => LocalSetsCompanion(
                localId: localId,
                sessionExerciseId: sessionExerciseId,
                exerciseId: exerciseId,
                setNumber: setNumber,
                type: type,
                weightKg: weightKg,
                reps: reps,
                rpe: rpe,
                isCompleted: isCompleted,
                performedAt: performedAt,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String localId,
                required int sessionExerciseId,
                required String exerciseId,
                required int setNumber,
                Value<String> type = const Value.absent(),
                Value<double?> weightKg = const Value.absent(),
                Value<int?> reps = const Value.absent(),
                Value<double?> rpe = const Value.absent(),
                Value<bool> isCompleted = const Value.absent(),
                required DateTime performedAt,
                Value<int> rowid = const Value.absent(),
              }) => LocalSetsCompanion.insert(
                localId: localId,
                sessionExerciseId: sessionExerciseId,
                exerciseId: exerciseId,
                setNumber: setNumber,
                type: type,
                weightKg: weightKg,
                reps: reps,
                rpe: rpe,
                isCompleted: isCompleted,
                performedAt: performedAt,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable(table),
                  $$LocalSetsTableReferences(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: ({sessionExerciseId = false}) {
            return PrefetchHooks(
              db: db,
              explicitlyWatchedTables: [],
              addJoins:
                  <
                    T extends TableManagerState<
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic
                    >
                  >(state) {
                    if (sessionExerciseId) {
                      state =
                          state.withJoin(
                                currentTable: table,
                                currentColumn: table.sessionExerciseId,
                                referencedTable: $$LocalSetsTableReferences
                                    ._sessionExerciseIdTable(db),
                                referencedColumn: $$LocalSetsTableReferences
                                    ._sessionExerciseIdTable(db)
                                    .id,
                              )
                              as T;
                    }

                    return state;
                  },
              getPrefetchedDataCallback: (items) async {
                return [];
              },
            );
          },
        ),
      );
}

typedef $$LocalSetsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $LocalSetsTable,
      LocalSetRow,
      $$LocalSetsTableFilterComposer,
      $$LocalSetsTableOrderingComposer,
      $$LocalSetsTableAnnotationComposer,
      $$LocalSetsTableCreateCompanionBuilder,
      $$LocalSetsTableUpdateCompanionBuilder,
      (LocalSetRow, $$LocalSetsTableReferences),
      LocalSetRow,
      PrefetchHooks Function({bool sessionExerciseId})
    >;

class $AppDatabaseManager {
  final _$AppDatabase _db;
  $AppDatabaseManager(this._db);
  $$LocalSessionsTableTableManager get localSessions =>
      $$LocalSessionsTableTableManager(_db, _db.localSessions);
  $$LocalSessionExercisesTableTableManager get localSessionExercises =>
      $$LocalSessionExercisesTableTableManager(_db, _db.localSessionExercises);
  $$LocalSetsTableTableManager get localSets =>
      $$LocalSetsTableTableManager(_db, _db.localSets);
}
