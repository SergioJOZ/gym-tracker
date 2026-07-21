import 'package:flutter/material.dart';

import '../theme/app_colors.dart';
import '../theme/app_theme.dart';

/// A shimmer placeholder used while content loads.
///
/// Renders a rectangular [SizedBox] of [width] × [height] with a linear
/// gradient that sweeps across the surface when [isLoading] is true. When
/// [isLoading] is false (or reduced-motion is active) the box renders as a
/// flat [AppColors.surface2] rounded rectangle.
///
/// Covers Requirement: SkeletonBox from the `ui-motion` spec.
class SkeletonBox extends StatefulWidget {
  final double width;
  final double height;
  final bool isLoading;
  final double radius;
  final Color baseColor;
  final Color highlightColor;

  const SkeletonBox({
    super.key,
    required this.width,
    required this.height,
    this.isLoading = true,
    this.radius = 6.0,
    this.baseColor = AppColors.surface2,
    this.highlightColor = AppColors.border,
  });

  @override
  State<SkeletonBox> createState() => _SkeletonBoxState();
}

class _SkeletonBoxState extends State<SkeletonBox>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  bool _shimmering = false;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1000),
    );
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _syncShimmer();
  }

  @override
  void didUpdateWidget(covariant SkeletonBox oldWidget) {
    super.didUpdateWidget(oldWidget);
    _syncShimmer();
  }

  void _syncShimmer() {
    final shouldShimmer =
        widget.isLoading && !MotionTokens.disabled(context);
    if (shouldShimmer == _shimmering) return;
    _shimmering = shouldShimmer;
    if (shouldShimmer) {
      _controller.repeat();
    } else {
      _controller.stop();
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final showShimmer = _shimmering;

    return SizedBox(
      width: widget.width,
      height: widget.height,
      child: showShimmer
          ? AnimatedBuilder(
              animation: _controller,
              builder: (context, _) {
                return CustomPaint(
                  painter: _ShimmerPainter(
                    progress: _controller.value,
                    baseColor: widget.baseColor,
                    highlightColor: widget.highlightColor,
                    radius: widget.radius,
                  ),
                );
              },
            )
          : DecoratedBox(
              decoration: BoxDecoration(
                color: widget.baseColor,
                borderRadius: BorderRadius.all(Radius.circular(widget.radius)),
              ),
            ),
    );
  }
}

/// Draws the gradient sweep used by [SkeletonBox] when loading.
class _ShimmerPainter extends CustomPainter {
  final double progress;
  final Color baseColor;
  final Color highlightColor;
  final double radius;

  _ShimmerPainter({
    required this.progress,
    required this.baseColor,
    required this.highlightColor,
    required this.radius,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final rrect = RRect.fromRectAndRadius(
      Offset.zero & size,
      Radius.circular(radius),
    );

    // Base color across the box.
    canvas.drawRRect(
      rrect,
      Paint()..color = baseColor,
    );

    // A highlight band sweeps from left to right once per cycle.
    final bandWidth = size.width * 0.4;
    final span = size.width + bandWidth;
    final dx = (progress * span) - bandWidth;
    final bandRect = Rect.fromLTWH(dx, 0, bandWidth, size.height);

    const gradient = LinearGradient(
      begin: Alignment.topLeft,
      end: Alignment.topRight,
      colors: [
        Color(0x00000000),
        Color(0x66FFFFFF),
        Color(0x00000000),
      ],
      stops: [0.0, 0.5, 1.0],
    );

    canvas.save();
    canvas.clipRRect(rrect);
    canvas.drawRect(
      bandRect,
      Paint()
        ..shader = gradient.createShader(bandRect)
        ..blendMode = BlendMode.srcATop,
    );
    canvas.restore();
  }

  @override
  bool shouldRepaint(covariant _ShimmerPainter old) =>
      old.progress != progress ||
      old.baseColor != baseColor ||
      old.highlightColor != highlightColor ||
      old.radius != radius;
}