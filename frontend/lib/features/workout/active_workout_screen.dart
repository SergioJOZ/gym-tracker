import 'dart:async';
import 'dart:math' as math;
import 'dart:ui' show FontFeature;

import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_radius.dart';
import '../../core/widgets/app_feedback.dart';
import '../../data/models/models.dart';

/// Full-screen active workout logger.
///
/// Pushed as a new route (no bottom nav) with an optional [Routine] whose
/// exercises pre-fill the set grids. Pass `null` for an empty workout.
class ActiveWorkoutScreen extends StatefulWidget {
  final Routine? routine;

  const ActiveWorkoutScreen({super.key, this.routine});

  @override
  State<ActiveWorkoutScreen> createState() => _ActiveWorkoutScreenState();
}

class _ActiveWorkoutScreenState extends State<ActiveWorkoutScreen> {
  late final Timer _timer;
  int _elapsedSeconds = 0;
  late final List<_WorkoutExercise> _exercises;

  @override
  void initState() {
    super.initState();
    _exercises = _buildExercises();
    _timer = Timer.periodic(const Duration(seconds: 1), (_) {
      setState(() => _elapsedSeconds++);
    });
  }

  @override
  void dispose() {
    _timer.cancel();
    super.dispose();
  }

  /// Builds the editable exercise state from the routine.
  ///
  /// The first exercise mirrors the approved mockup: 4 sets with the first
  /// two done (60/10, 60/9), the third pending with values (62.5/8) and the
  /// fourth empty. Remaining exercises get their planned set count, all
  /// pending and empty.
  List<_WorkoutExercise> _buildExercises() {
    final routine = widget.routine;
    if (routine == null) {
      return const [];
    }
    return [
      for (var i = 0; i < routine.exercises.length; i++)
        _WorkoutExercise(
          name: routine.exercises[i].name,
          sets: i == 0
              ? [
                  WorkoutSet(kg: 60, reps: 10, done: true),
                  WorkoutSet(kg: 60, reps: 9, done: true),
                  WorkoutSet(kg: 62.5, reps: 8),
                  WorkoutSet(),
                ]
              : [
                  for (var s = 0; s < routine.exercises[i].sets; s++)
                    WorkoutSet(),
                ],
        ),
    ];
  }

  String get _elapsedLabel {
    final minutes = (_elapsedSeconds ~/ 60).toString().padLeft(2, '0');
    final seconds = (_elapsedSeconds % 60).toString().padLeft(2, '0');
    return '$minutes:$seconds';
  }

  void _finish() {
    // Capture the app-level messenger before popping this route.
    final messenger = ScaffoldMessenger.of(context);
    Navigator.of(context).pop();
    messenger.showSnackBar(
      const SnackBar(content: Text('Workout saved (mock)')),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      widget.routine?.name ?? 'Workout',
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ),
                  Text(
                    _elapsedLabel,
                    style: const TextStyle(
                      fontSize: 26,
                      fontWeight: FontWeight.w800,
                      color: AppColors.accent,
                      fontFeatures: [FontFeature.tabularFigures()],
                    ),
                  ),
                  const SizedBox(width: 12),
                  ElevatedButton(
                    onPressed: _finish,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppColors.green,
                      foregroundColor: AppColors.onGreen,
                      elevation: 0,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 9,
                      ),
                      minimumSize: Size.zero,
                      tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                      ),
                      textStyle: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                    child: const Text('Finish'),
                  ),
                ],
              ),
            ),
            Expanded(
              child: ListView(
                padding: const EdgeInsets.only(top: 2),
                children: [
                  for (final exercise in _exercises)
                    _ExerciseCard(
                      exercise: exercise,
                      onChanged: () => setState(() {}),
                    ),
                ],
              ),
            ),
            Container(
              decoration: const BoxDecoration(
                border: Border(top: BorderSide(color: AppColors.border)),
              ),
              padding: const EdgeInsets.fromLTRB(16, 10, 16, 18),
              child: CustomPaint(
                painter: const _DashedRRectPainter(color: AppColors.border),
                child: GestureDetector(
                  onTap: () => showComingSoonSnackBar(context),
                  behavior: HitTestBehavior.opaque,
                  child: const Padding(
                    padding: EdgeInsets.symmetric(vertical: 13),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.add, size: 16, color: AppColors.accent),
                        SizedBox(width: 8),
                        Text(
                          'Add Exercise',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w700,
                            color: AppColors.accent,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Editable state for one exercise inside the active workout.
class _WorkoutExercise {
  final String name;
  final List<WorkoutSet> sets;

  _WorkoutExercise({required this.name, required this.sets});
}

/// Card with the SET / KG / REPS grid for one exercise.
class _ExerciseCard extends StatelessWidget {
  final _WorkoutExercise exercise;

  /// Called after any mutation so the parent rebuilds.
  final VoidCallback onChanged;

  const _ExerciseCard({required this.exercise, required this.onChanged});

  static const List<String> _headers = ['SET', 'KG', 'REPS', ''];

  String _formatKg(double? kg) {
    if (kg == null) {
      return '—';
    }
    return kg == kg.roundToDouble() ? kg.toInt().toString() : kg.toString();
  }

  String _formatReps(int? reps) => reps?.toString() ?? '—';

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.fromLTRB(16, 0, 16, 14),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(14, 14, 14, 10),
        child: Column(
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    exercise.name,
                    style: Theme.of(
                      context,
                    ).textTheme.titleSmall?.copyWith(color: AppColors.accent),
                  ),
                ),
                const Icon(
                  Icons.more_vert,
                  size: 20,
                  color: AppColors.textTertiary,
                ),
              ],
            ),
            const SizedBox(height: 6),
            _GridRow(
              cells: [
                for (final header in _headers)
                  _GridCell(
                    width: header == 'SET' || header == '' ? 44 : null,
                    child: Text(
                      header,
                      textAlign: TextAlign.center,
                      style: const TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        letterSpacing: 0.5,
                        color: AppColors.textTertiary,
                      ),
                    ),
                  ),
              ],
            ),
            for (var i = 0; i < exercise.sets.length; i++)
              _buildSetRow(i, exercise.sets[i]),
            GestureDetector(
              onTap: () {
                exercise.sets.add(WorkoutSet());
                onChanged();
              },
              behavior: HitTestBehavior.opaque,
              child: const Padding(
                padding: EdgeInsets.only(top: 10, bottom: 4),
                child: Center(
                  child: Text(
                    '+ Add Set',
                    style: TextStyle(
                      fontSize: 13.5,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSetRow(int index, WorkoutSet set) {
    final done = set.done;
    final valueColor = done
        ? AppColors.green
        : (set.kg == null ? AppColors.textTertiary : AppColors.text);

    return _GridRow(
      cells: [
        _GridCell(
          width: 44,
          child: Text(
            '${index + 1}',
            textAlign: TextAlign.center,
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w700,
              color: done ? AppColors.green : AppColors.textSecondary,
            ),
          ),
        ),
        _GridCell(
          child: _ValueBox(
            text: _formatKg(set.kg),
            done: done,
            textColor: valueColor,
          ),
        ),
        _GridCell(
          child: _ValueBox(
            text: _formatReps(set.reps),
            done: done,
            textColor: done
                ? AppColors.green
                : (set.reps == null ? AppColors.textTertiary : AppColors.text),
          ),
        ),
        _GridCell(
          width: 44,
          child: GestureDetector(
            onTap: () {
              set.done = !set.done;
              onChanged();
            },
            child: Container(
              width: 30,
              height: 30,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: done ? AppColors.green : Colors.transparent,
                border: Border.all(
                  color: done ? AppColors.green : AppColors.border,
                  width: 2,
                ),
              ),
              child: done
                  ? const Icon(
                      Icons.check,
                      size: 16,
                      color: AppColors.onGreen,
                    )
                  : null,
            ),
          ),
        ),
      ],
    );
  }
}

/// A grid row: fixed 44px edge columns, flexible KG/REPS columns, 8px gaps.
class _GridRow extends StatelessWidget {
  final List<_GridCell> cells;

  const _GridRow({required this.cells});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 5),
      child: Row(
        children: [
          for (var i = 0; i < cells.length; i++) ...[
            if (i > 0) const SizedBox(width: 8),
            cells[i],
          ],
        ],
      ),
    );
  }
}

class _GridCell extends StatelessWidget {
  final double? width;
  final Widget child;

  const _GridCell({this.width, required this.child});

  @override
  Widget build(BuildContext context) {
    final width = this.width;
    if (width != null) {
      return SizedBox(width: width, child: Center(child: child));
    }
    return Expanded(child: Center(child: child));
  }
}

/// Read-only value box styled like the mockup's set inputs.
class _ValueBox extends StatelessWidget {
  final String text;
  final bool done;
  final Color textColor;

  const _ValueBox({
    required this.text,
    required this.done,
    required this.textColor,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 84,
      padding: const EdgeInsets.symmetric(vertical: 8),
      decoration: BoxDecoration(
        color: done ? AppColors.greenDim : AppColors.surface2,
        borderRadius: BorderRadius.circular(8),
        border: done ? null : Border.all(color: AppColors.border),
      ),
      child: Text(
        text,
        textAlign: TextAlign.center,
        style: TextStyle(
          fontSize: 15,
          fontWeight: FontWeight.w700,
          color: textColor,
        ),
      ),
    );
  }
}

/// Paints a dashed rounded-rect border for the "Add Exercise" ghost button.
class _DashedRRectPainter extends CustomPainter {
  final Color color;
  final double radius;
  final double strokeWidth;
  final double dashLength;
  final double dashGap;

  const _DashedRRectPainter({
    required this.color,
    this.radius = AppRadius.card,
    this.strokeWidth = 1,
    this.dashLength = 6,
    this.dashGap = 4,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = color
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth;
    final rrect = RRect.fromRectAndRadius(
      Offset.zero & size,
      Radius.circular(radius),
    ).deflate(strokeWidth / 2);
    final path = Path()..addRRect(rrect);
    for (final metric in path.computeMetrics()) {
      var distance = 0.0;
      while (distance < metric.length) {
        final length = math.min(dashLength, metric.length - distance);
        canvas.drawPath(metric.extractPath(distance, distance + length), paint);
        distance += dashLength + dashGap;
      }
    }
  }

  @override
  bool shouldRepaint(covariant _DashedRRectPainter oldDelegate) {
    return oldDelegate.color != color ||
        oldDelegate.radius != radius ||
        oldDelegate.strokeWidth != strokeWidth ||
        oldDelegate.dashLength != dashLength ||
        oldDelegate.dashGap != dashGap;
  }
}
