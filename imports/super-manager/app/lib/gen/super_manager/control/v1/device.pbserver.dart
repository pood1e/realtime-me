// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/device.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'device.pb.dart' as $1;
import 'device.pbjson.dart';

export 'device.pb.dart';

abstract class PairingServiceBase extends $pb.GeneratedService {
  $async.Future<$1.PairDeviceResponse> pairDevice(
      $pb.ServerContext ctx, $1.PairDeviceRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'PairDevice':
        return $1.PairDeviceRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'PairDevice':
        return pairDevice(ctx, request as $1.PairDeviceRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => PairingServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => PairingServiceBase$messageJson;
}

abstract class DeviceServiceBase extends $pb.GeneratedService {
  $async.Future<$1.ListDevicesResponse> listDevices(
      $pb.ServerContext ctx, $1.ListDevicesRequest request);
  $async.Future<$1.DeleteDeviceResponse> deleteDevice(
      $pb.ServerContext ctx, $1.DeleteDeviceRequest request);

  $pb.GeneratedMessage createRequest($core.String methodName) {
    switch (methodName) {
      case 'ListDevices':
        return $1.ListDevicesRequest();
      case 'DeleteDevice':
        return $1.DeleteDeviceRequest();
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $async.Future<$pb.GeneratedMessage> handleCall($pb.ServerContext ctx,
      $core.String methodName, $pb.GeneratedMessage request) {
    switch (methodName) {
      case 'ListDevices':
        return listDevices(ctx, request as $1.ListDevicesRequest);
      case 'DeleteDevice':
        return deleteDevice(ctx, request as $1.DeleteDeviceRequest);
      default:
        throw $core.ArgumentError('Unknown method: $methodName');
    }
  }

  $core.Map<$core.String, $core.dynamic> get $json => DeviceServiceBase$json;
  $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
      get $messageJson => DeviceServiceBase$messageJson;
}
