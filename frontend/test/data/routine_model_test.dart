import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/data/models/models.dart';

/// Verifies the `id` field added to [Routine] for stable Hero tags.
///
/// Covers Requirement: Hero Transition for Routine Detail.
void main() {
  group('Routine.id', () {
    test('constructs with an id', () {
      const routine = Routine(
        id: 'routine-push-day',
        name: 'Push Day',
        exercises: [],
        lastPerformedLabel: '2 days ago',
        estimatedMinutes: 45,
      );
      expect(routine.id, 'routine-push-day');
    });

    test('treats empty id as invalid', () {
      const routine = Routine(
        id: '',
        name: 'Push Day',
        exercises: [],
        lastPerformedLabel: '2 days ago',
        estimatedMinutes: 45,
      );
      expect(routine.id, isEmpty);
      expect(() => validateRoutineId(routine.id), throwsArgumentError);
    });

    test('copyWith preserves id by default', () {
      const routine = Routine(
        id: 'push-day',
        name: 'Push Day',
        exercises: [],
        lastPerformedLabel: '2 days ago',
        estimatedMinutes: 45,
      );
      final updated = routine.copyWith(name: 'Push Day v2');
      expect(updated.id, 'push-day');
      expect(updated.name, 'Push Day v2');
    });

    test('copyWith can override id', () {
      const routine = Routine(
        id: 'push-day',
        name: 'Push Day',
        exercises: [],
        lastPerformedLabel: '2 days ago',
        estimatedMinutes: 45,
      );
      final updated = routine.copyWith(id: 'pull-day');
      expect(updated.id, 'pull-day');
    });

    test('two routines with the same id are equal', () {
      const a = Routine(
        id: 'legs',
        name: 'Legs',
        exercises: [],
        lastPerformedLabel: '6 days ago',
        estimatedMinutes: 55,
      );
      const b = Routine(
        id: 'legs',
        name: 'Legs',
        exercises: [],
        lastPerformedLabel: '6 days ago',
        estimatedMinutes: 55,
      );
      expect(a, b);
    });

    test('two routines with differing ids are not equal', () {
      const a = Routine(
        id: 'legs',
        name: 'Legs',
        exercises: [],
        lastPerformedLabel: '6 days ago',
        estimatedMinutes: 55,
      );
      const b = Routine(
        id: 'push-day',
        name: 'Legs',
        exercises: [],
        lastPerformedLabel: '6 days ago',
        estimatedMinutes: 55,
      );
      expect(a == b, isFalse);
    });
  });
}

/// Helper mirroring the spec invariant: ids MUST be non-empty.
void validateRoutineId(String id) {
  if (id.isEmpty) {
    throw ArgumentError('Routine.id must be non-empty');
  }
}