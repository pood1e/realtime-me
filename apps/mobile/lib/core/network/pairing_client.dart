import 'dart:io';
import 'dart:typed_data';

import 'package:crypto/crypto.dart';

import 'package:realtime_me_manager_contracts/gen/realtime/me/manager/control/v1/device.connect.client.dart';
import 'package:realtime_me_manager_contracts/gen/realtime/me/manager/control/v1/device.pb.dart';
import '../security/device_credentials.dart';
import 'secure_client.dart';

final class PairingClient {
  static const _timeout = Duration(seconds: 30);

  Future<DeviceCredentials> pair(PairingOffer offer, String displayName) async {
    if (!offer.expireTime.isAfter(DateTime.now().toUtc())) {
      throw const FormatException('配对内容已过期');
    }
    final normalizedName = displayName.trim();
    if (normalizedName.isEmpty || normalizedName.length > 128) {
      throw const FormatException('设备名称需为 1–128 个字符');
    }
    final httpClient = createBoundHttpClient(
      createPairingSecurityContext(offer),
    );
    try {
      final transport = createConnectTransport(
        baseUri: offer.pairingUri,
        client: httpClient,
      );
      final response = await PairingServiceClient(transport)
          .pairDevice(
            PairDeviceRequest(
              pairingSecret: offer.secret,
              displayName: normalizedName,
            ),
          )
          .timeout(_timeout);
      final responseCa = Uint8List.fromList(response.caCertificatePem);
      if (sha256.convert(responseCa) !=
          sha256.convert(offer.caCertificatePem)) {
        throw HandshakeException('配对响应返回了不同的证书颁发机构');
      }
      if (!response.hasDevice() ||
          response.devicePkcs12.isEmpty ||
          response.pkcs12Password.isEmpty ||
          response.deviceToken.isEmpty) {
        throw const FormatException('配对响应缺少设备凭据');
      }
      return DeviceCredentials(
        serviceUri: offer.serviceUri,
        caCertificatePem: responseCa,
        devicePkcs12: Uint8List.fromList(response.devicePkcs12),
        pkcs12Password: response.pkcs12Password,
        deviceToken: response.deviceToken,
        deviceUid: response.device.uid,
        displayName: response.device.displayName,
      );
    } finally {
      httpClient.close(force: true);
    }
  }
}
