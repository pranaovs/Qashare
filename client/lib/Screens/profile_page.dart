import 'package:flutter/material.dart';
import 'package:qashare/Config/token_storage.dart';
import 'package:qashare/Service/api_service.dart';
import 'package:qashare/Models/user_models.dart';

class ProfilePage extends StatefulWidget {
  const ProfilePage({super.key});

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  UserResult? _result;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadProfile();
  }

  Future<void> _loadProfile() async {
    final token = await TokenStorage.getToken();

    if (token == null) {
      setState(() {
        _result = UserResult.error("Not logged in");
        _loading = false;
      });
      return;
    }

    final res = await ApiService.getCurrentUser(token);

    setState(() {
      _result = res;
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text("Profile")),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _result!.isSuccess
          ? _profileView()
          : _errorView(),
    );
  }

  Widget _profileView() {
    final joinedDate = DateTime.fromMillisecondsSinceEpoch(
      _result!.createdAt! * 1000,
    );

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            const SizedBox(height: 20),
            const CircleAvatar(radius: 60, child: Icon(Icons.person, size: 80)),
            const SizedBox(height: 25),
            Text(
              _result!.name ?? "",
              style: Theme.of(
                context,
              ).textTheme.headlineMedium?.copyWith(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 10),
            Text(
              _result!.email ?? "",
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: Theme.of(context).colorScheme.outline,
              ),
            ),
            const SizedBox(height: 30),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  Icons.calendar_month,
                  color: Theme.of(context).colorScheme.outline,
                  size: 18,
                ),
                const SizedBox(width: 8),
                Text(
                  "Member since ${_formatDate(joinedDate)}",
                  style: TextStyle(
                    color: Theme.of(context).colorScheme.outline,
                  ),
                ),
              ],
            ),
            const Spacer(),
            SizedBox(
              width: double.infinity,
              height: 45,
              child: OutlinedButton.icon(
                onPressed: _confirmLogout,
                icon: Icon(
                  Icons.logout,
                  color: Theme.of(context).colorScheme.error,
                ),
                label: Text(
                  "Logout",
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
                style: OutlinedButton.styleFrom(
                  side: BorderSide(color: Theme.of(context).colorScheme.error),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _confirmLogout() {
    showDialog(
      context: context,
      builder: (_) => AlertDialog(
        title: const Text("Logout"),
        content: const Text("Are you sure you want to logout?"),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text("Cancel"),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _handleLogout();
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Theme.of(context).colorScheme.error,
              foregroundColor: Theme.of(context).colorScheme.onError,
            ),
            child: const Text("Logout"),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime dt) {
    return "${dt.day}/${dt.month}/${dt.year}";
  }

  Future<void> _handleLogout() async {
    await TokenStorage.clear();
    if (!mounted) return;
    Navigator.pushNamedAndRemoveUntil(context, '/login', (route) => false);
  }

  Widget _errorView() {
    return Center(
      child: Text(
        _result!.errorMessage ?? "Error",
        style: TextStyle(color: Theme.of(context).colorScheme.error),
      ),
    );
  }
}
