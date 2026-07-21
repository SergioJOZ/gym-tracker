import 'dart:async';

import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

/// Wraps a list of children so they fade in and slide up with a per-item
/// stagger delay.
///
/// Child `i` animates with delay `i × stagger`. When the parent remounts,
/// each Tween runs once. Under reduced-motion ([MotionTokens.disabled]),
/// the fade-slide becomes an instant no-op.
///
/// Covers Requirement: StaggeredEntrance from the `ui-motion` spec.
class StaggeredEntrance {
  StaggeredEntrance._();

  /// Per-item stagger default sourced from [MotionTokens.stagger].
  static const Duration defaultStagger = MotionTokens.stagger;

  /// Slide-up offset used during the entrance tween.
  static const double kSlideOffset = 16.0;

  /// Computes the start delay for child [index] given [stagger].
  static Duration delayFor(int index, Duration stagger) {
    return Duration(microseconds: stagger.inMicroseconds * index);
  }

  /// Wraps [children] in [StaggeredEntranceItem] widgets that fade+slide
  /// each child in with a per-index stagger delay.
  ///
  /// Always returns a list of the same length as [children]; an empty list
  /// in produces an empty list out.
  static List<Widget> wrap(
    List<Widget> children, {
    Duration stagger = defaultStagger,
  }) {
    return List<Widget>.generate(children.length, (index) {
      return StaggeredEntranceItem(
        key: ValueKey('staggered-entrance-$index'),
        delay: delayFor(index, stagger),
        child: children[index],
      );
    });
  }
}

/// A single delayed-entrance wrapper exposed so tests and feature code can
/// read its [delay] for assertions. Use [StaggeredEntrance.wrap] instead of
/// constructing this widget directly.
class StaggeredEntranceItem extends StatefulWidget {
  /// Pre-computed delay before this child begins its entrance tween.
  final Duration delay;

  final Widget child;

  const StaggeredEntranceItem({
    super.key,
    required this.delay,
    required this.child,
  });

  @override
  State<StaggeredEntranceItem> createState() => _StaggeredEntranceItemState();
}

class _StaggeredEntranceItemState extends State<StaggeredEntranceItem> {
  bool _visible = false;
  Timer? _timer;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _scheduleVisible();
  }

  void _scheduleVisible() {
    _timer?.cancel();
    final effectiveDelay = MotionTokens.disabled(context)
        ? Duration.zero
        : widget.delay;
    _timer = Timer(effectiveDelay, () {
      if (mounted) setState(() => _visible = true);
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    // Reduced-motion: snap to visible without animating.
    if (MotionTokens.disabled(context)) return widget.child;

    return TweenAnimationBuilder<double>(
      tween: Tween<double>(begin: 0.0, end: _visible ? 1.0 : 0.0),
      duration: MotionTokens.resolve(context, MotionTokens.medium),
      curve: MotionTokens.standard,
      builder: (context, value, child) {
        final clamped = value.clamp(0.0, 1.0);
        return Opacity(
          opacity: clamped,
          child: Transform.translate(
            offset: Offset(0, (1 - clamped) * StaggeredEntrance.kSlideOffset),
            child: child,
          ),
        );
      },
      child: widget.child,
    );
  }
}