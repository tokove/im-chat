import 'package:file_picker/file_picker.dart';

Future<List<PlatformFile>> pickFiles() async {
  final result = await FilePicker.platform.pickFiles(
    allowMultiple: true,
    type: FileType.image,
    withData: true,
  );

  if (result == null || result.files.isEmpty) return [];
  return result.files;
}