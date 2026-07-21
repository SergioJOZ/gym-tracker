import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/core/widgets/staggered_entrance.dart';

/// Verifies [StaggeredEntrance].
///
/// Covers Requirement: StaggeredEntrance from the `ui-motion` spec.
void main() {
  group('StaggeredEntrance.wrap', () {
    test('returns an empty list for zero children', () {
      final wrapped = StaggeredEntrance.wrap(const <Widget>[]);
      expect(wrapped, isEmpty);
    });

    test('returns one StaggeredEntranceItem for one child', () {
      final wrapped = StaggeredEntrance.wrap(
        const [SizedBox(key: ValueKey('only'))],
      );
      expect(wrapped, hasLength(1));
      expect(wrapped.first, isA<StaggeredEntranceItem>());
    });

    test('returns N items for N children', () {
      final children = List.generate(
        5,
        (i) => SizedBox(key: ValueKey('c$i')),
      );
      final wrapped = StaggeredEntrance.wrap(children);
      expect(wrapped, hasLength(5));
      for (final w in wrapped) {
        expect(w, isA<StaggeredEntranceItem>());
      }
    });

    testWidgets('per-child delay equals i × stagger', (tester) async {
      const stagger = Duration(milliseconds: 40);
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Column(
              children: StaggeredEntrance.wrap(
                [
                  const SizedBox(),
                  const SizedBox(),
                  const SizedBox(),
                ],
                stagger: stagger,
              ),
            ),
          ),
        ),
      );

      final items = tester
          .widgetList<StaggeredEntranceItem>(find.byType(StaggeredEntranceItem))
          .map((w) => w.delay)
          .toList()
        ..sort();

      expect(items, hasLength(3));
      expect(items[0], Duration.zero);
      expect(items[1], stagger);
      expect(items[2], stagger * 2);
    });

    testWidgets('mounts three TweenAnimationBuilder widgets', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Column(
              children: StaggeredEntrance.wrap(const [
                SizedBox(),
                SizedBox(),
                SizedBox(),
              ]),
            ),
          ),
        ),
      );

      expect(
        find.byType(TweenAnimationBuilder<double>),
        findsNWidgets(3),
      );
    });

    testWidgets('handles rebuild with a different child count', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Column(
              children: StaggeredEntrance.wrap(const [SizedBox()]),
            ),
          ),
        ),
      );
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: Column(
              children: StaggeredEntrance.wrap(const [SizedBox(), SizedBox()]),
            ),
          ),
        ),
      );
      expect(
        find.byType(StaggeredEntranceItem),
        findsNWidgets(2),
      );
    });
  });

  group('StaggeredEntrance.delayFor', () {
    test('i × stagger formula', () {
      const stagger = Duration(milliseconds: 40);
      expect(StaggeredEntrance.delayFor(0, stagger), Duration.zero);
      expect(
        StaggeredEntrance.delayFor(1, stagger),
        const Duration(milliseconds: 40),
      );
      expect(
        StaggeredEntrance.delayFor(4, stagger),
        const Duration(milliseconds: 160),
      );
    });
  });
}