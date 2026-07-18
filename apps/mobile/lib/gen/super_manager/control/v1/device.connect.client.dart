//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/device.proto
//

import "package:connectrpc/connect.dart" as connect;
import "device.pb.dart" as super_managercontrolv1device;
import "device.connect.spec.dart" as specs;

/// PairingService exposes the only API available without an existing device certificate.
extension type PairingServiceClient (connect.Transport _transport) {
  /// PairDevice atomically redeems a locally generated one-time secret.
  Future<super_managercontrolv1device.PairDeviceResponse> pairDevice(
    super_managercontrolv1device.PairDeviceRequest input, {
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
  Future<super_managercontrolv1device.ListDevicesResponse> listDevices(
    super_managercontrolv1device.ListDevicesRequest input, {
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
  Future<super_managercontrolv1device.DeleteDeviceResponse> deleteDevice(
    super_managercontrolv1device.DeleteDeviceRequest input, {
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
