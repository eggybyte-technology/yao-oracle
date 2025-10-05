// Basic widget test for Yao-Oracle Dashboard
import 'package:flutter_test/flutter_test.dart';
import 'package:dashboard/main.dart';

void main() {
  testWidgets('Dashboard app smoke test', (WidgetTester tester) async {
    // Build our app and trigger a frame.
    await tester.pumpWidget(const YaoOracleApp());

    // Verify that the app builds successfully
    expect(find.text('Yao-Oracle Dashboard'), findsOneWidget);
  });
}
