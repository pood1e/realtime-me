//
//  Generated code. Do not modify.
//  source: realtime/me/manager/control/v1/device.proto
//

import "package:connectrpc/connect.dart" as connect;
import "device.pb.dart" as realtimememanagercontrolv1device;

/// PairingService exposes the only API available without an existing device certificate.
abstract final class PairingService {
  /// Fully-qualified name of the PairingService service.
  static const name = 'realtime.me.manager.control.v1.PairingService';

  /// PairDevice atomically redeems a locally generated one-time secret.
  static const pairDevice = connect.Spec(
    '/$name/PairDevice',
    connect.StreamType.unary,
    realtimememanagercontrolv1device.PairDeviceRequest.new,
    realtimememanagercontrolv1device.PairDeviceResponse.new,
  );
}
/// DeviceService manages paired device revocation.
abstract final class DeviceService {
  /// Fully-qualified name of the DeviceService service.
  static const name = 'realtime.me.manager.control.v1.DeviceService';

  /// ListDevices returns paired devices without credentials.
  static const listDevices = connect.Spec(
    '/$name/ListDevices',
    connect.StreamType.unary,
    realtimememanagercontrolv1device.ListDevicesRequest.new,
    realtimememanagercontrolv1device.ListDevicesResponse.new,
  );

  /// DeleteDevice revokes a device immediately through the inner token layer.
  static const deleteDevice = connect.Spec(
    '/$name/DeleteDevice',
    connect.StreamType.unary,
    realtimememanagercontrolv1device.DeleteDeviceRequest.new,
    realtimememanagercontrolv1device.DeleteDeviceResponse.new,
  );
}
