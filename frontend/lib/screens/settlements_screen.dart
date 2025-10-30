import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/groups_provider.dart';
import '../models/settlement.dart';
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

  @override
  void initState() {
    super.initState();
    _loadSettlements();
  }

  Future<void> _loadSettlements() async {
    setState(() {
      _isLoading = true;
      _error = null;
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
                        ..._settlements!.map((settlement) {
                          final fromName = _getUserName(settlement.fromUserId);
                          final toName = _getUserName(settlement.toUserId);
                          
                          return Card(
                            margin: const EdgeInsets.only(bottom: 12),
                            child: ListTile(
                              contentPadding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 8,
                              ),
                              leading: CircleAvatar(
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
                              trailing: Text(
                                Formatters.formatCurrency(settlement.amount),
                                style: TextStyle(
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
                                  color: Theme.of(context).colorScheme.primary,
                                ),
                              ),
                            ),
                          );
                        }).toList(),
                      ],
                    ),
    );
  }
}
