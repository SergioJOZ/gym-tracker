/// Domain models for the gym-tracker frontend.
///
/// Everything is immutable with const constructors, except [WorkoutSet],
/// which is intentionally mutable because the active workout screen edits
/// sets in place.
library;

/// An exercise in the catalog.
class Exercise {
  final String id;
  final Map<String, String> nameByLang;

  /// English name (shorthand for [nameByLang] access).
  String get name => nameByLang['en'] ?? '';

  final String muscleGroup;
  final String equipment;

  const Exercise({
    required this.id,
    required this.nameByLang,
    required this.muscleGroup,
    required this.equipment,
  });

  factory Exercise.fromJson(Map<String, dynamic> json) => Exercise(
    id: json['id'] as String,
    nameByLang: Map<String, String>.from(json['name_by_lang'] as Map),
    muscleGroup: json['muscle_group'] as String? ?? '',
    equipment: json['equipment'] as String? ?? '',
  );
}

/// An exercise slot inside a routine, with its target scheme.
class RoutineExercise {
  final String name;
  final int sets;

  /// Target reps as shown in the UI, e.g. "8" or "10–12".
  final String reps;
  final int restSeconds;

  const RoutineExercise({
    required this.name,
    required this.sets,
    required this.reps,
    required this.restSeconds,
  });

  String get scheme => '$sets sets × $reps reps';
  String get restLabel => 'Rest ${restSeconds}s';
}

/// A saved workout routine.
class Routine {
  /// Stable unique identifier used as the Hero tag for shared-element
  /// transitions to [RoutineDetailScreen]. MUST be non-empty.
  final String id;
  final String name;
  final List<RoutineExercise> exercises;

  /// Display label such as "2 days ago".
  final String lastPerformedLabel;
  final int estimatedMinutes;

  const Routine({
    required this.id,
    required this.name,
    required this.exercises,
    required this.lastPerformedLabel,
    required this.estimatedMinutes,
  });

  Routine copyWith({
    String? id,
    String? name,
    List<RoutineExercise>? exercises,
    String? lastPerformedLabel,
    int? estimatedMinutes,
  }) {
    return Routine(
      id: id ?? this.id,
      name: name ?? this.name,
      exercises: exercises ?? this.exercises,
      lastPerformedLabel: lastPerformedLabel ?? this.lastPerformedLabel,
      estimatedMinutes: estimatedMinutes ?? this.estimatedMinutes,
    );
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is Routine &&
          runtimeType == other.runtimeType &&
          id == other.id &&
          name == other.name &&
          exercises == other.exercises &&
          lastPerformedLabel == other.lastPerformedLabel &&
          estimatedMinutes == other.estimatedMinutes;

  @override
  int get hashCode => Object.hash(
        id,
        name,
        exercises,
        lastPerformedLabel,
        estimatedMinutes,
      );
}

/// A single set inside an active workout. Mutable on purpose: kg, reps and
/// the done flag change as the user logs the session.
class WorkoutSet {
  double? kg;
  int? reps;
  bool done;

  WorkoutSet({this.kg, this.reps, this.done = false});
}

/// A personal record entry for the progress screen.
class PersonalRecord {
  final String exerciseName;

  /// Display label such as "Max weight · Jul 18".
  final String detail;

  /// Display value such as "80 kg" or "15".
  final String value;

  const PersonalRecord({
    required this.exerciseName,
    required this.detail,
    required this.value,
  });
}

/// Aggregated stats shown on the progress screen.
class ProgressSummary {
  final String totalWorkouts;
  final String workoutsThisWeek;
  final String totalVolume;

  /// Bar heights as fractions (0.0–1.0) of the chart area, Monday first.
  final List<double> weeklyFractions;
  final List<String> dayLabels;

  const ProgressSummary({
    required this.totalWorkouts,
    required this.workoutsThisWeek,
    required this.totalVolume,
    required this.weeklyFractions,
    required this.dayLabels,
  });
}
