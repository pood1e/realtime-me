//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/runtime.proto
//

import "package:connectrpc/connect.dart" as connect;
import "runtime.pb.dart" as super_managercontrolv1runtime;
import "runtime.connect.spec.dart" as specs;

/// RuntimeService exposes installed runtime status and account telemetry.
extension type RuntimeServiceClient (connect.Transport _transport) {
  /// GetRuntime returns one installed runtime.
  Future<super_managercontrolv1runtime.GetRuntimeResponse> getRuntime(
    super_managercontrolv1runtime.GetRuntimeRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RuntimeService.getRuntime,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListRuntimes returns all installed runtime slots.
  Future<super_managercontrolv1runtime.ListRuntimesResponse> listRuntimes(
    super_managercontrolv1runtime.ListRuntimesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RuntimeService.listRuntimes,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetRuntimeQuota returns the latest account-level quota observation.
  Future<super_managercontrolv1runtime.GetRuntimeQuotaResponse> getRuntimeQuota(
    super_managercontrolv1runtime.GetRuntimeQuotaRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.RuntimeService.getRuntimeQuota,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
