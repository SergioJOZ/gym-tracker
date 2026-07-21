/// Login request body.
class LoginRequest {
  final String email;
  final String password;

  const LoginRequest({required this.email, required this.password});

  Map<String, dynamic> toJson() => {
    'email': email,
    'password': password,
  };
}

/// Login / refresh response body.
class AuthResponse {
  final String accessToken;
  final String refreshToken;

  const AuthResponse({required this.accessToken, required this.refreshToken});

  factory AuthResponse.fromJson(Map<String, dynamic> json) => AuthResponse(
    accessToken: json['access_token'] as String,
    refreshToken: json['refresh_token'] as String,
  );
}

/// Registration request body.
class RegisterRequest {
  final String email;
  final String password;

  const RegisterRequest({required this.email, required this.password});

  Map<String, dynamic> toJson() => {
    'email': email,
    'password': password,
  };
}

/// Registration response — returns the created user.
class RegisterResponse {
  final String id;
  final String email;
  final String createdAt;

  const RegisterResponse({
    required this.id,
    required this.email,
    required this.createdAt,
  });

  factory RegisterResponse.fromJson(Map<String, dynamic> json) => RegisterResponse(
    id: json['id'] as String,
    email: json['email'] as String,
    createdAt: json['created_at'] as String,
  );
}

/// Refresh token request body.
class RefreshRequest {
  final String refreshToken;

  const RefreshRequest({required this.refreshToken});

  Map<String, dynamic> toJson() => {
    'refresh_token': refreshToken,
  };
}
