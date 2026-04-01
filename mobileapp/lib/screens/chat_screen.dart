import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobileapp/core/utils/pick_files.dart';

import '../models/message.dart';
import '../models/user.dart';
import '../notifiers/privates_notifier.dart';
import '../notifiers/ws_notifier.dart';
import '../notifiers/users_notifier.dart';
import '../widgets/auth_image.dart';

class ChatScreen extends ConsumerStatefulWidget {
  final int privateId;
  final int currentUserId;
  final int receiverId;

  const ChatScreen({
    super.key,
    required this.privateId,
    required this.currentUserId,
    required this.receiverId,
  });

  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  final _controller = TextEditingController();
  final _scrollController = ScrollController();

  Timer? _typingTimer;
  bool _isTyping = false;

  @override
  void initState() {
    super.initState();
    _scrollController.addListener(_onScroll);
    // Ensure the other user's full info is loaded
    ref.read(usersProvider.notifier).getUserById(widget.receiverId);
  }

  @override
  void dispose() {
    _controller.dispose();
    _scrollController.dispose();
    _typingTimer?.cancel();
    super.dispose();
  }

  // ---------------- PAGINATION ----------------
  void _onScroll() {
    if (_scrollController.position.pixels <=
        _scrollController.position.minScrollExtent + 80) {
      ref
          .read(privatesProvider.notifier)
          .fetchPrivateMessages(widget.privateId);
    }
  }

  // ---------------- SEND MESSAGE ----------------
  void _sendText(String text) {
    if (text.trim().isEmpty) return;
    _controller.clear();
    _stopTyping();
    _sendMessage(text, 'text');
  }

  Future<void> _pickFilesAndSend() async {
    final files = await pickFiles();
    if (files.isEmpty) return;

    for (final file in files) {
      final url = await ref
          .read(privatesProvider.notifier)
          .uploadAttachment(file, widget.privateId);
      if (url != null) _sendMessage(url, 'attachment');
    }
  }

  void _sendMessage(String content, String type) {
    ref
        .read(wsNotifierProvider.notifier)
        .sendMessage(
          privateId: widget.privateId,
          receiverId: widget.receiverId,
          content: content,
          messageType: type,
        );
  }

  // ---------------- TYPING ----------------
  void _onTyping(String _) {
    if (!_isTyping) {
      _isTyping = true;
      ref
          .read(wsNotifierProvider.notifier)
          .sendTyping(
            privateId: widget.privateId,
            receiverId: widget.receiverId,
            isTyping: true,
          );
    }
    _typingTimer?.cancel();
    _typingTimer = Timer(const Duration(seconds: 2), _stopTyping);
  }

  void _stopTyping() {
    if (!_isTyping) return;
    _isTyping = false;
    ref
        .read(wsNotifierProvider.notifier)
        .sendTyping(
          privateId: widget.privateId,
          receiverId: widget.receiverId,
          isTyping: false,
        );
  }

  // ---------------- READ RECEIPTS ----------------
  void _markReadIfNeeded(Message msg) {
    if (msg.fromId != widget.currentUserId && !msg.read) {
      ref.read(wsNotifierProvider.notifier).sendRead(msg.id);
    }
  }

  Icon _buildStatusIcon(Message msg) {
    if (msg.fromId != widget.currentUserId) return const Icon(null);
    if (msg.read) {
      return const Icon(Icons.done_all, size: 16, color: Colors.white);
    }
    if (msg.delivered) {
      return const Icon(Icons.done, size: 16, color: Colors.white);
    }
    return const Icon(Icons.access_time, size: 16, color: Colors.white);
  }

  @override
  Widget build(BuildContext context) {
    final privatesState = ref.watch(privatesProvider);
    final conv = privatesState.value?[widget.privateId];
    final messages = conv?.messages ?? const <Message>[];

    final isTyping =
        conv?.typing.entries.any(
          (e) => e.key != widget.currentUserId && e.value,
        ) ??
        false;

    // Get full receiver user info
    final otherUser = ref.watch(usersProvider)[widget.receiverId];
    final displayName = otherUser?.name ?? 'User ${widget.receiverId}';

    return Scaffold(
      appBar: AppBar(title: Text(displayName)),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              controller: _scrollController,
              reverse: true,
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: messages.length + 1,
              itemBuilder: (_, index) {
                if (index == messages.length) {
                  return conv?.isLoadingMore == true
                      ? const Padding(
                          padding: EdgeInsets.all(12),
                          child: Center(
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                        )
                      : const SizedBox.shrink();
                }

                final msg = messages[messages.length - 1 - index];
                final isMe = msg.fromId == widget.currentUserId;
                WidgetsBinding.instance.addPostFrameCallback(
                  (_) => _markReadIfNeeded(msg),
                );

                return _MessageBubble(
                  message: msg,
                  isMe: isMe,
                  statusIcon: _buildStatusIcon(msg),
                  otherUser: isMe ? null : otherUser,
                );
              },
            ),
          ),
          if (isTyping)
            const Padding(
              padding: EdgeInsets.only(left: 16, bottom: 6),
              child: Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  'Typing…',
                  style: TextStyle(
                    fontSize: 12,
                    color: Colors.grey,
                    fontStyle: FontStyle.italic,
                  ),
                ),
              ),
            ),
          _ChatInput(
            controller: _controller,
            onChanged: _onTyping,
            onSend: _sendText,
            onAttach: _pickFilesAndSend,
          ),
        ],
      ),
    );
  }
}

// ---------------- CHAT INPUT ----------------
class _ChatInput extends StatelessWidget {
  final TextEditingController controller;
  final ValueChanged<String> onChanged;
  final ValueChanged<String> onSend;
  final VoidCallback onAttach;

  const _ChatInput({
    required this.controller,
    required this.onChanged,
    required this.onSend,
    required this.onAttach,
  });

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: controller,
                onChanged: onChanged,
                onSubmitted: onSend,
                decoration: InputDecoration(
                  hintText: 'Type a message…',
                  filled: true,
                  fillColor: Colors.grey.shade100,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(24),
                    borderSide: BorderSide.none,
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 12,
                  ),
                ),
              ),
            ),
            const SizedBox(width: 8),
            CircleAvatar(
              radius: 22,
              child: IconButton(
                icon: const Icon(Icons.add, size: 18),
                onPressed: onAttach,
              ),
            ),
            const SizedBox(width: 8),
            CircleAvatar(
              radius: 22,
              child: IconButton(
                icon: const Icon(Icons.send, size: 18),
                onPressed: () => onSend(controller.text),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ---------------- MESSAGE BUBBLE ----------------
class _MessageBubble extends StatelessWidget {
  final Message message;
  final bool isMe;
  final Icon statusIcon;
  final User? otherUser;

  const _MessageBubble({
    required this.message,
    required this.isMe,
    required this.statusIcon,
    this.otherUser,
  });

  @override
  Widget build(BuildContext context) {
    return Align(
      alignment: isMe ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        padding: const EdgeInsets.all(12),
        constraints: const BoxConstraints(maxWidth: 300),
        decoration: BoxDecoration(
          color: isMe ? Colors.blue : Colors.grey.shade300,
          borderRadius: BorderRadius.circular(14),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Flexible(
              child: message.messageType == 'attachment'
                  ? AuthImage(message.content)
                  : Text(
                      message.content,
                      style: TextStyle(
                        color: isMe ? Colors.white : Colors.black,
                      ),
                    ),
            ),
            if (isMe) ...[const SizedBox(width: 4), statusIcon],
          ],
        ),
      ),
    );
  }
}
