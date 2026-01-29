class GroupExpense {
  final String expenseId;
  final String title;
  final String? description;
  final double amount;
  final int createdAt;
  final bool isIncompleteAmount;
  final bool isIncompleteSplit;
  final String addedBy;

  GroupExpense({
    required this.expenseId,
    required this.title,
    this.description,
    required this.amount,
    required this.createdAt,
    required this.isIncompleteAmount,
    required this.isIncompleteSplit,
    required this.addedBy,
  });

  factory GroupExpense.fromJson(Map<String, dynamic> json) {
    return GroupExpense(
      expenseId: json["expense_id"],
      title: json["title"],
      description: json["description"],
      amount: (json["amount"] as num).toDouble(),
      createdAt: json["created_at"],
      isIncompleteAmount: json["is_incomplete_amount"],
      isIncompleteSplit: json["is_incomplete_split"],
      addedBy: json["added_by"],
    );
  }
}

class ExpenseListResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<GroupExpense>? expenses;

  ExpenseListResult._({
    required this.isSuccess,
    this.errorMessage,
    this.expenses,
  });

  factory ExpenseListResult.success(List<GroupExpense> list) {
    return ExpenseListResult._(isSuccess: true, expenses: list);
  }

  factory ExpenseListResult.error(String msg) {
    return ExpenseListResult._(isSuccess: false, errorMessage: msg);
  }
}
