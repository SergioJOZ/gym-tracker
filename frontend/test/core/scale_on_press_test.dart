import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/core/theme/app_theme.dart';
import 'package:gym_tracker/core/widgets/scale_on_press.dart';

/// Verifies [ScaleOnPress] micro-interaction.
///
/// Covers Requirement: ScaleOnPress from the `ui-motion` spec.
void main() {
  testWidgets('scales to 0.97 on tap-down and back to 1.0 on release',
      (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Center(
            child: ScaleOnPress(
              child: Container(width: 100, height: 100, color: Colors.red),
            ),
          ),
        ),
      ),
    );

    final center = tester.getCenter(find.byType(ScaleOnPress));

    final gesture = await tester.startGesture(center);
    await tester.pump();

    final pressed = tester.widget<AnimatedScale>(find.byType(AnimatedScale));
    expect(pressed.scale, ScaleOnPress.kPressedScale);

    await gesture.up();
    await tester.pump();

    final released = tester.widget<AnimatedScale>(find.byType(AnimatedScale));
    expect(released.scale, 1.0);
  });

  testWidgets('restores 1.0 when tap is cancelled', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Center(
            child: ScaleOnPress(
              child: Container(width: 100, height: 100, color: Colors.red),
            ),
          ),
        ),
      ),
    );

    final center = tester.getCenter(find.byType(ScaleOnPress));

    final gesture = await tester.startGesture(center);
    await tester.pump();
    expect(
      tester.widget<AnimatedScale>(find.byType(AnimatedScale)).scale,
      ScaleOnPress.kPressedScale,
    );

    await gesture.cancel();
    await tester.pump();
    expect(
      tester.widget<AnimatedScale>(find.byType(AnimatedScale)).scale,
      1.0,
    );
  });

  testWidgets('does not scale under reduced-motion', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: MediaQuery(
          data: const MediaQueryData(disableAnimations: true),
          child: Scaffold(
            body: Center(
              child: ScaleOnPress(
                child: Container(width: 100, height: 100, color: Colors.red),
              ),
            ),
          ),
        ),
      ),
    );

    final center = tester.getCenter(find.byType(ScaleOnPress));

    final gesture = await tester.startGesture(center);
    await tester.pump();
    expect(
      tester.widget<AnimatedScale>(find.byType(AnimatedScale)).scale,
      1.0,
    );

    await gesture.up();
    await tester.pump();
    expect(
      tester.widget<AnimatedScale>(find.byType(AnimatedScale)).scale,
      1.0,
    );
  });

  testWidgets('invokes onTap callback when tapped', (tester) async {
    var taps = 0;
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Center(
            child: ScaleOnPress(
              onTap: () => taps++,
              child: Container(width: 100, height: 100, color: Colors.red),
            ),
          ),
        ),
      ),
    );

    await tester.tap(find.byType(ScaleOnPress));
    await tester.pump();
    expect(taps, 1);
  });

  test('MotionTokens integration: ScaleOnPress exposes pressed scale', () {
    expect(ScaleOnPress.kPressedScale, 0.97);
    expect(MotionTokens.spring, isA<Curve>());
  });
}