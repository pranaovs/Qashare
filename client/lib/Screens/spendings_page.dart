import 'package:flutter/material.dart';
import 'package:qashare/Config/token_storage.dart';
import 'package:qashare/Models/groupdetail_model.dart';
import 'package:qashare/Models/spending_model.dart';
import 'package:qashare/Service/api_service.dart';

class SpendingsPage extends StatefulWidget {
  final String groupId;
  final String groupName;
  final List<Member> members;

  const SpendingsPage({
    super.key,
    required this.groupId,
    required this.groupName,
    required this.members,
  });

  @override
  State<SpendingsPage> createState() => _SpendingsPageState();
}

class _SpendingsPageState extends State<SpendingsPage>
    with SingleTickerProviderStateMixin {
  bool _loading = true;
  SpendingResult? _result;
  late AnimationController _animController;
  late Animation<double> _fadeAnim;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 600),
    );
    _fadeAnim = CurvedAnimation(
      parent: _animController,
      curve: Curves.easeOutCubic,
    );
    _loadSpendings();
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  Future<void> _loadSpendings() async {
    final res = await ApiService.getUserSpendings(
      groupId: widget.groupId,
    );

    if (res.errorMessage == "Session expired") {
      if (!mounted) return;
      Navigator.pushNamedAndRemoveUntil(context, "/login", (_) => false);
      return;
    }

    if (!mounted) return;
    setState(() {
      _result = res;
      _loading = false;
    });
    _animController.forward();
  }

  String _formatDate(int epochSecs) {
    final dt = DateTime.fromMillisecondsSinceEpoch(epochSecs * 1000);
    const months = [
      'Jan',
      'Feb',
      'Mar',
      'Apr',
      'May',
      'Jun',
      'Jul',
      'Aug',
      'Sep',
      'Oct',
      'Nov',
      'Dec',
    ];
    return "${dt.day} ${months[dt.month - 1]} ${dt.year}";
  }

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;

    return Scaffold(
      body: _loading
          ? Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  CircularProgressIndicator(color: cs.primary),
                  const SizedBox(height: 16),
                  Text(
                    "Loading expenses…",
                    style: TextStyle(color: cs.outline),
                  ),
                ],
              ),
            )
          : !_result!.isSuccess
          ? _errorBody(cs)
          : _successBody(cs),
    );
  }

  // ─── ERROR BODY ───────────────────────────────────────────────
  Widget _errorBody(ColorScheme cs) {
    return CustomScrollView(
      slivers: [
        SliverAppBar.large(title: const Text("My Expenses")),
        SliverFillRemaining(
          hasScrollBody: false,
          child: Center(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  Icons.error_outline_rounded,
                  size: 64,
                  color: cs.error.withValues(alpha: 0.7),
                ),
                const SizedBox(height: 16),
                Text(
                  _result!.errorMessage ?? "Failed to load",
                  style: TextStyle(
                    fontSize: 16,
                    color: cs.error,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 24),
                FilledButton.icon(
                  onPressed: () {
                    setState(() => _loading = true);
                    _loadSpendings();
                  },
                  icon: const Icon(Icons.refresh_rounded),
                  label: const Text("Retry"),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  // ─── SUCCESS BODY ─────────────────────────────────────────────
  Widget _successBody(ColorScheme cs) {
    final spendings = _result!.spendings!;

    return FadeTransition(
      opacity: _fadeAnim,
      child: RefreshIndicator(
        onRefresh: _loadSpendings,
        child: CustomScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          slivers: [
            SliverAppBar.large(title: const Text("My Expenses")),

            // ── SUMMARY CARD ──
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(20, 0, 20, 8),
                child: _summaryCard(spendings, cs),
              ),
            ),

            // ── EMPTY STATE ──
            if (spendings.isEmpty)
              SliverFillRemaining(
                hasScrollBody: false,
                child: Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        Icons.receipt_long_outlined,
                        size: 72,
                        color: cs.outline.withValues(alpha: 0.4),
                      ),
                      const SizedBox(height: 16),
                      Text(
                        "No expenses yet",
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w600,
                          color: cs.onSurface,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        "Your expense shares will appear here",
                        style: TextStyle(fontSize: 14, color: cs.outline),
                      ),
                    ],
                  ),
                ),
              )
            else
              // ── EXPENSE LIST ──
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(20, 8, 20, 40),
                sliver: SliverList(
                  delegate: SliverChildBuilderDelegate((context, index) {
                    final s = spendings[index];
                    return _spendingCard(s, cs);
                  }, childCount: spendings.length),
                ),
              ),
          ],
        ),
      ),
    );
  }

  // ─── SUMMARY CARD ─────────────────────────────────────────────
  Widget _summaryCard(List<UserSpending> spendings, ColorScheme cs) {
    final totalOwed = spendings.fold<double>(0, (sum, s) => sum + s.userAmount);
    final totalExpense = spendings.fold<double>(0, (sum, s) => sum + s.amount);
    final count = spendings.where((s) => !s.isSettlement).length;

    return Card(
      elevation: 0,
      color: cs.primaryContainer,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 22, horizontal: 24),
        child: Column(
          children: [
            Text(
              widget.groupName,
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: cs.onPrimaryContainer,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              "$count expense${count == 1 ? '' : 's'}",
              style: TextStyle(
                fontSize: 13,
                color: cs.onPrimaryContainer.withValues(alpha: 0.7),
              ),
            ),
            const SizedBox(height: 18),
            Row(
              children: [
                Expanded(
                  child: _summaryItem(
                    icon: Icons.account_balance_wallet_rounded,
                    label: "Your share",
                    amount: totalOwed,
                    color: cs.error,
                    cs: cs,
                  ),
                ),
                Container(
                  width: 1,
                  height: 40,
                  color: cs.onPrimaryContainer.withValues(alpha: 0.15),
                ),
                Expanded(
                  child: _summaryItem(
                    icon: Icons.receipt_long_rounded,
                    label: "Total",
                    amount: totalExpense,
                    color: cs.onPrimaryContainer,
                    cs: cs,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _summaryItem({
    required IconData icon,
    required String label,
    required double amount,
    required Color color,
    required ColorScheme cs,
  }) {
    return Column(
      children: [
        Icon(icon, size: 22, color: color),
        const SizedBox(height: 4),
        Text(
          "₹${amount.toStringAsFixed(2)}",
          style: TextStyle(
            fontSize: 18,
            fontWeight: FontWeight.w700,
            color: cs.onPrimaryContainer,
          ),
        ),
        const SizedBox(height: 2),
        Text(
          label,
          style: TextStyle(
            fontSize: 12,
            color: cs.onPrimaryContainer.withValues(alpha: 0.7),
          ),
        ),
      ],
    );
  }

  // ─── SPENDING CARD ─────────────────────────────────────────────
  Widget _spendingCard(UserSpending s, ColorScheme cs) {
    return GestureDetector(
      onTap: () {
        Navigator.pushNamed(
          context,
          "/expense-details",
          arguments: {"expenseId": s.expenseId, "members": widget.members},
        );
      },
      child: Card(
        elevation: 0,
        margin: const EdgeInsets.only(bottom: 10),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: BorderSide(
            color: cs.outlineVariant.withValues(alpha: 0.3),
            width: 1,
          ),
        ),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          child: Row(
            children: [
              // ── Avatar ──
              CircleAvatar(
                radius: 22,
                backgroundColor: s.isSettlement
                    ? cs.tertiaryContainer
                    : cs.primaryContainer,
                child: Icon(
                  s.isSettlement
                      ? Icons.handshake_outlined
                      : Icons.receipt_long_rounded,
                  size: 22,
                  color: s.isSettlement
                      ? cs.onTertiaryContainer
                      : cs.onPrimaryContainer,
                ),
              ),
              const SizedBox(width: 14),

              // ── Info ──
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      s.title,
                      style: TextStyle(
                        fontWeight: FontWeight.w600,
                        fontSize: 14,
                        color: cs.onSurface,
                      ),
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 3),
                    Text(
                      _formatDate(s.createdAt),
                      style: TextStyle(color: cs.outline, fontSize: 12),
                    ),
                    if (s.isIncompleteAmount || s.isIncompleteSplit) ...[
                      const SizedBox(height: 3),
                      Text(
                        "Incomplete",
                        style: TextStyle(
                          color: Colors.orange[700],
                          fontSize: 11,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ],
                  ],
                ),
              ),

              // ── Amount ──
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Text(
                    "₹${s.userAmount.toStringAsFixed(2)}",
                    style: TextStyle(
                      fontWeight: FontWeight.w700,
                      fontSize: 17,
                      color: cs.error,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    "of ₹${s.amount.toStringAsFixed(2)}",
                    style: TextStyle(fontSize: 11, color: cs.outline),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
