//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/device.proto
//

import "package:connectrpc/connect.dart" as connect;
import "device.pb.dart" as super_managercontrolv1device;

/// PairingService exposes the only API available without an existing device certificate.
abstract final class PairingService {
  /// Fully-qualified name of the PairingService service.
  static const name = 'super_manager.control.v1.PairingService';

  /// PairDevice atomically redeems a locally generated one-time secret.
  static const pairDevice = connect.Spec(
    '/$name/PairDevice',
    connect.StreamType.unary,
    super_managercontrolv1device.PairDeviceRequest.new,
    super_managercontrolv1device.PairDeviceResponse.new,
  );
}
/// DeviceService manages paired device revocation.
abstract final class DeviceService {
  /// Fully-qualified name of the DeviceService service.
  static const name = 'super_manager.control.v1.DeviceService';

  /// ListDevices returns paired devices without credentials.
  static const listDevices = connect.Spec(
    '/$name/ListDevices',
    connect.StreamType.unary,
    super_managercontrolv1device.ListDevicesRequest.new,
    super_managercontrolv1device.ListDevicesResponse.new,
  );

  /// DeleteDevice revokes a device immediately through the inner token layer.
  static const deleteDevice = connect.Spec(
    '/$name/DeleteDevice',
    connect.StreamType.unary,
    super_managercontrolv1device.DeleteDeviceRequest.new,
    super_managercontrolv1device.DeleteDeviceResponse.new,
  );
}
