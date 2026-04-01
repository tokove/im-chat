import 'package:dio/dio.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import '../../core/services/dio_service.dart';

class UploadApi {
  final DioService _dioService = DioService.instance;

  Future<String?> uploadAttachment(PlatformFile file, int privateId) async {
    try {
      late MultipartFile multipartFile;

      if (kIsWeb) {
        if (file.bytes == null) return null;
        multipartFile = MultipartFile.fromBytes(
          file.bytes!,
          filename: file.name,
        );
      } else {
        if (file.path == null || file.path!.isEmpty) return null;
        multipartFile = await MultipartFile.fromFile(
          file.path!,
          filename: file.name,
        );
      }

      final formData = FormData.fromMap({
        'file': multipartFile,
      });

      final res = await _dioService.dio.post(
        '/files/$privateId',
        data: formData,
      );

      if (res.data['success'] == true) {
        return res.data['data']?.toString();
      }
      return null;
    } catch (_) {
      return null;
    }
  }
}
