// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/execution.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'execution.pb.dart' as $1;
import 'execution.pbjson.dart';

export 'execution.pb.dart';

abstract class ExecutionServiceBase extends $pb.GeneratedService {
  $async.Future<$1.GetExecutionResponse> getExecution(
      $pb.ServerContext ctx, $1.GetExecutionRequest request);
  $async.Future<$1.ListExecutionsResponse> listExecutions(
      $pb.ServerContext ctx, $1.ListExecutionsRequest request);
  $async.Future<$1.CancelExecutionResponse> cancelExecution(
      $pb.ServerContext ctx, $1.CancelExecutionRequest request);
  $async.Future<$1.SteerExecutionResponse> steerExecution(
      $pb.ServerContext ctx, $1.SteerExecutionRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'GetExecution':
        return $1.GetExecutionRequest();
      case 'ListExecutions':
        return $1.ListExecutionsRequest();
      case 'CancelExecution':
        return $1.CancelExecutionRequest();
      case 'SteerExecution':
        return $1.SteerExecutionRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'GetExecution':
        return getExecution(ctx, request as $1.GetExecutionRequest);
      case 'ListExecutions':
        return listExecutions(ctx, request as $1.ListExecutionsRequest);
      case 'CancelExecution':
        return cancelExecution(ctx, request as $1.CancelExecutionRequest);
      case 'SteerExecution':
        return steerExecution(ctx, request as $1.SteerExecutionRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => ExecutionServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => ExecutionServiceBase$messageJson;
}
