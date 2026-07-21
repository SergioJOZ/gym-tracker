import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../exercises/exercise_catalog_screen.dart';
import '../profile/profile_screen.dart';
import '../progress/progress_screen.dart';
import '../routines/routines_screen.dart';

/// App shell with the 4-tab Material 3 [NavigationBar].
///
/// Tabs crossfade via [AnimatedSwitcher]. Because [AnimatedSwitcher] disposes
/// its outgoing child once the transition completes, scroll-offset (and any
/// other ephemeral [State]) is restored via a persistent [PageStorageBucket]
/// owned by this [State]. Each screen's primary [Scrollable] SHOULD tag
/// itself with a stable [PageStorageKey] so the bucket can hand back its
/// last position on remount.
class MainShell extends StatefulWidget {
  const MainShell({super.key});

  @override
  State<MainShell> createState() => _MainShellState();
}

class _MainShellState extends State<MainShell> {
  int _index = 0;

  /// Persistent bucket used to restore per-screen scroll positions across
  /// tab switches. Lives with this [State] and survives the
  /// [AnimatedSwitcher] round-trip.
  final PageStorageBucket _bucket = PageStorageBucket();

  /// Stable [GlobalKey]s per tab so the [State] identity is preserved when a
  /// screen is reparented. Combined with the [PageStorageBucket] above, this
  /// guarantees in-active-tab state restoration across tab switches.
  static final List<GlobalKey<State<StatefulWidget>>> _screenKeys =
      List<GlobalKey<State<StatefulWidget>>>.generate(
    4,
    (_) => GlobalKey<State<StatefulWidget>>(),
  );

  Widget _screenFor(int index) {
    final key = _screenKeys[index];
    switch (index) {
      case 0:
        return RoutinesScreen(key: key);
      case 1:
        return ExerciseCatalogScreen(key: key);
      case 2:
        return ProgressScreen(key: key);
      case 3:
        return ProfileScreen(key: key);
      default:
        return RoutinesScreen(key: key);
    }
  }

  @override
  Widget build(BuildContext context) {
    final crossfade = MotionTokens.resolve(context, MotionTokens.medium);
    return Scaffold(
      body: SafeArea(
        bottom: false,
        child: PageStorage(
          bucket: _bucket,
          child: AnimatedSwitcher(
            duration: crossfade,
            switchInCurve: MotionTokens.standard,
            switchOutCurve: MotionTokens.standard,
            transitionBuilder: (child, animation) =>
                FadeTransition(opacity: animation, child: child),
            child: KeyedSubtree(
              key: ValueKey<int>(_index),
              child: _screenFor(_index),
            ),
          ),
        ),
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