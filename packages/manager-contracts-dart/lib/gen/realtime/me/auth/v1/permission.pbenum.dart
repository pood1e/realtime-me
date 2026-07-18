// This is a generated file - do not edit.
//
// Generated from realtime/me/auth/v1/permission.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// Permission identifies one owner capability enforced by a bounded context.
class Permission extends $pb.ProtobufEnum {
  /// The permission was not specified.
  static const Permission PERMISSION_UNSPECIFIED =
      Permission._(0, _omitEnumNames ? '' : 'PERMISSION_UNSPECIFIED');

  /// The caller may read internal status and metrics.
  static const Permission PERMISSION_STATUS_INTERNAL_READ =
      Permission._(1, _omitEnumNames ? '' : 'PERMISSION_STATUS_INTERNAL_READ');

  /// The caller may manage private Library resources.
  static const Permission PERMISSION_LIBRARY_MANAGE =
      Permission._(2, _omitEnumNames ? '' : 'PERMISSION_LIBRARY_MANAGE');

  /// The caller may observe and control Manager resources.
  static const Permission PERMISSION_MANAGER_CONTROL =
      Permission._(3, _omitEnumNames ? '' : 'PERMISSION_MANAGER_CONTROL');

  static const $core.List<Permission> values = <Permission>[
    PERMISSION_UNSPECIFIED,
    PERMISSION_STATUS_INTERNAL_READ,
    PERMISSION_LIBRARY_MANAGE,
    PERMISSION_MANAGER_CONTROL,
  ];

  static final $core.List<Permission?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static Permission? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const Permission._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
