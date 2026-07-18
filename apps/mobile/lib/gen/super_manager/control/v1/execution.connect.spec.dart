//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/execution.proto
//

import "package:connectrpc/connect.dart" as connect;
import "execution.pb.dart" as super_managercontrolv1execution;

/// ExecutionService observes and controls native provider executions.
abstract final class ExecutionService {
  /// Fully-qualified name of the ExecutionService service.
  static const name = 'super_manager.control.v1.ExecutionService';

  /// GetExecution returns one provider execution.
  static const getExecution = connect.Spec(
    '/$name/GetExecution',
    connect.StreamType.unary,
    super_managercontrolv1execution.GetExecutionRequest.new,
    super_managercontrolv1execution.GetExecutionResponse.new,
  );

  /// ListExecutions returns recent executions in one thread.
  static const listExecutions = connect.Spec(
    '/$name/ListExecutions',
    connect.StreamType.unary,
    super_managercontrolv1execution.ListExecutionsRequest.new,
    super_managercontrolv1execution.ListExecutionsResponse.new,
  );

  /// CancelExecution cancels one active execution.
  static const cancelExecution = connect.Spec(
    '/$name/CancelExecution',
    connect.StreamType.unary,
    super_managercontrolv1execution.CancelExecutionRequest.new,
    super_managercontrolv1execution.CancelExecutionResponse.new,
  );

  /// SteerExecution submits an instruction when the runtime declares support.
  static const steerExecution = connect.Spec(
    '/$name/SteerExecution',
    connect.StreamType.unary,
    super_managercontrolv1execution.SteerExecutionRequest.new,
    super_managercontrolv1execution.SteerExecutionResponse.new,
  );
}
