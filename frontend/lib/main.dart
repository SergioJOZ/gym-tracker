import 'package:flutter/material.dart';

import 'core/theme/app_theme.dart';
import 'features/shell/main_shell.dart';

void main() {
  runApp(const GymTrackerApp());
}

/// Root widget. Dark-only Material 3 app hosting the 4-tab shell.
class GymTrackerApp extends StatelessWidget {
  const GymTrackerApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'gym-tracker',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.dark,
      home: const MainShell(),
    );
  }
}
