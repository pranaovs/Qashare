class Settlement {
  final String fromUserId;
  final String toUserId;
  final double amount;

  Settlement({
    required this.fromUserId,
    required this.toUserId,
    required this.amount,
  });

  factory Settlement.fromJson(Map<String, dynamic> json) {
    return Settlement(
      fromUserId: json['from_user_id'] as String,
      toUserId: json['to_user_id'] as String,
      amount: (json['amount'] as num).toDouble(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'from_user_id': fromUserId,
      'to_user_id': toUserId,
      'amount': amount,
    };
  }
}
