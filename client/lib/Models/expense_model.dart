class ExpenseSplit {
  final String userId;
  final double amount;
  final bool isPaid;

  ExpenseSplit({
    required this.userId,
    required this.amount,
    required this.isPaid,
  });

  Map<String, dynamic> toJson() => {
    "user_id": userId,
    "amount": amount,
    "is_paid": isPaid,
  };
}

class ExpenseRequest {
  final String groupId;
  final String title;
  final String? description;
  final double amount;
  final bool isIncompleteAmount;
  final bool isIncompleteSplit;
  final double? latitude;
  final double? longitude;
  final List<ExpenseSplit> splits;

  ExpenseRequest({
    required this.groupId,
    required this.title,
    this.description,
    required this.amount,
    required this.isIncompleteAmount,
    required this.isIncompleteSplit,
    this.latitude,
    this.longitude,
    required this.splits,
  });

  Map<String, dynamic> toJson() => {
    "group_id": groupId,
    "title": title,
    "description": description,
    "amount": amount,
    "is_incomplete_amount": isIncompleteAmount,
    "is_incomplete_split": isIncompleteSplit,
    "latitude": latitude,
    "longitude": longitude,
    "splits": splits.map((e) => e.toJson()).toList(),
  };
}

class BasicResult {
  final bool isSuccess;
  final String? errorMessage;

  BasicResult._({required this.isSuccess, this.errorMessage});

  factory BasicResult.success() => BasicResult._(isSuccess: true);

  factory BasicResult.error(String msg) =>
      BasicResult._(isSuccess: false, errorMessage: msg);
}
