import 'package:flutter/material.dart';
import 'package:qashare/Config/api_config.dart';
import 'package:qashare/Config/token_storage.dart';
import 'package:qashare/Service/api_service.dart';

class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _formKey = GlobalKey<FormState>();
  final _usernameController = TextEditingController();
  final _passwordController = TextEditingController();
  String _serverUrl = "https://qashare.pranaovs.me";

  bool _obscure = true;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          child: Padding(
            padding: const EdgeInsets.symmetric(
              horizontal: 40,
              vertical: 100.0,
            ),
            child: Form(
              key: _formKey,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  const SizedBox(height: 80),

                  Text(
                    "Welcome back",
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.headlineLarge?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),

                  const SizedBox(height: 10),

                  Text(
                    "Login into your account",
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Theme.of(context).colorScheme.outline,
                    ),
                  ),

                  const SizedBox(height: 40),

                  // Username
                  TextFormField(
                    controller: _usernameController,
                    decoration: _inputDecoration("Email"),
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return "Enter Email";
                      }
                      return null;
                    },
                  ),

                  const SizedBox(height: 16),

                  // Password
                  TextFormField(
                    controller: _passwordController,
                    obscureText: _obscure,
                    decoration: _inputDecoration("Password").copyWith(
                      suffixIcon: IconButton(
                        icon: Icon(
                          _obscure
                              ? Icons.visibility_outlined
                              : Icons.visibility_off_outlined,
                        ),
                        onPressed: () {
                          setState(() => _obscure = !_obscure);
                        },
                      ),
                    ),
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return "Enter password";
                      }
                      return null;
                    },
                  ),

                  const SizedBox(height: 30),

                  // Login Button
                  SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: ElevatedButton(
                      onPressed: _handleLogin,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: Theme.of(context).colorScheme.primary,
                        foregroundColor: Theme.of(
                          context,
                        ).colorScheme.onPrimary,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16),
                        ),
                      ),
                      child: const Text(
                        "Login",
                        style: TextStyle(fontSize: 16),
                      ),
                    ),
                  ),

                  const SizedBox(height: 20),

                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        "Don't have an account?",
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.outline,
                        ),
                      ),
                      TextButton(
                        onPressed: () {
                          Navigator.pushNamed(context, '/signup');
                        },
                        child: const Text("Sign up"),
                      ),
                    ],
                  ),
                  const SizedBox(height: 100.0),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Theme.of(context).colorScheme.surfaceVariant,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Row(
                      children: [
                        Icon(
                          Icons.dns,
                          size: 16,
                          color: Theme.of(context).colorScheme.onSurfaceVariant,
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            'Server: ${_serverUrl.isEmpty ? "Not configured" : _serverUrl}',
                            style: Theme.of(context).textTheme.bodySmall
                                ?.copyWith(
                                  color: Theme.of(
                                    context,
                                  ).colorScheme.onSurfaceVariant,
                                ),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        TextButton(
                          onPressed: _showServerDialog,
                          style: TextButton.styleFrom(
                            padding: const EdgeInsets.symmetric(horizontal: 8),
                            minimumSize: const Size(0, 30),
                          ),
                          child: const Text('Change'),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  void _showServerDialog() {
    final controller = TextEditingController(text: _serverUrl);

    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text("Server Address"),
          content: TextField(
            controller: controller,
            decoration: const InputDecoration(
              hintText: "host:port (e.g. devbox:8080)",
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text("Cancel"),
            ),
            ElevatedButton(
              onPressed: () {
                final server = controller.text.trim();

                if (server.isNotEmpty) {
                  setState(() {
                    _serverUrl = server;
                  });
                  ApiConfig.setServer(server);
                  Navigator.pop(context);

                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(
                      content: Text("Server set to $server"),
                      backgroundColor: Theme.of(context).colorScheme.primary,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      duration: const Duration(milliseconds: 1000),
                      behavior: SnackBarBehavior.floating,
                    ),
                  );
                }
              },
              child: const Text("Save"),
            ),
          ],
        );
      },
    );
  }

  InputDecoration _inputDecoration(String label) {
    return InputDecoration(
      labelText: label,
      enabledBorder: OutlineInputBorder(
        borderSide: BorderSide(color: Theme.of(context).colorScheme.outline),
        borderRadius: BorderRadius.circular(12),
      ),
      focusedBorder: OutlineInputBorder(
        borderSide: BorderSide(color: Theme.of(context).colorScheme.primary),
        borderRadius: BorderRadius.circular(12),
      ),
      errorBorder: OutlineInputBorder(
        borderSide: BorderSide(color: Theme.of(context).colorScheme.error),
        borderRadius: BorderRadius.circular(12),
      ),
      focusedErrorBorder: OutlineInputBorder(
        borderSide: BorderSide(color: Theme.of(context).colorScheme.error),
        borderRadius: BorderRadius.circular(12),
      ),
    );
  }

  void _showSuccess(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: Theme.of(context).colorScheme.primary,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        duration: Duration(milliseconds: 1000),
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        duration: Duration(milliseconds: 1000),
        backgroundColor: Theme.of(context).colorScheme.error,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  Future<void> _handleLogin() async {
    if (!_formKey.currentState!.validate()) return;

    final result = await ApiService.loginUser(
      email: _usernameController.text.trim(),
      password: _passwordController.text,
    );

    if (result.isSuccess) {
      _showSuccess("Login successful");

      // Store JWT token securely via TokenStorage (FlutterSecureStorage).
      // TODO: Consider additional hardening (e.g., biometrics or token rotation) if required.
      await TokenStorage.saveToken(result.token!);

      Future.delayed(const Duration(milliseconds: 800), () {
        if (!mounted) return;
        Navigator.pushReplacementNamed(context, '/home');
      });
    } else {
      _showError(result.errorMessage ?? "Login failed");
    }
  }

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    super.dispose();
  }
}
