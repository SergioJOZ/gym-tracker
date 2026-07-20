import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../exercises/exercise_catalog_screen.dart';
import '../profile/profile_screen.dart';
import '../progress/progress_screen.dart';
import '../routines/routines_screen.dart';

/// App shell with the 4-tab Material 3 [NavigationBar].
///
/// An [IndexedStack] keeps every tab's state alive while switching.
class MainShell extends StatefulWidget {
  const MainShell({super.key});

  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _index = 0;

  static const List<Widget> _screens = [
    RoutinesScreen(),
    ExerciseCatalogScreen(),
    ProgressScreen(),
    ProfileScreen(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        bottom: false,
        child: IndexedStack(index: _index, children: _screens),
      ),
      bottomNavigationBar: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Divider(height: 1, thickness: 1, color: AppColors.border),
          NavigationBar(
            selectedIndex: _index,
            onDestinationSelected: (index) => setState(() => _index = index),
            destinations: const [
              NavigationDestination(
                icon: Icon(Icons.format_list_bulleted),
                label: 'Routines',
              ),
              NavigationDestination(
                icon: Icon(Icons.fitness_center),
                label: 'Exercises',
              ),
              NavigationDestination(
                icon: Icon(Icons.bar_chart),
                label: 'Progress',
              ),
              NavigationDestination(
                icon: Icon(Icons.person_outline),
                label: 'Profile',
              ),
            ],
          ),
        ],
      ),
    );
  }
}
