class ExpenseDetailResult {
  final bool isSuccess;
  final ExpenseDetail? expense;
  final String? errorMessage;

  ExpenseDetailResult._({
    required this.isSuccess,
    this.expense,
    this.errorMessage,
  });

  factory ExpenseDetailResult.success(ExpenseDetail e) {
    return ExpenseDetailResult._(isSuccess: true, expense: e);
  }

  factory ExpenseDetailResult.error(String msg) {
    return ExpenseDetailResult._(isSuccess: false, errorMessage: msg);
  }
}

class ExpenseDetail {
  final String expenseId;
  final String title;
  final String? description;
  final double amount;
  final String addedBy;
  final String groupId;
  final bool isSettlement;
  final bool isIncompleteAmount;
  final bool isIncompleteSplit;
  final double? latitude;
  final double? longitude;
  final int createdAt;
  final int transactedAt;
  final List<ExpenseSplit> splits;

  ExpenseDetail({
    required this.expenseId,
    required this.title,
    this.description,
    required this.amount,
    required this.addedBy,
    required this.groupId,
    required this.isSettlement,
    required this.isIncompleteAmount,
    required this.isIncompleteSplit,
    this.latitude,
    this.longitude,
    required this.createdAt,
    required this.transactedAt,
    required this.splits,
  });

  factory ExpenseDetail.fromJson(Map<String, dynamic> json) {
    return ExpenseDetail(
      expenseId: json["expense_id"],
      title: json["title"],
      description: json["description"],
      amount: (json["amount"] as num).toDouble(),
      addedBy: json["added_by"] as String,
      groupId: json["group_id"] as String,
      isSettlement: json["is_settlement"] ?? false,
      isIncompleteAmount: json["is_incomplete_amount"] ?? false,
      isIncompleteSplit: json["is_incomplete_split"] ?? false,
      latitude: json["latitude"] != null
          ? (json["latitude"] as num).toDouble()
          : null,
      longitude: json["longitude"] != null
          ? (json["longitude"] as num).toDouble()
          : null,
      createdAt: json["created_at"] ?? 0,
      transactedAt: json["transacted_at"] ?? 0,
      splits: (json["splits"] as List? ?? [])
          .map((e) => ExpenseSplit.fromJson(e))
          .toList(),
    );
  }
}

class ExpenseSplit {
  final String userId;
  final double amount;
  final bool isPaid;

  ExpenseSplit({
    required this.userId,
    required this.amount,
    required this.isPaid,
  });

  factory ExpenseSplit.fromJson(Map<String, dynamic> json) {
    return ExpenseSplit(
      userId: json["user_id"],
      amount: (json["amount"] as num).toDouble(),
      isPaid: json["is_paid"],
    );
  }
}
