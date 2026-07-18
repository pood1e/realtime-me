//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/execution.proto
//

import "package:connectrpc/connect.dart" as connect;
import "execution.pb.dart" as super_managercontrolv1execution;
import "execution.connect.spec.dart" as specs;

/// ExecutionService observes and controls native provider executions.
extension type ExecutionServiceClient (connect.Transport _transport) {
  /// GetExecution returns one provider execution.
  Future<super_managercontrolv1execution.GetExecutionResponse> getExecution(
    super_managercontrolv1execution.GetExecutionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ExecutionService.getExecution,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListExecutions returns recent executions in one thread.
  Future<super_managercontrolv1execution.ListExecutionsResponse> listExecutions(
    super_managercontrolv1execution.ListExecutionsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ExecutionService.listExecutions,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// CancelExecution cancels one active execution.
  Future<super_managercontrolv1execution.CancelExecutionResponse> cancelExecution(
    super_managercontrolv1execution.CancelExecutionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ExecutionService.cancelExecution,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// SteerExecution submits an instruction when the runtime declares support.
  Future<super_managercontrolv1execution.SteerExecutionResponse> steerExecution(
    super_managercontrolv1execution.SteerExecutionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ExecutionService.steerExecution,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
