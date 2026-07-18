// This is a generated file - do not edit.
//
// Generated from realtime/me/auth/v1/permission.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use permissionDescriptor instead')
const Permission$json = {
  '1': 'Permission',
  '2': [
    {'1': 'PERMISSION_UNSPECIFIED', '2': 0},
    {'1': 'PERMISSION_STATUS_INTERNAL_READ', '2': 1},
    {'1': 'PERMISSION_LIBRARY_MANAGE', '2': 2},
    {'1': 'PERMISSION_MANAGER_CONTROL', '2': 3},
  ],
};

/// Descriptor for `Permission`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List permissionDescriptor = $convert.base64Decode(
    'CgpQZXJtaXNzaW9uEhoKFlBFUk1JU1NJT05fVU5TUEVDSUZJRUQQABIjCh9QRVJNSVNTSU9OX1'
    'NUQVRVU19JTlRFUk5BTF9SRUFEEAESHQoZUEVSTUlTU0lPTl9MSUJSQVJZX01BTkFHRRACEh4K'
    'GlBFUk1JU1NJT05fTUFOQUdFUl9DT05UUk9MEAM=');
