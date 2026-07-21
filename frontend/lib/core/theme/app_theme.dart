import 'package:flutter/material.dart';

import 'app_colors.dart';
import 'app_radius.dart';

/// Builds the single dark-only [ThemeData] for the app.
abstract final class AppTheme {
  static ThemeData get dark {
    const colorScheme = ColorScheme.dark(
      primary: AppColors.accent,
      onPrimary: Colors.white,
      secondary: AppColors.green,
      onSecondary: AppColors.onGreen,
      surface: AppColors.card,
      onSurface: AppColors.text,
      outline: AppColors.border,
    );

    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      colorScheme: colorScheme,
      scaffoldBackgroundColor: AppColors.scaffold,
      textTheme: _textTheme,
      cardTheme: _cardTheme,
      dividerTheme: _dividerTheme,
      snackBarTheme: _snackBarTheme,
      navigationBarTheme: _navigationBarTheme,
    );
  }

  static const TextTheme _textTheme = TextTheme(
    headlineLarge: TextStyle(
      fontSize: 28,
      fontWeight: FontWeight.w800,
      letterSpacing: -0.5,
      color: AppColors.text,
    ),
    titleLarge: TextStyle(
      fontSize: 17,
      fontWeight: FontWeight.w600,
      color: AppColors.text,
    ),
    titleMedium: TextStyle(
      fontSize: 16,
      fontWeight: FontWeight.w700,
      color: AppColors.text,
    ),
    titleSmall: TextStyle(
      fontSize: 15,
      fontWeight: FontWeight.w700,
      color: AppColors.text,
    ),
    bodyLarge: TextStyle(
      fontSize: 15,
      fontWeight: FontWeight.w600,
      color: AppColors.text,
    ),
    bodyMedium: TextStyle(
      fontSize: 14,
      fontWeight: FontWeight.w500,
      color: AppColors.text,
    ),
    bodySmall: TextStyle(
      fontSize: 12.5,
      color: AppColors.textSecondary,
    ),
    labelLarge: TextStyle(
      fontSize: 14,
      fontWeight: FontWeight.w700,
      color: AppColors.text,
    ),
    labelSmall: TextStyle(
      fontSize: 11,
      fontWeight: FontWeight.w600,
      letterSpacing: 0.4,
      color: AppColors.textSecondary,
    ),
  );

  static const CardThemeData _cardTheme = CardThemeData(
    color: AppColors.card,
    elevation: 0,
    margin: EdgeInsets.zero,
    clipBehavior: Clip.antiAlias,
    shape: RoundedRectangleBorder(
      borderRadius: BorderRadius.all(Radius.circular(AppRadius.card)),
      side: BorderSide(color: AppColors.border),
    ),
  );

  static const DividerThemeData _dividerTheme = DividerThemeData(
    color: AppColors.border,
    thickness: 1,
    space: 1,
  );

  static const SnackBarThemeData _snackBarTheme = SnackBarThemeData(
    behavior: SnackBarBehavior.floating,
    backgroundColor: AppColors.surface2,
    contentTextStyle: TextStyle(color: AppColors.text, fontSize: 14),
  );

  static final NavigationBarThemeData _navigationBarTheme =
      NavigationBarThemeData(
    backgroundColor: AppColors.scaffold,
    elevation: 0,
    height: 64,
    indicatorColor: AppColors.accentDim,
    surfaceTintColor: Colors.transparent,
    labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
    iconTheme: WidgetStateProperty.resolveWith(
      (states) => IconThemeData(
        size: 22,
        color: states.contains(WidgetState.selected)
            ? AppColors.accent
            : AppColors.textTertiary,
      ),
    ),
    labelTextStyle: WidgetStateProperty.resolveWith(
      (states) => TextStyle(
        fontSize: 11,
        fontWeight: FontWeight.w600,
        color: states.contains(WidgetState.selected)
            ? AppColors.accent
            : AppColors.textTertiary,
      ),
    ),
  );
}

/// Centralized motion tokens and the single reduced-motion gate.
///
/// All motion components MUST consume durations/curves from here and MUST
/// consult [disabled] (or [resolve]) before animating. Raw `Duration` or
/// curve literals in feature code are disallowed by the `ui-motion` spec.
abstract final class MotionTokens {
  static const Duration fast = Duration(milliseconds: 150);
  static const Duration medium = Duration(milliseconds: 250);
  static const Duration slow = Duration(milliseconds: 400);

  /// Per-item stagger delay for list entrance animations.
  static const Duration stagger = Duration(milliseconds: 40);

  /// Subtle spring curve for press feedback and bar entrance.
  static const Curve spring = Curves.easeOutBack;

  /// Default ease for crossfades and general transitions.
  static const Curve standard = Curves.easeOutCubic;

  /// Single reduced-motion gate. Returns true when the platform requested
  /// `disableAnimations` or `accessibleNavigation`. Returns false when no
  /// [MediaQuery] is available (no ancestor context).
  static bool disabled(BuildContext context) {
    final mq = MediaQuery.maybeOf(context);
    if (mq == null) return false;
    return mq.disableAnimations || mq.accessibleNavigation;
  }

  /// Returns [duration] clamped to [Duration.zero] when reduced-motion is
  /// active. Otherwise returns [duration] unchanged.
  static Duration resolve(BuildContext context, Duration duration) {
    return disabled(context) ? Duration.zero : duration;
  }
}
