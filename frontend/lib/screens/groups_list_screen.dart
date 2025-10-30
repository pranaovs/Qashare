import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import '../providers/groups_provider.dart';
import '../providers/auth_provider.dart';
import '../widgets/group_card.dart';

class GroupsListScreen extends StatefulWidget {
  const GroupsListScreen({super.key});

  @override
  State<GroupsListScreen> createState() => _GroupsListScreenState();
}

class _GroupsListScreenState extends State<GroupsListScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() => context.read<GroupsProvider>().loadGroups());
  }

  Future<void> _refresh() async {
    await context.read<GroupsProvider>().loadGroups();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('My Groups'),
        actions: [
          IconButton(
            icon: const Icon(Icons.person),
            onPressed: () => context.push('/profile'),
          ),
        ],
      ),
      body: Consumer<GroupsProvider>(
        builder: (context, groupsProvider, child) {
          if (groupsProvider.isLoading && groupsProvider.groups.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }

          if (groupsProvider.error != null && groupsProvider.groups.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error_outline, size: 64, color: Colors.red),
                  const SizedBox(height: 16),
                  Text(groupsProvider.error!),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: _refresh,
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (groupsProvider.groups.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.group_outlined,
                    size: 64,
                    color: Colors.grey[400],
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'No groups yet',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Create a group to start sharing expenses',
                    style: TextStyle(color: Colors.grey[600]),
                  ),
                ],
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: _refresh,
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: groupsProvider.groups.length,
              itemBuilder: (context, index) {
                final group = groupsProvider.groups[index];
                return GroupCard(
                  group: group,
                  onTap: () => context.push('/groups/${group.groupId}'),
                );
              },
            ),
          );
        },
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.push('/groups/create'),
        icon: const Icon(Icons.add),
        label: const Text('New Group'),
      ),
    );
  }
}
