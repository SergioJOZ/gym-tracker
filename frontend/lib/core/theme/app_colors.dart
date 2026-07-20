import 'package:flutter/material.dart';

/// Design tokens from the approved UI mockup (frontend/mockups/preview.html).
///
/// All colors carry a full alpha byte so no opacity helpers are needed.
abstract final class AppColors {
  static const Color scaffold = Color(0xFF141417);
  static const Color card = Color(0xFF1E1E23);
  static const Color surface2 = Color(0xFF26262C);
  static const Color border = Color(0xFF303036);

  static const Color text = Color(0xFFFFFFFF);
  static const Color textSecondary = Color(0xFF9C9CA5);
  static const Color textTertiary = Color(0xFF6B6B74);

  static const Color accent = Color(0xFF4A9EFF);

  /// Accent at 14% opacity, used for the PR icon background and the
  /// NavigationBar indicator.
  static const Color accentDim = Color(0x244A9EFF);

  static const Color green = Color(0xFF4ADE80);

  /// Green at 12% opacity, used for completed set value boxes.
  static const Color greenDim = Color(0x1F4ADE80);

  /// Dark green for text/icons rendered on top of [green] fills.
  static const Color onGreen = Color(0xFF0B3B1E);
}
