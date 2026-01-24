class UserResult {
  final bool isSuccess;
  final String? errorMessage;

  final String? userId;
  final String? name;
  final String? email;
  final bool? guest;
  final int? createdAt;

  UserResult._({
    required this.isSuccess,
    this.errorMessage,
    this.userId,
    this.name,
    this.email,
    this.guest,
    this.createdAt,
  });

  factory UserResult.success({
    required String userId,
    required String name,
    required String email,
    required bool guest,
    required int createdAt,
  }) {
    return UserResult._(
      isSuccess: true,
      userId: userId,
      name: name,
      email: email,
      guest: guest,
      createdAt: createdAt,
    );
  }

  factory UserResult.error(String msg) {
    return UserResult._(isSuccess: false, errorMessage: msg);
  }
}
