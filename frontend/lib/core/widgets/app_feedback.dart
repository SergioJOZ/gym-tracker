import 'package:flutter/material.dart';

/// Shows the standard placeholder snackbar for actions that are not
/// implemented yet.
void showComingSoonSnackBar(BuildContext context) {
  ScaffoldMessenger.of(
    context,
  ).showSnackBar(const SnackBar(content: Text('Coming soon')));
}
