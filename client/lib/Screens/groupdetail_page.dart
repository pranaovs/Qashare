import 'package:flutter/material.dart';
import 'package:qashare/Screens/members_page.dart';
import '../Config/token_storage.dart';
import '../Service/api_service.dart';
import 'package:qashare/Models/groupdetail_model.dart';

class GroupDetailsPage extends StatefulWidget {
  final String groupId;
  final String groupName;

  const GroupDetailsPage({
    super.key,
    required this.groupId,
    required this.groupName,
  });

  @override
  State<GroupDetailsPage> createState() => _GroupDetailsPageState();
}

class _GroupDetailsPageState extends State<GroupDetailsPage> {
  bool _loading = true;
  GroupDetailsResult? _result;

  @override
  void initState() {
    super.initState();
    _loadDetails();
  }

  Future<void> _loadDetails() async {
    final token = await TokenStorage.getToken();

    if (token == null) {
      setState(() {
        _result = GroupDetailsResult.error("Not logged in");
        _loading = false;
      });
      return;
    }

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
      appBar: AppBar(
        title: Text(widget.groupName),
        actions: [
          IconButton(
            icon: const Icon(Icons.group_add_rounded),
            tooltip: "Members",
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => MembersPage(groupId: widget.groupId),
                ),
              );
            },
          ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _result == null
          ? const Center(child: Text("Something went wrong"))
          : _result!.isSuccess
          ? _content() // ✅ HERE it is used
          : _errorView(),

      floatingActionButton: FloatingActionButton.extended(
        onPressed: () async {
          final group = _result!.group!;

          final created = await Navigator.pushNamed(
            context,
            "/create-expense",
            arguments: {"groupId": widget.groupId, "members": group.members},
          );

          if (created == true) {
            _loadDetails(); // refresh expenses after adding
          }
        },
        icon: const Icon(Icons.add),
        label: const Text("Add Expense"),
      ),
    );
  }

  Widget _content() {
    final group = _result!.group!;
    final created = DateTime.fromMillisecondsSinceEpoch(group.createdAt * 1000);

    return RefreshIndicator(
      onRefresh: _loadDetails,
      child: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          // -------- TRIP CARD --------
          Card(
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(14),
            ),
            child: Padding(
              padding: const EdgeInsets.all(18),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    group.name,
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 8),

                  Text(
                    group.description.isEmpty
                        ? "No description provided"
                        : group.description,
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),

                  const SizedBox(height: 14),

                  Row(
                    children: [
                      const Icon(Icons.calendar_month, size: 18),
                      const SizedBox(width: 8),
                      Text(
                        "Trip started on "
                        "${created.day}/${created.month}/${created.year}",
                      ),
                    ],
                  ),

                  const SizedBox(height: 10),
                  Row(
                    children: [
                      const Icon(Icons.group, size: 18),
                      const SizedBox(width: 8),
                      Text("${group.members.length} members"),
                    ],
                  ),
                ],
              ),
            ),
          ),

          const SizedBox(height: 30),

          // -------- QUICK INFO --------
          const Text(
            "Trip Overview",
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 12),

          _infoTile(Icons.receipt_long, "Expenses", "Coming soon"),
          _infoTile(Icons.account_balance_wallet, "Balance", "Coming soon"),

          const SizedBox(height: 40),
        ],
      ),
    );
  }

  Widget _infoTile(IconData icon, String title, String value) {
    return Card(
      child: ListTile(
        leading: Icon(icon),
        title: Text(title),
        trailing: Text(value),
      ),
    );
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
