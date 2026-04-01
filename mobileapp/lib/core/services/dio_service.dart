import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../../core/consts/base_url.dart';
import '../../notifiers/auth_notifier.dart';

class DioService {
  static final DioService instance = DioService._internal();
  static late WidgetRef _ref;
  static late SharedPreferences _prefs;

  static Future<void> init(WidgetRef ref) async {
    _ref = ref;
    _prefs = await SharedPreferences.getInstance();
  }

  late final Dio dio;

  DioService._internal() {
    dio = Dio(
      BaseOptions(
        baseUrl: baseUrlHTTP,
        headers: {'Content-Type': 'application/json', 'X-Platform': 'mobile'},
      ),
    );

    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final token = _prefs.getString('access_token');
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }

          if (options.data is! FormData) {
            options.headers['Content-Type'] = 'application/json';
          }

          handler.next(options);
        },
        onError: (error, handler) async {
          if (error.response?.statusCode == 401) {
            final authNotifier = _ref.read(authProvider.notifier);
            final refreshed = await authNotifier.refreshSession();

            if (refreshed) {
              final token = _prefs.getString('access_token');
              if (token != null) {
                error.requestOptions.headers['Authorization'] = 'Bearer $token';
                final retry = await dio.fetch(error.requestOptions);
                return handler.resolve(retry);
              }
            }
          }

          handler.next(error);
        },
      ),
    );
  }

  static Dio get client => instance.dio;
}
