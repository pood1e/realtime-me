// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/watch.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// ChargeState describes whether the watch is charging.
class ChargeState extends $pb.ProtobufEnum {
  /// Charge state is not known.
  static const ChargeState CHARGE_STATE_UNSPECIFIED =
      ChargeState._(0, _omitEnumNames ? '' : 'CHARGE_STATE_UNSPECIFIED');

  /// The watch is not charging.
  static const ChargeState CHARGE_STATE_NOT_CHARGING =
      ChargeState._(1, _omitEnumNames ? '' : 'CHARGE_STATE_NOT_CHARGING');

  /// The watch is charging or full while plugged in.
  static const ChargeState CHARGE_STATE_CHARGING =
      ChargeState._(2, _omitEnumNames ? '' : 'CHARGE_STATE_CHARGING');

  static const $core.List<ChargeState> values = <ChargeState>[
    CHARGE_STATE_UNSPECIFIED,
    CHARGE_STATE_NOT_CHARGING,
    CHARGE_STATE_CHARGING,
  ];

  static final $core.List<ChargeState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static ChargeState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ChargeState._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
