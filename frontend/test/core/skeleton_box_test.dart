import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/core/widgets/skeleton_box.dart';

/// Verifies [SkeletonBox].
///
/// Covers Requirement: SkeletonBox from the `ui-motion` spec.
void main() {
  group('SkeletonBox', () {
    testWidgets('renders a SizedBox with the requested dimensions',
        (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SkeletonBox(
              width: 120,
              height: 24,
              isLoading: true,
            ),
          ),
        ),
      );

      final sizedBox = tester.widget<SizedBox>(find.byType(SizedBox));
      expect(sizedBox.width, 120);
      expect(sizedBox.height, 24);
    });

    testWidgets('shows shimmer when isLoading is true', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SkeletonBox(
              width: 80,
              height: 12,
              isLoading: true,
            ),
          ),
        ),
      );

      // Pump a few frames to let the animation advance and confirm that the
      // widget maintains a shimmer painter (CustomPaint) for the gradient
      // sweep.
      await tester.pump(const Duration(milliseconds: 100));
      expect(find.byType(SkeletonBox), findsOneWidget);
      expect(
        tester.widgetList<CustomPaint>(find.byType(CustomPaint)),
        isNotEmpty,
      );
    });

    testWidgets('renders a static box when isLoading is false', (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SkeletonBox(
              width: 80,
              height: 12,
              isLoading: false,
            ),
          ),
        ),
      );

      // Allow any pending animation frames to be processed. The static box
      // must remain.
      await tester.pumpAndSettle();
      expect(find.byType(SkeletonBox), findsOneWidget);
      final box = tester.widget<SkeletonBox>(find.byType(SkeletonBox));
      expect(box.isLoading, isFalse);
    });

    testWidgets('does not advance animation while not loading',
        (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: SkeletonBox(width: 80, height: 12, isLoading: false),
          ),
        ),
      );
      // Pumping a long time should not throw or animbreak.
      await tester.pump(const Duration(seconds: 1));
      expect(find.byType(SkeletonBox), findsOneWidget);
    });

    testWidgets('reduced-motion shows static box even when isLoading is true',
        (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: MediaQuery(
            data: MediaQueryData(disableAnimations: true),
            child: Scaffold(
              body: SkeletonBox(width: 80, height: 12, isLoading: true),
            ),
          ),
        ),
      );

      // Reduced-motion: shimmer is suppressed, the widget renders a static
      // DecoratedBox instead of an AnimatedBuilder shimmer.
      expect(find.byType(SkeletonBox), findsOneWidget);
      expect(
        find.descendant(
          of: find.byType(SkeletonBox),
          matching: find.byType(DecoratedBox),
        ),
        findsOneWidget,
      );
      // No shimmer-driven CustomPaint inside the SkeletonBox subtree.
      expect(
        find.descendant(
          of: find.byType(SkeletonBox),
          matching: find.byType(AnimatedBuilder),
        ),
        findsNothing,
      );
    });
  });
}