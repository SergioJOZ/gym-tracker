/// Domain models for the gym-tracker frontend.
///
/// Everything is immutable with const constructors, except [WorkoutSet],
/// which is intentionally mutable because the active workout screen edits
/// sets in place.
library;

/// An exercise in the catalog.
class Exercise {
  final String name;
  final String muscleGroup;
  final String equipment;

  const Exercise({
    required this.name,
    required this.muscleGroup,
    required this.equipment,
  });
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
  final String name;
  final List<RoutineExercise> exercises;

  /// Display label such as "2 days ago".
  final String lastPerformedLabel;
  final int estimatedMinutes;

  const Routine({
    required this.name,
    required this.exercises,
    required this.lastPerformedLabel,
    required this.estimatedMinutes,
  });
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
