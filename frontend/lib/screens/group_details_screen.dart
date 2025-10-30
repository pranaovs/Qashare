import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import '../providers/groups_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/expenses_provider.dart';
import '../widgets/member_list_tile.dart';
import '../widgets/expense_list_tile.dart';
import '../utils/formatters.dart';

class GroupDetailsScreen extends StatefulWidget {
  final String groupId;

  const GroupDetailsScreen({super.key, required this.groupId});

  @override
  State<GroupDetailsScreen> createState() => _GroupDetailsScreenState();
}

class _GroupDetailsScreenState extends State<GroupDetailsScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      context.read<GroupsProvider>().loadGroup(widget.groupId);
    });
  }

  Future<void> _showAddMemberDialog() async {
    final emailController = TextEditingController();
    final result = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Add Member'),
        content: TextField(
          controller: emailController,
          decoration: const InputDecoration(
            labelText: 'Email',
            hintText: 'Enter user email',
          ),
          keyboardType: TextInputType.emailAddress,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () async {
              final email = emailController.text.trim();
              if (email.isNotEmpty) {
                final user = await context.read<GroupsProvider>().searchUserByEmail(email);
                if (user != null && mounted) {
                  final success = await context
                      .read<GroupsProvider>()
                      .addMembers(widget.groupId, [user.userId]);
                  if (mounted) {
                    Navigator.pop(context, success);
                  }
                } else if (mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('User not found')),
                  );
                }
              }
            },
            child: const Text('Add'),
          ),
        ],
      ),
    );

    if (result == true && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Member added successfully')),
      );
    }
  }

  Future<void> _removeMember(String userId, String userName) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Remove Member'),
        content: Text('Remove $userName from this group?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Remove'),
          ),
        ],
      ),
    );

    if (confirm == true && mounted) {
      final success = await context.read<GroupsProvider>().removeMember(widget.groupId, userId);
      if (success && mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Member removed')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Group Details'),
      ),
      body: Consumer2<GroupsProvider, AuthProvider>(
        builder: (context, groupsProvider, authProvider, child) {
          if (groupsProvider.isLoading && groupsProvider.selectedGroup == null) {
            return const Center(child: CircularProgressIndicator());
          }

          final group = groupsProvider.selectedGroup;
          if (group == null) {
            return const Center(child: Text('Group not found'));
          }

          final isAdmin = groupsProvider.isUserAdmin(authProvider.currentUser?.userId ?? '');

          return ListView(
            children: [
              // Group Info Card
              Card(
                margin: const EdgeInsets.all(16),
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        group.name,
                        style: Theme.of(context).textTheme.headlineSmall,
                      ),
                      if (group.description != null && group.description!.isNotEmpty) ...[
                        const SizedBox(height: 8),
                        Text(
                          group.description!,
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                      ],
                      const SizedBox(height: 16),
                      Row(
                        children: [
                          const Icon(Icons.people, size: 16),
                          const SizedBox(width: 4),
                          Text('${group.memberCount} members'),
                          const SizedBox(width: 16),
                          const Icon(Icons.calendar_today, size: 16),
                          const SizedBox(width: 4),
                          Text(Formatters.formatDate(group.createdAtDateTime)),
                        ],
                      ),
                    ],
                  ),
                ),
              ),

              // Members Section
              Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Members',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                    if (isAdmin)
                      IconButton(
                        icon: const Icon(Icons.person_add),
                        onPressed: _showAddMemberDialog,
                      ),
                  ],
                ),
              ),
              ...group.members.map((member) => MemberListTile(
                    member: member,
                    isAdmin: member.userId == group.createdBy,
                    canRemove: isAdmin && member.userId != group.createdBy,
                    onRemove: () => _removeMember(member.userId, member.name),
                  )),

              // Expenses Section
              Padding(
                padding: const EdgeInsets.all(16),
                child: Text(
                  'Recent Expenses',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
              ),
              Consumer<ExpensesProvider>(
                builder: (context, expensesProvider, child) {
                  final expenses = expensesProvider.getExpensesForGroup(widget.groupId);
                  if (expenses.isEmpty) {
                    return const Padding(
                      padding: EdgeInsets.all(16),
                      child: Text('No expenses yet'),
                    );
                  }
                  return Column(
                    children: expenses
                        .map((expense) => ExpenseListTile(
                              expense: expense,
                              onTap: () => context.push('/expenses/${expense.expenseId}'),
                            ))
                        .toList(),
                  );
                },
              ),
            ],
          );
        },
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => context.push('/expenses/create', extra: widget.groupId),
        icon: const Icon(Icons.add),
        label: const Text('Add Expense'),
      ),
    );
  }
}
