import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/user.dart';

class ClientsNotifier extends Notifier<Map<int, User>> {
  @override
  Map<int, User> build() {
    // initial state: nobody online
    return {};
  }

  void userOnline(User user) {
    if (state[user.id] == user) return;
    state = {...state, user.id: user};
  }

  void userOffline(int userId) {
    final newState = Map<int, User>.from(state);
    newState.remove(userId);
    state = newState;
  }

  void clear() {
    state = {};
  }

  void setAll(List<User> clients) {
    state = {for (final c in clients) c.id: c};
  }
}

final clientsProvider = NotifierProvider<ClientsNotifier, Map<int, User>>(
  ClientsNotifier.new,
);
