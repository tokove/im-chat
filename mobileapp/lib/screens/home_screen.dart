import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:timeago/timeago.dart' as timeago;

import '../models/private.dart';
import '../models/user.dart';
import '../notifiers/auth_notifier.dart';
import '../notifiers/clients_notifier.dart';
import '../notifiers/privates_notifier.dart';
import '../notifiers/users_notifier.dart';
import '../notifiers/ws_notifier.dart';

import 'auth_screen.dart';
import 'chat_screen.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  @override
  void initState() {
    super.initState();

    WidgetsBinding.instance.addPostFrameCallback((_) async {
      final authState = ref.read(authProvider);
      final authUser = authState.value;

      if (authUser == null && mounted) {
        _goToAuthScreen();
        return;
      }

      // Initialize WebSocket
      ref
          .read(wsNotifierProvider.notifier)
          .init(
            currentUser: authState,
            clientsNotifier: ref.read(clientsProvider.notifier),
            privatesNotifier: ref.read(privatesProvider.notifier),
          );

      // Fetch private chats
      await ref.read(privatesProvider.notifier).fetchPrivates();
    });
  }

  void _goToAuthScreen() {
    Navigator.pushReplacement(
      context,
      MaterialPageRoute(builder: (_) => const AuthScreen()),
    );
  }

  void _logout() async {
    await ref.read(authProvider.notifier).logout();
    if (mounted) _goToAuthScreen();
  }

  @override
  void dispose() {
    ref.read(wsNotifierProvider.notifier).dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authProvider);

    if (authState.isLoading) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }

    final user = authState.value;
    if (user == null) return const SizedBox.shrink();

    final clients = ref.watch(clientsProvider);
    final privatesState = ref.watch(privatesProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Chats'),
        actions: [
          IconButton(onPressed: _logout, icon: const Icon(Icons.logout)),
        ],
      ),
      body: Column(
        children: [
          _OnlineUsersList(userId: user.id, clients: clients, ref: ref),
          const Divider(height: 1),
          Expanded(
            child: _PrivateChatsList(
              currentUserId: user.id,
              clients: clients,
              privatesState: privatesState,
              ref: ref,
            ),
          ),
        ],
      ),
    );
  }
}

class _OnlineUsersList extends StatelessWidget {
  final int userId;
  final Map<int, dynamic> clients;
  final WidgetRef ref;

  const _OnlineUsersList({
    required this.userId,
    required this.clients,
    required this.ref,
  });

  @override
  Widget build(BuildContext context) {
    if (clients.isEmpty) {
      return const SizedBox(
        height: 96,
        child: Center(child: Text('No users online')),
      );
    }

    return SizedBox(
      height: 96,
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        itemCount: clients.length,
        itemBuilder: (_, i) {
          final u = clients.values.elementAt(i) as User;

          return GestureDetector(
            onTap: () async {
              final private = await ref
                  .read(privatesProvider.notifier)
                  .joinPrivate(otherUserId: u.id, currentUserId: userId);
              if (!context.mounted) return;
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => ChatScreen(
                    privateId: private.id,
                    currentUserId: userId,
                    receiverId: u.id,
                  ),
                ),
              );
            },
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Stack(
                    children: [
                      CircleAvatar(
                        radius: 26,
                        child: Text(
                          u.name.isNotEmpty ? u.name[0].toUpperCase() : '?',
                          style: const TextStyle(fontSize: 20),
                        ),
                      ),
                      const Positioned(
                        bottom: 0,
                        right: 0,
                        child: _OnlineIndicator(),
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  SizedBox(
                    width: 64,
                    child: Text(
                      u.name,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      textAlign: TextAlign.center,
                      style: const TextStyle(fontSize: 12),
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }
}

class _OnlineIndicator extends StatelessWidget {
  const _OnlineIndicator();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 10,
      height: 10,
      decoration: const BoxDecoration(
        color: Colors.green,
        shape: BoxShape.circle,
      ),
    );
  }
}

class _PrivateChatsList extends ConsumerWidget {
  final int currentUserId;
  final Map<int, dynamic> clients;
  final AsyncValue<Map<int, dynamic>> privatesState;
  final WidgetRef ref;

  const _PrivateChatsList({
    required this.currentUserId,
    required this.clients,
    required this.privatesState,
    required this.ref,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return privatesState.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (_, _) => const Center(child: Text('Failed to load chats')),
      data: (map) {
        if (map.isEmpty) return const Center(child: Text('No conversations'));

        final chats = map.values.toList()
          ..sort((a, b) {
            final at = a.messages.isNotEmpty
                ? a.messages.last.createdAt
                : DateTime.fromMillisecondsSinceEpoch(0);
            final bt = b.messages.isNotEmpty
                ? b.messages.last.createdAt
                : DateTime.fromMillisecondsSinceEpoch(0);
            return bt.compareTo(at);
          });

        return ListView.separated(
          itemCount: chats.length,
          separatorBuilder: (_, _) => const Divider(height: 1),
          itemBuilder: (_, i) {
            final p = chats[i] as Private;
            final otherUserId = p.user1 == currentUserId ? p.user2 : p.user1;

            // Only fetch once per user
            ref.read(usersProvider.notifier).getUserById(otherUserId);
            final otherUser = ref.watch(usersProvider)[otherUserId];

            final lastMsg = p.messages.isNotEmpty ? p.messages.last : null;
            final isUnread =
                lastMsg != null &&
                !lastMsg.read &&
                lastMsg.fromId != currentUserId;

            final otherUserDisplayName = otherUser?.name ?? 'User $otherUserId';
            final lastMsgText = lastMsg?.content ?? 'No messages yet';
            final lastMsgTime = lastMsg != null
                ? timeago.format(lastMsg.createdAt)
                : '';

            return ListTile(
              leading: CircleAvatar(
                radius: 24,
                child: Text(
                  otherUserDisplayName.isNotEmpty
                      ? otherUserDisplayName[0].toUpperCase()
                      : '?',
                ),
              ),
              title: Text(
                otherUserDisplayName,
                style: TextStyle(
                  fontWeight: isUnread ? FontWeight.bold : FontWeight.normal,
                ),
              ),
              subtitle: Text(
                lastMsgText,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(fontWeight: isUnread ? FontWeight.w900 : null),
              ),
              trailing: Text(
                lastMsgTime,
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: isUnread ? FontWeight.bold : FontWeight.normal,
                  color: Colors.grey[600],
                ),
              ),
              onTap: () {
                if (!context.mounted) return;
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => ChatScreen(
                      privateId: p.id,
                      currentUserId: currentUserId,
                      receiverId: otherUserId,
                    ),
                  ),
                );
              },
            );
          },
        );
      },
    );
  }
}
