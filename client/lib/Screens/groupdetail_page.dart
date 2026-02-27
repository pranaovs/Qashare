import 'package:flutter/material.dart';
import 'package:qashare/Models/expense_list_model.dart';
import 'package:qashare/Models/settle_model.dart';
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
  ExpenseListResult? _expenseResult;
  bool _expenseLoading = true;

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

    final expenseRes = await ApiService.getGroupExpenses(
      token: token,
      groupId: widget.groupId,
    );

    setState(() {
      _expenseResult = expenseRes;
      _expenseLoading = false;
      _result = res;
      _loading = false;
    });
  }

  // -------- BALANCE SHEET --------
  Future<void> _showBalanceSheet() async {
    final token = await TokenStorage.getToken();
    if (token == null) return;

    // Show loading bottom sheet
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => _BalanceSheet(
        token: token,
        groupId: widget.groupId,
        members: _result!.group!.members,
      ),
    );
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
          ? _content()
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
          _balanceTile(),

          const SizedBox(height: 40),

          const Text(
            "Expenses",
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),

          const SizedBox(height: 12),

          _expenseLoading
              ? const Center(child: CircularProgressIndicator())
              : !_expenseResult!.isSuccess
              ? Text(
                  _expenseResult!.errorMessage ?? "Failed to load expenses",
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                )
              : _expenseResult!.expenses!.isEmpty
              ? const Text("No expenses yet")
              : ListView.builder(
                  padding: const EdgeInsets.fromLTRB(16, 16, 16, 90),
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  itemCount: _expenseResult!.expenses!.length,
                  itemBuilder: (context, index) {
                    final e = _expenseResult!.expenses![index];
                    final date = DateTime.fromMillisecondsSinceEpoch(
                      e.createdAt * 1000,
                    );

                    return Card(
                      child: ListTile(
                        onTap: () async {
                          final group = _result!.group!;
                          final changed = await Navigator.pushNamed(
                            context,
                            "/expense-details",
                            arguments: {
                              "expenseId": e.expenseId,
                              "members": group.members,
                            },
                          );
                          if (changed == true) {
                            _loadDetails();
                          }
                        },
                        leading: const Icon(Icons.receipt_long),
                        title: Text(e.title),
                        subtitle: Text(
                          "${date.day}/${date.month}/${date.year}"
                          "${e.description != null ? " • ${e.description}" : ""}",
                        ),
                        trailing: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Text(
                              "₹${e.amount.toStringAsFixed(2)}",
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            if (e.isIncompleteAmount || e.isIncompleteSplit)
                              const Text(
                                "Incomplete",
                                style: TextStyle(
                                  fontSize: 11,
                                  color: Colors.orange,
                                ),
                              ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
        ],
      ),
    );
  }

  Widget _balanceTile() {
    final hasExpenses = _expenseResult != null &&
        _expenseResult!.isSuccess &&
        _expenseResult!.expenses != null &&
        _expenseResult!.expenses!.isNotEmpty;

    return Card(
      child: ListTile(
        leading: const Icon(Icons.account_balance_wallet),
        title: const Text("Balance"),
        trailing: hasExpenses
            ? const Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text("View"),
                  SizedBox(width: 4),
                  Icon(Icons.chevron_right, size: 20),
                ],
              )
            : const Text(
                "No settlements",
                style: TextStyle(color: Colors.grey),
              ),
        onTap: hasExpenses ? _showBalanceSheet : null,
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

// ================= BALANCE BOTTOM SHEET =================

class _BalanceSheet extends StatefulWidget {
  final String token;
  final String groupId;
  final List<Member> members;

  const _BalanceSheet({
    required this.token,
    required this.groupId,
    required this.members,
  });

  @override
  State<_BalanceSheet> createState() => _BalanceSheetState();
}

class _BalanceSheetState extends State<_BalanceSheet> {
  bool _loading = true;
  SettleResult? _settleResult;

  @override
  void initState() {
    super.initState();
    _fetchSettlements();
  }

  Future<void> _fetchSettlements() async {
    final res = await ApiService.getGroupSettlements(
      token: widget.token,
      groupId: widget.groupId,
    );

    setState(() {
      _settleResult = res;
      _loading = false;
    });
  }

  String _memberName(String userId) {
    final match = widget.members.where((m) => m.userId == userId);
    return match.isNotEmpty ? match.first.name : userId;
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return DraggableScrollableSheet(
      initialChildSize: 0.5,
      minChildSize: 0.3,
      maxChildSize: 0.85,
      expand: false,
      builder: (context, scrollController) {
        return Padding(
          padding: const EdgeInsets.fromLTRB(20, 16, 20, 20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // ---- Handle bar ----
              Container(
                width: 40,
                height: 4,
                margin: const EdgeInsets.only(bottom: 16),
                decoration: BoxDecoration(
                  color: Colors.grey[400],
                  borderRadius: BorderRadius.circular(2),
                ),
              ),

              Text(
                "Balances",
                style: theme.textTheme.titleLarge,
              ),
              const SizedBox(height: 4),
              const SizedBox(height: 16),

              if (_loading)
                const Expanded(
                  child: Center(child: CircularProgressIndicator()),
                )
              else if (!_settleResult!.isSuccess)
                Expanded(
                  child: Center(
                    child: Text(
                      _settleResult!.errorMessage ?? "Failed to load",
                      style: TextStyle(color: theme.colorScheme.error),
                    ),
                  ),
                )
              else if (_settleResult!.settlements!.isEmpty)
                const Expanded(
                  child: Center(
                    child: Text("All settled up! 🎉"),
                  ),
                )
              else
                Expanded(
                  child: ListView.separated(
                    controller: scrollController,
                    itemCount: _settleResult!.settlements!.length,
                    separatorBuilder: (_, __) => const Divider(height: 1),
                    itemBuilder: (context, index) {
                      final s = _settleResult!.settlements![index];
                      final isPositive = s.amount >= 0;
                      final name = _memberName(s.userId);

                      return ListTile(
                        leading: CircleAvatar(
                          backgroundColor: isPositive
                              ? Colors.green.withValues(alpha: 0.15)
                              : Colors.red.withValues(alpha: 0.15),
                          child: Icon(
                            isPositive
                                ? Icons.arrow_downward_rounded
                                : Icons.arrow_upward_rounded,
                            color: isPositive ? Colors.green : Colors.red,
                          ),
                        ),
                        title: Text(name),
                        subtitle: Text(
                          isPositive
                              ? "$name owes you"
                              : "You owe $name",
                          style: TextStyle(
                            color: isPositive ? Colors.green : Colors.red,
                            fontSize: 12,
                          ),
                        ),
                        trailing: Text(
                          "₹${s.amount.abs().toStringAsFixed(2)}",
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: isPositive ? Colors.green : Colors.red,
                          ),
                        ),
                      );
                    },
                  ),
                ),
            ],
          ),
        );
      },
    );
  }
}
