import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../apis/remote/auth_api.dart';
import '../models/user.dart';

class AuthNotifier extends AsyncNotifier<User?> {
  final AuthApi _api = AuthApi();

  @override
  User? build() => null;

  Future<bool> registerEmail({
    required String name,
    required String email,
    required String password,
  }) async {
    state = const AsyncValue.loading();
    try {
      final user = await _api.registerEmail(
        name: name,
        email: email,
        password: password,
      );
      state = AsyncValue.data(user);
      return true;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      return false;
    }
  }

  Future<bool> loginEmail({
    required String email,
    required String password,
  }) async {
    state = const AsyncValue.loading();
    try {
      final user = await _api.loginEmail(email: email, password: password);
      state = AsyncValue.data(user);
      return true;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      return false;
    }
  }

  Future<bool> refreshSession() async {
    try {
      return await _api.refreshSession();
    } catch (_) {
      return false;
    }
  }

  Future<bool> getCurrentUser() async {
    state = const AsyncValue.loading();

    try {
      final user = await _api.getCurrentUser();

      if (user == null) {
        state = const AsyncValue.data(null);
        return false;
      }

      state = AsyncValue.data(user);
      return true;
    } catch (e, st) {
      state = AsyncValue.error(e, st);
      return false;
    }
  }

  Future<void> logout() async {
    await _api.logout();
    state = const AsyncValue.data(null);
  }
}

final authProvider = AsyncNotifierProvider<AuthNotifier, User?>(
  () => AuthNotifier(),
);
