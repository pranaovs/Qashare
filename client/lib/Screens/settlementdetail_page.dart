import 'package:flutter/material.dart';
import 'package:qashare/Models/groupdetail_model.dart';
import 'package:qashare/Models/settle_model.dart';
import 'package:qashare/Service/api_service.dart';

class SettlementDetailPage extends StatefulWidget {
  final String settlementId;
  final List<Member> members;

  const SettlementDetailPage({
    super.key,
    required this.settlementId,
    required this.members,
  });

  @override
  State<SettlementDetailPage> createState() => _SettlementDetailPageState();
}

class _SettlementDetailPageState extends State<SettlementDetailPage>
    with SingleTickerProviderStateMixin {
  bool _loading = true;
  SettlementDetailResult? _result;
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
    _loadDetails();
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  String _resolveName(String userId) {
    return _memberNames[userId] ?? userId;
  }

  Future<void> _loadDetails() async {
    final res = await ApiService.getSettlementDetails(
      settlementId: widget.settlementId,
    );

    if (!mounted) return;
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
                    "Loading settlement…",
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
        SliverAppBar.large(title: const Text("Settlement Details")),
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
                    _loadDetails();
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
    final s = _result!.settlement!;
    final isPositive = s.amount >= 0;
    final accent = isPositive ? cs.error : const Color(0xFF2E7D32);
    final name = _resolveName(s.userId);

    final transactedDate = s.transactedAt != null
        ? DateTime.fromMillisecondsSinceEpoch(s.transactedAt! * 1000)
        : null;
    final createdDate = s.createdAt != null
        ? DateTime.fromMillisecondsSinceEpoch(s.createdAt! * 1000)
        : null;

    return FadeTransition(
      opacity: _fadeAnim,
      child: RefreshIndicator(
        onRefresh: _loadDetails,
        child: CustomScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          slivers: [
            // ── COLLAPSING HEADER ──
            SliverAppBar.large(title: const Text("Settlement Details")),

            // ── CONTENT ──
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(20, 0, 20, 40),
              sliver: SliverList(
                delegate: SliverChildListDelegate([
                  // ── HERO AMOUNT CARD ──
                  _heroCard(s, cs, accent),

                  const SizedBox(height: 24),

                  // ── PARTICIPANTS SECTION ──
                  _sectionHeader(
                    "Participants",
                    Icons.people_outline_rounded,
                    cs,
                  ),
                  const SizedBox(height: 12),
                  _participantCard(
                    label: isPositive ? "You paid" : "Paid by",
                    name: isPositive ? "You" : name,
                    icon: Icons.arrow_upward_rounded,
                    color: cs.error,
                    cs: cs,
                  ),
                  const SizedBox(height: 8),
                  // Direction arrow
                  Center(
                    child: Icon(
                      Icons.arrow_downward_rounded,
                      size: 28,
                      color: cs.outline.withValues(alpha: 0.5),
                    ),
                  ),
                  const SizedBox(height: 8),
                  _participantCard(
                    label: isPositive ? "Received by" : "Received by",
                    name: isPositive ? name : "You",
                    icon: Icons.arrow_downward_rounded,
                    color: const Color(0xFF2E7D32),
                    cs: cs,
                  ),

                  const SizedBox(height: 28),

                  // ── DETAILS SECTION ──
                  _sectionHeader("Details", Icons.info_outline_rounded, cs),
                  const SizedBox(height: 12),

                  if (transactedDate != null) ...[
                    _detailRow(
                      cs,
                      icon: Icons.calendar_today_rounded,
                      label: "Transaction Date",
                      value: _formatDate(transactedDate),
                    ),
                    _detailRow(
                      cs,
                      icon: Icons.access_time_rounded,
                      label: "Time",
                      value: _formatTime(transactedDate),
                    ),
                  ],

                  if (createdDate != null)
                    _detailRow(
                      cs,
                      icon: Icons.schedule_rounded,
                      label: "Created",
                      value: _formatDate(createdDate),
                    ),

                  _detailRow(
                    cs,
                    icon: Icons.group_outlined,
                    label: "Group",
                    value: s.groupId,
                  ),
                ]),
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ─── HERO AMOUNT CARD ─────────────────────────────────────────
  Widget _heroCard(Settlement s, ColorScheme cs, Color accent) {
    final isPositive = s.amount >= 0;

    return Card(
      elevation: 0,
      color: cs.primaryContainer,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 28, horizontal: 24),
        child: Column(
          children: [
            // Settlement badge
            Container(
              margin: const EdgeInsets.only(bottom: 14),
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 5),
              decoration: BoxDecoration(
                color: cs.tertiary.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.handshake_outlined, size: 16, color: cs.tertiary),
                  const SizedBox(width: 6),
                  Text(
                    "Settlement",
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      color: cs.tertiary,
                      fontSize: 13,
                    ),
                  ),
                ],
              ),
            ),

            // Title
            Text(
              s.title.isNotEmpty ? s.title : "Settlement",
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.w600,
                color: cs.onPrimaryContainer,
              ),
            ),
            const SizedBox(height: 12),

            // Amount
            Text(
              "₹${s.amount.abs().toStringAsFixed(2)}",
              style: TextStyle(
                fontSize: 38,
                fontWeight: FontWeight.w800,
                color: cs.onPrimaryContainer,
                letterSpacing: -1,
              ),
            ),
            const SizedBox(height: 8),

            // Direction label
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
              decoration: BoxDecoration(
                color: accent.withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                isPositive ? "Money Paid" : "Money Received",
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: accent,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ─── PARTICIPANT CARD ─────────────────────────────────────────
  Widget _participantCard({
    required String label,
    required String name,
    required IconData icon,
    required Color color,
    required ColorScheme cs,
  }) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(14),
        side: BorderSide(color: color.withValues(alpha: 0.2), width: 1),
      ),
      color: color.withValues(alpha: 0.08),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Row(
          children: [
            CircleAvatar(
              radius: 20,
              backgroundColor: color.withValues(alpha: 0.15),
              child: Icon(icon, color: color, size: 20),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    name,
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      fontSize: 15,
                      color: cs.onSurface,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 2),
                  Text(
                    label,
                    style: TextStyle(
                      color: color,
                      fontSize: 12,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ─── SECTION HEADER ───────────────────────────────────────────
  Widget _sectionHeader(String title, IconData icon, ColorScheme cs) {
    return Row(
      children: [
        Icon(icon, size: 20, color: cs.primary),
        const SizedBox(width: 8),
        Text(
          title,
          style: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w700,
            color: cs.onSurface,
            letterSpacing: 0.2,
          ),
        ),
      ],
    );
  }

  // ─── DETAIL ROW ───────────────────────────────────────────────
  Widget _detailRow(
    ColorScheme cs, {
    required IconData icon,
    required String label,
    required String value,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, size: 18, color: cs.outline),
          const SizedBox(width: 12),
          SizedBox(
            width: 120,
            child: Text(
              label,
              style: TextStyle(
                color: cs.outline,
                fontSize: 13,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: TextStyle(
                color: cs.onSurface,
                fontSize: 14,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
