// This is a generated file - do not edit.
//
// Generated from realtime/me/manager/control/v1/terminal.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'terminal.pb.dart' as $1;
import 'terminal.pbjson.dart';

export 'terminal.pb.dart';

abstract class TerminalServiceBase extends $pb.GeneratedService {
  $async.Future<$1.CreateTerminalSessionResponse> createTerminalSession(
      $pb.ServerContext ctx, $1.CreateTerminalSessionRequest request);
  $async.Future<$1.GetTerminalSessionResponse> getTerminalSession(
      $pb.ServerContext ctx, $1.GetTerminalSessionRequest request);
  $async.Future<$1.ListTerminalSessionsResponse> listTerminalSessions(
      $pb.ServerContext ctx, $1.ListTerminalSessionsRequest request);
  $async.Future<$1.DeleteTerminalSessionResponse> deleteTerminalSession(
      $pb.ServerContext ctx, $1.DeleteTerminalSessionRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'CreateTerminalSession':
        return $1.CreateTerminalSessionRequest();
      case 'GetTerminalSession':
        return $1.GetTerminalSessionRequest();
      case 'ListTerminalSessions':
        return $1.ListTerminalSessionsRequest();
      case 'DeleteTerminalSession':
        return $1.DeleteTerminalSessionRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'CreateTerminalSession':
        return createTerminalSession(
            ctx, request as $1.CreateTerminalSessionRequest);
      case 'GetTerminalSession':
        return getTerminalSession(ctx, request as $1.GetTerminalSessionRequest);
      case 'ListTerminalSessions':
        return listTerminalSessions(
            ctx, request as $1.ListTerminalSessionsRequest);
      case 'DeleteTerminalSession':
        return deleteTerminalSession(
            ctx, request as $1.DeleteTerminalSessionRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => TerminalServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => TerminalServiceBase$messageJson;
}
