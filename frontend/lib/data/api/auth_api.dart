import '../../data/api/api_client.dart';
import '../../data/models/auth_models.dart';

/// API client for authentication endpoints.
class AuthApi {
  final ApiClient _client;

  AuthApi(this._client);

  /// POST /auth/register
  Future<RegisterResponse> register(RegisterRequest request) async {
    final response = await _client.dio.post('/auth/register', data: request.toJson());
    return RegisterResponse.fromJson(response.data as Map<String, dynamic>);
  }

  /// POST /auth/login
  /// Automatically persists the returned tokens via [ApiClient.saveTokens].
  Future<AuthResponse> login(LoginRequest request) async {
    final response = await _client.dio.post('/auth/login', data: request.toJson());
    final authResp = AuthResponse.fromJson(response.data as Map<String, dynamic>);
    await _client.saveTokens(
      accessToken: authResp.accessToken,
      refreshToken: authResp.refreshToken,
    );
    return authResp;
  }

  /// POST /auth/refresh
  /// Automatically persists new tokens via [ApiClient.saveTokens].
  Future<AuthResponse> refresh(String refreshToken) async {
    final response = await _client.dio.post('/auth/refresh', data: {
      'refresh_token': refreshToken,
    });
    final authResp = AuthResponse.fromJson(response.data as Map<String, dynamic>);
    await _client.saveTokens(
      accessToken: authResp.accessToken,
      refreshToken: authResp.refreshToken,
    );
    return authResp;
  }

  /// POST /auth/logout
  Future<void> logout() async {
    final token = await _client.refreshToken;
    if (token != null) {
      try {
        await _client.dio.post('/auth/logout', data: {
          'refresh_token': token,
        });
      } catch (_) {
        // Ignore network errors during logout — tokens are cleared locally.
      }
    }
    await _client.clearTokens();
  }

  /// Check whether a stored session exists (tokens present).
  Future<bool> get isLoggedIn => _client.isAuthenticated;

  /// Restore the access token from secure storage (call at app startup).
  Future<void> tryRestoreSession() => _client.restoreAccessToken();
}
