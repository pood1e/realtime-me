import 'dart:io';

import '../security/device_credentials.dart';
import 'agui_transport.dart';
import 'control_api.dart';
import 'secure_client.dart';
import 'terminal_socket.dart';

final class RemoteSession {
  final DeviceCredentials credentials;
  final SecurityContext _securityContext;
  final HttpClient _httpClient;
  late final ControlApi control;
  late final AguiTransport agui;

  factory RemoteSession(DeviceCredentials credentials) {
    final securityContext = createDeviceSecurityContext(credentials);
    return RemoteSession._(
      credentials,
      securityContext,
      createBoundHttpClient(securityContext),
    );
  }

  RemoteSession._(this.credentials, this._securityContext, this._httpClient) {
    control = ControlApi(
      createConnectTransport(
        baseUri: credentials.serviceUri,
        client: _httpClient,
        bearerToken: credentials.deviceToken,
      ),
    );
    agui = AguiTransport(
      baseUri: credentials.serviceUri,
      token: credentials.deviceToken,
      client: _httpClient,
    );
  }

  Future<TerminalSocket> connectTerminal(String terminalUid) =>
      TerminalSocket.connect(
        serviceUri: credentials.serviceUri,
        token: credentials.deviceToken,
        terminalUid: terminalUid,
        httpClient: createBoundHttpClient(_securityContext),
      );

  void close() => _httpClient.close(force: true);
}
