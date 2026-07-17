// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/device.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// DeviceStatus describes whether a paired client may authenticate.
class DeviceStatus extends $pb.ProtobufEnum {
  /// The device status was not specified.
  static const DeviceStatus DEVICE_STATUS_UNSPECIFIED =
      DeviceStatus._(0, _omitEnumNames ? '' : 'DEVICE_STATUS_UNSPECIFIED');

  /// The device credentials are active.
  static const DeviceStatus DEVICE_STATUS_ACTIVE =
      DeviceStatus._(1, _omitEnumNames ? '' : 'DEVICE_STATUS_ACTIVE');

  /// The device has been revoked.
  static const DeviceStatus DEVICE_STATUS_REVOKED =
      DeviceStatus._(2, _omitEnumNames ? '' : 'DEVICE_STATUS_REVOKED');

  static const $core.List<DeviceStatus> values = <DeviceStatus>[
    DEVICE_STATUS_UNSPECIFIED,
    DEVICE_STATUS_ACTIVE,
    DEVICE_STATUS_REVOKED,
  ];

  static final $core.List<DeviceStatus?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static DeviceStatus? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DeviceStatus._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
