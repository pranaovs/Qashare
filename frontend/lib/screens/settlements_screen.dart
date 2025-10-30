import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/groups_provider.dart';
import '../models/settlement.dart';
import '../models/expense.dart';
import '../services/api_service.dart';
import '../utils/formatters.dart';

class SettlementsScreen extends StatefulWidget {
  final String groupId;

  const SettlementsScreen({super.key, required this.groupId});

  @override
  State<SettlementsScreen> createState() => _SettlementsScreenState();
}

class _SettlementsScreenState extends State<SettlementsScreen> {
  final ApiService _apiService = ApiService();
  List<Settlement>? _settlements;
  bool _isLoading = true;
  String? _error;
  final Map<int, bool> _selectedSettlements = {};
  bool _isRecording = false;

  @override
  void initState() {
    super.initState();
    _loadSettlements();
  }

  Future<void> _loadSettlements() async {
    setState(() {
      _isLoading = true;
      _error = null;
      _selectedSettlements.clear();
    });

    try {
      await _apiService.ensureInitialized();
      final settlements = await _apiService.getGroupSettlements(widget.groupId);
      setState(() {
        _settlements = settlements;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = _apiService.getErrorMessage(e);
        _isLoading = false;
      });
    }
  }

  Future<void> _recordSettlements() async {
    final selectedIndices = _selectedSettlements.entries
        .where((entry) => entry.value)
        .map((entry) => entry.key)
        .toList();

    if (selectedIndices.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select settlements to record')),
      );
      return;
    }

    final selectedSettlementsList = selectedIndices
        .map((index) => _settlements![index])
        .toList();

    // Show confirmation dialog
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Record Settlements'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('Record the following settlements as paid?'),
            const SizedBox(height: 16),
            ...selectedSettlementsList.map((settlement) {
              final fromName = _getUserName(settlement.fromUserId);
              final toName = _getUserName(settlement.toUserId);
              return Padding(
                padding: const EdgeInsets.symmetric(vertical: 4),
                child: Text(
                  '• $fromName pays $toName ${Formatters.formatCurrency(settlement.amount)}',
                  style: const TextStyle(fontSize: 14),
                ),
              );
            }).toList(),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Confirm'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    setState(() {
      _isRecording = true;
    });

    try {
      await _apiService.ensureInitialized();
      
      // Create settlement expenses for each selected settlement
      for (final settlement in selectedSettlementsList) {
        final expense = ExpenseRequest(
          groupId: widget.groupId,
          title: 'Settlement',
          description: 'Settlement payment',
          amount: settlement.amount,
          splits: [
            ExpenseSplit(
              userId: settlement.fromUserId,
              amount: settlement.amount,
              isPaid: true,
            ),
            ExpenseSplit(
              userId: settlement.toUserId,
              amount: settlement.amount,
              isPaid: false,
            ),
          ],
        );

        await _apiService.createExpense(expense);
      }

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              '${selectedSettlementsList.length} settlement(s) recorded successfully',
            ),
          ),
        );
        _loadSettlements(); // Reload to show updated settlements
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(_apiService.getErrorMessage(e) ?? 'Failed to record settlements')),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isRecording = false;
        });
      }
    }
  }

  String _getUserName(String userId) {
    final group = context.read<GroupsProvider>().selectedGroup;
    if (group == null) return 'User ${userId.substring(0, 8)}';
    
    final member = group.members.firstWhere(
      (m) => m.userId == userId,
      orElse: () => null as dynamic,
    );
    
    return member?.name ?? 'User ${userId.substring(0, 8)}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Settle Expenses'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadSettlements,
            tooltip: 'Refresh',
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Padding(
                    padding: const EdgeInsets.all(16.0),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(Icons.error_outline, size: 48, color: Colors.red),
                        const SizedBox(height: 16),
                        Text(_error!, textAlign: TextAlign.center),
                        const SizedBox(height: 16),
                        ElevatedButton(
                          onPressed: _loadSettlements,
                          child: const Text('Retry'),
                        ),
                      ],
                    ),
                  ),
                )
              : _settlements == null || _settlements!.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            Icons.check_circle_outline,
                            size: 64,
                            color: Colors.green.shade300,
                          ),
                          const SizedBox(height: 16),
                          Text(
                            'All Settled!',
                            style: Theme.of(context).textTheme.headlineSmall,
                          ),
                          const SizedBox(height: 8),
                          const Text(
                            'No pending settlements in this group',
                            style: TextStyle(color: Colors.grey),
                          ),
                        ],
                      ),
                    )
                  : ListView(
                      padding: const EdgeInsets.all(16),
                      children: [
                        Card(
                          color: Colors.blue.shade50,
                          child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Row(
                              children: [
                                Icon(Icons.info_outline, color: Colors.blue.shade700),
                                const SizedBox(width: 12),
                                Expanded(
                                  child: Text(
                                    'These are simplified settlements to minimize transactions',
                                    style: TextStyle(
                                      color: Colors.blue.shade700,
                                      fontSize: 13,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                        const SizedBox(height: 16),
                        ..._settlements!.asMap().entries.map((entry) {
                          final index = entry.key;
                          final settlement = entry.value;
                          final fromName = _getUserName(settlement.fromUserId);
                          final toName = _getUserName(settlement.toUserId);
                          final isSelected = _selectedSettlements[index] ?? false;
                          
                          return Card(
                            margin: const EdgeInsets.only(bottom: 12),
                            child: CheckboxListTile(
                              value: isSelected,
                              onChanged: (value) {
                                setState(() {
                                  _selectedSettlements[index] = value ?? false;
                                });
                              },
                              contentPadding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 8,
                              ),
                              secondary: CircleAvatar(
                                backgroundColor: Colors.orange.shade100,
                                child: Text(
                                  fromName.substring(0, 1).toUpperCase(),
                                  style: TextStyle(
                                    color: Colors.orange.shade900,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                              title: Text.rich(
                                TextSpan(
                                  children: [
                                    TextSpan(
                                      text: fromName,
                                      style: const TextStyle(
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                    const TextSpan(text: ' pays '),
                                    TextSpan(
                                      text: toName,
                                      style: const TextStyle(
                                        fontWeight: FontWeight.bold,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              subtitle: Text(
                                Formatters.formatCurrency(settlement.amount),
                                style: TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.bold,
                                  color: Theme.of(context).colorScheme.primary,
                                ),
                              ),
                            ),
                          );
                        }).toList(),
                        const SizedBox(height: 16),
                        if (_settlements!.isNotEmpty)
                          SizedBox(
                            width: double.infinity,
                            child: FilledButton.icon(
                              onPressed: _isRecording ? null : _recordSettlements,
                              icon: _isRecording
                                  ? const SizedBox(
                                      width: 20,
                                      height: 20,
                                      child: CircularProgressIndicator(
                                        strokeWidth: 2,
                                        color: Colors.white,
                                      ),
                                    )
                                  : const Icon(Icons.check_circle),
                              label: Text(_isRecording ? 'Recording...' : 'Record Settlements'),
                              style: FilledButton.styleFrom(
                                padding: const EdgeInsets.symmetric(vertical: 16),
                              ),
                            ),
                          ),
                      ],
                    ),
    );
  }
}
