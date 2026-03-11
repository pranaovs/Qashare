import 'package:flutter/material.dart';
import 'package:qashare/Config/token_storage.dart';
import 'package:qashare/Service/api_service.dart';

class AuthcheckPage extends StatefulWidget {
  const AuthcheckPage({super.key});

  @override
  State<AuthcheckPage> createState() => _AuthcheckPageState();
}

class _AuthcheckPageState extends State<AuthcheckPage> {
  @override
  void initState() {
    super.initState();
    _check();
  }

  Future<void> _check() async {
    final accessToken = await TokenStorage.getAccessToken();

    if (accessToken == null) {
      // No tokens at all → go to login
      if (!mounted) return;
      Navigator.pushReplacementNamed(context, '/login');
      return;
    }

    // Try fetching the current user to validate the access token.
    // The internal _authenticatedRequest will auto-refresh if needed.
    final userResult = await ApiService.getCurrentUser();

    if (!mounted) return;

    if (userResult.isSuccess) {
      Navigator.pushReplacementNamed(context, '/home');
    } else {
      // Tokens are fully expired / invalid → clear and go to login
      await TokenStorage.clear();
      if (!mounted) return;
      Navigator.pushReplacementNamed(context, '/login');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: CircularProgressIndicator(
          color: Theme.of(context).colorScheme.primary,
        ),
      ),
    );
  }
}
