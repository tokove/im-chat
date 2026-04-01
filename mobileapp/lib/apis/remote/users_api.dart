import '../../core/services/dio_service.dart';
import '../../models/user.dart';

class UsersApi {
  final DioService _dioService = DioService.instance;

  // ---------------- GET USER BY ID ----------------
  Future<User> getUserById(int userId) async {
    final response = await _dioService.dio.get('/users/$userId');

    final body = response.data;
    if (body['success'] != true) {
      throw Exception(body['message'] ?? 'Failed to fetch user');
    }

    return User.fromMap(body['data']);
  }
}
