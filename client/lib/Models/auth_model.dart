// ================= REGISTER RESULT =================

class RegisterResult {
  final bool isSuccess;
  final String? errorMessage;

  final String? userId;
  final String? name;
  final String? email;
  final bool? guest;
  final int? createdAt;

  RegisterResult._({
    required this.isSuccess,
    this.errorMessage,
    this.userId,
    this.name,
    this.email,
    this.guest,
    this.createdAt,
  });

  factory RegisterResult.success({
    required String userId,
    required String name,
    required String email,
    required bool guest,
    required int createdAt,
  }) {
    return RegisterResult._(
      isSuccess: true,
      userId: userId,
      name: name,
      email: email,
      guest: guest,
      createdAt: createdAt,
    );
  }

  factory RegisterResult.error(String message) {
    return RegisterResult._(
      isSuccess: false,
      errorMessage: message,
    );
  }
}

// ================= LOGIN RESULT =================

class LoginResult {
  final bool isSuccess;
  final String? errorMessage;

  final String? token;
  final String? message;

  LoginResult._({
    required this.isSuccess,
    this.errorMessage,
    this.token,
    this.message,
  });

  factory LoginResult.success({
    required String token,
    required String message,
  }) {
    return LoginResult._(
      isSuccess: true,
      token: token,
      message: message,
    );
  }

  factory LoginResult.error(String message) {
    return LoginResult._(
      isSuccess: false,
      errorMessage: message,
    );
  }
}
