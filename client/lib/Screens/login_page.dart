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
  String _serverUrl = "devbox:8080";

  bool _obscure = true;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
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

                  const Text(
                    "Welcome back",
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 40,
                      fontWeight: FontWeight.bold,
                    ),
                  ),

                  const SizedBox(height: 10),

                  const Text(
                    "Login into your account",
                    textAlign: TextAlign.center,
                    style: TextStyle(color: Colors.white70, fontSize: 12),
                  ),

                  const SizedBox(height: 40),

                  // Username
                  TextFormField(
                    controller: _usernameController,
                    style: const TextStyle(color: Colors.white),
                    decoration: _inputDecoration("Email"),
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return "Enter Email";
                      }
                      final emailRegex = RegExp(r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$');
                      if (!emailRegex.hasMatch(value)) {
                        return "Enter a valid email";
                      }
                      return null;
                    },
                  ),

                  const SizedBox(height: 16),

                  // Password
                  TextFormField(
                    controller: _passwordController,
                    obscureText: _obscure,
                    style: const TextStyle(color: Colors.white),
                    decoration: _inputDecoration("Password").copyWith(
                      suffixIcon: IconButton(
                        icon: Icon(
                          _obscure
                              ? Icons.visibility_outlined
                              : Icons.visibility_off_outlined,
                          color: Colors.white70,
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
                        backgroundColor: Colors.blue,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      child: const Text(
                        "Login",
                        style: TextStyle(color: Colors.white, fontSize: 16),
                      ),
                    ),
                  ),

                  const SizedBox(height: 20),

                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Text(
                        "Don't have an account?",
                        style: TextStyle(color: Colors.white70),
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
                      color: Colors.black26,
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
    final controller = TextEditingController(text: "$_serverUrl");

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
      labelStyle: const TextStyle(color: Colors.white70),
      enabledBorder: OutlineInputBorder(
        borderSide: const BorderSide(color: Colors.white38),
        borderRadius: BorderRadius.circular(12),
      ),
      focusedBorder: OutlineInputBorder(
        borderSide: const BorderSide(color: Colors.blue),
        borderRadius: BorderRadius.circular(12),
      ),
      errorBorder: OutlineInputBorder(
        borderSide: const BorderSide(color: Colors.redAccent),
        borderRadius: BorderRadius.circular(12),
      ),
      focusedErrorBorder: OutlineInputBorder(
        borderSide: const BorderSide(color: Colors.redAccent),
        borderRadius: BorderRadius.circular(12),
      ),
    );
  }

  void _showSuccess(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: Colors.green,
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: Colors.redAccent,
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

      await TokenStorage.saveToken(result.token!);

      Future.delayed(const Duration(milliseconds: 800), () {
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
