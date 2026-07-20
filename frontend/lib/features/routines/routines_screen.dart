import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_radius.dart';
import '../../core/widgets/app_feedback.dart';
import '../../core/widgets/primary_cta_button.dart';
import '../../data/mock/mock_data.dart';
import '../../data/models/models.dart';
import '../workout/active_workout_screen.dart';
import 'routine_detail_screen.dart';

/// Home tab: CTA to start an empty workout plus the saved routine cards.
class RoutinesScreen extends StatelessWidget {
  const RoutinesScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final routines = MockData.routines;

    return ListView(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 6, 16, 2),
          child: Text(
            'Routines',
            style: Theme.of(context).textTheme.headlineLarge,
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 14, 16, 4),
          child: PrimaryCtaButton(
            label: 'Start Empty Workout',
            onPressed: () => Navigator.of(context).push(
              MaterialPageRoute<void>(
                builder: (_) => const ActiveWorkoutScreen(),
              ),
            ),
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 18, 16, 10),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'My Routines',
                style: Theme.of(context).textTheme.titleSmall,
              ),
              GestureDetector(
                onTap: () => showComingSoonSnackBar(context),
                child: const Text(
                  '+ New',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: AppColors.accent,
                  ),
                ),
              ),
            ],
          ),
        ),
        for (final routine in routines) _RoutineCard(routine: routine),
        const SizedBox(height: 24),
      ],
    );
  }
}

class _RoutineCard extends StatelessWidget {
  final Routine routine;

  const _RoutineCard({required this.routine});

  @override
  Widget build(BuildContext context) {
    final meta =
        '${routine.exercises.length} exercises · Last: ${routine.lastPerformedLabel}';
    final preview = routine.exercises.map((e) => e.name).join(' · ');

    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
      child: Card(
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadius.card),
          onTap: () => Navigator.of(context).push(
            MaterialPageRoute<void>(
              builder: (_) => RoutineDetailScreen(routine: routine),
            ),
          ),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  routine.name,
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 4),
                Text(
                  meta,
                  style: const TextStyle(
                    fontSize: 13,
                    color: AppColors.textSecondary,
                  ),
                ),
                const SizedBox(height: 10),
                Text(
                  preview,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    fontSize: 12.5,
                    height: 1.5,
                    color: AppColors.textTertiary,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
