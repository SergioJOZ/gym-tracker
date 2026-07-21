import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/widgets/exercise_thumb.dart';
import '../../core/widgets/scale_on_press.dart';
import '../../core/widgets/staggered_entrance.dart';
import '../../data/mock/mock_data.dart';
import '../../data/models/models.dart';

/// Exercises tab: searchable, filterable read-only catalog.
class ExerciseCatalogScreen extends StatefulWidget {
  const ExerciseCatalogScreen({super.key});

  @override
  State<ExerciseCatalogScreen> createState() => _ExerciseCatalogScreenState();
}

class _ExerciseCatalogScreenState extends State<ExerciseCatalogScreen> {
  static const List<String> _groups = [
    'All',
    'Chest',
    'Back',
    'Legs',
    'Shoulders',
    'Arms',
  ];

  String _query = '';
  String _selectedGroup = 'All';

  List<Exercise> get _filtered {
    final query = _query.trim().toLowerCase();
    return MockData.exerciseCatalog.where((exercise) {
      final matchesGroup =
          _selectedGroup == 'All' || exercise.muscleGroup == _selectedGroup;
      final matchesQuery =
          query.isEmpty || exercise.name.toLowerCase().contains(query);
      return matchesGroup && matchesQuery;
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    final exercises = _filtered;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 6, 16, 2),
          child: Text(
            'Exercises',
            style: Theme.of(context).textTheme.headlineLarge,
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 10, 16, 12),
          child: TextField(
            onChanged: (value) => setState(() => _query = value),
            style: const TextStyle(fontSize: 14, color: AppColors.text),
            cursorColor: AppColors.accent,
            decoration: InputDecoration(
              hintText: 'Search 1,324 exercises',
              hintStyle: const TextStyle(
                fontSize: 14,
                color: AppColors.textTertiary,
              ),
              prefixIcon: const Icon(
                Icons.search,
                size: 20,
                color: AppColors.textTertiary,
              ),
              filled: true,
              fillColor: AppColors.card,
              isDense: true,
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 14,
                vertical: 12,
              ),
              border: _searchBorder(AppColors.border),
              enabledBorder: _searchBorder(AppColors.border),
              focusedBorder: _searchBorder(AppColors.accent),
            ),
          ),
        ),
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.fromLTRB(16, 0, 16, 14),
          child: Row(
            children: [
              for (var i = 0; i < _groups.length; i++)
                Padding(
                  padding: EdgeInsets.only(
                    right: i < _groups.length - 1 ? 8 : 0,
                  ),
                  child: _FilterChip(
                    label: _groups[i],
                    selected: _groups[i] == _selectedGroup,
                    onTap: () =>
                        setState(() => _selectedGroup = _groups[i]),
                  ),
                ),
            ],
          ),
        ),
        Expanded(
          child: ListView(
            key: const PageStorageKey<String>('exercises-list'),
            children: [
              ...StaggeredEntrance.wrap(
                [for (final exercise in exercises) _CatalogRow(exercise: exercise)],
              ),
              const SizedBox(height: 16),
            ],
          ),
        ),
      ],
    );
  }

  OutlineInputBorder _searchBorder(Color color) {
    return OutlineInputBorder(
      borderRadius: BorderRadius.circular(12),
      borderSide: BorderSide(color: color),
    );
  }
}

class _FilterChip extends StatelessWidget {
  final String label;
  final bool selected;
  final VoidCallback onTap;

  const _FilterChip({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
        decoration: BoxDecoration(
          color: selected ? AppColors.accent : AppColors.card,
          borderRadius: BorderRadius.circular(999),
          border: Border.all(
            color: selected ? AppColors.accent : AppColors.border,
          ),
        ),
        child: Text(
          label,
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: selected ? Colors.white : AppColors.textSecondary,
          ),
        ),
      ),
    );
  }
}

class _CatalogRow extends StatelessWidget {
  final Exercise exercise;

  const _CatalogRow({required this.exercise});

  @override
  Widget build(BuildContext context) {
    return ScaleOnPress(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 11),
        child: Row(
          children: [
            const ExerciseThumb(),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    exercise.name,
                    style: const TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    '${exercise.muscleGroup} · ${exercise.equipment}',
                    style: const TextStyle(
                      fontSize: 12.5,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
            ),
            const Icon(
              Icons.chevron_right,
              size: 18,
              color: AppColors.textTertiary,
            ),
          ],
        ),
      ),
    );
  }
}
