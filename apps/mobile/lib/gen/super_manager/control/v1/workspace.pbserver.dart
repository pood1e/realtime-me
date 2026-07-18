// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/workspace.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'workspace.pb.dart' as $1;
import 'workspace.pbjson.dart';

export 'workspace.pb.dart';

abstract class WorkspaceServiceBase extends $pb.GeneratedService {
  $async.Future<$1.CreateWorkspaceResponse> createWorkspace(
      $pb.ServerContext ctx, $1.CreateWorkspaceRequest request);
  $async.Future<$1.GetWorkspaceResponse> getWorkspace(
      $pb.ServerContext ctx, $1.GetWorkspaceRequest request);
  $async.Future<$1.ListWorkspacesResponse> listWorkspaces(
      $pb.ServerContext ctx, $1.ListWorkspacesRequest request);
  $async.Future<$1.DeleteWorkspaceResponse> deleteWorkspace(
      $pb.ServerContext ctx, $1.DeleteWorkspaceRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'CreateWorkspace':
        return $1.CreateWorkspaceRequest();
      case 'GetWorkspace':
        return $1.GetWorkspaceRequest();
      case 'ListWorkspaces':
        return $1.ListWorkspacesRequest();
      case 'DeleteWorkspace':
        return $1.DeleteWorkspaceRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'CreateWorkspace':
        return createWorkspace(ctx, request as $1.CreateWorkspaceRequest);
      case 'GetWorkspace':
        return getWorkspace(ctx, request as $1.GetWorkspaceRequest);
      case 'ListWorkspaces':
        return listWorkspaces(ctx, request as $1.ListWorkspacesRequest);
      case 'DeleteWorkspace':
        return deleteWorkspace(ctx, request as $1.DeleteWorkspaceRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => WorkspaceServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => WorkspaceServiceBase$messageJson;
}
