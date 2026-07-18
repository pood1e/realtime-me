//
//  Generated code. Do not modify.
//  source: realtime/me/manager/control/v1/workspace.proto
//

import "package:connectrpc/connect.dart" as connect;
import "workspace.pb.dart" as realtimememanagercontrolv1workspace;

/// WorkspaceService manages approved server-local project directories.
abstract final class WorkspaceService {
  /// Fully-qualified name of the WorkspaceService service.
  static const name = 'realtime.me.manager.control.v1.WorkspaceService';

  /// CreateWorkspace registers a path contained by an allowed workspace root.
  static const createWorkspace = connect.Spec(
    '/$name/CreateWorkspace',
    connect.StreamType.unary,
    realtimememanagercontrolv1workspace.CreateWorkspaceRequest.new,
    realtimememanagercontrolv1workspace.CreateWorkspaceResponse.new,
  );

  /// GetWorkspace returns one registered workspace.
  static const getWorkspace = connect.Spec(
    '/$name/GetWorkspace',
    connect.StreamType.unary,
    realtimememanagercontrolv1workspace.GetWorkspaceRequest.new,
    realtimememanagercontrolv1workspace.GetWorkspaceResponse.new,
  );

  /// ListWorkspaces returns registered workspaces.
  static const listWorkspaces = connect.Spec(
    '/$name/ListWorkspaces',
    connect.StreamType.unary,
    realtimememanagercontrolv1workspace.ListWorkspacesRequest.new,
    realtimememanagercontrolv1workspace.ListWorkspacesResponse.new,
  );

  /// DeleteWorkspace removes only the registration record.
  static const deleteWorkspace = connect.Spec(
    '/$name/DeleteWorkspace',
    connect.StreamType.unary,
    realtimememanagercontrolv1workspace.DeleteWorkspaceRequest.new,
    realtimememanagercontrolv1workspace.DeleteWorkspaceResponse.new,
  );
}
