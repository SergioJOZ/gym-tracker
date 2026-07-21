import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/core/theme/app_theme.dart';

/// Verifies the centralized reduced-motion gate [MotionTokens].
///
/// Covers Requirement: Reduced-Motion Gate and Requirement: MotionTokens
/// from the `ui-motion` spec.
void main() {
  group('MotionTokens.disabled', () {
    testWidgets(
      'returns true when MediaQuery.disableAnimations is true',
      (tester) async {
        bool? captured;
        await tester.pumpWidget(
          MaterialApp(
            home: MediaQuery(
              data: const MediaQueryData(disableAnimations: true),
              child: Builder(
                builder: (context) {
                  captured = MotionTokens.disabled(context);
                  return const SizedBox.shrink();
                },
              ),
            ),
          ),
        );
        expect(captured, isTrue);
      },
    );

    testWidgets(
      'returns true when MediaQuery.accessibleNavigation is true',
      (tester) async {
        bool? captured;
        await tester.pumpWidget(
          MaterialApp(
            home: MediaQuery(
              data: const MediaQueryData(accessibleNavigation: true),
              child: Builder(
                builder: (context) {
                  captured = MotionTokens.disabled(context);
                  return const SizedBox.shrink();
                },
              ),
            ),
          ),
        );
        expect(captured, isTrue);
      },
    );

    testWidgets(
      'returns false when both flags are false/null',
      (tester) async {
        bool? captured;
        await tester.pumpWidget(
          MaterialApp(
            home: MediaQuery(
              data: const MediaQueryData(),
              child: Builder(
                builder: (context) {
                  captured = MotionTokens.disabled(context);
                  return const SizedBox.shrink();
                },
              ),
            ),
          ),
        );
        expect(captured, isFalse);
      },
    );

    testWidgets(
      'returns false when MediaQuery is absent (maybeOf == null)',
      (tester) async {
        bool? captured;
        await tester.pumpWidget(
          Builder(
            builder: (context) {
              captured = MotionTokens.disabled(context);
              return const SizedBox.shrink();
            },
          ),
        );
        expect(captured, isFalse);
      },
    );
  });

  group('MotionTokens.resolve', () {
    testWidgets(
      'returns Duration.zero when reduced-motion is active',
      (tester) async {
        Duration? captured;
        const duration = Duration(milliseconds: 250);
        await tester.pumpWidget(
          MaterialApp(
            home: MediaQuery(
              data: const MediaQueryData(disableAnimations: true),
              child: Builder(
                builder: (context) {
                  captured = MotionTokens.resolve(context, duration);
                  return const SizedBox.shrink();
                },
              ),
            ),
          ),
        );
        expect(captured, Duration.zero);
      },
    );

    testWidgets(
      'returns the provided duration when reduced-motion is inactive',
      (tester) async {
        Duration? captured;
        const duration = Duration(milliseconds: 250);
        await tester.pumpWidget(
          MaterialApp(
            home: MediaQuery(
              data: const MediaQueryData(),
              child: Builder(
                builder: (context) {
                  captured = MotionTokens.resolve(context, duration);
                  return const SizedBox.shrink();
                },
              ),
            ),
          ),
        );
        expect(captured, duration);
      },
    );
  });

  test('MotionTokens exposes duration and curve tokens', () {
    expect(MotionTokens.fast, const Duration(milliseconds: 150));
    expect(MotionTokens.medium, const Duration(milliseconds: 250));
    expect(MotionTokens.slow, const Duration(milliseconds: 400));
    expect(MotionTokens.stagger, const Duration(milliseconds: 40));
    expect(MotionTokens.standard, isA<Curve>());
    expect(MotionTokens.spring, isA<Curve>());
  });
}