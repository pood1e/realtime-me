//
//  Generated code. Do not modify.
//  source: realtime/me/manager/control/v1/device.proto
//

import "package:connectrpc/connect.dart" as connect;
import "device.pb.dart" as realtimememanagercontrolv1device;
import "device.connect.spec.dart" as specs;

/// PairingService exposes the only API available without an existing device certificate.
extension type PairingServiceClient (connect.Transport _transport) {
  /// PairDevice atomically redeems a locally generated one-time secret.
  Future<realtimememanagercontrolv1device.PairDeviceResponse> pairDevice(
    realtimememanagercontrolv1device.PairDeviceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.PairingService.pairDevice,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
/// DeviceService manages paired device revocation.
extension type DeviceServiceClient (connect.Transport _transport) {
  /// ListDevices returns paired devices without credentials.
  Future<realtimememanagercontrolv1device.ListDevicesResponse> listDevices(
    realtimememanagercontrolv1device.ListDevicesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DeviceService.listDevices,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteDevice revokes a device immediately through the inner token layer.
  Future<realtimememanagercontrolv1device.DeleteDeviceResponse> deleteDevice(
    realtimememanagercontrolv1device.DeleteDeviceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.DeviceService.deleteDevice,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
