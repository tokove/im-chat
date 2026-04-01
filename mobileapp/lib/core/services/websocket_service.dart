import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../consts/base_url.dart';

typedef OnWsEvent = void Function(Map<String, dynamic> event);

class WsService {
  final OnWsEvent onEvent;
  WebSocketChannel? _channel;
  bool _isConnecting = false;

  WsService({required this.onEvent});

  void connect() async {
    if (_isConnecting) return;
    _isConnecting = true;

    final prefs = await SharedPreferences.getInstance();
    final accessToken = prefs.getString('access_token');
    if (accessToken == null || accessToken.isEmpty) {
      _isConnecting = false;
      return;
    }

    final platform = kIsWeb ? 'web' : 'mobile';
    final separator = baseUrlWs.contains('?') ? '&' : '?';
    final wsUrl = '${baseUrlWs}${separator}token=$accessToken&platform=$platform';

    try {
      _channel = WebSocketChannel.connect(Uri.parse(wsUrl));

      _channel!.stream.listen(
        (data) {
          try {
            final event = jsonDecode(data as String) as Map<String, dynamic>;
            onEvent(event);
          } catch (_) {}
        },
        onDone: () {
          _isConnecting = false;
          reconnect();
        },
        onError: (_) {
          _isConnecting = false;
          reconnect();
        },
        cancelOnError: true,
      );
    } catch (_) {
      _isConnecting = false;
      reconnect();
      return;
    }

    _isConnecting = false;
  }

  void reconnect() {
    _channel?.sink.close();
    Future.delayed(const Duration(seconds: 2), connect);
  }

  void send(Map<String, dynamic> event) {
    _channel?.sink.add(jsonEncode(event));
  }

  void disconnect() {
    _channel?.sink.close();
    _channel = null;
    _isConnecting = false;
  }
}
