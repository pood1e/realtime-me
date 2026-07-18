//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/workspace.proto
//

import "package:connectrpc/connect.dart" as connect;
import "workspace.pb.dart" as super_managercontrolv1workspace;

/// WorkspaceService manages approved server-local project directories.
abstract final class WorkspaceService {
  /// Fully-qualified name of the WorkspaceService service.
  static const name = 'super_manager.control.v1.WorkspaceService';

  /// CreateWorkspace registers a path contained by an allowed workspace root.
  static const createWorkspace = connect.Spec(
    '/$name/CreateWorkspace',
    connect.StreamType.unary,
    super_managercontrolv1workspace.CreateWorkspaceRequest.new,
    super_managercontrolv1workspace.CreateWorkspaceResponse.new,
  );

  /// GetWorkspace returns one registered workspace.
  static const getWorkspace = connect.Spec(
    '/$name/GetWorkspace',
    connect.StreamType.unary,
    super_managercontrolv1workspace.GetWorkspaceRequest.new,
    super_managercontrolv1workspace.GetWorkspaceResponse.new,
  );

  /// ListWorkspaces returns registered workspaces.
  static const listWorkspaces = connect.Spec(
    '/$name/ListWorkspaces',
    connect.StreamType.unary,
    super_managercontrolv1workspace.ListWorkspacesRequest.new,
    super_managercontrolv1workspace.ListWorkspacesResponse.new,
  );

  /// DeleteWorkspace removes only the registration record.
  static const deleteWorkspace = connect.Spec(
    '/$name/DeleteWorkspace',
    connect.StreamType.unary,
    super_managercontrolv1workspace.DeleteWorkspaceRequest.new,
    super_managercontrolv1workspace.DeleteWorkspaceResponse.new,
  );
}
