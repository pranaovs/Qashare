import 'package:flutter/material.dart';
import 'package:qashare/Config/token_storage.dart';
import 'package:qashare/Models/expensedetail_model.dart';
import 'package:qashare/Models/groupdetail_model.dart';
import 'package:qashare/Service/api_service.dart';

class ExpenseDetailsPage extends StatefulWidget {
  final String expenseId;
  final List<Member> members;

  const ExpenseDetailsPage({
    super.key,
    required this.expenseId,
    required this.members,
  });

  @override
  State<ExpenseDetailsPage> createState() => _ExpenseDetailsPageState();
}

class _ExpenseDetailsPageState extends State<ExpenseDetailsPage>
    with SingleTickerProviderStateMixin {
  bool _loading = true;
  ExpenseDetailResult? _result;
  late AnimationController _animController;
  late Animation<double> _fadeAnim;
  late Map<String, String> _memberNames;

  /// Resolve a userId to a display name using the members list.
  String _resolveName(String userId) {
    return _memberNames[userId] ?? userId;
  }

  @override
  void initState() {
    super.initState();
    // Build userId -> name lookup map from the members list
    _memberNames = {for (final m in widget.members) m.userId: m.name};
    _animController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 600),
    );
    _fadeAnim = CurvedAnimation(
      parent: _animController,
      curve: Curves.easeOutCubic,
    );
    _loadExpense();
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
  }

  Future<void> _loadExpense() async {
    final token = await TokenStorage.getToken();
    if (token == null) return;

    final res = await ApiService.getExpenseDetails(
      token: token,
      expenseId: widget.expenseId,
    );

    // JWT expired handling
    if (res.errorMessage == "Session expired") {
      await TokenStorage.clear();
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

  // ─── DELETE EXPENSE ────────────────────────────────────────────
  Future<void> _confirmDelete() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text("Delete Expense"),
        content: const Text(
          "Are you sure you want to delete this expense? This cannot be undone.",
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text("Cancel"),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: FilledButton.styleFrom(
              backgroundColor: Theme.of(context).colorScheme.error,
            ),
            child: const Text("Delete"),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    final token = await TokenStorage.getToken();
    if (token == null) return;

    final res = await ApiService.deleteExpense(
      token: token,
      expenseId: widget.expenseId,
    );

    if (!mounted) return;

    if (res.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("Expense deleted")),
      );
      Navigator.pop(context, true); // true = signal parent to refresh
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(res.errorMessage ?? "Failed to delete"),
          backgroundColor: Theme.of(context).colorScheme.error,
        ),
      );
    }
  }

  // ─── EDIT EXPENSE ──────────────────────────────────────────────
  Future<void> _showEditSheet() async {
    final e = _result!.expense!;

    final edited = await showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => _EditExpenseSheet(
        expenseId: e.expenseId,
        title: e.title,
        description: e.description ?? "",
        amount: e.amount,
      ),
    );

    if (edited == true) {
      setState(() => _loading = true);
      _loadExpense();
    }
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
                  Text("Loading expense…", style: TextStyle(color: cs.outline)),
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
        SliverAppBar.large(title: const Text("Expense Details")),
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
                  _result!.errorMessage!,
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
                    _loadExpense();
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
    final e = _result!.expense!;
    final transactedDate = DateTime.fromMillisecondsSinceEpoch(
      e.transactedAt * 1000,
    );
    final createdDate = DateTime.fromMillisecondsSinceEpoch(e.createdAt * 1000);

    final paidSplits = e.splits.where((s) => s.isPaid).toList();
    final owedSplits = e.splits.where((s) => !s.isPaid).toList();

    return FadeTransition(
      opacity: _fadeAnim,
      child: RefreshIndicator(
        onRefresh: _loadExpense,
        child: CustomScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          slivers: [
            // ── COLLAPSING HEADER ──
            SliverAppBar.large(
              title: Text(e.isSettlement ? "Settlement" : "Expense Details"),
              actions: [
                // Edit button
                IconButton(
                  icon: const Icon(Icons.edit_outlined),
                  tooltip: "Edit",
                  onPressed: _showEditSheet,
                ),
                // Delete button
                IconButton(
                  icon: Icon(Icons.delete_outline, color: cs.error),
                  tooltip: "Delete",
                  onPressed: _confirmDelete,
                ),
                if (e.isIncompleteAmount || e.isIncompleteSplit)
                  Padding(
                    padding: const EdgeInsets.only(right: 8),
                    child: Tooltip(
                      message: e.isIncompleteAmount
                          ? "Amount is incomplete"
                          : "Split is incomplete",
                      child: Icon(Icons.warning_amber_rounded, color: cs.error),
                    ),
                  ),
              ],
            ),

            // ── CONTENT ──
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(20, 0, 20, 40),
              sliver: SliverList(
                delegate: SliverChildListDelegate([
                  // ── HERO AMOUNT CARD ──
                  _heroCard(e, cs),

                  const SizedBox(height: 20),

                  // ── INFO SECTION ──
                  _sectionHeader("Details", Icons.info_outline_rounded, cs),
                  const SizedBox(height: 10),
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
                  if (e.description != null && e.description!.isNotEmpty)
                    _detailRow(
                      cs,
                      icon: Icons.notes_rounded,
                      label: "Note",
                      value: e.description!,
                    ),
                  _detailRow(
                    cs,
                    icon: Icons.person_outline_rounded,
                    label: "Added by",
                    value: _resolveName(e.addedBy),
                  ),
                  if (e.latitude != null && e.longitude != null)
                    _detailRow(
                      cs,
                      icon: Icons.location_on_outlined,
                      label: "Location",
                      value:
                          "${e.latitude!.toStringAsFixed(4)}, ${e.longitude!.toStringAsFixed(4)}",
                    ),
                  _detailRow(
                    cs,
                    icon: Icons.schedule_rounded,
                    label: "Created",
                    value: _formatDate(createdDate),
                  ),

                  // ── WARNING CHIPS ──
                  if (e.isIncompleteAmount || e.isIncompleteSplit) ...[
                    const SizedBox(height: 14),
                    Wrap(
                      spacing: 8,
                      children: [
                        if (e.isIncompleteAmount)
                          _warningChip("Incomplete amount", cs),
                        if (e.isIncompleteSplit)
                          _warningChip("Incomplete split", cs),
                      ],
                    ),
                  ],

                  const SizedBox(height: 28),

                  // ── PAID BY SECTION ──
                  if (paidSplits.isNotEmpty) ...[
                    _sectionHeader(
                      e.isSettlement ? "Settled by" : "Paid by",
                      Icons.arrow_upward_rounded,
                      cs,
                    ),
                    const SizedBox(height: 10),
                    ...paidSplits.map((s) => _splitCard(s, true, cs)),
                    const SizedBox(height: 20),
                  ],

                  // ── OWES SECTION ──
                  if (owedSplits.isNotEmpty) ...[
                    _sectionHeader(
                      e.isSettlement ? "Received by" : "Owes",
                      Icons.arrow_downward_rounded,
                      cs,
                    ),
                    const SizedBox(height: 10),
                    ...owedSplits.map((s) => _splitCard(s, false, cs)),
                  ],
                ]),
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ─── HERO AMOUNT CARD ─────────────────────────────────────────
  Widget _heroCard(ExpenseDetail e, ColorScheme cs) {
    return Card(
      elevation: 0,
      color: cs.primaryContainer,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 28, horizontal: 24),
        child: Column(
          children: [
            // Settlement badge
            if (e.isSettlement)
              Container(
                margin: const EdgeInsets.only(bottom: 14),
                padding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 5,
                ),
                decoration: BoxDecoration(
                  color: cs.tertiary.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      Icons.handshake_outlined,
                      size: 16,
                      color: cs.tertiary,
                    ),
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
              e.title,
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
              "₹${e.amount.toStringAsFixed(2)}",
              style: TextStyle(
                fontSize: 38,
                fontWeight: FontWeight.w800,
                color: cs.onPrimaryContainer,
                letterSpacing: -1,
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

  // ─── SPLIT CARD ───────────────────────────────────────────────
  Widget _splitCard(ExpenseSplit split, bool isPayer, ColorScheme cs) {
    final accent = isPayer
        ? const Color(0xFF2E7D32) // green 800
        : cs.error;
    final accentBg = isPayer
        ? const Color(0xFF2E7D32).withValues(alpha: 0.08)
        : cs.error.withValues(alpha: 0.08);

    return Card(
      elevation: 0,
      margin: const EdgeInsets.only(bottom: 8),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(14),
        side: BorderSide(color: accent.withValues(alpha: 0.2), width: 1),
      ),
      color: accentBg,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Row(
          children: [
            // Avatar
            CircleAvatar(
              radius: 20,
              backgroundColor: accent.withValues(alpha: 0.15),
              child: Icon(
                isPayer
                    ? Icons.arrow_upward_rounded
                    : Icons.arrow_downward_rounded,
                color: accent,
                size: 20,
              ),
            ),
            const SizedBox(width: 14),

            // User ID + label
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    _resolveName(split.userId),
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      fontSize: 14,
                      color: cs.onSurface,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 2),
                  Text(
                    isPayer ? "Paid" : "Owes",
                    style: TextStyle(
                      color: accent,
                      fontSize: 12,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
            ),

            // Amount
            Text(
              "₹${split.amount.toStringAsFixed(2)}",
              style: TextStyle(
                fontWeight: FontWeight.w700,
                fontSize: 16,
                color: accent,
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ─── WARNING CHIP ─────────────────────────────────────────────
  Widget _warningChip(String label, ColorScheme cs) {
    return Chip(
      avatar: Icon(Icons.warning_amber_rounded, size: 16, color: cs.error),
      label: Text(
        label,
        style: TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.w500,
          color: cs.error,
        ),
      ),
      backgroundColor: cs.errorContainer.withValues(alpha: 0.4),
      side: BorderSide(color: cs.error.withValues(alpha: 0.2)),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
    );
  }
}

// ================= EDIT EXPENSE BOTTOM SHEET =================

class _EditExpenseSheet extends StatefulWidget {
  final String expenseId;
  final String title;
  final String description;
  final double amount;

  const _EditExpenseSheet({
    required this.expenseId,
    required this.title,
    required this.description,
    required this.amount,
  });

  @override
  State<_EditExpenseSheet> createState() => _EditExpenseSheetState();
}

class _EditExpenseSheetState extends State<_EditExpenseSheet> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _titleCtrl;
  late final TextEditingController _descCtrl;
  late final TextEditingController _amountCtrl;
  bool _saving = false;

  @override
  void initState() {
    super.initState();
    _titleCtrl = TextEditingController(text: widget.title);
    _descCtrl = TextEditingController(text: widget.description);
    _amountCtrl = TextEditingController(
      text: widget.amount.toStringAsFixed(2),
    );
  }

  @override
  void dispose() {
    _titleCtrl.dispose();
    _descCtrl.dispose();
    _amountCtrl.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _saving = true);

    final token = await TokenStorage.getToken();
    if (token == null) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("Session expired")),
      );
      setState(() => _saving = false);
      return;
    }

    final body = <String, dynamic>{
      "title": _titleCtrl.text.trim(),
      "description": _descCtrl.text.trim().isEmpty
          ? null
          : _descCtrl.text.trim(),
      "amount": double.tryParse(_amountCtrl.text) ?? widget.amount,
    };

    final res = await ApiService.updateExpense(
      token: token,
      expenseId: widget.expenseId,
      body: body,
    );

    if (!mounted) return;
    setState(() => _saving = false);

    if (res.isSuccess) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text("Expense updated")),
      );
      Navigator.pop(context, true); // true = signal to refresh
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(res.errorMessage ?? "Failed to update"),
          backgroundColor: Theme.of(context).colorScheme.error,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.fromLTRB(
        20,
        16,
        20,
        MediaQuery.of(context).viewInsets.bottom + 20,
      ),
      child: Form(
        key: _formKey,
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
              "Edit Expense",
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 20),

            TextFormField(
              controller: _titleCtrl,
              decoration: const InputDecoration(
                labelText: "Title",
                prefixIcon: Icon(Icons.title),
                border: OutlineInputBorder(),
              ),
              validator: (v) =>
                  (v == null || v.trim().isEmpty) ? "Title required" : null,
            ),
            const SizedBox(height: 14),

            TextFormField(
              controller: _descCtrl,
              decoration: const InputDecoration(
                labelText: "Description (optional)",
                prefixIcon: Icon(Icons.notes),
                border: OutlineInputBorder(),
              ),
              maxLines: 2,
            ),
            const SizedBox(height: 14),

            TextFormField(
              controller: _amountCtrl,
              decoration: const InputDecoration(
                labelText: "Amount",
                prefixIcon: Icon(Icons.currency_rupee),
                border: OutlineInputBorder(),
              ),
              keyboardType:
                  const TextInputType.numberWithOptions(decimal: true),
              validator: (v) {
                if (v == null || v.trim().isEmpty) return "Amount required";
                final n = double.tryParse(v);
                if (n == null || n <= 0) return "Enter a valid amount";
                return null;
              },
            ),
            const SizedBox(height: 20),

            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                onPressed: _saving ? null : _save,
                icon: _saving
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.white,
                        ),
                      )
                    : const Icon(Icons.save_rounded),
                label: Text(_saving ? "Saving…" : "Save Changes"),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
