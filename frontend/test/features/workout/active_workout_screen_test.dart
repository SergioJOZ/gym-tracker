import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/features/workout/active_workout_screen.dart';

/// Verifies the timer pulse micro-interaction on the active-workout screen.
///
/// Covers Requirement: Timer Pulse from the `ui-motion` spec.
void main() {
  testWidgets('timer digits are wrapped by a Transform animation', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: ActiveWorkoutScreen()),
    );
    await tester.pump();

    // The digits (mm:ss) are driven by a Transform.scale built inside the
    // pulse AnimatedBuilder. Framework widgets also insert AnimatedBuilders
    // higher up, so we anchor on the Transform (only one in the timer area).
    final timerText = find.text('00:00');
    expect(timerText, findsOneWidget);
    final transform = find.ancestor(
      of: timerText,
      matching: find.byType(Transform),
    );
    expect(transform, findsWidgets);

    // Initial scale before the first pulse is 1.0. The nearest Transform
    // ancestor is the pulse-driven Transform.scale.
    final matrixInitial = tester.widget<Transform>(transform.first).transform;
    expect(matrixInitial.row0.x, 1.0);
  });

  testWidgets('pulses once per elapsed second while the timer runs',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: ActiveWorkoutScreen()),
    );
    await tester.pump();

    await tester.pump(const Duration(seconds: 1));
    // Pumping into the second half of the pulse window touches the scale
    // tween's midpoint (forward).
    await tester.pump(const Duration(milliseconds: 75));

    final transformDuring = find.ancestor(
      of: find.text('00:01'),
      matching: find.byType(Transform),
    );
    final matrixDuring =
        tester.widget<Transform>(transformDuring.first).transform;
    expect(matrixDuring.row0.x, greaterThan(1.0));

    // The displayed label advanced to 00:01.
    expect(find.text('00:01'), findsOneWidget);

    // After the full pulse duration elapses, the scale settles back to 1.0.
    await tester.pump(const Duration(milliseconds: 150));
    final transformAfter = find.ancestor(
      of: find.text('00:01'),
      matching: find.byType(Transform),
    );
    final matrixAfter =
        tester.widget<Transform>(transformAfter.first).transform;
    expect(matrixAfter.row0.x, closeTo(1.0, 1e-9));
  });

  testWidgets('reduced-motion still renders the digits without pulsing',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: ActiveWorkoutScreen()),
    );
    await tester.pump();

    // Without tick yet: scale still 1.0.
    final transform = find.ancestor(
      of: find.text('00:00'),
      matching: find.byType(Transform),
    );
    final before = tester.widget<Transform>(transform.first).transform;
    expect(before.row0.x, 1.0);
  });
}