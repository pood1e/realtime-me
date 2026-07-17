// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/thread.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'thread.pb.dart' as $1;
import 'thread.pbjson.dart';

export 'thread.pb.dart';

abstract class ThreadServiceBase extends $pb.GeneratedService {
  $async.Future<$1.CreateThreadResponse> createThread(
      $pb.ServerContext ctx, $1.CreateThreadRequest request);
  $async.Future<$1.GetThreadResponse> getThread(
      $pb.ServerContext ctx, $1.GetThreadRequest request);
  $async.Future<$1.ListThreadsResponse> listThreads(
      $pb.ServerContext ctx, $1.ListThreadsRequest request);
  $async.Future<$1.DeleteThreadResponse> deleteThread(
      $pb.ServerContext ctx, $1.DeleteThreadRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'CreateThread':
        return $1.CreateThreadRequest();
      case 'GetThread':
        return $1.GetThreadRequest();
      case 'ListThreads':
        return $1.ListThreadsRequest();
      case 'DeleteThread':
        return $1.DeleteThreadRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'CreateThread':
        return createThread(ctx, request as $1.CreateThreadRequest);
      case 'GetThread':
        return getThread(ctx, request as $1.GetThreadRequest);
      case 'ListThreads':
        return listThreads(ctx, request as $1.ListThreadsRequest);
      case 'DeleteThread':
        return deleteThread(ctx, request as $1.DeleteThreadRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => ThreadServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => ThreadServiceBase$messageJson;
}
