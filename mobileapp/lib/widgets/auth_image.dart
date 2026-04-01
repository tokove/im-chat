import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class AuthImage extends StatelessWidget {
  final String url;
  final double? width;

  const AuthImage(this.url, {this.width, super.key});

  Future<String?> _getToken() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString('access_token');
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<String?>(
      future: _getToken(),
      builder: (context, snapshot) {
        if (!snapshot.hasData) {
          return const SizedBox(
            width: 200,
            height: 200,
            child: Center(child: CircularProgressIndicator()),
          );
        }

        return Image.network(
          url,
          width: width,
          headers: {
            'Authorization': 'Bearer ${snapshot.data}',
            'X-Platform': 'mobile',
          },
        );
      },
    );
  }
}
