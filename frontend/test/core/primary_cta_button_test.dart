import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/core/widgets/scale_on_press.dart';
import 'package:gym_tracker/core/widgets/primary_cta_button.dart';

/// Verifies the [PrimaryCtaButton] micro-interaction integration.
///
/// Covers Requirement: ScaleOnPress as applied to primary CTAs in the
/// `ui-motion` spec.
void main() {
  testWidgets('wraps the ElevatedButton in ScaleOnPress', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Center(
            child: PrimaryCtaButton(label: 'Start', onPressed: () {}),
          ),
        ),
      ),
    );

    expect(find.byType(ScaleOnPress), findsOneWidget);
    expect(find.byType(ElevatedButton), findsOneWidget);
    // ScaleOnPress is the parent of the ElevatedButton.
    expect(
      find.descendant(of: find.byType(ScaleOnPress), matching: find.byType(ElevatedButton)),
      findsOneWidget,
    );
  });

  testWidgets('invokes onPressed on tap', (tester) async {
    var taps = 0;
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Center(
            child:
                PrimaryCtaButton(label: 'Start', onPressed: () => taps++),
          ),
        ),
      ),
    );

    await tester.tap(find.byType(PrimaryCtaButton));
    await tester.pump();
    expect(taps, 1);
  });

  testWidgets('scales toward 0.97 on press-down', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Center(
            child: PrimaryCtaButton(label: 'Start', onPressed: () {}),
          ),
        ),
      ),
    );

    final center = tester.getCenter(find.byType(PrimaryCtaButton));
    final gesture = await tester.startGesture(center);
    await tester.pump();

    expect(
      tester.widget<AnimatedScale>(find.byType(AnimatedScale)).scale,
      ScaleOnPress.kPressedScale,
    );

    await gesture.up();
    await tester.pump();
    expect(
      tester.widget<AnimatedScale>(find.byType(AnimatedScale)).scale,
      1.0,
    );
  });
}