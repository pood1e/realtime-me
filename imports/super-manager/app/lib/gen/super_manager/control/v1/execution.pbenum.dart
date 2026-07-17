// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/execution.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// ExecutionState describes the native provider process lifecycle.
class ExecutionState extends $pb.ProtobufEnum {
  /// The execution state was not specified.
  static const ExecutionState EXECUTION_STATE_UNSPECIFIED =
      ExecutionState._(0, _omitEnumNames ? '' : 'EXECUTION_STATE_UNSPECIFIED');

  /// The provider execution is running.
  static const ExecutionState EXECUTION_STATE_RUNNING =
      ExecutionState._(1, _omitEnumNames ? '' : 'EXECUTION_STATE_RUNNING');

  /// The provider execution is waiting for structured input.
  static const ExecutionState EXECUTION_STATE_INPUT_REQUIRED = ExecutionState._(
      2, _omitEnumNames ? '' : 'EXECUTION_STATE_INPUT_REQUIRED');

  /// The provider execution completed successfully.
  static const ExecutionState EXECUTION_STATE_SUCCEEDED =
      ExecutionState._(3, _omitEnumNames ? '' : 'EXECUTION_STATE_SUCCEEDED');

  /// The provider execution failed.
  static const ExecutionState EXECUTION_STATE_FAILED =
      ExecutionState._(4, _omitEnumNames ? '' : 'EXECUTION_STATE_FAILED');

  /// The user canceled the provider execution.
  static const ExecutionState EXECUTION_STATE_CANCELED =
      ExecutionState._(5, _omitEnumNames ? '' : 'EXECUTION_STATE_CANCELED');

  /// The provider process disappeared and cannot be resumed in memory.
  static const ExecutionState EXECUTION_STATE_LOST =
      ExecutionState._(6, _omitEnumNames ? '' : 'EXECUTION_STATE_LOST');

  static const $core.List<ExecutionState> values = <ExecutionState>[
    EXECUTION_STATE_UNSPECIFIED,
    EXECUTION_STATE_RUNNING,
    EXECUTION_STATE_INPUT_REQUIRED,
    EXECUTION_STATE_SUCCEEDED,
    EXECUTION_STATE_FAILED,
    EXECUTION_STATE_CANCELED,
    EXECUTION_STATE_LOST,
  ];

  static final $core.List<ExecutionState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 6);
  static ExecutionState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ExecutionState._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
