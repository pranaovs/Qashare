import 'package:flutter/material.dart';
import 'package:qashare/Config/token_storage.dart';
import 'package:qashare/Models/groupdetail_model.dart';
import 'package:qashare/Models/settle_model.dart';
import 'package:qashare/Service/api_service.dart';

class SettlementHistoryPage extends StatefulWidget {
  final String groupId;
  final String groupName;
  final List<Member> members;

  const SettlementHistoryPage({
    super.key,
    required this.groupId,
    required this.groupName,
    required this.members,
  });

  @override
  State<SettlementHistoryPage> createState() => _SettlementHistoryPageState();
}

class _SettlementHistoryPageState extends State<SettlementHistoryPage>
    with SingleTickerProviderStateMixin {
  bool _loading = true;
  SettleResult? _result;
  late AnimationController _animController;
  late Animation<double> _fadeAnim;
  late Map<String, String> _memberNames;

  @override
  void initState() {
    super.initState();
    _memberNames = {for (final m in widget.members) m.userId: m.name};
    _animController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 600),
    );
    _fadeAnim = CurvedAnimation(
      parent: _animController,
      curve: Curves.easeOutCubic,
    );
    _loadHistory();
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  String _resolveName(String userId) {
    return _memberNames[userId] ?? userId;
  }

  Future<void> _loadHistory() async {
    // Handle session expiry
    final res = await ApiService.getSettlementHistory(groupId: widget.groupId);

    if (res.errorMessage == "Session expired") {
      if (!mounted) return;
      Navigator.pushNamedAndRemoveUntil(context, "/login", (_) => false);
      return;
    }

    setState(() {
      _result = res;
      _loading = false;
    });
    _animController.forward();
  }

  String _formatDate(DateTime dt) {
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

  String _formatTime(DateTime dt) {
    final hour = dt.hour > 12 ? dt.hour - 12 : (dt.hour == 0 ? 12 : dt.hour);
    final period = dt.hour >= 12 ? "PM" : "AM";
    final minute = dt.minute.toString().padLeft(2, '0');
    return "$hour:$minute $period";
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
                    "Loading settlements…",
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
        SliverAppBar.large(title: const Text("Settlement History")),
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
                    _loadHistory();
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
    final settlements = _result!.settlements!;

    return FadeTransition(
      opacity: _fadeAnim,
      child: RefreshIndicator(
        onRefresh: _loadHistory,
        child: CustomScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          slivers: [
            // ── COLLAPSING HEADER ──
            SliverAppBar.large(title: const Text("Settlement History")),

            // ── SUMMARY CARD ──
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(20, 0, 20, 8),
                child: _summaryCard(settlements, cs),
              ),
            ),

            // ── EMPTY STATE ──
            if (settlements.isEmpty)
              SliverFillRemaining(
                hasScrollBody: false,
                child: Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        Icons.handshake_outlined,
                        size: 72,
                        color: cs.outline.withValues(alpha: 0.4),
                      ),
                      const SizedBox(height: 16),
                      Text(
                        "No settlements yet",
                        style: TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w600,
                          color: cs.onSurface,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        "Settlement transactions will appear here",
                        style: TextStyle(fontSize: 14, color: cs.outline),
                      ),
                    ],
                  ),
                ),
              )
            else
              // ── SETTLEMENT LIST ──
              SliverPadding(
                padding: const EdgeInsets.fromLTRB(20, 8, 20, 40),
                sliver: SliverList(
                  delegate: SliverChildBuilderDelegate((context, index) {
                    final s = settlements[index];
                    return _settlementCard(s, cs);
                  }, childCount: settlements.length),
                ),
              ),
          ],
        ),
      ),
    );
  }

  // ─── SUMMARY CARD ─────────────────────────────────────────────
  Widget _summaryCard(List<Settlement> settlements, ColorScheme cs) {
    double totalPaid = 0;
    double totalReceived = 0;

    for (final s in settlements) {
      if (s.amount >= 0) {
        totalPaid += s.amount;
      } else {
        totalReceived += s.amount.abs();
      }
    }

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
              "${settlements.length} settlement${settlements.length == 1 ? '' : 's'}",
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
                    icon: Icons.arrow_upward_rounded,
                    label: "Paid",
                    amount: totalPaid,
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
                    icon: Icons.arrow_downward_rounded,
                    label: "Received",
                    amount: totalReceived,
                    color: const Color(0xFF2E7D32),
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

  // ─── SETTLEMENT CARD ──────────────────────────────────────────
  Widget _settlementCard(Settlement s, ColorScheme cs) {
    final isPositive = s.amount >= 0;
    final accent = isPositive ? cs.error : const Color(0xFF2E7D32);
    final accentBg = accent.withValues(alpha: 0.08);

    final name = _resolveName(s.userId);

    final transactedDate = s.transactedAt != null
        ? DateTime.fromMillisecondsSinceEpoch(s.transactedAt! * 1000)
        : null;
    final createdDate = s.createdAt != null
        ? DateTime.fromMillisecondsSinceEpoch(s.createdAt! * 1000)
        : null;

    final displayDate = transactedDate ?? createdDate;

    return GestureDetector(
      onTap: s.settlementId != null
          ? () {
              Navigator.pushNamed(
                context,
                "/settlement-details",
                arguments: {
                  "settlementId": s.settlementId,
                  "members": widget.members,
                },
              );
            }
          : null,
      child: Card(
        elevation: 0,
        margin: const EdgeInsets.only(bottom: 10),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: BorderSide(color: accent.withValues(alpha: 0.2), width: 1),
        ),
        color: accentBg,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          child: Row(
            children: [
              // ── Avatar ──
              CircleAvatar(
                radius: 22,
                backgroundColor: accent.withValues(alpha: 0.15),
                child: Icon(
                  isPositive
                      ? Icons.arrow_upward_rounded
                      : Icons.arrow_downward_rounded,
                  color: accent,
                  size: 22,
                ),
              ),
              const SizedBox(width: 14),

              // ── Info ──
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const SizedBox(height: 3),
                    Text(
                      isPositive ? "You → $name" : "$name → You",
                      style: TextStyle(
                        color: accent,
                        fontSize: 12,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                    if (displayDate != null) ...[
                      const SizedBox(height: 3),
                      Text(
                        "${_formatDate(displayDate)}  •  ${_formatTime(displayDate)}",
                        style: TextStyle(color: cs.outline, fontSize: 11),
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
                    "₹${s.amount.abs().toStringAsFixed(2)}",
                    style: TextStyle(
                      fontWeight: FontWeight.w700,
                      fontSize: 17,
                      color: accent,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    isPositive ? "Paid" : "Received",
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w500,
                      color: accent.withValues(alpha: 0.8),
                    ),
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
