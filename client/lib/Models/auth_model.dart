// ================= REGISTER RESULT =================

class RegisterResult {
  final bool isSuccess;
  final bool isPendingVerification;
  final String? errorMessage;

  final String? userId;
  final String? name;
  final String? email;
  final bool? guest;
  final int? createdAt;

  RegisterResult._({
    required this.isSuccess,
    this.isPendingVerification = false,
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

  factory RegisterResult.pending() {
    return RegisterResult._(
      isSuccess: true,
      isPendingVerification: true,
    );
  }

  factory RegisterResult.error(String message) {
    return RegisterResult._(isSuccess: false, errorMessage: message);
  }
}

// ================= LOGIN RESULT =================

class LoginResult {
  final bool isSuccess;
  final String? errorMessage;

  final String? accessToken;
  final String? refreshToken;
  final String? tokenType;

  LoginResult._({
    required this.isSuccess,
    this.errorMessage,
    this.accessToken,
    this.refreshToken,
    this.tokenType,
  });

  factory LoginResult.success({
    required String accessToken,
    required String refreshToken,
    String tokenType = "Bearer",
  }) {
    return LoginResult._(
      isSuccess: true,
      accessToken: accessToken,
      refreshToken: refreshToken,
      tokenType: tokenType,
    );
  }

  factory LoginResult.error(String message) {
    return LoginResult._(isSuccess: false, errorMessage: message);
  }
}

// ================= REFRESH RESULT =================

class RefreshResult {
  final bool isSuccess;
  final String? errorMessage;

  final String? accessToken;
  final String? refreshToken;
  final String? tokenType;

  RefreshResult._({
    required this.isSuccess,
    this.errorMessage,
    this.accessToken,
    this.refreshToken,
    this.tokenType,
  });

  factory RefreshResult.success({
    required String accessToken,
    required String refreshToken,
    String tokenType = "Bearer",
  }) {
    return RefreshResult._(
      isSuccess: true,
      accessToken: accessToken,
      refreshToken: refreshToken,
      tokenType: tokenType,
    );
  }

  factory RefreshResult.error(String message) {
    return RefreshResult._(isSuccess: false, errorMessage: message);
  }
}


