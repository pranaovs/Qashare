class Expense {
  final String expenseId;
  final String groupId;
  final String addedBy;
  final String title;
  final String? description;
  final int createdAt;
  final double amount;
  final bool isIncompleteAmount;
  final bool isIncompleteSplit;
  final double? latitude;
  final double? longitude;
  final List<ExpenseSplit> splits;

  Expense({
    required this.expenseId,
    required this.groupId,
    required this.addedBy,
    required this.title,
    this.description,
    required this.createdAt,
    required this.amount,
    required this.isIncompleteAmount,
    required this.isIncompleteSplit,
    this.latitude,
    this.longitude,
    this.splits = const [],
  });

  factory Expense.fromJson(Map<String, dynamic> json) {
    return Expense(
      expenseId: json['expense_id'] as String,
      groupId: json['group_id'] as String,
      addedBy: json['added_by'] as String,
      title: json['title'] as String,
      description: json['description'] as String?,
      createdAt: json['created_at'] as int,
      amount: (json['amount'] as num).toDouble(),
      isIncompleteAmount: json['is_incomplete_amount'] as bool? ?? false,
      isIncompleteSplit: json['is_incomplete_split'] as bool? ?? false,
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      splits: (json['splits'] as List<dynamic>?)
              ?.map((s) => ExpenseSplit.fromJson(s as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'expense_id': expenseId,
      'group_id': groupId,
      'added_by': addedBy,
      'title': title,
      'description': description,
      'created_at': createdAt,
      'amount': amount,
      'is_incomplete_amount': isIncompleteAmount,
      'is_incomplete_split': isIncompleteSplit,
      'latitude': latitude,
      'longitude': longitude,
      'splits': splits.map((s) => s.toJson()).toList(),
    };
  }

  DateTime get createdAtDateTime => DateTime.fromMillisecondsSinceEpoch(createdAt * 1000);

  List<ExpenseSplit> get paidSplits => splits.where((s) => s.isPaid).toList();
  List<ExpenseSplit> get owedSplits => splits.where((s) => !s.isPaid).toList();
}

class ExpenseSplit {
  final String userId;
  final double amount;
  final bool isPaid; // true = paid (lender), false = owes (borrower)

  ExpenseSplit({
    required this.userId,
    required this.amount,
    required this.isPaid,
  });

  factory ExpenseSplit.fromJson(Map<String, dynamic> json) {
    return ExpenseSplit(
      userId: json['user_id'] as String,
      amount: (json['amount'] as num).toDouble(),
      isPaid: json['is_paid'] as bool,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'user_id': userId,
      'amount': amount,
      'is_paid': isPaid,
    };
  }
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
    this.isIncompleteAmount = false,
    this.isIncompleteSplit = false,
    this.latitude,
    this.longitude,
    required this.splits,
  });

  Map<String, dynamic> toJson() {
    return {
      'group_id': groupId,
      'title': title,
      'description': description,
      'amount': amount,
      'is_incomplete_amount': isIncompleteAmount,
      'is_incomplete_split': isIncompleteSplit,
      'latitude': latitude,
      'longitude': longitude,
      'splits': splits.map((s) => s.toJson()).toList(),
    };
  }
}
