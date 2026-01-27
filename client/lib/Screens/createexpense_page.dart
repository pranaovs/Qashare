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
      final v = double.tryParse(c.text) ?? 0;
      if (v > 0) splits.add(ExpenseSplit(userId: id, amount: v, isPaid: true));
    });

    _owed.forEach((id, c) {
      final v = double.tryParse(c.text) ?? 0;
      if (v > 0) splits.add(ExpenseSplit(userId: id, amount: v, isPaid: false));
    });

    if (splits.isEmpty) {
      _snack("Add at least one split", true);
      return;
    }

    setState(() => _loading = true);

    final token = await TokenStorage.getToken();

    final req = ExpenseRequest(
      groupId: widget.groupId,
      title: _titleController.text.trim(),
      description:
      _descController.text.trim().isEmpty ? null : _descController.text,
      amount: double.tryParse(_amountController.text) ?? 0,
      isIncompleteAmount: _incompleteAmount,
      isIncompleteSplit: _incompleteSplit,
      splits: splits,
    );

    final res = await ApiService.createExpenseAdvanced(
      token: token!,
      request: req,
    );

    setState(() => _loading = false);

    if (res.isSuccess && mounted) {
      Navigator.pop(context, true);
    } else {
      _snack(res.errorMessage ?? "Failed", true);
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
    return Scaffold(
      appBar: AppBar(title: const Text("Add Expense")),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
              TextFormField(
                controller: _titleController,
                decoration: const InputDecoration(
                  labelText: "Title",
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.all(Radius.circular(10)),
                  ),
                ),
                validator: (v) => v == null || v.isEmpty ? "Required" : null,
              ),

              const SizedBox(height: 12),

              TextFormField(
                controller: _descController,
                decoration: const InputDecoration(
                  labelText: "Description (optional)",
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.all(Radius.circular(10)),
                  ),
                ),
                maxLines: 2,
              ),

              const SizedBox(height: 12),

              TextFormField(
                controller: _amountController,
                decoration: const InputDecoration(
                  labelText: "Amount",
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.all(Radius.circular(10)),
                  ),
                ),
                keyboardType: const TextInputType.numberWithOptions(decimal: true),
                inputFormatters: [
                  FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,2}'))
                ],
                validator: (v) {
                  if (_incompleteAmount) return null;
                  final n = double.tryParse(v ?? "");
                  if (n == null || n <= 0) return "Invalid amount";
                  return null;
                },
              ),
            CheckboxListTile(
              title: const Text("Amount incomplete"),
              value: _incompleteAmount,
              onChanged: (v) => setState(() => _incompleteAmount = v ?? false),
            ),
            CheckboxListTile(
              title: const Text("Split incomplete"),
              value: _incompleteSplit,
              onChanged: (v) => setState(() => _incompleteSplit = v ?? false),
            ),

            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text("Split Details",
                    style:
                    TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                TextButton.icon(
                  onPressed: _equalSplit,
                  icon: const Icon(Icons.calculate),
                  label: const Text("Equal"),
                )
              ],
            ),

            ...widget.members.map((m) => Card(
              child: Column(
                children: [
                  CheckboxListTile(
                    title: Text(m.name),
                    subtitle: Text(m.email),
                    value: _selected[m.userId],
                    onChanged: (v) =>
                        setState(() => _selected[m.userId] = v ?? false),
                  ),
                  if (_selected[m.userId] == true)
                    Padding(
                      padding: const EdgeInsets.all(12),
                      child: Row(
                        children: [
                          Expanded(
                            child: TextField(
                              controller: _paid[m.userId],
                              decoration:
                              const InputDecoration(labelText: "Paid"),
                              keyboardType:
                              const TextInputType.numberWithOptions(
                                  decimal: true),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Expanded(
                            child: TextField(
                              controller: _owed[m.userId],
                              decoration:
                              const InputDecoration(labelText: "Owes"),
                              keyboardType:
                              const TextInputType.numberWithOptions(
                                  decimal: true),
                            ),
                          ),
                        ],
                      ),
                    )
                ],
              ),
            )),

            const SizedBox(height: 20),
            ElevatedButton(
              onPressed: _loading ? null : _submit,
              child: _loading
                  ? const CircularProgressIndicator(strokeWidth: 2)
                  : const Text("Create Expense"),
            )
          ],
        ),
      ),
    );
  }
}
