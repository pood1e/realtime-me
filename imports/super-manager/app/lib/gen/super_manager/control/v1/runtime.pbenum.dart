// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/runtime.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// RuntimeKind identifies an installed coding-agent runtime.
class RuntimeKind extends $pb.ProtobufEnum {
  /// The runtime kind was not specified.
  static const RuntimeKind RUNTIME_KIND_UNSPECIFIED =
      RuntimeKind._(0, _omitEnumNames ? '' : 'RUNTIME_KIND_UNSPECIFIED');

  /// The runtime is OpenAI Codex CLI.
  static const RuntimeKind RUNTIME_KIND_CODEX =
      RuntimeKind._(1, _omitEnumNames ? '' : 'RUNTIME_KIND_CODEX');

  /// The runtime is Anthropic Claude Code CLI.
  static const RuntimeKind RUNTIME_KIND_CLAUDE_CODE =
      RuntimeKind._(2, _omitEnumNames ? '' : 'RUNTIME_KIND_CLAUDE_CODE');

  static const $core.List<RuntimeKind> values = <RuntimeKind>[
    RUNTIME_KIND_UNSPECIFIED,
    RUNTIME_KIND_CODEX,
    RUNTIME_KIND_CLAUDE_CODE,
  ];

  static final $core.List<RuntimeKind?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static RuntimeKind? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RuntimeKind._(super.value, super.name);
}

/// RuntimeAvailability describes whether a runtime can accept new executions.
class RuntimeAvailability extends $pb.ProtobufEnum {
  /// The runtime availability was not determined.
  static const RuntimeAvailability RUNTIME_AVAILABILITY_UNSPECIFIED =
      RuntimeAvailability._(
          0, _omitEnumNames ? '' : 'RUNTIME_AVAILABILITY_UNSPECIFIED');

  /// The runtime is installed, authenticated, and compatible.
  static const RuntimeAvailability RUNTIME_AVAILABILITY_AVAILABLE =
      RuntimeAvailability._(
          1, _omitEnumNames ? '' : 'RUNTIME_AVAILABILITY_AVAILABLE');

  /// The runtime executable is not installed.
  static const RuntimeAvailability RUNTIME_AVAILABILITY_NOT_INSTALLED =
      RuntimeAvailability._(
          2, _omitEnumNames ? '' : 'RUNTIME_AVAILABILITY_NOT_INSTALLED');

  /// The runtime is not authenticated with a supported subscription login.
  static const RuntimeAvailability RUNTIME_AVAILABILITY_NOT_AUTHENTICATED =
      RuntimeAvailability._(
          3, _omitEnumNames ? '' : 'RUNTIME_AVAILABILITY_NOT_AUTHENTICATED');

  /// The installed runtime version is not supported.
  static const RuntimeAvailability RUNTIME_AVAILABILITY_INCOMPATIBLE =
      RuntimeAvailability._(
          4, _omitEnumNames ? '' : 'RUNTIME_AVAILABILITY_INCOMPATIBLE');

  /// The runtime health check failed.
  static const RuntimeAvailability RUNTIME_AVAILABILITY_UNHEALTHY =
      RuntimeAvailability._(
          5, _omitEnumNames ? '' : 'RUNTIME_AVAILABILITY_UNHEALTHY');

  static const $core.List<RuntimeAvailability> values = <RuntimeAvailability>[
    RUNTIME_AVAILABILITY_UNSPECIFIED,
    RUNTIME_AVAILABILITY_AVAILABLE,
    RUNTIME_AVAILABILITY_NOT_INSTALLED,
    RUNTIME_AVAILABILITY_NOT_AUTHENTICATED,
    RUNTIME_AVAILABILITY_INCOMPATIBLE,
    RUNTIME_AVAILABILITY_UNHEALTHY,
  ];

  static final $core.List<RuntimeAvailability?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static RuntimeAvailability? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RuntimeAvailability._(super.value, super.name);
}

/// RuntimeCapability is a user-facing structured-control capability.
class RuntimeCapability extends $pb.ProtobufEnum {
  /// The runtime capability was not specified.
  static const RuntimeCapability RUNTIME_CAPABILITY_UNSPECIFIED =
      RuntimeCapability._(
          0, _omitEnumNames ? '' : 'RUNTIME_CAPABILITY_UNSPECIFIED');

  /// The runtime streams assistant text.
  static const RuntimeCapability RUNTIME_CAPABILITY_TEXT_STREAMING =
      RuntimeCapability._(
          1, _omitEnumNames ? '' : 'RUNTIME_CAPABILITY_TEXT_STREAMING');

  /// The runtime streams structured tool calls and results.
  static const RuntimeCapability RUNTIME_CAPABILITY_TOOL_STREAMING =
      RuntimeCapability._(
          2, _omitEnumNames ? '' : 'RUNTIME_CAPABILITY_TOOL_STREAMING');

  /// The runtime can request structured user input.
  static const RuntimeCapability RUNTIME_CAPABILITY_STRUCTURED_QUESTIONS =
      RuntimeCapability._(
          3, _omitEnumNames ? '' : 'RUNTIME_CAPABILITY_STRUCTURED_QUESTIONS');

  /// The runtime can cancel an active execution.
  static const RuntimeCapability RUNTIME_CAPABILITY_CANCEL =
      RuntimeCapability._(4, _omitEnumNames ? '' : 'RUNTIME_CAPABILITY_CANCEL');

  /// The runtime can steer an active execution.
  static const RuntimeCapability RUNTIME_CAPABILITY_STEER =
      RuntimeCapability._(5, _omitEnumNames ? '' : 'RUNTIME_CAPABILITY_STEER');

  /// The runtime exposes account quota telemetry.
  static const RuntimeCapability RUNTIME_CAPABILITY_QUOTA =
      RuntimeCapability._(6, _omitEnumNames ? '' : 'RUNTIME_CAPABILITY_QUOTA');

  /// The runtime emits safe reasoning summaries.
  static const RuntimeCapability RUNTIME_CAPABILITY_REASONING_SUMMARIES =
      RuntimeCapability._(
          7, _omitEnumNames ? '' : 'RUNTIME_CAPABILITY_REASONING_SUMMARIES');

  static const $core.List<RuntimeCapability> values = <RuntimeCapability>[
    RUNTIME_CAPABILITY_UNSPECIFIED,
    RUNTIME_CAPABILITY_TEXT_STREAMING,
    RUNTIME_CAPABILITY_TOOL_STREAMING,
    RUNTIME_CAPABILITY_STRUCTURED_QUESTIONS,
    RUNTIME_CAPABILITY_CANCEL,
    RUNTIME_CAPABILITY_STEER,
    RUNTIME_CAPABILITY_QUOTA,
    RUNTIME_CAPABILITY_REASONING_SUMMARIES,
  ];

  static final $core.List<RuntimeCapability?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 7);
  static RuntimeCapability? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const RuntimeCapability._(super.value, super.name);
}

/// QuotaFreshness describes the reliability of a quota observation.
class QuotaFreshness extends $pb.ProtobufEnum {
  /// Quota freshness was not determined.
  static const QuotaFreshness QUOTA_FRESHNESS_UNSPECIFIED =
      QuotaFreshness._(0, _omitEnumNames ? '' : 'QUOTA_FRESHNESS_UNSPECIFIED');

  /// The quota was observed recently from a structured source.
  static const QuotaFreshness QUOTA_FRESHNESS_FRESH =
      QuotaFreshness._(1, _omitEnumNames ? '' : 'QUOTA_FRESHNESS_FRESH');

  /// A prior observation exists but is no longer fresh.
  static const QuotaFreshness QUOTA_FRESHNESS_STALE =
      QuotaFreshness._(2, _omitEnumNames ? '' : 'QUOTA_FRESHNESS_STALE');

  /// The runtime does not currently provide quota telemetry.
  static const QuotaFreshness QUOTA_FRESHNESS_UNAVAILABLE =
      QuotaFreshness._(3, _omitEnumNames ? '' : 'QUOTA_FRESHNESS_UNAVAILABLE');

  static const $core.List<QuotaFreshness> values = <QuotaFreshness>[
    QUOTA_FRESHNESS_UNSPECIFIED,
    QUOTA_FRESHNESS_FRESH,
    QUOTA_FRESHNESS_STALE,
    QUOTA_FRESHNESS_UNAVAILABLE,
  ];

  static final $core.List<QuotaFreshness?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static QuotaFreshness? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const QuotaFreshness._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
