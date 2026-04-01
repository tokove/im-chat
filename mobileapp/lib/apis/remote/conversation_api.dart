import '../../core/services/dio_service.dart';
import '../../models/message.dart';
import '../../models/private.dart';

class ConversationApi {
  final DioService _dioService = DioService.instance;

  // ---------------- GET ALL PRIVATES ----------------
  Future<List<Private>> fetchPrivates() async {
    try {
      final res = await _dioService.dio.get('/conversations');

      if (res.data['success'] != true) {
        throw Exception(res.data['message']);
      }

      return (res.data['data'] as List).map((e) => Private.fromMap(e)).toList();
    } catch (_) {
      rethrow;
    }
  }

  // ---------------- GET MESSAGES (PAGINATED) ----------------
  Future<List<Message>> fetchPrivateMessages(
    int privateId, {
    int page = 1,
    int limit = 20,
  }) async {
    try {
      final res = await _dioService.dio.get(
        '/conversations/privates/$privateId/messages',
        queryParameters: {'page': page, 'limit': limit},
      );
      print(res);
      if (res.data['success'] != true) return [];

      final data = res.data['data'];

      final messages = (data['messages'] as List)
          .map((e) => Message.fromMap(e))
          .toList();

      messages.sort((a, b) => a.createdAt.compareTo(b.createdAt));
      return messages;
    } catch (_) {
      return [];
    }
  }

  // ---------------- GET SINGLE PRIVATE ----------------
  Future<Private> fetchPrivateById(int privateId) async {
    final res = await _dioService.dio.get('/conversations/privates/$privateId');

    if (res.data['success'] != true) throw Exception(res.data['message']);

    final private = Private.fromMap(res.data['data']);
    final messages = await fetchPrivateMessages(private.id);

    return private.copyWith(messages: messages);
  }

  // ---------------- JOIN / CREATE PRIVATE ----------------
  Future<Private> joinPrivate({required int otherUserId}) async {
    final res = await _dioService.dio.post(
      '/conversations/privates/create',
      data: {'receiver_id': otherUserId},
    );

    if (res.data['success'] != true) throw Exception(res.data['message']);

    return Private.fromMap(res.data['data']);
  }
}
