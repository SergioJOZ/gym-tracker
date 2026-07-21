import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/core/widgets/staggered_entrance.dart';
import 'package:gym_tracker/features/profile/profile_screen.dart';

/// Verifies the Profile screen motion integration:
/// - Centered column children are wrapped by StaggeredEntrance.
/// - The column carries a stable PageStorageKey for tab state restoration.
///
/// Covers Requirement: StaggeredEntrance from the `ui-motion` spec.
void main() {
  testWidgets('exposes a stable PageStorageKey on the centered column',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: ProfileScreen())),
    );
    await tester.pump();

    final column = tester.widget<Column>(find.byType(Column));
    expect(column.key, isA<PageStorageKey<String>>());
    expect((column.key as PageStorageKey<String>).value, 'profile-content');
  });

  testWidgets('wraps centered content children by StaggeredEntrance',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: ProfileScreen())),
    );
    await tester.pump();

    expect(find.byType(StaggeredEntranceItem), findsWidgets);
  });
}