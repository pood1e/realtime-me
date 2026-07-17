// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/thread.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// ThreadState describes whether a structured conversation can accept new runs.
class ThreadState extends $pb.ProtobufEnum {
  /// The thread state was not specified.
  static const ThreadState THREAD_STATE_UNSPECIFIED =
      ThreadState._(0, _omitEnumNames ? '' : 'THREAD_STATE_UNSPECIFIED');

  /// The thread can accept a new run.
  static const ThreadState THREAD_STATE_IDLE =
      ThreadState._(1, _omitEnumNames ? '' : 'THREAD_STATE_IDLE');

  /// The thread has an active provider execution.
  static const ThreadState THREAD_STATE_RUNNING =
      ThreadState._(2, _omitEnumNames ? '' : 'THREAD_STATE_RUNNING');

  /// The thread is waiting for structured user input.
  static const ThreadState THREAD_STATE_INPUT_REQUIRED =
      ThreadState._(3, _omitEnumNames ? '' : 'THREAD_STATE_INPUT_REQUIRED');

  /// The thread cannot continue because its provider execution was lost.
  static const ThreadState THREAD_STATE_LOST =
      ThreadState._(4, _omitEnumNames ? '' : 'THREAD_STATE_LOST');

  static const $core.List<ThreadState> values = <ThreadState>[
    THREAD_STATE_UNSPECIFIED,
    THREAD_STATE_IDLE,
    THREAD_STATE_RUNNING,
    THREAD_STATE_INPUT_REQUIRED,
    THREAD_STATE_LOST,
  ];

  static final $core.List<ThreadState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static ThreadState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ThreadState._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
