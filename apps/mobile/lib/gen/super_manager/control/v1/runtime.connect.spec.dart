//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/runtime.proto
//

import "package:connectrpc/connect.dart" as connect;
import "runtime.pb.dart" as super_managercontrolv1runtime;

/// RuntimeService exposes installed runtime status and account telemetry.
abstract final class RuntimeService {
  /// Fully-qualified name of the RuntimeService service.
  static const name = 'super_manager.control.v1.RuntimeService';

  /// GetRuntime returns one installed runtime.
  static const getRuntime = connect.Spec(
    '/$name/GetRuntime',
    connect.StreamType.unary,
    super_managercontrolv1runtime.GetRuntimeRequest.new,
    super_managercontrolv1runtime.GetRuntimeResponse.new,
  );

  /// ListRuntimes returns all installed runtime slots.
  static const listRuntimes = connect.Spec(
    '/$name/ListRuntimes',
    connect.StreamType.unary,
    super_managercontrolv1runtime.ListRuntimesRequest.new,
    super_managercontrolv1runtime.ListRuntimesResponse.new,
  );

  /// GetRuntimeQuota returns the latest account-level quota observation.
  static const getRuntimeQuota = connect.Spec(
    '/$name/GetRuntimeQuota',
    connect.StreamType.unary,
    super_managercontrolv1runtime.GetRuntimeQuotaRequest.new,
    super_managercontrolv1runtime.GetRuntimeQuotaResponse.new,
  );
}
