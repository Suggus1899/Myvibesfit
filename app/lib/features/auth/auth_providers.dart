import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/providers.dart';

enum AuthStatus { unknown, authenticated, unauthenticated }

class AuthState {
  final AuthStatus status;
  final String? error;
  const AuthState({this.status = AuthStatus.unknown, this.error});
}

class AuthController extends StateNotifier<AuthState> {
  AuthController(this._ref) : super(const AuthState()) {
    _ref.read(apiClientProvider).onSessionExpired = _handleSessionExpired;
    _bootstrap();
  }

  final Ref _ref;

  void _handleSessionExpired() {
    if (state.status != AuthStatus.unauthenticated) {
      state = const AuthState(status: AuthStatus.unauthenticated);
    }
  }

  Future<void> _bootstrap() async {
    final hasSession = await _ref.read(apiRepositoryProvider).hasSession();
    state = AuthState(status: hasSession ? AuthStatus.authenticated : AuthStatus.unauthenticated);
  }

  Future<void> login(String email, String password) async {
    try {
      await _ref.read(apiRepositoryProvider).login(email: email, password: password);
      state = const AuthState(status: AuthStatus.authenticated);
    } catch (e) {
      state = AuthState(status: AuthStatus.unauthenticated, error: _friendlyError(e));
      rethrow;
    }
  }

  Future<void> register(String email, String password, String fullName) async {
    try {
      await _ref.read(apiRepositoryProvider).register(email: email, password: password, fullName: fullName);
      state = const AuthState(status: AuthStatus.authenticated);
    } catch (e) {
      state = AuthState(status: AuthStatus.unauthenticated, error: _friendlyError(e));
      rethrow;
    }
  }

  Future<void> logout() async {
    await _ref.read(apiRepositoryProvider).logout();
    state = const AuthState(status: AuthStatus.unauthenticated);
  }

  String _friendlyError(Object e) => e.toString();
}

final authControllerProvider = StateNotifierProvider<AuthController, AuthState>((ref) => AuthController(ref));
