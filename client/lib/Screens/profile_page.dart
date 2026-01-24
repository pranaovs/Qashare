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
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        title: const Text("Profile", style: TextStyle(color: Colors.white)),
      ),
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

            const CircleAvatar(radius: 60, child: Icon(Icons.person, size: 50)),

            const SizedBox(height: 25),

            // -------- NAME --------
            Text(
              _result!.name ?? "",
              style: const TextStyle(
                color: Colors.white,
                fontSize: 30,
                fontWeight: FontWeight.bold,
              ),
            ),

            const SizedBox(height: 10),

            // -------- EMAIL --------
            Text(
              _result!.email ?? "",
              style: const TextStyle(color: Colors.white70, fontSize: 14),
            ),

            const SizedBox(height: 30),

            // -------- MEMBER SINCE --------
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(
                  Icons.calendar_month,
                  color: Colors.white70,
                  size: 18,
                ),
                const SizedBox(width: 8),
                Text(
                  "Member since ${_formatDate(joinedDate)}",
                  style: const TextStyle(color: Colors.white70),
                ),
              ],
            ),

            const Spacer(),
            // -------- LOGOUT BUTTON --------
            SizedBox(
              width: double.infinity,
              height: 45,
              child: OutlinedButton.icon(
                onPressed: _confirmLogout,
                icon: const Icon(Icons.logout, color: Colors.redAccent),
                label: const Text(
                  "Logout",
                  style: TextStyle(color: Colors.redAccent),
                ),
                style: OutlinedButton.styleFrom(
                  side: const BorderSide(color: Colors.redAccent),
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
        backgroundColor: Colors.black,
        title: const Text("Logout", style: TextStyle(color: Colors.white)),
        content: const Text(
          "Are you sure you want to logout?",
          style: TextStyle(color: Colors.white),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text("Cancel", style: TextStyle(color: Colors.white)),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context); // close dialog
              _handleLogout(); // perform logout
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.redAccent),
            child: const Text("Logout", style: TextStyle(color: Colors.white)),
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
        style: const TextStyle(color: Colors.red),
      ),
    );
  }
}
