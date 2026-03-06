class UserSpending {
  final String expenseId;
  final String groupId;
  final String title;
  final String? description;
  final double amount;
  final double userAmount;
  final int createdAt;
  final int? transactedAt;
  final bool isIncompleteAmount;
  final bool isIncompleteSplit;
  final bool isSettlement;
  final String addedBy;

  UserSpending({
    required this.expenseId,
    required this.groupId,
    required this.title,
    this.description,
    required this.amount,
    required this.userAmount,
    required this.createdAt,
    this.transactedAt,
    required this.isIncompleteAmount,
    required this.isIncompleteSplit,
    required this.isSettlement,
    required this.addedBy,
  });

  factory UserSpending.fromJson(Map<String, dynamic> json) {
    return UserSpending(
      expenseId: json["expense_id"],
      groupId: json["group_id"],
      title: json["title"] ?? "",
      description: json["description"],
      amount: (json["amount"] as num).toDouble(),
      userAmount: (json["user_amount"] as num).toDouble(),
      createdAt: json["created_at"],
      transactedAt: json["transacted_at"],
      isIncompleteAmount: json["is_incomplete_amount"] ?? false,
      isIncompleteSplit: json["is_incomplete_split"] ?? false,
      isSettlement: json["is_settlement"] ?? false,
      addedBy: json["added_by"] ?? "",
    );
  }
}

class SpendingResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<UserSpending>? spendings;

  SpendingResult._({
    required this.isSuccess,
    this.errorMessage,
    this.spendings,
  });

  factory SpendingResult.success(List<UserSpending> list) {
    return SpendingResult._(isSuccess: true, spendings: list);
  }

  factory SpendingResult.error(String msg) {
    return SpendingResult._(isSuccess: false, errorMessage: msg);
  }
}
