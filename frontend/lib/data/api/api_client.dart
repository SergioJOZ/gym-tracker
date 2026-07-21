import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Thrown when the API returns a structured error.
class ApiException implements Exception {
  final int? statusCode;
  final String code;
  final String message;
  final List<Map<String, dynamic>>? details;

  const ApiException({
    this.statusCode,
    required this.code,
    required this.message,
    this.details,
  });

  factory ApiException.fromResponse(Response response) {
    final body = response.data is Map ? response.data as Map : null;
    final error = body?['error'] as Map?;
    return ApiException(
      statusCode: response.statusCode,
      code: error?['code'] as String? ?? 'UNKNOWN',
      message: error?['message'] as String? ?? response.statusMessage ?? 'Unknown error',
      details: (error?['details'] as List?)?.cast<Map<String, dynamic>>(),
    );
  }

  @override
  String toString() => 'ApiException($statusCode): $code — $message';
}

/// Thrown when the user needs to re-authenticate.
class UnauthenticatedException extends ApiException {
  const UnauthenticatedException() : super(code: 'UNAUTHENTICATED', message: 'Session expired');
}

/// Central API client with JWT auto-refresh and secure token storage.
class ApiClient {
  static const String _accessKey = 'jwt_access';
  static const String _refreshKey = 'jwt_refresh';

  final Dio _dio;
  final FlutterSecureStorage _storage;

  String? _accessToken;

  ApiClient({
    required String baseUrl,
    FlutterSecureStorage? storage,
    Duration connectTimeout = const Duration(seconds: 10),
    Duration receiveTimeout = const Duration(seconds: 15),
  }) : _storage = storage ?? const FlutterSecureStorage(),
       _dio = Dio(BaseOptions(
         baseUrl: baseUrl,
         connectTimeout: connectTimeout,
         receiveTimeout: receiveTimeout,
         headers: {'Content-Type': 'application/json'},
       )) {
    _dio.interceptors.add(_AuthInterceptor(this));
  }

  /// Direct access to Dio for custom requests.
  Dio get dio => _dio;

  /// The stored refresh token (for use by AuthApi).
  Future<String?> get refreshToken => _storage.read(key: _refreshKey);

  /// Whether the user has a stored session.
  Future<bool> get isAuthenticated async {
    final token = await _storage.read(key: _accessKey);
    return token != null && token.isNotEmpty;
  }

  /// Persist tokens after login or refresh.
  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    _accessToken = accessToken;
    await _storage.write(key: _accessKey, value: accessToken);
    await _storage.write(key: _refreshKey, value: refreshToken);
  }

  /// Clear stored tokens (logout).
  Future<void> clearTokens() async {
    _accessToken = null;
    await _storage.delete(key: _accessKey);
    await _storage.delete(key: _refreshKey);
  }

  /// Internal: called by the interceptor to add the auth header.
  String? get accessToken => _accessToken;

  /// Internal: restore access token from storage (called at startup).
  Future<void> restoreAccessToken() async {
    _accessToken = await _storage.read(key: _accessKey);
  }
}

/// Dio interceptor that injects the JWT and handles 401 → refresh → retry.
class _AuthInterceptor extends Interceptor {
  final ApiClient _client;

  _AuthInterceptor(this._client);

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final token = _client.accessToken;
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode != 401) {
      return handler.next(err);
    }

    final refreshToken = await _client.refreshToken;
    if (refreshToken == null) {
      return handler.reject(DioException(
        requestOptions: err.requestOptions,
        error: const UnauthenticatedException(),
        type: DioExceptionType.badResponse,
        response: err.response,
      ));
    }

    try {
      // Attempt token refresh
      final refreshResp = await _client.dio.post('/auth/refresh', data: {
        'refresh_token': refreshToken,
      });

      final data = refreshResp.data as Map;
      final newAccess = data['access_token'] as String;
      final newRefresh = data['refresh_token'] as String;

      await _client.saveTokens(accessToken: newAccess, refreshToken: newRefresh);

      // Retry the original request with the new token
      final retryOptions = err.requestOptions;
      retryOptions.headers['Authorization'] = 'Bearer $newAccess';

      final response = await _client.dio.fetch(retryOptions);
      return handler.resolve(response);
    } catch (_) {
      await _client.clearTokens();
      return handler.reject(DioException(
        requestOptions: err.requestOptions,
        error: const UnauthenticatedException(),
        type: DioExceptionType.badResponse,
      ));
    }
  }
}
