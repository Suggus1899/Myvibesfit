class AuthUser {
  final String id;
  final String email;
  final String fullName;
  final String? orgId;
  final String? role;

  AuthUser.fromJson(Map<String, dynamic> j)
      : id = j['id'],
        email = j['email'],
        fullName = j['full_name'],
        orgId = j['org_id'],
        role = j['role'];
}

class AuthResult {
  final String accessToken;
  final String refreshToken;
  final AuthUser user;

  AuthResult.fromJson(Map<String, dynamic> j)
      : accessToken = j['access_token'],
        refreshToken = j['refresh_token'],
        user = AuthUser.fromJson(j['user']);
}

class ExerciseSummary {
  final String id;
  final String name;
  final String pattern;
  final String primaryMuscle;

  ExerciseSummary.fromJson(Map<String, dynamic> j)
      : id = j['id'],
        name = j['name'],
        pattern = j['pattern'],
        primaryMuscle = j['primary_muscle'];
}

class AssignedExerciseInfo {
  final String id;
  final String exerciseId;
  final int orderIndex;
  final int targetSets;
  final int? targetRepsMin;
  final int? targetRepsMax;
  final double? targetRPE;
  final int restSeconds;

  AssignedExerciseInfo.fromJson(Map<String, dynamic> j)
      : id = j['id'],
        exerciseId = j['exercise_id'],
        orderIndex = j['order_index'],
        targetSets = j['target_sets'],
        targetRepsMin = j['target_reps_min'],
        targetRepsMax = j['target_reps_max'],
        targetRPE = (j['target_rpe'] as num?)?.toDouble(),
        restSeconds = j['rest_seconds'];
}

class AssignedWorkoutInfo {
  final String id;
  final int weekNumber;
  final int dayIndex;
  final String name;
  final String scheduledOn;
  final String status;
  final List<AssignedExerciseInfo> exercises;

  AssignedWorkoutInfo.fromJson(Map<String, dynamic> j)
      : id = j['id'],
        weekNumber = j['week_number'],
        dayIndex = j['day_index'],
        name = j['name'],
        scheduledOn = j['scheduled_on'] ?? '',
        status = j['status'],
        exercises = (j['exercises'] as List).map((e) => AssignedExerciseInfo.fromJson(e)).toList();
}

class CurrentAssignment {
  final String id;
  final String name;
  final List<AssignedWorkoutInfo> workouts;

  CurrentAssignment.fromJson(Map<String, dynamic> j)
      : id = j['assignment']['id'],
        name = j['assignment']['name'],
        workouts = (j['workouts'] as List).map((w) => AssignedWorkoutInfo.fromJson(w)).toList();
}

class StreakInfo {
  final String kind;
  final int currentCount;
  final String? lastActiveOn;

  StreakInfo.fromJson(Map<String, dynamic> j)
      : kind = j['kind'],
        currentCount = j['current_count'],
        lastActiveOn = j['last_active_on'];
}

class UserStats {
  final int totalXP;
  final int level;
  final int totalSessions;
  final List<StreakInfo> streaks;

  UserStats.fromJson(Map<String, dynamic> j)
      : totalXP = j['total_xp'],
        level = j['level'],
        totalSessions = j['total_sessions'],
        streaks = (j['streaks'] as List? ?? []).map((s) => StreakInfo.fromJson(s)).toList();

  int streakFor(String kind) =>
      streaks.firstWhere((s) => s.kind == kind, orElse: () => StreakInfo.fromJson({'kind': kind, 'current_count': 0})).currentCount;
}

class MyHabit {
  final String id;
  final String habitId;
  final String name;
  final String icon;

  MyHabit.fromJson(Map<String, dynamic> j)
      : id = j['id'],
        habitId = j['habit_id'],
        name = j['name'],
        icon = j['icon'];
}

class PersonalRecordInfo {
  final String exerciseId;
  final String type;
  final double value;

  PersonalRecordInfo.fromJson(Map<String, dynamic> j)
      : exerciseId = j['exercise_id'],
        type = j['type'],
        value = (j['value'] as num).toDouble();
}

class SetLogPoint {
  final double? weightKg;
  final int? reps;
  final String performedAt;

  SetLogPoint.fromJson(Map<String, dynamic> j)
      : weightKg = (j['weight_kg'] as num?)?.toDouble(),
        reps = j['reps'],
        performedAt = j['performed_at'];
}

/// Serie de volumen por sesion que devuelve GET /v1/progress/volume. No se
/// calcula en local: Drift solo guarda lo que sincronizo *este* telefono, asi
/// que el historico completo tiene que venir del servidor.
class VolumePoint {
  final String sessionId;
  final DateTime startedAt;
  final double totalVolumeKg;

  VolumePoint.fromJson(Map<String, dynamic> j)
      : sessionId = j['session_id'],
        startedAt = DateTime.parse(j['started_at']),
        totalVolumeKg = (j['total_volume_kg'] as num).toDouble();
}

class ClientProfile {
  final String sex;
  final double? heightCm;
  final String experience;
  final String primaryGoal;
  final int daysPerWeek;
  final int sessionMinutes;
  final List<String> availableEquipment;
  final String limitations;
  final String unitSystem;
  final bool isOnboarded;

  ClientProfile.fromJson(Map<String, dynamic> j)
      : sex = j['sex'] ?? 'unspecified',
        heightCm = (j['height_cm'] as num?)?.toDouble(),
        experience = j['experience'] ?? 'beginner',
        primaryGoal = j['primary_goal'] ?? 'general_health',
        daysPerWeek = j['days_per_week'] ?? 3,
        sessionMinutes = j['session_minutes'] ?? 60,
        availableEquipment = (j['available_equipment'] as List? ?? []).cast<String>(),
        limitations = j['limitations'] ?? '',
        unitSystem = j['unit_system'] ?? 'metric',
        isOnboarded = j['is_onboarded'] ?? false;

  bool get isImperial => unitSystem == 'imperial';
}

class ExerciseDetail {
  final String id;
  final String name;
  final String description;
  final List<String> instructions;
  final String pattern;
  final String primaryMuscle;
  final List<String> secondaryMuscles;
  final List<String> equipment;
  final String difficulty;
  final bool isUnilateral;
  final String videoUrl;

  ExerciseDetail.fromJson(Map<String, dynamic> j)
      : id = j['id'],
        name = j['name'],
        description = j['description'] ?? '',
        instructions = (j['instructions'] as List? ?? []).cast<String>(),
        pattern = j['pattern'] ?? '',
        primaryMuscle = j['primary_muscle'] ?? '',
        secondaryMuscles = (j['secondary_muscles'] as List? ?? []).cast<String>(),
        equipment = (j['equipment'] as List? ?? []).cast<String>(),
        difficulty = j['difficulty'] ?? '',
        isUnilateral = j['is_unilateral'] ?? false,
        videoUrl = j['video_url'] ?? '';
}

class BodyMetric {
  final String measuredOn;
  final double? weightKg;
  final double? bodyFatPct;
  final String note;

  BodyMetric.fromJson(Map<String, dynamic> j)
      : measuredOn = j['measured_on'],
        weightKg = (j['weight_kg'] as num?)?.toDouble(),
        bodyFatPct = (j['body_fat_pct'] as num?)?.toDouble(),
        note = j['note'] ?? '';
}

class SyncResult {
  final int newPersonalRecords;
  final List<UserAchievementInfo> unlockedAchievements;

  SyncResult.fromJson(Map<String, dynamic> j)
      : newPersonalRecords = j['new_personal_records'] ?? 0,
        unlockedAchievements =
            (j['unlocked_achievements'] as List? ?? []).map((a) => UserAchievementInfo.fromJson(a)).toList();

  bool get hasRewards => newPersonalRecords > 0 || unlockedAchievements.isNotEmpty;
}

class UserAchievementInfo {
  final String code;
  final String name;
  final String description;
  final String icon;
  final String? earnedAt;

  UserAchievementInfo.fromJson(Map<String, dynamic> j)
      : code = j['code'],
        name = j['name'],
        description = j['description'],
        icon = j['icon'],
        earnedAt = (j['earned_at'] as String?)?.isEmpty == true ? null : j['earned_at'];

  bool get isEarned => earnedAt != null;
}
