import 'package:flutter/material.dart';
import '../models/expense.dart';
import '../utils/formatters.dart';

class ExpenseListTile extends StatelessWidget {
  final Expense expense;
  final VoidCallback onTap;

  const ExpenseListTile({
    super.key,
    required this.expense,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: ListTile(
        onTap: onTap,
        leading: CircleAvatar(
          backgroundColor: Theme.of(context).colorScheme.secondaryContainer,
          child: Icon(
            Icons.receipt,
            color: Theme.of(context).colorScheme.onSecondaryContainer,
          ),
        ),
        title: Text(
          expense.title,
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Text(Formatters.formatRelativeTime(expense.createdAtDateTime)),
        trailing: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text(
              Formatters.formatCurrency(expense.amount),
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    color: Theme.of(context).colorScheme.primary,
                    fontWeight: FontWeight.bold,
                  ),
            ),
            if (expense.isIncompleteAmount || expense.isIncompleteSplit)
              const Icon(Icons.warning_amber, size: 16, color: Colors.orange),
          ],
        ),
      ),
    );
  }
}
