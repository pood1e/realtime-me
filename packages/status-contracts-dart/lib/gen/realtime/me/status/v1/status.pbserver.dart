// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/status.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'status.pb.dart' as $3;
import 'status.pbjson.dart';

export 'status.pb.dart';

abstract class StatusServiceBase extends $pb.GeneratedService {
  $async.Future<$3.GetPublicStatusResponse> getPublicStatus(
      $pb.ServerContext ctx, $3.GetPublicStatusRequest request);
  $async.Future<$3.GetInternalStatusResponse> getInternalStatus(
      $pb.ServerContext ctx, $3.GetInternalStatusRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'GetPublicStatus':
        return $3.GetPublicStatusRequest();
      case 'GetInternalStatus':
        return $3.GetInternalStatusRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'GetPublicStatus':
        return getPublicStatus(ctx, request as $3.GetPublicStatusRequest);
      case 'GetInternalStatus':
        return getInternalStatus(ctx, request as $3.GetInternalStatusRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => StatusServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => StatusServiceBase$messageJson;
}
