class UserLookup {
  final String userId;
  final String name;
  final String email;
  final bool guest;
  final int createdAt;

  UserLookup({
    required this.userId,
    required this.name,
    required this.email,
    required this.guest,
    required this.createdAt,
  });

  factory UserLookup.fromJson(Map<String, dynamic> json) {
    return UserLookup(
      userId: json["user_id"],
      name: json["name"],
      email: json["email"],
      guest: json["guest"],
      createdAt: json["created_at"],
    );
  }
}

//=============RESULT WRAPPER================
class UserLookupResult {
  final bool isSuccess;
  final String? errorMessage;
  final UserLookup? user;

  UserLookupResult._({
    required this.isSuccess,
    this.errorMessage,
    this.user,
  });

  factory UserLookupResult.success(UserLookup user) {
    return UserLookupResult._(
      isSuccess: true,
      user: user,
    );
  }

  factory UserLookupResult.error(String msg) {
    return UserLookupResult._(
      isSuccess: false,
      errorMessage: msg,
    );
  }
}
