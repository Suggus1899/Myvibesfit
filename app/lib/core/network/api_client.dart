import 'package:dio/dio.dart';

import '../config.dart';
import 'token_storage.dart';

/// Cliente dio compartido por toda la app. El interceptor adjunta el access
/// token a cada request y, si la API responde 401, intenta refrescar una
/// sola vez antes de reintentar (evita loops si el refresh tambien expiro).
class ApiClient {
  ApiClient(this._tokenStorage) : dio = Dio(BaseOptions(baseUrl: apiBaseUrl, connectTimeout: const Duration(seconds: 10))) {
    dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final token = await _tokenStorage.readAccessToken();
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (error, handler) async {
        final isAuthEndpoint = error.requestOptions.path.startsWith('/v1/auth/');
        if (error.response?.statusCode == 401 && !isAuthEndpoint && !_isRetry(error.requestOptions)) {
          final refreshed = await _tryRefresh();
          if (refreshed != null) {
            final retryOptions = error.requestOptions;
            retryOptions.headers['Authorization'] = 'Bearer $refreshed';
            retryOptions.extra['retried'] = true;
            try {
              final response = await dio.fetch(retryOptions);
              return handler.resolve(response);
            } catch (_) {
              // cae al error original
            }
          }
        }
        handler.next(error);
      },
    ));
  }

  final Dio dio;
  final TokenStorage _tokenStorage;
  Future<String?>? _refreshing;

  bool _isRetry(RequestOptions options) => options.extra['retried'] == true;

  Future<String?> _tryRefresh() {
    // Coalesce: si ya hay un refresh en vuelo, todas las requests 401
    // simultaneas esperan el mismo resultado en vez de disparar N refreshes.
    return _refreshing ??= _doRefresh().whenComplete(() => _refreshing = null);
  }

  Future<String?> _doRefresh() async {
    final refreshToken = await _tokenStorage.readRefreshToken();
    if (refreshToken == null) return null;
    try {
      final response = await Dio(BaseOptions(baseUrl: apiBaseUrl)).post(
        '/v1/auth/refresh',
        data: {'refresh_token': refreshToken},
      );
      final access = response.data['access_token'] as String;
      final refresh = response.data['refresh_token'] as String;
      await _tokenStorage.save(accessToken: access, refreshToken: refresh);
      return access;
    } catch (_) {
      await _tokenStorage.clear();
      return null;
    }
  }
}
