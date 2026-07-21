import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

/// Wraps [child] with a subtle press-scale micro-interaction.
///
/// On tap-down the child scales toward [kPressedScale] (0.97) using
/// [MotionTokens.spring]; on release, cancel, or tap-up it returns to 1.0.
///
/// When reduced-motion is active ([MotionTokens.disabled] returns true) the
/// widget stays at scale 1.0 — no animation runs.
///
/// Covers Requirement: ScaleOnPress from the `ui-motion` spec.
class ScaleOnPress extends StatefulWidget {
  static const double kPressedScale = 0.97;

  final Widget child;
  final VoidCallback? onTap;
  final GestureTapDownCallback? onTapDown;

  const ScaleOnPress({
    super.key,
    required this.child,
    this.onTap,
    this.onTapDown,
  });

  @override
  State<ScaleOnPress> createState() => _ScaleOnPressState();
}

class _ScaleOnPressState extends State<ScaleOnPress> {
  double _targetScale = 1.0;

  bool get _reducedMotion => MotionTokens.disabled(context);

  void _onPointerDown(PointerDownEvent event) {
    widget.onTapDown?.call(TapDownDetails(
      globalPosition: event.position,
      localPosition: event.localPosition,
      kind: event.kind,
    ));
    if (_reducedMotion) return;
    setState(() => _targetScale = ScaleOnPress.kPressedScale);
  }

  void _onPointerUp(PointerUpEvent event) {
    if (!mounted) return;
    setState(() => _targetScale = 1.0);
  }

  void _onPointerCancel(PointerCancelEvent event) {
    if (!mounted) return;
    setState(() => _targetScale = 1.0);
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      behavior: HitTestBehavior.opaque,
      onTap: widget.onTap,
      child: Listener(
        onPointerDown: _onPointerDown,
        onPointerUp: _onPointerUp,
        onPointerCancel: _onPointerCancel,
        child: AnimatedScale(
          scale: _targetScale,
          duration: MotionTokens.medium,
          curve: MotionTokens.spring,
          child: widget.child,
        ),
      ),
    );
  }
}