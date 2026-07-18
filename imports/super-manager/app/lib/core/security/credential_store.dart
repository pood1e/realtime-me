import 'dart:convert';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'device_credentials.dart';

final class CredentialStore {
  static const _key = 'device_credentials_v1';
  static const _storage = FlutterSecureStorage(
    aOptions: AndroidOptions(
      storageNamespace: 'super_manager',
      resetOnError: false,
      migrateWithBackup: false,
    ),
  );

  Future<DeviceCredentials?> read() async {
    final encoded = await _storage.read(key: _key);
    if (encoded == null) {
      return null;
    }
    final value = jsonDecode(encoded);
    if (value is! Map<String, dynamic>) {
      throw const FormatException('本地设备凭据格式无效');
    }
    return DeviceCredentials.fromJson(value);
  }

  Future<void> write(DeviceCredentials credentials) =>
      _storage.write(key: _key, value: jsonEncode(credentials.toJson()));

  Future<void> clear() => _storage.delete(key: _key);
}
