// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/runtime.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'runtime.pb.dart' as $1;
import 'runtime.pbjson.dart';

export 'runtime.pb.dart';

abstract class RuntimeServiceBase extends $pb.GeneratedService {
  $async.Future<$1.GetRuntimeResponse> getRuntime(
      $pb.ServerContext ctx, $1.GetRuntimeRequest request);
  $async.Future<$1.ListRuntimesResponse> listRuntimes(
      $pb.ServerContext ctx, $1.ListRuntimesRequest request);
  $async.Future<$1.GetRuntimeQuotaResponse> getRuntimeQuota(
      $pb.ServerContext ctx, $1.GetRuntimeQuotaRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'GetRuntime':
        return $1.GetRuntimeRequest();
      case 'ListRuntimes':
        return $1.ListRuntimesRequest();
      case 'GetRuntimeQuota':
        return $1.GetRuntimeQuotaRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'GetRuntime':
        return getRuntime(ctx, request as $1.GetRuntimeRequest);
      case 'ListRuntimes':
        return listRuntimes(ctx, request as $1.ListRuntimesRequest);
      case 'GetRuntimeQuota':
        return getRuntimeQuota(ctx, request as $1.GetRuntimeQuotaRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => RuntimeServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => RuntimeServiceBase$messageJson;
}
