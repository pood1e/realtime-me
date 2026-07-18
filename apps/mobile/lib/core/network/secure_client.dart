import 'dart:io';

import 'package:connectrpc/connect.dart' as connect;
import 'package:connectrpc/io.dart' as connect_io;
import 'package:connectrpc/protobuf.dart' as connect_protobuf;
import 'package:connectrpc/protocol/connect.dart' as connect_protocol;

import '../security/device_credentials.dart';

SecurityContext createPairingSecurityContext(PairingOffer offer) {
  return SecurityContext(withTrustedRoots: false)
    ..setTrustedCertificatesBytes(offer.caCertificatePem);
}

SecurityContext createDeviceSecurityContext(DeviceCredentials credentials) {
  return SecurityContext(withTrustedRoots: false)
    ..setTrustedCertificatesBytes(credentials.caCertificatePem)
    ..useCertificateChainBytes(
      credentials.devicePkcs12,
      password: credentials.pkcs12Password,
    )
    ..usePrivateKeyBytes(
      credentials.devicePkcs12,
      password: credentials.pkcs12Password,
    );
}

HttpClient createBoundHttpClient(SecurityContext context) {
  return HttpClient(context: context)
    ..connectionTimeout = const Duration(seconds: 15)
    ..idleTimeout = const Duration(seconds: 30)
    ..maxConnectionsPerHost = 8
    ..userAgent = 'SuperManager/0.1.0';
}

connect.Transport createConnectTransport({
  required Uri baseUri,
  required HttpClient client,
  String? bearerToken,
}) {
  final delegate = connect_io.createHttpClient(client);
  final authenticated = bearerToken == null
      ? delegate
      : (connect.HttpRequest request) {
          request.header['authorization'] = 'Bearer $bearerToken';
          return delegate(request);
        };
  return connect_protocol.Transport(
    baseUrl: baseUri.toString().replaceFirst(RegExp(r'/$'), ''),
    codec: const connect_protobuf.ProtoCodec(),
    httpClient: authenticated,
  );
}
