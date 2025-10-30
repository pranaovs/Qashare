import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import '../providers/expenses_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/groups_provider.dart';
import '../utils/formatters.dart';

class ExpenseDetailsScreen extends StatefulWidget {
  final String expenseId;

  const ExpenseDetailsScreen({super.key, required this.expenseId});

  @override
  State<ExpenseDetailsScreen> createState() => _ExpenseDetailsScreenState();
}

class _ExpenseDetailsScreenState extends State<ExpenseDetailsScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      context.read<ExpensesProvider>().loadExpense(widget.expenseId);
    });
  }

  Future<void> _deleteExpense() async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete Expense'),
        content: const Text('Are you sure you want to delete this expense?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Delete'),
          ),
        ],
      ),
    );

    if (confirm == true && mounted) {
      final expense = context.read<ExpensesProvider>().selectedExpense;
      if (expense != null) {
        final success = await context
            .read<ExpensesProvider>()
            .deleteExpense(widget.expenseId, expense.groupId);
        if (success && mounted) {
          context.pop();
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Expense deleted')),
          );
        }
      }
    }
  }

  bool _canEdit(String? currentUserId, String addedBy, String createdBy) {
    return currentUserId == addedBy || currentUserId == createdBy;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Expense Details'),
        actions: [
          Consumer3<ExpensesProvider, AuthProvider, GroupsProvider>(
            builder: (context, expensesProvider, authProvider, groupsProvider, child) {
              final expense = expensesProvider.selectedExpense;
              final currentUserId = authProvider.currentUser?.userId;
              
              if (expense == null || currentUserId == null) return const SizedBox();

              // Load group to check if user is admin
              Future.microtask(() {
                if (groupsProvider.selectedGroup?.groupId != expense.groupId) {
                  groupsProvider.loadGroup(expense.groupId);
                }
              });

              final group = groupsProvider.selectedGroup;
              final canEdit = _canEdit(
                currentUserId,
                expense.addedBy,
                group?.createdBy ?? '',
              );

              if (!canEdit) return const SizedBox();

              return PopupMenuButton(
                itemBuilder: (context) => [
                  const PopupMenuItem(
                    value: 'delete',
                    child: Text('Delete', style: TextStyle(color: Colors.red)),
                  ),
                ],
                onSelected: (value) {
                  if (value == 'delete') {
                    _deleteExpense();
                  }
                },
              );
            },
          ),
        ],
      ),
      body: Consumer<ExpensesProvider>(
        builder: (context, provider, child) {
          if (provider.isLoading && provider.selectedExpense == null) {
            return const Center(child: CircularProgressIndicator());
          }

          final expense = provider.selectedExpense;
          if (expense == null) {
            return const Center(child: Text('Expense not found'));
          }

          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        expense.title,
                        style: Theme.of(context).textTheme.headlineSmall,
                      ),
                      if (expense.description != null && expense.description!.isNotEmpty) ...[
                        const SizedBox(height: 8),
                        Text(expense.description!),
                      ],
                      const SizedBox(height: 16),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            'Amount',
                            style: Theme.of(context).textTheme.titleMedium,
                          ),
                          Text(
                            Formatters.formatCurrency(expense.amount),
                            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                                  color: Theme.of(context).colorScheme.primary,
                                  fontWeight: FontWeight.bold,
                                ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      Text(
                        Formatters.formatDateTime(expense.createdAtDateTime),
                        style: Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Colors.grey,
                            ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
              
              // Who Paid Section
              Text(
                'Who Paid',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 8),
              Card(
                child: expense.paidSplits.isEmpty
                    ? const Padding(
                        padding: EdgeInsets.all(16),
                        child: Text('No payers'),
                      )
                    : Column(
                        children: expense.paidSplits.map((split) {
                          return ListTile(
                            leading: CircleAvatar(
                              child: Text(split.userId.substring(0, 2).toUpperCase()),
                            ),
                            title: Text('User ${split.userId.substring(0, 8)}'),
                            trailing: Text(
                              Formatters.formatCurrency(split.amount),
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                                color: Colors.green,
                              ),
                            ),
                          );
                        }).toList(),
                      ),
              ),
              const SizedBox(height: 16),
              
              // Who Owes Section
              Text(
                'Who Owes',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: 8),
              Card(
                child: expense.owedSplits.isEmpty
                    ? const Padding(
                        padding: EdgeInsets.all(16),
                        child: Text('No borrowers'),
                      )
                    : Column(
                        children: expense.owedSplits.map((split) {
                          return ListTile(
                            leading: CircleAvatar(
                              child: Text(split.userId.substring(0, 2).toUpperCase()),
                            ),
                            title: Text('User ${split.userId.substring(0, 8)}'),
                            trailing: Text(
                              Formatters.formatCurrency(split.amount),
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                                color: Colors.orange,
                              ),
                            ),
                          );
                        }).toList(),
                      ),
              ),
              
              if (expense.isIncompleteAmount || expense.isIncompleteSplit) ...[
                const SizedBox(height: 16),
                Card(
                  color: Colors.orange.shade50,
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Row(
                      children: [
                        const Icon(Icons.warning_amber, color: Colors.orange),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            expense.isIncompleteAmount
                                ? 'This expense has incomplete amount data'
                                : 'This expense has incomplete split data',
                            style: const TextStyle(color: Colors.orange),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ],
          );
        },
      ),
    );
  }
}
