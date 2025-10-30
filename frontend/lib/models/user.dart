class User {
  final String userId;
  final String name;
  final String email;
  final bool guest;
  final int createdAt;

  User({
    required this.userId,
    required this.name,
    required this.email,
    required this.guest,
    required this.createdAt,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      userId: json['user_id'] as String,
      name: json['name'] as String,
      email: json['email'] as String,
      guest: json['guest'] as bool? ?? false,
      createdAt: json['created_at'] as int,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'user_id': userId,
      'name': name,
      'email': email,
      'guest': guest,
      'created_at': createdAt,
    };
  }

  DateTime get createdAtDateTime => DateTime.fromMillisecondsSinceEpoch(createdAt * 1000);
}

class GroupUser {
  final String userId;
  final String name;
  final String email;
  final bool guest;
  final int joinedAt;

  GroupUser({
    required this.userId,
    required this.name,
    required this.email,
    required this.guest,
    required this.joinedAt,
  });

  factory GroupUser.fromJson(Map<String, dynamic> json) {
    return GroupUser(
      userId: json['user_id'] as String,
      name: json['name'] as String,
      email: json['email'] as String,
      guest: json['guest'] as bool? ?? false,
      joinedAt: json['joined_at'] as int,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'user_id': userId,
      'name': name,
      'email': email,
      'guest': guest,
      'joined_at': joinedAt,
    };
  }

  DateTime get joinedAtDateTime => DateTime.fromMillisecondsSinceEpoch(joinedAt * 1000);
}
