import 'package:flutter/material.dart';

import '../theme/app_colors.dart';

/// 44x44 rounded placeholder thumbnail used by exercise rows.
class ExerciseThumb extends StatelessWidget {
  const ExerciseThumb({super.key});

  @override
  Widget build(BuildContext context) {
    return const SizedBox(
      width: 44,
      height: 44,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: AppColors.surface2,
          borderRadius: BorderRadius.all(Radius.circular(10)),
          border: Border.fromBorderSide(BorderSide(color: AppColors.border)),
        ),
        child: Icon(
          Icons.fitness_center,
          size: 22,
          color: AppColors.textTertiary,
        ),
      ),
    );
  }
}
