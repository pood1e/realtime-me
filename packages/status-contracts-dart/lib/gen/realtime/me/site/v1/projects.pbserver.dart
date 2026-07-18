// This is a generated file - do not edit.
//
// Generated from realtime/me/site/v1/projects.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'projects.pb.dart' as $1;
import 'projects.pbjson.dart';

export 'projects.pb.dart';

abstract class ProjectsServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListProjectsResponse> listProjects(
      $pb.ServerContext ctx, $1.ListProjectsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListProjects':
        return $1.ListProjectsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListProjects':
        return listProjects(ctx, request as $1.ListProjectsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => ProjectsServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => ProjectsServiceBase$messageJson;
}
