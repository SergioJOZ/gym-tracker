import 'package:flutter/material.dart';

import '../theme/app_colors.dart';
import '../theme/app_radius.dart';
import 'scale_on_press.dart';

/// Full-width accent call-to-action button with the play icon, shared by
/// the routines and routine-detail screens.
///
/// The button is wrapped in [ScaleOnPress] so the whole surface scales
/// subtly on press-down (0.97 spring) and snaps back on release. Under
/// reduced-motion the scale is a no-op (see [ScaleOnPress]).
///
/// Covers Requirement: ScaleOnPress as applied to primary CTAs in the
/// `ui-motion` spec.
class PrimaryCtaButton extends StatelessWidget {
  final String label;
  final VoidCallback onPressed;

  const PrimaryCtaButton({
    super.key,
    required this.label,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    return ScaleOnPress(
      onTap: onPressed,
      child: ElevatedButton.icon(
        // The button's own onPressed is also wired so disabled semantics
        // (onPressed == null) still surface correctly through ScaleOnPress.
        onPressed: onPressed,
        icon: const Icon(Icons.play_arrow, size: 18),
        label: Text(label),
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.accent,
          foregroundColor: Colors.white,
          elevation: 0,
          minimumSize: const Size.fromHeight(50),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(AppRadius.card),
          ),
          textStyle: const TextStyle(fontSize: 15, fontWeight: FontWeight.w700),
        ),
      ),
    );
  }
}