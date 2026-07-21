import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_radius.dart';
import '../../core/widgets/app_feedback.dart';
import '../../core/widgets/staggered_entrance.dart';

/// Profile tab: minimal placeholder with a mock identity and logout action.
///
/// The centered column's children fade+slide in via [StaggeredEntrance]. The
/// column is tagged with a stable [PageStorageKey] so the [MainShell]'s
/// [PageStorageBucket] can restore any future scroll state across tab
/// switches.
class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    const avatar = SizedBox(
      width: 84,
      height: 84,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: AppColors.surface2,
          shape: BoxShape.circle,
          border: Border.fromBorderSide(
            BorderSide(color: AppColors.border),
          ),
        ),
        child: Icon(
          Icons.person_outline,
          size: 40,
          color: AppColors.textTertiary,
        ),
      ),
    );
    const email = Text(
      'athlete@gymtracker.app',
      style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
    );
    final logout = Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Card(
        child: ListTile(
          leading: const Icon(
            Icons.logout,
            color: AppColors.textSecondary,
          ),
          title: const Text(
            'Logout',
            style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
          ),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(AppRadius.card),
          ),
          onTap: () => showComingSoonSnackBar(context),
        ),
      ),
    );

    final children = <Widget>[
      avatar,
      const SizedBox(height: 16),
      email,
      const SizedBox(height: 32),
      logout,
    ];

    return Center(
      child: Column(
        key: const PageStorageKey<String>('profile-content'),
        mainAxisSize: MainAxisSize.min,
        children: StaggeredEntrance.wrap(children),
      ),
    );
  }
}