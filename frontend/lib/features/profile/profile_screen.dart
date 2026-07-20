import 'package:flutter/material.dart';

import '../../core/theme/app_colors.dart';
import '../../core/theme/app_radius.dart';
import '../../core/widgets/app_feedback.dart';

/// Profile tab: minimal placeholder with a mock identity and logout action.
class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const SizedBox(
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
          ),
          const SizedBox(height: 16),
          const Text(
            'athlete@gymtracker.app',
            style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 32),
          Padding(
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
          ),
        ],
      ),
    );
  }
}
