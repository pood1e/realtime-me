// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/ingest.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'ingest.pb.dart' as $3;
import 'ingest.pbjson.dart';

export 'ingest.pb.dart';

abstract class EnrollmentServiceBase extends $pb.GeneratedService {
  $async.Future<$3.EnrollDeviceResponse> enrollDevice(
      $pb.ServerContext ctx, $3.EnrollDeviceRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'EnrollDevice':
        return $3.EnrollDeviceRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'EnrollDevice':
        return enrollDevice(ctx, request as $3.EnrollDeviceRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json =>
      EnrollmentServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => EnrollmentServiceBase$messageJson;
}

abstract class IngestServiceBase extends $pb.GeneratedService {
  $async.Future<$3.ReportMobileStatusResponse> reportMobileStatus(
      $pb.ServerContext ctx, $3.ReportMobileStatusRequest request);
  $async.Future<$3.RegisterScrapeTargetsResponse> registerScrapeTargets(
      $pb.ServerContext ctx, $3.RegisterScrapeTargetsRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ReportMobileStatus':
        return $3.ReportMobileStatusRequest();
      case 'RegisterScrapeTargets':
        return $3.RegisterScrapeTargetsRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ReportMobileStatus':
        return reportMobileStatus(ctx, request as $3.ReportMobileStatusRequest);
      case 'RegisterScrapeTargets':
        return registerScrapeTargets(
            ctx, request as $3.RegisterScrapeTargetsRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => IngestServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => IngestServiceBase$messageJson;
}
