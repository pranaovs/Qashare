import 'package:flutter/material.dart';
import '../Config/token_storage.dart';
import 'package:qashare/Models/groupdetail_model.dart';
import '../Service/api_service.dart';

class MembersPage extends StatefulWidget {
  final String groupId;

  const MembersPage({super.key, required this.groupId});

  @override
  State<MembersPage> createState() => _MembersPageState();
}

class _MembersPageState extends State<MembersPage> {
  bool _loading = true;
  GroupDetailsResult? _result;

  @override
  void initState() {
    super.initState();
    _loadMembers();
  }

  Future<void> _loadMembers() async {
    final token = await TokenStorage.getToken();
    if (token == null) return;

    final res = await ApiService.getGroupDetails(
      token: token,
      groupId: widget.groupId,
    );

    setState(() {
      _result = res;
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text("Members")),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _result!.isSuccess
          ? _memberList()
          : _errorView(),

      floatingActionButton: FloatingActionButton(
        onPressed: _showAddMemberPopup,
        child: const Icon(Icons.person_add),
      ),
    );
  }

  Widget _memberList() {
    final members = _result!.group!.members;

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: members.length,
      itemBuilder: (_, i) {
        final m = members[i];
        return Card(
          child: ListTile(
            leading: const CircleAvatar(child: Icon(Icons.person)),
            title: Text(m.name),
            subtitle: Text(m.email),
            trailing: m.guest
                ? const Text("Guest", style: TextStyle(fontSize: 12))
                : null,
          ),
        );
      },
    );
  }

  Widget _errorView() {
    return Center(
      child: Text(
        _result!.errorMessage ?? "Error",
        style: const TextStyle(color: Colors.red),
      ),
    );
  }

  // ---------------- ADD MEMBER POPUP ----------------

  void _showAddMemberPopup() {
    final controller = TextEditingController();
    bool loading = false;

    showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setLocal) => AlertDialog(
          title: const Text("Add Member by Email"),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: controller,
                keyboardType: TextInputType.emailAddress,
                decoration: const InputDecoration(
                  labelText: "Email address",
                ),
              ),
              if (loading)
                const Padding(
                  padding: EdgeInsets.only(top: 12),
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text("Cancel"),
            ),
            ElevatedButton(
              onPressed: loading
                  ? null
                  : () async {
                final email = controller.text.trim();
                if (email.isEmpty) return;

                setLocal(() => loading = true);

                final token = await TokenStorage.getToken();
                if (token == null) return;

                final lookup = await ApiService.searchUserByEmail(
                  token: token,
                  email: email,
                );

                if (!lookup.isSuccess) {
                  Navigator.pop(ctx);
                  _showSnack(lookup.errorMessage!, true);
                  return;
                }

                final addResult = await ApiService.addMembersToGroup(
                  token: token,
                  groupId: widget.groupId,
                  userIds: [lookup.user!.userId],
                );

                Navigator.pop(ctx);

                if (addResult.isSuccess) {
                  _showSnack("Member added", false);
                  _loadMembers();
                } else {
                  _showSnack(addResult.errorMessage!, true);
                }
              },
              child: const Text("Add"),
            ),
          ],
        ),
      ),
    );
  }

  void _showSnack(String msg, bool error) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor:
        error ? Theme.of(context).colorScheme.error : null,
        duration: const Duration(milliseconds: 900),
        behavior: SnackBarBehavior.floating,
      ),
    );
  }
}

