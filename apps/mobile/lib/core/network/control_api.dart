import 'dart:async';

import 'package:connectrpc/connect.dart' as connect;

import '../../gen/super_manager/control/v1/device.connect.client.dart';
import '../../gen/super_manager/control/v1/device.pb.dart';
import '../../gen/super_manager/control/v1/execution.connect.client.dart';
import '../../gen/super_manager/control/v1/execution.pb.dart';
import '../../gen/super_manager/control/v1/runtime.connect.client.dart';
import '../../gen/super_manager/control/v1/runtime.pb.dart';
import '../../gen/super_manager/control/v1/terminal.connect.client.dart';
import '../../gen/super_manager/control/v1/terminal.pb.dart';
import '../../gen/super_manager/control/v1/thread.connect.client.dart';
import '../../gen/super_manager/control/v1/thread.pb.dart';
import '../../gen/super_manager/control/v1/workspace.connect.client.dart';
import '../../gen/super_manager/control/v1/workspace.pb.dart';

final class ControlApi {
  static const _timeout = Duration(seconds: 30);

  final RuntimeServiceClient _runtimes;
  final WorkspaceServiceClient _workspaces;
  final ThreadServiceClient _threads;
  final ExecutionServiceClient _executions;
  final TerminalServiceClient _terminals;
  final DeviceServiceClient _devices;

  ControlApi(connect.Transport transport)
    : _runtimes = RuntimeServiceClient(transport),
      _workspaces = WorkspaceServiceClient(transport),
      _threads = ThreadServiceClient(transport),
      _executions = ExecutionServiceClient(transport),
      _terminals = TerminalServiceClient(transport),
      _devices = DeviceServiceClient(transport);

  Future<List<Runtime>> listRuntimes() async {
    final response = await _runtimes
        .listRuntimes(ListRuntimesRequest())
        .timeout(_timeout);
    return List.unmodifiable(response.runtimes);
  }

  Future<QuotaSnapshot> getRuntimeQuota(String runtimeUid) async {
    final response = await _runtimes
        .getRuntimeQuota(GetRuntimeQuotaRequest(runtimeUid: runtimeUid))
        .timeout(_timeout);
    return response.quotaSnapshot;
  }

  Future<List<Workspace>> listWorkspaces() async {
    final response = await _workspaces
        .listWorkspaces(ListWorkspacesRequest(pageSize: 100))
        .timeout(_timeout);
    return List.unmodifiable(response.workspaces);
  }

  Future<Workspace> getWorkspace(String uid) async {
    final response = await _workspaces
        .getWorkspace(GetWorkspaceRequest(uid: uid))
        .timeout(_timeout);
    return response.workspace;
  }

  Future<Workspace> createWorkspace(String displayName, String path) async {
    final response = await _workspaces
        .createWorkspace(
          CreateWorkspaceRequest(
            displayName: displayName.trim(),
            path: path.trim(),
          ),
        )
        .timeout(_timeout);
    return response.workspace;
  }

  Future<void> deleteWorkspace(String uid) async {
    await _workspaces
        .deleteWorkspace(DeleteWorkspaceRequest(uid: uid))
        .timeout(_timeout);
  }

  Future<List<Thread>> listThreads(String workspaceUid) async {
    final response = await _threads
        .listThreads(
          ListThreadsRequest(workspaceUid: workspaceUid, pageSize: 100),
        )
        .timeout(_timeout);
    return List.unmodifiable(response.threads);
  }

  Future<Thread> getThread(String uid) async {
    final response = await _threads
        .getThread(GetThreadRequest(uid: uid))
        .timeout(_timeout);
    return response.thread;
  }

  Future<Thread> createThread({
    required String workspaceUid,
    required String runtimeUid,
    required String displayName,
  }) async {
    final response = await _threads
        .createThread(
          CreateThreadRequest(
            workspaceUid: workspaceUid,
            runtimeUid: runtimeUid,
            displayName: displayName.trim(),
          ),
        )
        .timeout(_timeout);
    return response.thread;
  }

  Future<void> deleteThread(String uid) async {
    await _threads
        .deleteThread(DeleteThreadRequest(uid: uid))
        .timeout(_timeout);
  }

  Future<List<Execution>> listExecutions(String threadUid) async {
    final response = await _executions
        .listExecutions(
          ListExecutionsRequest(threadUid: threadUid, pageSize: 20),
        )
        .timeout(_timeout);
    return List.unmodifiable(response.executions);
  }

  Future<Execution?> cancelActiveExecution(String threadUid) async {
    final executions = await listExecutions(threadUid);
    final active = executions.where(
      (execution) =>
          execution.state == ExecutionState.EXECUTION_STATE_RUNNING ||
          execution.state == ExecutionState.EXECUTION_STATE_INPUT_REQUIRED,
    );
    if (active.isEmpty) {
      return null;
    }
    final response = await _executions
        .cancelExecution(CancelExecutionRequest(uid: active.first.uid))
        .timeout(_timeout);
    return response.execution;
  }

  Future<Execution?> steerActiveExecution(
    String threadUid,
    String instruction,
  ) async {
    final executions = await listExecutions(threadUid);
    final active = executions.where(
      (execution) => execution.state == ExecutionState.EXECUTION_STATE_RUNNING,
    );
    if (active.isEmpty) {
      return null;
    }
    final response = await _executions
        .steerExecution(
          SteerExecutionRequest(
            uid: active.first.uid,
            instruction: instruction.trim(),
          ),
        )
        .timeout(_timeout);
    return response.execution;
  }

  Future<List<TerminalSession>> listTerminalSessions(
    String workspaceUid,
  ) async {
    final response = await _terminals
        .listTerminalSessions(
          ListTerminalSessionsRequest(workspaceUid: workspaceUid),
        )
        .timeout(_timeout);
    return List.unmodifiable(response.terminalSessions);
  }

  Future<TerminalSession> getTerminalSession(String uid) async {
    final response = await _terminals
        .getTerminalSession(GetTerminalSessionRequest(uid: uid))
        .timeout(_timeout);
    return response.terminalSession;
  }

  Future<TerminalSession> createTerminalSession({
    required String workspaceUid,
    required String displayName,
    int columns = 80,
    int rows = 24,
  }) async {
    final response = await _terminals
        .createTerminalSession(
          CreateTerminalSessionRequest(
            workspaceUid: workspaceUid,
            displayName: displayName.trim(),
            columns: columns,
            rows: rows,
          ),
        )
        .timeout(_timeout);
    return response.terminalSession;
  }

  Future<void> deleteTerminalSession(String uid) async {
    await _terminals
        .deleteTerminalSession(DeleteTerminalSessionRequest(uid: uid))
        .timeout(_timeout);
  }

  Future<List<Device>> listDevices() async {
    final response = await _devices
        .listDevices(ListDevicesRequest())
        .timeout(_timeout);
    return List.unmodifiable(response.devices);
  }

  Future<void> revokeDevice(String uid) async {
    await _devices
        .deleteDevice(DeleteDeviceRequest(uid: uid))
        .timeout(_timeout);
  }
}
