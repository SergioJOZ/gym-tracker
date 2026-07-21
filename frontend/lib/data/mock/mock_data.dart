import '../models/models.dart';

/// Static mock data. The backend is not wired up yet; every screen reads
/// from here.
abstract final class MockData {
  static const List<Routine> routines = [
    Routine(
      id: 'routine-push-day',
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
      id: 'routine-pull-day',
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
      id: 'routine-legs',
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
    Exercise(id: 'ex-001', nameByLang: {'en': 'Bench Press'}, muscleGroup: 'Chest', equipment: 'Barbell'),
    Exercise(id: 'ex-002', nameByLang: {'en': 'Back Squat'}, muscleGroup: 'Legs', equipment: 'Barbell'),
    Exercise(id: 'ex-003', nameByLang: {'en': 'Deadlift'}, muscleGroup: 'Back', equipment: 'Barbell'),
    Exercise(id: 'ex-004', nameByLang: {'en': 'Pull-Up'}, muscleGroup: 'Back', equipment: 'Bodyweight'),
    Exercise(id: 'ex-005', nameByLang: {'en': 'Overhead Press'}, muscleGroup: 'Shoulders', equipment: 'Barbell'),
    Exercise(id: 'ex-006', nameByLang: {'en': 'Bicep Curl'}, muscleGroup: 'Arms', equipment: 'Dumbbell'),
    Exercise(id: 'ex-007', nameByLang: {'en': 'Barbell Row'}, muscleGroup: 'Back', equipment: 'Barbell'),
    Exercise(id: 'ex-008', nameByLang: {'en': 'Incline Dumbbell Press'}, muscleGroup: 'Chest', equipment: 'Dumbbell'),
    Exercise(id: 'ex-009', nameByLang: {'en': 'Lateral Raise'}, muscleGroup: 'Shoulders', equipment: 'Dumbbell'),
    Exercise(id: 'ex-010', nameByLang: {'en': 'Romanian Deadlift'}, muscleGroup: 'Legs', equipment: 'Barbell'),
    Exercise(id: 'ex-011', nameByLang: {'en': 'Leg Press'}, muscleGroup: 'Legs', equipment: 'Machine'),
    Exercise(id: 'ex-012', nameByLang: {'en': 'Calf Raise'}, muscleGroup: 'Legs', equipment: 'Machine'),
    Exercise(id: 'ex-013', nameByLang: {'en': 'Face Pull'}, muscleGroup: 'Back', equipment: 'Cable'),
    Exercise(id: 'ex-014', nameByLang: {'en': 'Tricep Pushdown'}, muscleGroup: 'Arms', equipment: 'Cable'),
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
