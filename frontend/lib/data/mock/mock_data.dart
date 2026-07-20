import '../models/models.dart';

/// Static mock data. The backend is not wired up yet; every screen reads
/// from here.
abstract final class MockData {
  static const List<Routine> routines = [
    Routine(
      name: 'Push Day',
      lastPerformedLabel: '2 days ago',
      estimatedMinutes: 45,
      exercises: [
        RoutineExercise(
          name: 'Bench Press (Barbell)',
          sets: 4,
          reps: '8',
          restSeconds: 90,
        ),
        RoutineExercise(
          name: 'Overhead Press (Barbell)',
          sets: 3,
          reps: '10',
          restSeconds: 90,
        ),
        RoutineExercise(
          name: 'Incline Dumbbell Press',
          sets: 3,
          reps: '10–12',
          restSeconds: 60,
        ),
        RoutineExercise(
          name: 'Lateral Raise (Dumbbell)',
          sets: 3,
          reps: '15',
          restSeconds: 60,
        ),
      ],
    ),
    Routine(
      name: 'Pull Day',
      lastPerformedLabel: '4 days ago',
      estimatedMinutes: 50,
      exercises: [
        RoutineExercise(
          name: 'Deadlift (Barbell)',
          sets: 3,
          reps: '5',
          restSeconds: 180,
        ),
        RoutineExercise(
          name: 'Pull-Up',
          sets: 3,
          reps: '8',
          restSeconds: 90,
        ),
        RoutineExercise(
          name: 'Barbell Row',
          sets: 4,
          reps: '10',
          restSeconds: 90,
        ),
        RoutineExercise(
          name: 'Face Pull',
          sets: 3,
          reps: '15',
          restSeconds: 60,
        ),
        RoutineExercise(
          name: 'Bicep Curl (Dumbbell)',
          sets: 3,
          reps: '12',
          restSeconds: 60,
        ),
      ],
    ),
    Routine(
      name: 'Legs',
      lastPerformedLabel: '6 days ago',
      estimatedMinutes: 55,
      exercises: [
        RoutineExercise(
          name: 'Back Squat (Barbell)',
          sets: 4,
          reps: '8',
          restSeconds: 120,
        ),
        RoutineExercise(
          name: 'Romanian Deadlift',
          sets: 3,
          reps: '10',
          restSeconds: 90,
        ),
        RoutineExercise(
          name: 'Leg Press',
          sets: 3,
          reps: '12',
          restSeconds: 90,
        ),
        RoutineExercise(
          name: 'Calf Raise',
          sets: 4,
          reps: '15',
          restSeconds: 45,
        ),
      ],
    ),
  ];

  static const List<Exercise> exerciseCatalog = [
    Exercise(name: 'Bench Press', muscleGroup: 'Chest', equipment: 'Barbell'),
    Exercise(name: 'Back Squat', muscleGroup: 'Legs', equipment: 'Barbell'),
    Exercise(name: 'Deadlift', muscleGroup: 'Back', equipment: 'Barbell'),
    Exercise(name: 'Pull-Up', muscleGroup: 'Back', equipment: 'Bodyweight'),
    Exercise(
      name: 'Overhead Press',
      muscleGroup: 'Shoulders',
      equipment: 'Barbell',
    ),
    Exercise(name: 'Bicep Curl', muscleGroup: 'Arms', equipment: 'Dumbbell'),
    Exercise(name: 'Barbell Row', muscleGroup: 'Back', equipment: 'Barbell'),
    Exercise(
      name: 'Incline Dumbbell Press',
      muscleGroup: 'Chest',
      equipment: 'Dumbbell',
    ),
    Exercise(
      name: 'Lateral Raise',
      muscleGroup: 'Shoulders',
      equipment: 'Dumbbell',
    ),
    Exercise(
      name: 'Romanian Deadlift',
      muscleGroup: 'Legs',
      equipment: 'Barbell',
    ),
    Exercise(name: 'Leg Press', muscleGroup: 'Legs', equipment: 'Machine'),
    Exercise(name: 'Calf Raise', muscleGroup: 'Legs', equipment: 'Machine'),
    Exercise(name: 'Face Pull', muscleGroup: 'Back', equipment: 'Cable'),
    Exercise(
      name: 'Tricep Pushdown',
      muscleGroup: 'Arms',
      equipment: 'Cable',
    ),
  ];

  static const ProgressSummary progress = ProgressSummary(
    totalWorkouts: '47',
    workoutsThisWeek: '3',
    totalVolume: '128t',
    weeklyFractions: [0.35, 0.55, 0.25, 0.70, 0.90, 0.40, 0.15],
    dayLabels: ['M', 'T', 'W', 'T', 'F', 'S', 'S'],
  );

  static const List<PersonalRecord> personalRecords = [
    PersonalRecord(
      exerciseName: 'Bench Press',
      detail: 'Max weight · Jul 18',
      value: '80 kg',
    ),
    PersonalRecord(
      exerciseName: 'Back Squat',
      detail: 'Max weight · Jul 15',
      value: '100 kg',
    ),
    PersonalRecord(
      exerciseName: 'Pull-Up',
      detail: 'Max reps · Jul 12',
      value: '15',
    ),
  ];
}
