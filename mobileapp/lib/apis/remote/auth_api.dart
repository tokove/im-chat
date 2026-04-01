import 'package:shared_preferences/shared_preferences.dart';

import '../../core/services/dio_service.dart';
import '../../models/user.dart';

class AuthApi {
  final DioService _dioService = DioService.instance;

  // ---------------- REGISTER EMAIL ----------------
  Future<User> registerEmail({
    required String name,
    required String email,
    required String password,
  }) async {
    final response = await _dioService.dio.post(
      '/auth/register-email',
      data: {'name': name, 'email': email, 'password': password},
    );

    final body = response.data;
    if (body['success'] != true) {
      throw Exception(body['message'] ?? 'Registration failed');
    }

    return User.fromMap(body['data']);
  }

  // ---------------- LOGIN ----------------
  Future<User> loginEmail({
    required String email,
    required String password,
  }) async {
    final response = await _dioService.dio.post(
      '/auth/login-email',
      data: {'email': email, 'password': password},
    );

    final body = response.data;
    if (body['success'] != true) {
      throw Exception(body['message'] ?? 'Login failed');
    }

    final data = body['data'];
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('access_token', data['access_token']);
    await prefs.setString('refresh_token', data['refresh_token']);

    return User.fromMap(data['user']);
  }

  // ---------------- REFRESH SESSION ----------------
  Future<bool> refreshSession() async {
    final prefs = await SharedPreferences.getInstance();
    final refreshToken = prefs.getString('refresh_token');

    if (refreshToken == null) return false;

    final response = await _dioService.dio.post(
      '/auth/refresh-session',
      data: {'refresh_token': refreshToken},
    );

    final body = response.data;

    if (body['success'] != true) {
      await prefs.remove('access_token');
      await prefs.remove('refresh_token');
      return false;
    }

    final data = body['data'];
    await prefs.setString('access_token', data['access_token']);
    await prefs.setString('refresh_token', data['refresh_token']);

    return true;
  }

  // ---------------- CURRENT USER ----------------
  Future<User?> getCurrentUser() async {
    final prefs = await SharedPreferences.getInstance();
    final refreshToken = prefs.getString('refresh_token');

    if (refreshToken == null) return null;

    final response = await _dioService.dio.post(
      '/auth/current-user',
      data: {'refresh_token': refreshToken},
    );

    final body = response.data;

    if (body['success'] != true) return null;

    return User.fromMap(body['data']);
  }

  // ---------------- LOGOUT ----------------
  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();

    final accessToken = prefs.getString('access_token');
    if (accessToken != null && accessToken.isNotEmpty) {
      await _dioService.dio.post('/auth/logout');
    }
    await prefs.remove('access_token');
    await prefs.remove('refresh_token');
  }
}
