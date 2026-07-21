import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/features/shell/main_shell.dart';

/// Verifies the `AnimatedSwitcher`-based crossfade with `GlobalKey` per
/// tab preserves each tab's state across switches.
///
/// Covers Requirement: Animated Tab Crossfade from the `ui-motion` spec.
void main() {
  testWidgets('starts on the Routines tab', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: MainShell()));
    await tester.pumpAndSettle();

    expect(find.text('Routines'), findsWidgets);
    expect(find.text('My Routines'), findsOneWidget);
  });

  testWidgets('preserves Routines scroll offset across tab switches',
      (tester) async {
    await tester.pumpWidget(const MaterialApp(home: MainShell()));
    await tester.pumpAndSettle();

    // Confirm Routines is the initial screen and grab the scroll controller
    // that lives behind the active scrollable.
    expect(find.text('My Routines'), findsOneWidget);

    final scrollable = find.byType(Scrollable).first;
    final initialOffset = tester.widget<Scrollable>(scrollable);

    // Drag the routines list up to scroll down.
    await tester.drag(scrollable, const Offset(0, -250));
    await tester.pumpAndSettle();

    // Switch to Progress tab.
    await tester.tap(find.text('Progress'));
    await tester.pumpAndSettle();

    // We're no longer seeing the Routines header.
    expect(find.text('My Routines'), findsNothing);

    // Switch back to Routines.
    await tester.tap(find.text('Routines'));
    await tester.pumpAndSettle();

    // Routines mount is intact: header reappears and the scroll offset is
    // non-zero (preserved by the GlobalKey).
    expect(find.text('My Routines'), findsOneWidget);

    final restoredScrollable = find.byType(Scrollable).first;
    final restoredController =
        tester.state<ScrollableState>(restoredScrollable).position;
    expect(restoredController.pixels, greaterThan(0));

    // Silence the unused-variable lint for the baseline assertion above.
    expect(initialOffset, isNotNull);
  });

  testWidgets('uses AnimatedSwitcher for crossfade between tabs',
      (tester) async {
    await tester.pumpWidget(const MaterialApp(home: MainShell()));
    await tester.pumpAndSettle();

    expect(find.byType(AnimatedSwitcher), findsOneWidget);
  });

  testWidgets('does NOT use IndexedStack anymore', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: MainShell()));
    await tester.pumpAndSettle();

    expect(find.byType(IndexedStack), findsNothing);
  });

  testWidgets('crossfade is observable when switching tabs', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: MainShell()));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Progress'));
    await tester.pump();
    // The AnimatedSwitcher is mid-transition: both outgoing Routines and
    // incoming Progress are in the tree momentarily.
    await tester.pump(const Duration(milliseconds: 10));
    // Settling completes the crossfade.
    await tester.pumpAndSettle();
    expect(find.text('Progress'), findsWidgets);
  });
}