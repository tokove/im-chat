import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/message.dart';
import '../models/user.dart';
import '../core/services/websocket_service.dart';
import 'clients_notifier.dart';
import 'privates_notifier.dart';

class WsNotifier extends Notifier<void> {
  late final WsService _wsService;
  late final AsyncValue<User?> _currentUser;
  late final ClientsNotifier _clientsNotifier;
  late final PrivatesNotifier _privatesNotifier;

  @override
  void build() {}

  void init({
    required AsyncValue<User?> currentUser,
    required ClientsNotifier clientsNotifier,
    required PrivatesNotifier privatesNotifier,
  }) {
    _currentUser = currentUser;
    _clientsNotifier = clientsNotifier;
    _privatesNotifier = privatesNotifier;

    _wsService = WsService(onEvent: _handleEvent);
    _wsService.connect();
  }

  void _handleEvent(Map<String, dynamic> event) {
    final type = event['event_type'];
    final payload = event['payload'];

    switch (type) {
      case 'current_users':
        final users = (payload as List)
            .map((e) => User.fromMap(e as Map<String, dynamic>))
            .toList();

        _clientsNotifier.setAll(users);
        break;

      case 'online':
        final user = User.fromMap(payload as Map<String, dynamic>);
        _clientsNotifier.userOnline(user);
        break;

      case 'offline':
        final user = User.fromMap(payload as Map<String, dynamic>);
        _clientsNotifier.userOffline(user.id);
        break;

      case 'message':
        final msgMap = payload['message'] as Map<String, dynamic>;
        final msg = Message.fromMap(msgMap);

        _privatesNotifier.addMessage(msg);

        // Receiver requests delivery confirmation
        if (msg.fromId != _currentUser.value!.id) {
          sendEvent({
            'event_type': 'delivered',
            'payload': {'message_id': msg.id},
          });
        }
        break;

      case 'typing':
        final privateId = payload['private_id'] as int;
        final userId = payload['user_id'] as int;
        final isTyping = payload['is_typing'] as bool;

        _privatesNotifier.setTyping(
          privateId: privateId,
          userId: userId,
          isTyping: isTyping,
        );
        break;

      case 'delivered':
        final messageId =
            (payload as Map<String, dynamic>)['message_id'] as int;
        _privatesNotifier.markMessageDelivered(messageId);
        break;

      case 'read':
        final messageId =
            (payload as Map<String, dynamic>)['message_id'] as int;
        _privatesNotifier.markMessageRead(messageId);
        break;

      case 'heartbeat':
        break;

      default:
        break;
    }
  }

  void sendEvent(Map<String, dynamic> event) => _wsService.send(event);

  void sendMessage({
    required int privateId,
    required int receiverId,
    required String content,
    required String messageType,
  }) {
    sendEvent({
      "event_type": "message",
      "payload": {
        "private_id": privateId,
        "receiver_id": receiverId,
        "content": content,
        "message_type": messageType,
      },
    });
  }

  void sendRead(int messageId) {
    sendEvent({
      'event_type': 'read',
      'payload': {'message_id': messageId},
    });
  }

  void sendTyping({
    required int privateId,
    required int receiverId,
    required bool isTyping,
  }) {
    sendEvent({
      'event_type': 'typing',
      'payload': {
        'private_id': privateId,
        'receiver_id': receiverId,
        'is_typing': isTyping,
      },
    });
  }

  void dispose() {
    _wsService.disconnect();
  }
}

final wsNotifierProvider = NotifierProvider<WsNotifier, void>(WsNotifier.new);
