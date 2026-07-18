//
//  Generated code. Do not modify.
//  source: realtime/me/manager/control/v1/execution.proto
//

import "package:connectrpc/connect.dart" as connect;
import "execution.pb.dart" as realtimememanagercontrolv1execution;

/// ExecutionService observes and controls native provider executions.
abstract final class ExecutionService {
  /// Fully-qualified name of the ExecutionService service.
  static const name = 'realtime.me.manager.control.v1.ExecutionService';

  /// GetExecution returns one provider execution.
  static const getExecution = connect.Spec(
    '/$name/GetExecution',
    connect.StreamType.unary,
    realtimememanagercontrolv1execution.GetExecutionRequest.new,
    realtimememanagercontrolv1execution.GetExecutionResponse.new,
  );

  /// ListExecutions returns recent executions in one thread.
  static const listExecutions = connect.Spec(
    '/$name/ListExecutions',
    connect.StreamType.unary,
    realtimememanagercontrolv1execution.ListExecutionsRequest.new,
    realtimememanagercontrolv1execution.ListExecutionsResponse.new,
  );

  /// CancelExecution cancels one active execution.
  static const cancelExecution = connect.Spec(
    '/$name/CancelExecution',
    connect.StreamType.unary,
    realtimememanagercontrolv1execution.CancelExecutionRequest.new,
    realtimememanagercontrolv1execution.CancelExecutionResponse.new,
  );

  /// SteerExecution submits an instruction when the runtime declares support.
  static const steerExecution = connect.Spec(
    '/$name/SteerExecution',
    connect.StreamType.unary,
    realtimememanagercontrolv1execution.SteerExecutionRequest.new,
    realtimememanagercontrolv1execution.SteerExecutionResponse.new,
  );
}
