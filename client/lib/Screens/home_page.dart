import 'package:flutter/material.dart';
import 'package:qashare/Screens/groupdetail_page.dart';
import '../Config/token_storage.dart';
import '/Models/group_model.dart';
import '../Service/api_service.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  bool _loading = true;
  GroupListResult? _result;

  @override
  void initState() {
    super.initState();
    _loadGroups();
  }

  Future<void> _loadGroups() async {
    final token = await TokenStorage.getToken();

    if (token == null) {
      setState(() {
        _result = GroupListResult.error("Not logged in");
        _loading = false;
      });
      return;
    }

    final res = await ApiService.displayGroup(token);

    setState(() {
      _result = res;
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text("My Groups"),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 15),
            child: GestureDetector(
              onTap: () => Navigator.pushNamed(context, "/profile"),
              child: const CircleAvatar(
                radius: 18,
                child: Icon(Icons.person, size: 25),
              ),
            ),
          ),
        ],
      ),

      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _result!.isSuccess
          ? _groupListView()
          : _errorView(),
      floatingActionButton: FloatingActionButton(
        onPressed: () async {
          final created = await Navigator.pushNamed(context, "/creategroup");

          if (created == true) {
            _loadGroups(); // refresh list after creating group
          }
        },
        child: const Icon(Icons.add),
      ),
    );
  }

  Widget _groupListView() {
    final groups = _result!.groups!;

    if (groups.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.group,
              size: 100,
              color: Theme.of(context).colorScheme.outline,
            ),
            Text(
              "You are not part of any group yet",
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: Theme.of(context).colorScheme.outline,
              ),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadGroups,
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: groups.length,
        itemBuilder: (context, index) {
          final g = groups[index];

          return Card(
            margin: const EdgeInsets.only(bottom: 12),
            child: ListTile(
              leading: const Icon(Icons.group),
              title: Text(g.name),
              subtitle: Text(g.description),
              trailing: const Icon(Icons.chevron_right),
              onTap: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) =>
                        GroupDetailsPage(groupId: g.groupId, groupName: g.name),
                  ),
                );
              },
            ),
          );
        },
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
