import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../apis/remote/users_api.dart';
import '../models/user.dart';

class UsersNotifier extends Notifier<Map<int, User>> {
  final UsersApi _usersApi = UsersApi();

  @override
  Map<int, User> build() {
    return {};
  }

  Future<User?> getUserById(int userId) async {
    if (state.containsKey(userId)) {
      return state[userId];
    }

    try {
      final user = await _usersApi.getUserById(userId);
      state = {...state, user.id: user};
      return user;
    } catch (e, _) {
      return null;
    }
  }

  void setUsers(List<User> users) {
    state = {...state, for (final u in users) u.id: u};
  }

  void clear() {
    state = {};
  }
}

final usersProvider = NotifierProvider<UsersNotifier, Map<int, User>>(
  UsersNotifier.new,
);
