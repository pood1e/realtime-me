// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/terminal.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// TerminalSessionState describes a tmux-backed shell lifecycle.
class TerminalSessionState extends $pb.ProtobufEnum {
  /// The terminal session state was not specified.
  static const TerminalSessionState TERMINAL_SESSION_STATE_UNSPECIFIED =
      TerminalSessionState._(
          0, _omitEnumNames ? '' : 'TERMINAL_SESSION_STATE_UNSPECIFIED');

  /// The tmux-backed shell is running.
  static const TerminalSessionState TERMINAL_SESSION_STATE_RUNNING =
      TerminalSessionState._(
          1, _omitEnumNames ? '' : 'TERMINAL_SESSION_STATE_RUNNING');

  /// The shell exited or the user explicitly closed it.
  static const TerminalSessionState TERMINAL_SESSION_STATE_CLOSED =
      TerminalSessionState._(
          2, _omitEnumNames ? '' : 'TERMINAL_SESSION_STATE_CLOSED');

  static const $core.List<TerminalSessionState> values = <TerminalSessionState>[
    TERMINAL_SESSION_STATE_UNSPECIFIED,
    TERMINAL_SESSION_STATE_RUNNING,
    TERMINAL_SESSION_STATE_CLOSED,
  ];

  static final $core.List<TerminalSessionState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static TerminalSessionState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const TerminalSessionState._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
