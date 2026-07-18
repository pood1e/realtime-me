//
//  Generated code. Do not modify.
//  source: realtime/me/manager/control/v1/workspace.proto
//

import "package:connectrpc/connect.dart" as connect;
import "workspace.pb.dart" as realtimememanagercontrolv1workspace;
import "workspace.connect.spec.dart" as specs;

/// WorkspaceService manages approved server-local project directories.
extension type WorkspaceServiceClient (connect.Transport _transport) {
  /// CreateWorkspace registers a path contained by an allowed workspace root.
  Future<realtimememanagercontrolv1workspace.CreateWorkspaceResponse> createWorkspace(
    realtimememanagercontrolv1workspace.CreateWorkspaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.createWorkspace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetWorkspace returns one registered workspace.
  Future<realtimememanagercontrolv1workspace.GetWorkspaceResponse> getWorkspace(
    realtimememanagercontrolv1workspace.GetWorkspaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.getWorkspace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListWorkspaces returns registered workspaces.
  Future<realtimememanagercontrolv1workspace.ListWorkspacesResponse> listWorkspaces(
    realtimememanagercontrolv1workspace.ListWorkspacesRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.listWorkspaces,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteWorkspace removes only the registration record.
  Future<realtimememanagercontrolv1workspace.DeleteWorkspaceResponse> deleteWorkspace(
    realtimememanagercontrolv1workspace.DeleteWorkspaceRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.WorkspaceService.deleteWorkspace,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
