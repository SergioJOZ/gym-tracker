import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/widgets/app_feedback.dart';
import '../../data/mock/mock_data.dart';

/// Progress tab: headline stats, weekly volume bars and personal records.
class ProgressScreen extends StatelessWidget {
  const ProgressScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final summary = MockData.progress;
    final records = MockData.personalRecords;
    final stats = [
      (summary.totalWorkouts, 'Workouts'),
      (summary.workoutsThisWeek, 'This Week'),
      (summary.totalVolume, 'Volume'),
    ];

    return ListView(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 6, 16, 2),
          child: Text(
            'Progress',
            style: Theme.of(context).textTheme.headlineLarge,
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 10, 16, 4),
          child: Row(
            children: [
              for (var i = 0; i < stats.length; i++) ...[
                if (i > 0) const SizedBox(width: 10),
                Expanded(
                  child: _StatCard(
                    value: stats[i].$1,
                    label: stats[i].$2,
                  ),
                ),
              ],
            ],
          ),
        ),
        Card(
          margin: const EdgeInsets.fromLTRB(16, 14, 16, 0),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Weekly Volume',
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.w700),
                ),
                const SizedBox(height: 2),
                const Text(
                  'Last 7 days',
                  style: TextStyle(
                    fontSize: 12,
                    color: AppColors.textTertiary,
                  ),
                ),
                const SizedBox(height: 16),
                SizedBox(
                  height: 110,
                  child: _WeeklyBars(
                    fractions: summary.weeklyFractions,
                    labels: summary.dayLabels,
                  ),
                ),
              ],
            ),
          ),
        ),
        Card(
          margin: const EdgeInsets.fromLTRB(16, 14, 16, 0),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            child: Column(
              children: [
                Padding(
                  padding: const EdgeInsets.only(top: 10, bottom: 4),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Personal Records',
                        style: Theme.of(context).textTheme.titleSmall,
                      ),
                      GestureDetector(
                        onTap: () => showComingSoonSnackBar(context),
                        child: const Text(
                          'See all',
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
                for (var i = 0; i < records.length; i++) ...[
                  _RecordRow(
                    name: records[i].exerciseName,
                    detail: records[i].detail,
                    value: records[i].value,
                  ),
                  if (i < records.length - 1) const Divider(),
                ],
              ],
            ),
          ),
        ),
        const SizedBox(height: 24),
      ],
    );
  }
}

class _StatCard extends StatelessWidget {
  final String value;
  final String label;

  const _StatCard({required this.value, required this.label});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
        child: Column(
          children: [
            Text(
              value,
              style: const TextStyle(
                fontSize: 22,
                fontWeight: FontWeight.w800,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              label.toUpperCase(),
              style: const TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w600,
                letterSpacing: 0.4,
                color: AppColors.textSecondary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _WeeklyBars extends StatelessWidget {
  final List<double> fractions;
  final List<String> labels;

  const _WeeklyBars({required this.fractions, required this.labels});

  @override
  Widget build(BuildContext context) {
    final maxFraction = fractions.reduce((a, b) => a > b ? a : b);

    return LayoutBuilder(
      builder: (context, constraints) {
        // Reserve vertical room for the day label and its gap.
        final barArea = constraints.maxHeight - 22;
        return Row(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            for (var i = 0; i < fractions.length; i++)
              Expanded(
                child: Padding(
                  padding: EdgeInsets.only(
                    right: i < fractions.length - 1 ? 8 : 0,
                  ),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      Container(
                        height: barArea * fractions[i],
                        decoration: BoxDecoration(
                          color: fractions[i] == maxFraction
                              ? AppColors.accent
                              : AppColors.surface2,
                          borderRadius: const BorderRadius.vertical(
                            top: Radius.circular(6),
                            bottom: Radius.circular(3),
                          ),
                        ),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        labels[i],
                        style: const TextStyle(
                          fontSize: 10,
                          fontWeight: FontWeight.w700,
                          color: AppColors.textTertiary,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
          ],
        );
      },
    );
  }
}

class _RecordRow extends StatelessWidget {
  final String name;
  final String detail;
  final String value;

  const _RecordRow({
    required this.name,
    required this.detail,
    required this.value,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 12),
      child: Row(
        children: [
          const SizedBox(
            width: 36,
            height: 36,
            child: DecoratedBox(
              decoration: BoxDecoration(
                color: AppColors.accentDim,
                borderRadius: BorderRadius.all(Radius.circular(10)),
              ),
              child: Icon(
                Icons.emoji_events_outlined,
                size: 18,
                color: AppColors.accent,
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  style: const TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  detail,
                  style: const TextStyle(
                    fontSize: 12.5,
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ),
          ),
          Text(
            value,
            style: const TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}
