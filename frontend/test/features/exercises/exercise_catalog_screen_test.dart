import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/core/widgets/scale_on_press.dart';
import 'package:gym_tracker/core/widgets/staggered_entrance.dart';
import 'package:gym_tracker/features/exercises/exercise_catalog_screen.dart';

/// Verifies the Exercise catalog screen motion integration:
/// - Each catalog row is wrapped in ScaleOnPress.
/// - The catalog list is wrapped by StaggeredEntranceItem per row.
/// - The ListView carries a stable PageStorageKey for tab state restoration.
///
/// Covers Requirements: ScaleOnPress, StaggeredEntrance from the
/// `ui-motion` spec.
void main() {
  testWidgets('exposes a stable PageStorageKey on the catalog ListView',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: ExerciseCatalogScreen())),
    );
    await tester.pumpAndSettle();

    final list = tester.widget<ListView>(find.byType(ListView));
    expect(list.key, isA<PageStorageKey<String>>());
    expect((list.key as PageStorageKey<String>).value, 'exercises-list');
  });

  testWidgets('wraps catalog rows in ScaleOnPress + StaggeredEntrance',
      (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: ExerciseCatalogScreen())),
    );
    await tester.pumpAndSettle();

    // Catalog rows are wrapped by ScaleOnPress (one per row).
    expect(find.byType(ScaleOnPress), findsWidgets);
    // And by StaggeredEntranceItem (one per row).
    expect(find.byType(StaggeredEntranceItem), findsWidgets);
  });
}