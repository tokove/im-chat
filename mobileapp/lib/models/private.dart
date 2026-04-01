import 'message.dart';

class Private {
  final int id;
  final int user1;
  final int user2;
  final DateTime createdAt;
  final List<Message> messages;
  final Map<int, bool> typing;

  // 🔽 Pagination
  final int page;
  final int limit;
  final bool hasNextPage;
  final bool isLoadingMore;

  Private({
    required this.id,
    required this.user1,
    required this.user2,
    required this.createdAt,
    this.messages = const [],
    this.typing = const {},
    this.page = 1,
    this.limit = 20,
    this.hasNextPage = true,
    this.isLoadingMore = false,
  });

  factory Private.fromMap(Map<String, dynamic> map) {
    return Private(
      id: map['id'],
      user1: map['user1'],
      user2: map['user2'],
      createdAt: DateTime.parse(map['created_at']),
      messages: map['messages'] != null
          ? List<Message>.from(
              (map['messages'] as List).map((x) => Message.fromMap(x)),
            )
          : [],
      typing: map['typing'] != null ? Map<int, bool>.from(map['typing']) : {},
      page: map['page'] ?? 1,
      limit: map['limit'] ?? 20,
      hasNextPage: map['has_next_page'] ?? true,
      isLoadingMore: map['is_loading_more'] ?? false,
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'user1': user1,
      'user2': user2,
      'created_at': createdAt.toIso8601String(),
      'messages': messages.map((m) => m.toMap()).toList(),
      'typing': typing,
      'page': page,
      'limit': limit,
      'has_next_page': hasNextPage,
      'is_loading_more': isLoadingMore,
    };
  }

  Private copyWith({
    int? id,
    int? user1,
    int? user2,
    DateTime? createdAt,
    List<Message>? messages,
    Map<int, bool>? typing,
    int? page,
    int? limit,
    bool? hasNextPage,
    bool? isLoadingMore,
  }) {
    return Private(
      id: id ?? this.id,
      user1: user1 ?? this.user1,
      user2: user2 ?? this.user2,
      createdAt: createdAt ?? this.createdAt,
      messages: messages ?? this.messages,
      typing: typing ?? this.typing,
      page: page ?? this.page,
      limit: limit ?? this.limit,
      hasNextPage: hasNextPage ?? this.hasNextPage,
      isLoadingMore: isLoadingMore ?? this.isLoadingMore,
    );
  }
}
