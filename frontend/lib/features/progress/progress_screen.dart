import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../core/widgets/app_feedback.dart';
import '../../core/widgets/staggered_entrance.dart';
import '../../data/mock/mock_data.dart';

/// Progress tab: headline stats, weekly volume bars and personal records.
///
/// Stat cards and personal-record rows fade+slide in via [StaggeredEntrance].
/// The weekly volume bars grow from the bottom upward with a spring curve on
/// first appearance. The screen's [ListView] carries a stable
/// [PageStorageKey] so the [MainShell]'s [PageStorageBucket] restores scroll
/// position across tab switches.
class ProgressScreen extends StatelessWidget {
  const ProgressScreen({super.key});

  @override
  Widget build(BuildContext context) {
    const summary = MockData.progress;
    const records = MockData.personalRecords;
    final stats = [
      (summary.totalWorkouts, 'Workouts'),
      (summary.workoutsThisWeek, 'This Week'),
      (summary.totalVolume, 'Volume'),
    ];

    // Staggered stat cards: each card is wrapped so its delay is i × stagger.
    final staggeredStats = StaggeredEntrance.wrap(
      [for (final s in stats) _StatCard(value: s.$1, label: s.$2)],
    );

    // Staggered record rows. The dividers between them stay outside the
    // stagger wrappers so only the rows animate.
    final staggeredRecords = StaggeredEntrance.wrap(
      [
        for (final r in records)
          _RecordRow(name: r.exerciseName, detail: r.detail, value: r.value),
      ],
    );

    return ListView(
      key: const PageStorageKey<String>('progress-list'),
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
              for (var i = 0; i < staggeredStats.length; i++) ...[
                if (i > 0) const SizedBox(width: 10),
                Expanded(child: staggeredStats[i]),
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
                for (var i = 0; i < staggeredRecords.length; i++) ...[
                  staggeredRecords[i],
                  if (i < staggeredRecords.length - 1) const Divider(),
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

/// Weekly volume bars animating from the bottom upward on first appearance.
///
/// Driven by a single [AnimationController] using [MotionTokens.slow] and the
/// spring curve. Each bar grows from 0 to its target fraction in lockstep
/// (the spec only requires bottom-up growth via spring, not a per-bar
/// stagger). Under reduced-motion the controller resolves to a 0ms
/// jump-to-completion, so bars appear at full value instantly (per the
/// "Reduced-motion snaps to value" scenario).
class _WeeklyBars extends StatefulWidget {
  final List<double> fractions;
  final List<String> labels;

  const _WeeklyBars({required this.fractions, required this.labels});

  @override
  State<_WeeklyBars> createState() => _WeeklyBarsState();
}

class _WeeklyBarsState extends State<_WeeklyBars>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _growth;
  late final double _maxFraction;

  @override
  void initState() {
    super.initState();
    // Guard against an empty fractions list — `reduce` would throw.
    _maxFraction = widget.fractions.isEmpty
        ? 0.0
        : widget.fractions.reduce((a, b) => a > b ? a : b);
    _controller = AnimationController(
      duration: MotionTokens.slow,
      vsync: this,
    );
    _growth = CurvedAnimation(
      parent: _controller,
      // Spring overshoot gives the bars a subtle settle, matching the spec's
      // "via spring" wording.
      curve: MotionTokens.spring,
    );
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    // Forward is triggered here so the reduced-motion gate has a valid
    // context. `resolve` returns Duration.zero under reduced-motion, which
    // makes the CurvedAnimation complete instantly on first frame.
    if (MotionTokens.disabled(context)) {
      _controller.value = 1.0;
    } else {
      _controller.forward();
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        // Reserve vertical room for the day label and its gap.
        final barArea = constraints.maxHeight - 22;
        return AnimatedBuilder(
          animation: _growth,
          builder: (context, _) {
            final progress = _growth.value;
            return Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                for (var i = 0; i < widget.fractions.length; i++)
                  Expanded(
                    child: Padding(
                      padding: EdgeInsets.only(
                        right: i < widget.fractions.length - 1 ? 8 : 0,
                      ),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          Container(
                            height: barArea *
                                widget.fractions[i] *
                                progress,
                            decoration: BoxDecoration(
                              color: widget.fractions[i] == _maxFraction
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
                            widget.labels[i],
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