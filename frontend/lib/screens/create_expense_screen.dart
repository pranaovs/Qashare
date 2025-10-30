import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import '../providers/groups_provider.dart';
import '../providers/expenses_provider.dart';
import '../models/expense.dart';
import '../utils/validators.dart';

class CreateExpenseScreen extends StatefulWidget {
  final String groupId;

  const CreateExpenseScreen({super.key, required this.groupId});

  @override
  State<CreateExpenseScreen> createState() => _CreateExpenseScreenState();
}

class _CreateExpenseScreenState extends State<CreateExpenseScreen> {
  final _formKey = GlobalKey<FormState>();
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _amountController = TextEditingController();
  
  final Map<String, TextEditingController> _paidControllers = {};
  final Map<String, TextEditingController> _owedControllers = {};
  final Map<String, bool> _selectedMembers = {};
  
  bool _isIncompleteAmount = false;
  bool _isIncompleteSplit = false;

  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      final group = context.read<GroupsProvider>().selectedGroup;
      if (group != null) {
        for (var member in group.members) {
          _paidControllers[member.userId] = TextEditingController(text: '0');
          _owedControllers[member.userId] = TextEditingController(text: '0');
          _selectedMembers[member.userId] = false;
        }
      }
    });
  }

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    _amountController.dispose();
    _paidControllers.values.forEach((c) => c.dispose());
    _owedControllers.values.forEach((c) => c.dispose());
    super.dispose();
  }

  void _equalSplit() {
    final amount = double.tryParse(_amountController.text);
    if (amount == null || amount <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please enter a valid amount first')),
      );
      return;
    }

    final selectedCount = _selectedMembers.values.where((s) => s).length;
    if (selectedCount == 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select members first')),
      );
      return;
    }

    final splitAmount = (amount / selectedCount).toStringAsFixed(2);
    setState(() {
      _selectedMembers.forEach((userId, selected) {
        if (selected) {
          _owedControllers[userId]?.text = splitAmount;
        }
      });
    });
  }

  Future<void> _createExpense() async {
    if (_formKey.currentState!.validate()) {
      final group = context.read<GroupsProvider>().selectedGroup;
      if (group == null) return;

      final splits = <ExpenseSplit>[];
      
      // Add paid splits
      _paidControllers.forEach((userId, controller) {
        final amount = double.tryParse(controller.text) ?? 0;
        if (amount > 0) {
          splits.add(ExpenseSplit(userId: userId, amount: amount, isPaid: true));
        }
      });

      // Add owed splits
      _owedControllers.forEach((userId, controller) {
        final amount = double.tryParse(controller.text) ?? 0;
        if (amount > 0) {
          splits.add(ExpenseSplit(userId: userId, amount: amount, isPaid: false));
        }
      });

      if (splits.isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Please add at least one split')),
        );
        return;
      }

      final expense = ExpenseRequest(
        groupId: widget.groupId,
        title: _titleController.text.trim(),
        description: _descriptionController.text.trim().isEmpty
            ? null
            : _descriptionController.text.trim(),
        amount: double.tryParse(_amountController.text) ?? 0,
        isIncompleteAmount: _isIncompleteAmount,
        isIncompleteSplit: _isIncompleteSplit,
        splits: splits,
      );

      final success = await context.read<ExpensesProvider>().createExpense(expense);

      if (success && mounted) {
        context.pop();
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Expense created successfully')),
        );
      } else if (mounted) {
        final error = context.read<ExpensesProvider>().error;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(error ?? 'Failed to create expense')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Add Expense'),
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            TextFormField(
              controller: _titleController,
              decoration: const InputDecoration(
                labelText: 'Title',
                hintText: 'e.g., Groceries',
                border: OutlineInputBorder(),
              ),
              validator: (value) => Validators.validateRequired(value, 'Title'),
            ),
            const SizedBox(height: 16),
            TextFormField(
              controller: _descriptionController,
              decoration: const InputDecoration(
                labelText: 'Description (Optional)',
                border: OutlineInputBorder(),
              ),
              maxLines: 2,
            ),
            const SizedBox(height: 16),
            TextFormField(
              controller: _amountController,
              decoration: const InputDecoration(
                labelText: 'Amount',
                prefixText: '\$ ',
                border: OutlineInputBorder(),
              ),
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              inputFormatters: [
                FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,2}')),
              ],
              validator: _isIncompleteAmount ? null : Validators.validateAmount,
            ),
            const SizedBox(height: 8),
            CheckboxListTile(
              title: const Text('Amount is incomplete'),
              value: _isIncompleteAmount,
              onChanged: (value) => setState(() => _isIncompleteAmount = value ?? false),
            ),
            CheckboxListTile(
              title: const Text('Split is incomplete'),
              value: _isIncompleteSplit,
              onChanged: (value) => setState(() => _isIncompleteSplit = value ?? false),
            ),
            const SizedBox(height: 24),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Split Details',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                TextButton.icon(
                  onPressed: _equalSplit,
                  icon: const Icon(Icons.calculate),
                  label: const Text('Equal Split'),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Consumer<GroupsProvider>(
              builder: (context, provider, child) {
                final group = provider.selectedGroup;
                if (group == null || group.members.isEmpty) {
                  return const Text('No members in this group');
                }

                return Column(
                  children: group.members.map((member) {
                    return Card(
                      child: Padding(
                        padding: const EdgeInsets.all(8.0),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            CheckboxListTile(
                              title: Text(member.name),
                              subtitle: Text(member.email),
                              value: _selectedMembers[member.userId] ?? false,
                              onChanged: (value) {
                                setState(() {
                                  _selectedMembers[member.userId] = value ?? false;
                                });
                              },
                            ),
                            if (_selectedMembers[member.userId] == true) ...[
                              Padding(
                                padding: const EdgeInsets.symmetric(horizontal: 16),
                                child: Row(
                                  children: [
                                    Expanded(
                                      child: TextField(
                                        controller: _paidControllers[member.userId],
                                        decoration: const InputDecoration(
                                          labelText: 'Paid',
                                          prefixText: '\$ ',
                                          border: OutlineInputBorder(),
                                        ),
                                        keyboardType: const TextInputType.numberWithOptions(decimal: true),
                                        inputFormatters: [
                                          FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,2}')),
                                        ],
                                      ),
                                    ),
                                    const SizedBox(width: 8),
                                    Expanded(
                                      child: TextField(
                                        controller: _owedControllers[member.userId],
                                        decoration: const InputDecoration(
                                          labelText: 'Owes',
                                          prefixText: '\$ ',
                                          border: OutlineInputBorder(),
                                        ),
                                        keyboardType: const TextInputType.numberWithOptions(decimal: true),
                                        inputFormatters: [
                                          FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,2}')),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              const SizedBox(height: 8),
                            ],
                          ],
                        ),
                      ),
                    );
                  }).toList(),
                );
              },
            ),
            const SizedBox(height: 24),
            Consumer<ExpensesProvider>(
              builder: (context, provider, child) {
                return ElevatedButton(
                  onPressed: provider.isLoading ? null : _createExpense,
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 16),
                  ),
                  child: provider.isLoading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Create Expense'),
                );
              },
            ),
          ],
        ),
      ),
    );
  }
}
