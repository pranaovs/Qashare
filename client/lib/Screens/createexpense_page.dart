import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:qashare/Models/expense_model.dart';
import '../Models/groupdetail_model.dart';
import '../Config/token_storage.dart';
import '../Service/api_service.dart';

class CreateExpensePage extends StatefulWidget {
  final String groupId;
  final List<Member> members;

  const CreateExpensePage({
    super.key,
    required this.groupId,
    required this.members,
  });

  @override
  State<CreateExpensePage> createState() => _CreateExpensePageState();
}

class _CreateExpensePageState extends State<CreateExpensePage> {
  final _formKey = GlobalKey<FormState>();

  final _titleController = TextEditingController();
  final _descController = TextEditingController();
  final _amountController = TextEditingController();

  final Map<String, TextEditingController> _paid = {};
  final Map<String, TextEditingController> _owed = {};
  final Map<String, bool> _selected = {};

  bool _incompleteAmount = false;
  bool _incompleteSplit = false;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    for (var m in widget.members) {
      _paid[m.userId] = TextEditingController(text: '0');
      _owed[m.userId] = TextEditingController(text: '0');
      _selected[m.userId] = false;
    }
  }

  @override
  void dispose() {
    _titleController.dispose();
    _descController.dispose();
    _amountController.dispose();
    _paid.values.forEach((e) => e.dispose());
    _owed.values.forEach((e) => e.dispose());
    super.dispose();
  }

  void _equalSplit() {
    final amount = double.tryParse(_amountController.text);
    if (amount == null || amount <= 0) {
      _snack("Enter amount first", true);
      return;
    }

    final selected = _selected.values.where((e) => e).length;
    if (selected == 0) {
      _snack("Select members", true);
      return;
    }

    final each = (amount / selected).toStringAsFixed(2);
    setState(() {
      _selected.forEach((id, v) {
        if (v) _owed[id]!.text = each;
      });
    });
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    final splits = <ExpenseSplit>[];

    _paid.forEach((id, c) {
      // Only include splits for members that are currently selected.
      if (_selected[id] != true) return;
      final v = double.tryParse(c.text) ?? 0;
      if (v > 0) {
        splits.add(ExpenseSplit(userId: id, amount: v, isPaid: true));
      }
    });

    _owed.forEach((id, c) {
      // Only include splits for members that are currently selected.
      if (_selected[id] != true) return;
      final v = double.tryParse(c.text) ?? 0;
      if (v > 0) {
        splits.add(ExpenseSplit(userId: id, amount: v, isPaid: false));
      }
    });

    if (splits.isEmpty) {
      _snack("Add at least one split", true);
      return;
    }

    final amount = double.tryParse(_amountController.text) ?? 0;

    double totalPaid = 0;
    double totalOwed = 0;

    for (var s in splits) {
      if (s.isPaid) totalPaid += s.amount;
      if (!s.isPaid) totalOwed += s.amount;
    }

    if (!_incompleteAmount && !_incompleteSplit) {
      if ((totalPaid - amount).abs() > 0.01 ||
          (totalOwed - amount).abs() > 0.01) {
        _snack("Paid and owed must both equal total amount", true);
        return;
      }
    }

    setState(() => _loading = true);

    final token = await TokenStorage.getToken();
    if (token == null) {
      _snack("Session expired. Please login again.", true);
      setState(() => _loading = false);
      return;
    }

    final req = ExpenseRequest(
      groupId: widget.groupId,
      title: _titleController.text.trim(),
      description: _descController.text.trim().isEmpty
          ? null
          : _descController.text.trim(),
      amount: double.tryParse(_amountController.text) ?? 0,
      isIncompleteAmount: _incompleteAmount,
      isIncompleteSplit: _incompleteSplit,
      splits: splits,
    );

    final res = await ApiService.createExpenseAdvanced(
      token: token,
      groupId: widget.groupId,
      request: req,
    );

    setState(() => _loading = false);

    if (res.isSuccess && mounted) {
      Navigator.pop(context, true);
    } else {
      _snack(res.errorMessage ?? "Failed to create expense", true);
    }
  }

  void _snack(String m, bool e) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(m),
        backgroundColor: e ? Theme.of(context).colorScheme.error : null,
        duration: const Duration(milliseconds: 900),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(title: const Text("Add Expense")),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.fromLTRB(16, 8, 16, 40),
          children: [
            // ── EXPENSE INFO CARD ──
            Card(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Icon(Icons.receipt_long_rounded,
                            size: 20, color: cs.primary),
                        const SizedBox(width: 8),
                        Text(
                          "Expense Info",
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w700,
                            color: cs.primary,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),

                    // Title
                    TextFormField(
                      controller: _titleController,
                      decoration: InputDecoration(
                        labelText: "Title",
                        prefixIcon: const Icon(Icons.title_rounded),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      validator: (v) =>
                          v == null || v.isEmpty ? "Required" : null,
                    ),
                    const SizedBox(height: 12),

                    // Description
                    TextFormField(
                      controller: _descController,
                      decoration: InputDecoration(
                        labelText: "Description (optional)",
                        prefixIcon: const Icon(Icons.notes_rounded),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      maxLines: 2,
                    ),
                    const SizedBox(height: 12),

                    // Amount
                    TextFormField(
                      controller: _amountController,
                      decoration: InputDecoration(
                        labelText: "Amount",
                        prefixIcon: const Icon(Icons.currency_rupee_rounded),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                      ),
                      keyboardType: const TextInputType.numberWithOptions(
                        decimal: true,
                      ),
                      inputFormatters: [
                        FilteringTextInputFormatter.allow(
                            RegExp(r'^\d+\.?\d{0,2}')),
                      ],
                      validator: (v) {
                        if (_incompleteAmount) return null;
                        final n = double.tryParse(v ?? "");
                        if (n == null || n <= 0) return "Invalid amount";
                        return null;
                      },
                    ),
                  ],
                ),
              ),
            ),

            const SizedBox(height: 8),

            // ── OPTIONS CARD ──
            Card(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
              ),
              child: Column(
                children: [
                  SwitchListTile(
                    title: const Text("Amount incomplete"),
                    subtitle: Text(
                      "Total isn't finalised yet",
                      style: TextStyle(fontSize: 12, color: cs.outline),
                    ),
                    secondary: Icon(
                      Icons.hourglass_bottom_rounded,
                      color: _incompleteAmount ? cs.primary : cs.outline,
                    ),
                    value: _incompleteAmount,
                    onChanged: (v) =>
                        setState(() => _incompleteAmount = v),
                  ),
                  Divider(height: 1, indent: 16, endIndent: 16,
                      color: cs.outlineVariant.withValues(alpha: 0.3)),
                  SwitchListTile(
                    title: const Text("Split incomplete"),
                    subtitle: Text(
                      "Splits don't need to match total",
                      style: TextStyle(fontSize: 12, color: cs.outline),
                    ),
                    secondary: Icon(
                      Icons.call_split_rounded,
                      color: _incompleteSplit ? cs.primary : cs.outline,
                    ),
                    value: _incompleteSplit,
                    onChanged: (v) =>
                        setState(() => _incompleteSplit = v),
                  ),
                ],
              ),
            ),

            const SizedBox(height: 16),

            // ── SPLIT HEADER ──
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Icon(Icons.people_outline_rounded,
                        size: 20, color: cs.primary),
                    const SizedBox(width: 8),
                    const Text(
                      "Split Details",
                      style: TextStyle(
                          fontSize: 17, fontWeight: FontWeight.w700),
                    ),
                  ],
                ),
                Row(
                  children: [
                    TextButton.icon(
                      onPressed: () {
                        final allSelected =
                            _selected.values.every((v) => v);
                        setState(() {
                          for (var id in _selected.keys) {
                            _selected[id] = !allSelected;
                          }
                        });
                      },
                      icon: Icon(
                        _selected.values.every((v) => v)
                            ? Icons.deselect
                            : Icons.select_all,
                        size: 18,
                      ),
                      label: Text(
                        _selected.values.every((v) => v)
                            ? "None"
                            : "All",
                      ),
                    ),
                    TextButton.icon(
                      onPressed: _equalSplit,
                      icon: const Icon(Icons.calculate_rounded, size: 18),
                      label: const Text("Equal"),
                    ),
                  ],
                ),
              ],
            ),

            const SizedBox(height: 6),

            // ── MEMBER CARDS ──
            ...widget.members.map(
              (m) {
                final isSelected = _selected[m.userId] == true;
                return Card(
                  margin: const EdgeInsets.only(bottom: 8),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(14),
                    side: isSelected
                        ? BorderSide(
                            color: cs.primary.withValues(alpha: 0.4),
                            width: 1.5)
                        : BorderSide.none,
                  ),
                  child: Column(
                    children: [
                      // Member row
                      InkWell(
                        borderRadius: BorderRadius.circular(14),
                        onTap: () => setState(
                            () => _selected[m.userId] = !isSelected),
                        child: Padding(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 12, vertical: 10),
                          child: Row(
                            children: [
                              // Avatar
                              CircleAvatar(
                                radius: 20,
                                backgroundColor: isSelected
                                    ? cs.primary
                                    : cs.surfaceContainerHighest,
                                child: isSelected
                                    ? Icon(Icons.check_rounded,
                                        color: cs.onPrimary, size: 20)
                                    : Text(
                                        m.name.isNotEmpty
                                            ? m.name[0].toUpperCase()
                                            : "?",
                                        style: TextStyle(
                                          fontWeight: FontWeight.bold,
                                          fontSize: 16,
                                          color: cs.onSurface,
                                        ),
                                      ),
                              ),
                              const SizedBox(width: 12),

                              // Name + email
                              Expanded(
                                child: Column(
                                  crossAxisAlignment:
                                      CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      m.name,
                                      style: const TextStyle(
                                        fontWeight: FontWeight.w600,
                                        fontSize: 15,
                                      ),
                                    ),
                                    Text(
                                      m.email,
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: cs.outline,
                                      ),
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  ],
                                ),
                              ),

                              // Guest badge
                              if (m.guest)
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                      horizontal: 8, vertical: 3),
                                  decoration: BoxDecoration(
                                    color: cs.tertiaryContainer,
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: Text(
                                    "Guest",
                                    style: TextStyle(
                                      fontSize: 11,
                                      color: cs.onTertiaryContainer,
                                      fontWeight: FontWeight.w500,
                                    ),
                                  ),
                                ),

                              // Checkbox
                              Checkbox(
                                value: isSelected,
                                onChanged: (v) => setState(
                                    () => _selected[m.userId] = v ?? false),
                              ),
                            ],
                          ),
                        ),
                      ),

                      // Split fields
                      if (isSelected)
                        Container(
                          decoration: BoxDecoration(
                            color: cs.surfaceContainerHighest
                                .withValues(alpha: 0.3),
                            borderRadius: const BorderRadius.only(
                              bottomLeft: Radius.circular(14),
                              bottomRight: Radius.circular(14),
                            ),
                          ),
                          padding:
                              const EdgeInsets.fromLTRB(12, 8, 12, 14),
                          child: Row(
                            children: [
                              // Paid
                              Expanded(
                                child: TextField(
                                  controller: _paid[m.userId],
                                  decoration: InputDecoration(
                                    labelText: "Paid",
                                    labelStyle: const TextStyle(
                                        color: Color(0xFF2E7D32),
                                        fontSize: 13),
                                    prefixIcon: const Icon(
                                      Icons.arrow_upward_rounded,
                                      color: Color(0xFF2E7D32),
                                      size: 18,
                                    ),
                                    border: OutlineInputBorder(
                                      borderRadius:
                                          BorderRadius.circular(10),
                                    ),
                                    enabledBorder: OutlineInputBorder(
                                      borderRadius:
                                          BorderRadius.circular(10),
                                      borderSide: BorderSide(
                                        color: const Color(0xFF2E7D32)
                                            .withValues(alpha: 0.3),
                                      ),
                                    ),
                                    contentPadding:
                                        const EdgeInsets.symmetric(
                                            horizontal: 12, vertical: 10),
                                    isDense: true,
                                  ),
                                  keyboardType:
                                      const TextInputType.numberWithOptions(
                                          decimal: true),
                                  inputFormatters: [
                                    FilteringTextInputFormatter.allow(
                                        RegExp(r'^\d+\.?\d{0,2}')),
                                  ],
                                ),
                              ),
                              const SizedBox(width: 10),
                              // Owes
                              Expanded(
                                child: TextField(
                                  controller: _owed[m.userId],
                                  decoration: InputDecoration(
                                    labelText: "Owes",
                                    labelStyle: TextStyle(
                                        color: cs.error, fontSize: 13),
                                    prefixIcon: Icon(
                                      Icons.arrow_downward_rounded,
                                      color: cs.error,
                                      size: 18,
                                    ),
                                    border: OutlineInputBorder(
                                      borderRadius:
                                          BorderRadius.circular(10),
                                    ),
                                    enabledBorder: OutlineInputBorder(
                                      borderRadius:
                                          BorderRadius.circular(10),
                                      borderSide: BorderSide(
                                        color: cs.error
                                            .withValues(alpha: 0.3),
                                      ),
                                    ),
                                    contentPadding:
                                        const EdgeInsets.symmetric(
                                            horizontal: 12, vertical: 10),
                                    isDense: true,
                                  ),
                                  keyboardType:
                                      const TextInputType.numberWithOptions(
                                          decimal: true),
                                  inputFormatters: [
                                    FilteringTextInputFormatter.allow(
                                        RegExp(r'^\d+\.?\d{0,2}')),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ),
                    ],
                  ),
                );
              },
            ),

            const SizedBox(height: 24),

            // ── SUBMIT BUTTON ──
            SizedBox(
              height: 50,
              child: FilledButton.icon(
                onPressed: _loading ? null : _submit,
                icon: _loading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                            strokeWidth: 2, color: Colors.white),
                      )
                    : const Icon(Icons.check_rounded),
                label: Text(
                  _loading ? "Creating…" : "Create Expense",
                  style: const TextStyle(fontSize: 16),
                ),
                style: FilledButton.styleFrom(
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(14),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
