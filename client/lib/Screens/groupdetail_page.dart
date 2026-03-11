import 'package:flutter/material.dart';
import 'package:qashare/Models/expense_list_model.dart';
import 'package:qashare/Models/expense_model.dart';
import 'package:qashare/Models/settle_model.dart';
import 'package:qashare/Models/spending_model.dart';
import 'package:qashare/Screens/members_page.dart';
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
    final res = await ApiService.getGroupDetails(groupId: widget.groupId);

    final expenseRes = await ApiService.getGroupExpenses(
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
    // Show loading bottom sheet
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => _BalanceSheet(
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

          _spendingsTile(),
          _balanceTile(),
          _settlementHistoryTile(),

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
    final hasExpenses =
        _expenseResult != null &&
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

  Widget _settlementHistoryTile() {
    return Card(
      child: ListTile(
        leading: const Icon(Icons.history_rounded),
        title: const Text("Settlement History"),
        trailing: const Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text("View"),
            SizedBox(width: 4),
            Icon(Icons.chevron_right, size: 20),
          ],
        ),
        onTap: () {
          final group = _result!.group!;
          Navigator.pushNamed(
            context,
            "/settlement-history",
            arguments: {
              "groupId": widget.groupId,
              "groupName": widget.groupName,
              "members": group.members,
            },
          );
        },
      ),
    );
  }

  Widget _spendingsTile() {
    return Card(
      child: ListTile(
        leading: const Icon(Icons.receipt_long),
        title: const Text("My Expenses"),
        trailing: const Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text("View"),
            SizedBox(width: 4),
            Icon(Icons.chevron_right, size: 20),
          ],
        ),
        onTap: () {
          final group = _result!.group!;
          Navigator.pushNamed(
            context,
            "/spendings",
            arguments: {
              "groupId": widget.groupId,
              "groupName": widget.groupName,
              "members": group.members,
            },
          );
        },
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
  final String groupId;
  final List<Member> members;

  const _BalanceSheet({required this.groupId, required this.members});

  @override
  State<_BalanceSheet> createState() => _BalanceSheetState();
}

class _BalanceSheetState extends State<_BalanceSheet> {
  bool _loading = true;
  SettleResult? _settleResult;
  final Set<String> _settlingIds = {};

  @override
  void initState() {
    super.initState();
    _fetchSettlements();
  }

  Future<void> _fetchSettlements() async {
    final res = await ApiService.getGroupSettlements(groupId: widget.groupId);

    if (!mounted) return;
    setState(() {
      _settleResult = res;
      _loading = false;
    });
  }

  String _memberName(String userId) {
    final match = widget.members.where((m) => m.userId == userId);
    return match.isNotEmpty ? match.first.name : userId;
  }

  Future<void> _settleUp(Settlement s) async {
    final name = _memberName(s.userId);
    final isPositive = s.amount >= 0;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text("Settle Up"),
        content: Text(
          isPositive
              ? "Record that $name paid you ₹${s.amount.abs().toStringAsFixed(2)}?"
              : "Record that you paid $name ₹${s.amount.abs().toStringAsFixed(2)}?",
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text("Cancel"),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text("Settle"),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    setState(() => _settlingIds.add(s.userId));

    final res = await ApiService.settlePayment(
      groupId: widget.groupId,
      userId: s.userId,
      amount: -s.amount,
      title: "Settlement with $name",
    );

    if (!mounted) return;
    setState(() => _settlingIds.remove(s.userId));

    if (res.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text("Settled with $name ✓"),
          duration: const Duration(milliseconds: 1200),
        ),
      );
      // Refresh balances
      setState(() => _loading = true);
      await _fetchSettlements();
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(res.errorMessage ?? "Failed to settle"),
          backgroundColor: Theme.of(context).colorScheme.error,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;

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

              Text("Balances", style: theme.textTheme.titleLarge),
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
                      style: TextStyle(color: cs.error),
                    ),
                  ),
                )
              else if (_settleResult!.settlements!.isEmpty)
                const Expanded(child: Center(child: Text("All settled up! 🎉")))
              else
                Expanded(
                  child: ListView.separated(
                    controller: scrollController,
                    itemCount: _settleResult!.settlements!.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 8),
                    itemBuilder: (context, index) {
                      final s = _settleResult!.settlements![index];
                      final isPositive = s.amount >= 0;
                      final name = _memberName(s.userId);
                      final color = isPositive
                          ? const Color(0xFF2E7D32)
                          : cs.error;
                      final isSettling = _settlingIds.contains(s.userId);

                      return Card(
                        elevation: 0,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(14),
                          side: BorderSide(
                            color: color.withValues(alpha: 0.2),
                            width: 1,
                          ),
                        ),
                        child: Padding(
                          padding: const EdgeInsets.all(14),
                          child: Row(
                            children: [
                              // Avatar
                              CircleAvatar(
                                radius: 22,
                                backgroundColor: color.withValues(alpha: 0.12),
                                child: Icon(
                                  isPositive
                                      ? Icons.arrow_downward_rounded
                                      : Icons.arrow_upward_rounded,
                                  color: color,
                                  size: 22,
                                ),
                              ),
                              const SizedBox(width: 12),

                              // Name + description
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      name,
                                      style: const TextStyle(
                                        fontWeight: FontWeight.w600,
                                        fontSize: 15,
                                      ),
                                    ),
                                    const SizedBox(height: 2),
                                    Text(
                                      isPositive ? "owes you" : "you owe",
                                      style: TextStyle(
                                        color: color,
                                        fontSize: 12,
                                        fontWeight: FontWeight.w500,
                                      ),
                                    ),
                                  ],
                                ),
                              ),

                              // Amount
                              Text(
                                "₹${s.amount.abs().toStringAsFixed(2)}",
                                style: TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.bold,
                                  color: color,
                                ),
                              ),
                              const SizedBox(width: 10),

                              // Settle button
                              SizedBox(
                                height: 36,
                                child: FilledButton(
                                  onPressed: isSettling
                                      ? null
                                      : () => _settleUp(s),
                                  style: FilledButton.styleFrom(
                                    backgroundColor: color,
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 14,
                                    ),
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(10),
                                    ),
                                  ),
                                  child: isSettling
                                      ? const SizedBox(
                                          height: 16,
                                          width: 16,
                                          child: CircularProgressIndicator(
                                            strokeWidth: 2,
                                            color: Colors.white,
                                          ),
                                        )
                                      : const Text(
                                          "Settle",
                                          style: TextStyle(fontSize: 13),
                                        ),
                                ),
                              ),
                            ],
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

// ================= SPENDINGS BOTTOM SHEET =================

class _SpendingsSheet extends StatefulWidget {
  final String groupId;
  final List<Member> members;

  const _SpendingsSheet({required this.groupId, required this.members});

  @override
  State<_SpendingsSheet> createState() => _SpendingsSheetState();
}

class _SpendingsSheetState extends State<_SpendingsSheet> {
  bool _loading = true;
  SpendingResult? _result;

  @override
  void initState() {
    super.initState();
    _fetchSpendings();
  }

  Future<void> _fetchSpendings() async {
    final res = await ApiService.getUserSpendings(groupId: widget.groupId);

    if (!mounted) return;
    setState(() {
      _result = res;
      _loading = false;
    });
  }

  String _formatDate(int epochSecs) {
    final dt = DateTime.fromMillisecondsSinceEpoch(epochSecs * 1000);
    return "${dt.day}/${dt.month}/${dt.year}";
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final cs = theme.colorScheme;

    return DraggableScrollableSheet(
      initialChildSize: 0.55,
      minChildSize: 0.3,
      maxChildSize: 0.9,
      expand: false,
      builder: (context, scrollController) {
        return Padding(
          padding: const EdgeInsets.fromLTRB(20, 16, 20, 20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 40,
                height: 4,
                margin: const EdgeInsets.only(bottom: 16),
                decoration: BoxDecoration(
                  color: Colors.grey[400],
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              Text("My Expenses", style: theme.textTheme.titleLarge),
              const SizedBox(height: 4),
              Text(
                "Your share in each expense",
                style: TextStyle(fontSize: 13, color: cs.outline),
              ),
              const SizedBox(height: 16),
              if (_loading)
                const Expanded(
                  child: Center(child: CircularProgressIndicator()),
                )
              else if (!_result!.isSuccess)
                Expanded(
                  child: Center(
                    child: Text(
                      _result!.errorMessage ?? "Failed to load",
                      style: TextStyle(color: cs.error),
                    ),
                  ),
                )
              else if (_result!.spendings!.isEmpty)
                const Expanded(child: Center(child: Text("No expenses yet")))
              else
                Expanded(
                  child: Column(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 12,
                        ),
                        decoration: BoxDecoration(
                          color: cs.primaryContainer,
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              "Total you owe",
                              style: TextStyle(
                                fontWeight: FontWeight.w600,
                                color: cs.onPrimaryContainer,
                              ),
                            ),
                            Text(
                              "\u20b9${_result!.spendings!.fold<double>(0, (sum, s) => sum + s.userAmount).toStringAsFixed(2)}",
                              style: TextStyle(
                                fontWeight: FontWeight.w700,
                                fontSize: 16,
                                color: cs.onPrimaryContainer,
                              ),
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 12),
                      Expanded(
                        child: ListView.separated(
                          controller: scrollController,
                          itemCount: _result!.spendings!.length,
                          separatorBuilder: (_, __) =>
                              const SizedBox(height: 6),
                          itemBuilder: (context, index) {
                            final s = _result!.spendings![index];
                            return Card(
                              elevation: 0,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(14),
                                side: BorderSide(
                                  color: cs.outlineVariant.withValues(
                                    alpha: 0.3,
                                  ),
                                  width: 1,
                                ),
                              ),
                              child: ListTile(
                                contentPadding: const EdgeInsets.symmetric(
                                  horizontal: 14,
                                  vertical: 4,
                                ),
                                leading: CircleAvatar(
                                  radius: 20,
                                  backgroundColor: s.isSettlement
                                      ? cs.tertiaryContainer
                                      : cs.primaryContainer,
                                  child: Icon(
                                    s.isSettlement
                                        ? Icons.handshake_outlined
                                        : Icons.receipt_long_rounded,
                                    size: 20,
                                    color: s.isSettlement
                                        ? cs.onTertiaryContainer
                                        : cs.onPrimaryContainer,
                                  ),
                                ),
                                title: Text(
                                  s.title,
                                  style: const TextStyle(
                                    fontWeight: FontWeight.w600,
                                    fontSize: 14,
                                  ),
                                  overflow: TextOverflow.ellipsis,
                                ),
                                subtitle: Text(
                                  _formatDate(s.createdAt),
                                  style: TextStyle(
                                    fontSize: 12,
                                    color: cs.outline,
                                  ),
                                ),
                                trailing: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  crossAxisAlignment: CrossAxisAlignment.end,
                                  children: [
                                    Text(
                                      "\u20b9${s.userAmount.toStringAsFixed(2)}",
                                      style: TextStyle(
                                        fontWeight: FontWeight.w700,
                                        fontSize: 16,
                                        color: cs.error,
                                      ),
                                    ),
                                    Text(
                                      "of \u20b9${s.amount.toStringAsFixed(2)}",
                                      style: TextStyle(
                                        fontSize: 11,
                                        color: cs.outline,
                                      ),
                                    ),
                                  ],
                                ),
                                onTap: () {
                                  Navigator.pop(context);
                                  Navigator.pushNamed(
                                    context,
                                    "/expense-details",
                                    arguments: {
                                      "expenseId": s.expenseId,
                                      "members": widget.members,
                                    },
                                  );
                                },
                              ),
                            );
                          },
                        ),
                      ),
                    ],
                  ),
                ),
            ],
          ),
        );
      },
    );
  }
}
