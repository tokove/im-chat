// import 'package:flutter_secure_storage/flutter_secure_storage.dart';

// class SecureStorageService {
//   static final SecureStorageService instance = SecureStorageService._internal();
//   SecureStorageService._internal();

//   // final FlutterSecureStorage _storage = const FlutterSecureStorage();

//   final _storage = FlutterSecureStorage();

//   Future<void> writeToken(String key, String value) async {
//     await _storage.write(key: key, value: value);
//   }

//   Future<String?> readToken(String key) async {
//     return await _storage.read(key: key);
//   }

//   Future<void> deleteToken(String key) async {
//     await _storage.delete(key: key);
//   }
// }
