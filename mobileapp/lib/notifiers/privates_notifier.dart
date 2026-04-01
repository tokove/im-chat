import 'package:file_picker/file_picker.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../apis/remote/conversation_api.dart';
import '../apis/remote/upload_api.dart';
import '../models/private.dart';
import '../models/message.dart';

class PrivatesNotifier extends AsyncNotifier<Map<int, Private>> {
  final ConversationApi _api = ConversationApi();
  final UploadApi _uploadApi = UploadApi();

  @override
  Future<Map<int, Private>> build() async {
    await fetchPrivates();
    return state.value ?? {};
  }

  Future<void> fetchPrivates() async {
    state = const AsyncValue.loading();
    try {
      final privates = await _api.fetchPrivates();

      final enriched = await Future.wait(
        privates.map((p) async {
          final messages = await _api.fetchPrivateMessages(p.id);
          return p.copyWith(messages: messages);
        }),
      );

      final map = <int, Private>{for (final p in enriched) p.id: p};
      state = AsyncValue.data(map);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<Private> fetchPrivateById(int privateId) async {
    return await _api.fetchPrivateById(privateId);
  }

  Future<Private> joinPrivate({
    required int otherUserId,
    required int currentUserId,
  }) async {
    final current = state.value ?? {};
    for (final p in current.values) {
      if ((p.user1 == currentUserId && p.user2 == otherUserId) ||
          (p.user2 == currentUserId && p.user1 == otherUserId)) {
        return p;
      }
    }

    final private = await _api.joinPrivate(otherUserId: otherUserId);
    await addPrivate(private);
    return private;
  }

  Future<void> addMessage(Message msg) async {
    final current = state.value ?? {};
    Private? private = current[msg.privateId];

    if (private == null) {
      try {
        private = await fetchPrivateById(msg.privateId);
        await addPrivate(private);
      } catch (_) {
        return;
      }
    }

    final privateMessages = [...private.messages];
    if (privateMessages.any((m) => m.id == msg.id)) return;

    privateMessages.add(msg);
    privateMessages.sort((a, b) => a.createdAt.compareTo(b.createdAt));

    state = AsyncValue.data({
      ...state.value!,
      msg.privateId: private.copyWith(messages: privateMessages),
    });
  }

  Future<void> addPrivate(Private p) async {
    final current = state.value ?? {};
    if (current.containsKey(p.id)) return;

    List<Message> messages = p.messages;
    if (messages.isEmpty) {
      messages = await _api.fetchPrivateMessages(p.id);
    }

    state = AsyncValue.data({...current, p.id: p.copyWith(messages: messages)});
  }

  void setMessages(int privateId, List<Message> messages) {
    final current = state.value ?? {};
    final conv = current[privateId];
    if (conv == null) return;

    messages.sort((a, b) => a.createdAt.compareTo(b.createdAt));
    state = AsyncValue.data({
      ...current,
      privateId: conv.copyWith(messages: messages),
    });
  }

  Future<void> fetchPrivateMessages(int privateId) async {
    final current = state.value ?? {};
    final conv = current[privateId];

    if (conv == null) return;
    if (!conv.hasNextPage || conv.isLoadingMore) return;

    state = AsyncValue.data({
      ...current,
      privateId: conv.copyWith(isLoadingMore: true),
    });

    try {
      final messages = await _api.fetchPrivateMessages(
        privateId,
        page: conv.page,
        limit: conv.limit,
      );

      final hasNextPage = messages.length == conv.limit;

      final merged = {
        for (final m in [...messages, ...conv.messages]) m.id: m,
      }.values.toList()..sort((a, b) => a.createdAt.compareTo(b.createdAt));

      state = AsyncValue.data({
        ...current,
        privateId: conv.copyWith(
          messages: merged,
          page: conv.page + 1,
          hasNextPage: hasNextPage,
          isLoadingMore: false,
        ),
      });
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  void setTyping({
    required int privateId,
    required int userId,
    required bool isTyping,
  }) {
    final current = state.value ?? {};
    final conv = current[privateId];
    if (conv == null) return;

    final typing = Map<int, bool>.from(conv.typing);
    typing[userId] = isTyping;

    state = AsyncValue.data({
      ...current,
      privateId: conv.copyWith(typing: typing),
    });
  }

  List<Message> messages(int privateId) {
    return state.value?[privateId]?.messages ?? [];
  }

  void markMessageDelivered(int messageId) {
    final current = state.value ?? {};
    final updated = <int, Private>{};

    current.forEach((key, conv) {
      if (conv.messages.isNotEmpty) {
        final msgs = conv.messages
            .map((m) => m.id == messageId ? m.copyWith(delivered: true) : m)
            .toList();
        updated[key] = conv.copyWith(messages: msgs);
      }
    });

    state = AsyncValue.data(updated);
  }

  Future<void> markMessageRead(int messageId) async {
    final current = state.value ?? {};
    int? privateId;

    for (final entry in current.entries) {
      if (entry.value.messages.any((m) => m.id == messageId)) {
        privateId = entry.key;
        break;
      }
    }
    if (privateId == null) return;

    final conv = current[privateId]!;

    final updatedMessages = conv.messages.map((m) {
      return m.id == messageId ? m.copyWith(read: true) : m;
    }).toList();

    state = AsyncValue.data({
      ...current,
      privateId: conv.copyWith(messages: updatedMessages),
    });
  }

  Future<String?> uploadAttachment(PlatformFile file, int privateId) async {
    return await _uploadApi.uploadAttachment(file, privateId);
  }
}

final privatesProvider =
    AsyncNotifierProvider<PrivatesNotifier, Map<int, Private>>(
      PrivatesNotifier.new,
    );
