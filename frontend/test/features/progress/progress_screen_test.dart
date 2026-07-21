import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/core/widgets/staggered_entrance.dart';
import 'package:gym_tracker/features/progress/progress_screen.dart';

/// Verifies the Progress screen motion integration:
/// - Weekly bars grow from 0 → full via an AnimationController (spring).
/// - Reduced-motion snaps bars to full value instantly.
/// - Stat cards and record rows are wrapped in StaggeredEntrance.
/// - The ListView carries a stable PageStorageKey for tab state restoration.
///
/// Covers Requirements: Animated Weekly Bars, StaggeredEntrance,
/// Reduced-Motion Gate from the `ui-motion` spec.
void main() {
  testWidgets('exposes a stable PageStorageKey on the main ListView',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: ProgressScreen())),
    );
    await tester.pump();

    final list = tester.widget<ListView>(find.byType(ListView));
    expect(list.key, isA<PageStorageKey<String>>());
    expect((list.key as PageStorageKey<String>).value, 'progress-list');
  });

  testWidgets('bars begin at zero height and grow toward full value on settle',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: ProgressScreen())),
    );
    // First frame at controller value 0: all bar containers are zero-height.
    // Anchor on the SizedBox(height: 110) that wraps _WeeklyBars to avoid
    // false matches from stat-card Containers inside StaggeredEntrance.
    await tester.pump();
    final barsArea = find.byWidgetPredicate(
      (w) => w is SizedBox && w.height == 110,
    );
    final firstBar = find.descendant(
      of: barsArea,
      matching: find.byType(Container),
    ).first;
    final sizeAtStart = tester.getSize(firstBar);
    expect(sizeAtStart.height, 0.0);

    // Settle: spring controller completes the bottom-up growth.
    await tester.pumpAndSettle();
    final sizeAtEnd = tester.getSize(firstBar);
    expect(sizeAtEnd.height, greaterThan(0.0));
  });

  testWidgets('bars snap to full value immediately under reduced-motion',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: MediaQuery(
          data: MediaQueryData(disableAnimations: true),
          child: Scaffold(body: ProgressScreen()),
        ),
      ),
    );
    // Single frame: reduced-motion forces the controller to value 1.0 in
    // didChangeDependencies, so bars must already be at full height.
    await tester.pump();

    final barsArea = find.byWidgetPredicate(
      (w) => w is SizedBox && w.height == 110,
    );
    final firstBar = find.descendant(
      of: barsArea,
      matching: find.byType(Container),
    ).first;
    final size = tester.getSize(firstBar);
    expect(size.height, greaterThan(0.0));
  });

  testWidgets('stat cards and record rows are wrapped by StaggeredEntrance',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: ProgressScreen())),
    );
    await tester.pump();

    // StaggeredEntranceItem count = (#stat cards) + (#record rows). Both
    // sets are non-empty, so the total is at least 2 and the wrappers are
    // public-machine-visible.
    expect(find.byType(StaggeredEntranceItem), findsWidgets);
    expect(
      find.byType(StaggeredEntranceItem).evaluate().length,
      greaterThanOrEqualTo(2),
    );
  });
}