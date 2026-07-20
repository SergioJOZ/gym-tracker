import 'package:flutter_test/flutter_test.dart';
import 'package:gym_tracker/main.dart';

void main() {
  testWidgets('renders the Routines tab on launch', (tester) async {
    await tester.pumpWidget(const GymTrackerApp());

    // Big screen title plus the selected NavigationBar destination label.
    expect(find.text('Routines'), findsWidgets);
  });
}
