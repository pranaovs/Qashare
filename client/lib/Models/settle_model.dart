class Settlement {
  final double amount;
  final int? createdAt;
  final String groupId;
  final String title;
  final int? transactedAt;
  final String userId;

  Settlement({
    required this.amount,
    this.createdAt,
    required this.groupId,
    required this.title,
    this.transactedAt,
    required this.userId,
  });

  factory Settlement.fromJson(Map<String, dynamic> json) {
    return Settlement(
      amount: (json["amount"] as num).toDouble(),
      createdAt: json["created_at"] as int?,
      groupId: json["group_id"] as String,
      title: json["title"] ?? "",
      transactedAt: json["transacted_at"] as int?,
      userId: json["user_id"] as String,
    );
  }
}

class SettleResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<Settlement>? settlements;

  SettleResult._({
    required this.isSuccess,
    this.errorMessage,
    this.settlements,
  });

  factory SettleResult.success(List<Settlement> list) {
    return SettleResult._(isSuccess: true, settlements: list);
  }

  factory SettleResult.error(String msg) {
    return SettleResult._(isSuccess: false, errorMessage: msg);
  }
}
