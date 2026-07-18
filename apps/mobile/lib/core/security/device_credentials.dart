import 'dart:convert';
import 'dart:typed_data';

import 'package:crypto/crypto.dart';

final class PairingOffer {
  static const maxPayloadLength = 256 * 1024;

  final Uri serviceUri;
  final Uri pairingUri;
  final Uint8List secret;
  final Uint8List caCertificatePem;
  final DateTime expireTime;

  const PairingOffer({
    required this.serviceUri,
    required this.pairingUri,
    required this.secret,
    required this.caCertificatePem,
    required this.expireTime,
  });

  factory PairingOffer.parse(String payload, {DateTime? now}) {
    final text = payload.trim();
    if (text.isEmpty || text.length > maxPayloadLength) {
      throw const FormatException('配对内容为空或过大');
    }
    final decoded = jsonDecode(text);
    if (decoded is! Map<String, dynamic> || decoded['version'] != 1) {
      throw const FormatException('不支持的配对内容版本');
    }
    final serviceUri = _parseOrigin(decoded['serviceUrl'], 'serviceUrl');
    final pairingUri = _parseOrigin(decoded['pairingUrl'], 'pairingUrl');
    if (serviceUri.host != pairingUri.host) {
      throw const FormatException('服务和配对地址必须使用同一域名');
    }
    final secret = _decodeBase64Url(decoded['pairingSecret'], 'pairingSecret');
    if (secret.length != 32) {
      throw const FormatException('配对密钥长度无效');
    }
    final caCertificate = _decodeBase64(
      decoded['caCertificatePem'],
      'caCertificatePem',
    );
    if (!utf8
        .decode(caCertificate, allowMalformed: true)
        .contains('BEGIN CERTIFICATE')) {
      throw const FormatException('配对证书格式无效');
    }
    final expectedHash = _requireString(decoded, 'caSha256').toLowerCase();
    if (expectedHash != sha256.convert(caCertificate).toString()) {
      throw const FormatException('配对证书摘要不匹配');
    }
    final expireTime = DateTime.tryParse(
      _requireString(decoded, 'expireTime'),
    )?.toUtc();
    if (expireTime == null ||
        !expireTime.isAfter((now ?? DateTime.now()).toUtc())) {
      throw const FormatException('配对内容已过期');
    }
    return PairingOffer(
      serviceUri: serviceUri,
      pairingUri: pairingUri,
      secret: secret,
      caCertificatePem: caCertificate,
      expireTime: expireTime,
    );
  }
}

final class DeviceCredentials {
  final Uri serviceUri;
  final Uint8List caCertificatePem;
  final Uint8List devicePkcs12;
  final String pkcs12Password;
  final String deviceToken;
  final String deviceUid;
  final String displayName;

  const DeviceCredentials({
    required this.serviceUri,
    required this.caCertificatePem,
    required this.devicePkcs12,
    required this.pkcs12Password,
    required this.deviceToken,
    required this.deviceUid,
    required this.displayName,
  });

  factory DeviceCredentials.fromJson(Map<String, dynamic> json) =>
      DeviceCredentials(
        serviceUri: _parseOrigin(json['serviceUrl'], 'serviceUrl'),
        caCertificatePem: _decodeBase64(
          json['caCertificatePem'],
          'caCertificatePem',
        ),
        devicePkcs12: _decodeBase64(json['devicePkcs12'], 'devicePkcs12'),
        pkcs12Password: _requireString(json, 'pkcs12Password'),
        deviceToken: _requireString(json, 'deviceToken'),
        deviceUid: _requireString(json, 'deviceUid'),
        displayName: _requireString(json, 'displayName'),
      );

  Map<String, dynamic> toJson() => {
    'serviceUrl': serviceUri.toString(),
    'caCertificatePem': base64Encode(caCertificatePem),
    'devicePkcs12': base64Encode(devicePkcs12),
    'pkcs12Password': pkcs12Password,
    'deviceToken': deviceToken,
    'deviceUid': deviceUid,
    'displayName': displayName,
  };
}

Uri _parseOrigin(Object? value, String field) {
  final uri = Uri.tryParse(value is String ? value : '');
  if (uri == null ||
      uri.scheme != 'https' ||
      !uri.hasAuthority ||
      uri.host.isEmpty ||
      uri.userInfo.isNotEmpty ||
      (uri.path.isNotEmpty && uri.path != '/') ||
      uri.hasQuery ||
      uri.hasFragment) {
    throw FormatException('$field 必须是 HTTPS 服务根地址');
  }
  return uri.replace(path: '');
}

String _requireString(Map<String, dynamic> json, String field) {
  final value = json[field];
  if (value is! String || value.isEmpty) {
    throw FormatException('$field 缺失或格式无效');
  }
  return value;
}

Uint8List _decodeBase64(Object? value, String field) {
  try {
    return Uint8List.fromList(base64Decode(value is String ? value : ''));
  } on FormatException {
    throw FormatException('$field 不是有效的 Base64');
  }
}

Uint8List _decodeBase64Url(Object? value, String field) {
  try {
    return Uint8List.fromList(
      base64Url.decode(base64Url.normalize(value is String ? value : '')),
    );
  } on FormatException {
    throw FormatException('$field 不是有效的 Base64URL');
  }
}
