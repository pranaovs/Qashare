import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import '../providers/auth_provider.dart';
import '../utils/formatters.dart';

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Profile'),
      ),
      body: Consumer<AuthProvider>(
        builder: (context, authProvider, child) {
          final user = authProvider.currentUser;

          if (user == null) {
            return const Center(child: Text('No user data'));
          }

          return ListView(
            children: [
              const SizedBox(height: 32),
              Center(
                child: CircleAvatar(
                  radius: 48,
                  backgroundColor: Theme.of(context).colorScheme.primaryContainer,
                  child: Text(
                    Formatters.getInitials(user.name),
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                          color: Theme.of(context).colorScheme.onPrimaryContainer,
                        ),
                  ),
                ),
              ),
              const SizedBox(height: 16),
              Text(
                user.name,
                style: Theme.of(context).textTheme.headlineSmall,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              Text(
                user.email,
                style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      color: Colors.grey,
                    ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 32),
              ListTile(
                leading: const Icon(Icons.calendar_today),
                title: const Text('Member since'),
                subtitle: Text(Formatters.formatDate(user.createdAtDateTime)),
              ),
              const Divider(),
              ListTile(
                leading: const Icon(Icons.info_outline),
                title: const Text('About'),
                onTap: () {
                  showAboutDialog(
                    context: context,
                    applicationName: 'Shared Expenses',
                    applicationVersion: '1.0.0',
                    applicationLegalese: '© 2024 Shared Expenses\nLicensed under GPL v3',
                    children: [
                      const SizedBox(height: 16),
                      const Text(
                        'A FOSS expense sharing application - Splitwise alternative.\n\n'
                        'Features:\n'
                        '• No ads\n'
                        '• Open source\n'
                        '• Privacy focused\n'
                        '• Free forever',
                      ),
                    ],
                  );
                },
              ),
              const Divider(),
              ListTile(
                leading: const Icon(Icons.logout, color: Colors.red),
                title: const Text('Logout', style: TextStyle(color: Colors.red)),
                onTap: () async {
                  final confirm = await showDialog<bool>(
                    context: context,
                    builder: (context) => AlertDialog(
                      title: const Text('Logout'),
                      content: const Text('Are you sure you want to logout?'),
                      actions: [
                        TextButton(
                          onPressed: () => Navigator.pop(context, false),
                          child: const Text('Cancel'),
                        ),
                        TextButton(
                          onPressed: () => Navigator.pop(context, true),
                          style: TextButton.styleFrom(foregroundColor: Colors.red),
                          child: const Text('Logout'),
                        ),
                      ],
                    ),
                  );

                  if (confirm == true && context.mounted) {
                    await context.read<AuthProvider>().logout();
                    if (context.mounted) {
                      context.go('/login');
                    }
                  }
                },
              ),
            ],
          );
        },
      ),
    );
  }
}
