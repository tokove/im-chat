import '../core/consts/base_url.dart'; // make sure baseUrlHTTP is imported

class Message {
  final int id;
  final int fromId;
  final int privateId;
  final String messageType;
  final String content;
  final bool delivered;
  final bool read;
  final DateTime createdAt;

  Message({
    required this.id,
    required this.fromId,
    required this.privateId,
    required this.messageType,
    required this.content,
    required this.delivered,
    required this.read,
    required this.createdAt,
  });

  factory Message.fromMap(Map<String, dynamic> map) {
    String content = map['content'] ?? '';

    if ((map['message_type'] ?? '') == 'attachment' &&
        !content.startsWith('http')) {
      content = '$baseUrlHTTP$content';
    }

    return Message(
      id: map['id'],
      fromId: map['from_id'],
      privateId: map['private_id'],
      messageType: map['message_type'] ?? 'text',
      content: content,
      delivered: map['delivered'] ?? false,
      read: map['read'] ?? false,
      createdAt: DateTime.parse(map['created_at']),
    );
  }

  Map<String, dynamic> toMap() {
    return {
      'id': id,
      'from_id': fromId,
      'private_id': privateId,
      'message_type': messageType,
      'content': content,
      'delivered': delivered,
      'read': read,
      'created_at': createdAt.toIso8601String(),
    };
  }

  Message copyWith({
    int? id,
    int? fromId,
    int? privateId,
    String? messageType,
    String? content,
    bool? delivered,
    bool? read,
    DateTime? createdAt,
  }) {
    return Message(
      id: id ?? this.id,
      fromId: fromId ?? this.fromId,
      privateId: privateId ?? this.privateId,
      messageType: messageType ?? this.messageType,
      content: content ?? this.content,
      delivered: delivered ?? this.delivered,
      read: read ?? this.read,
      createdAt: createdAt ?? this.createdAt,
    );
  }
}
