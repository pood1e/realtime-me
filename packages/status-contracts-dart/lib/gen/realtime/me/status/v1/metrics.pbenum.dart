// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/metrics.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// MetricSeries names a chartable series. Each value maps to exactly one query
/// the gateway builds; callers never construct a query expression.
class MetricSeries extends $pb.ProtobufEnum {
  /// Series is not known.
  static const MetricSeries METRIC_SERIES_UNSPECIFIED =
      MetricSeries._(0, _omitEnumNames ? '' : 'METRIC_SERIES_UNSPECIFIED');

  /// Host CPU utilization as a percentage. Requires device_uid.
  static const MetricSeries METRIC_SERIES_HOST_CPU_UTILIZATION = MetricSeries._(
      1, _omitEnumNames ? '' : 'METRIC_SERIES_HOST_CPU_UTILIZATION');

  /// Host memory in use, in bytes. Requires device_uid.
  static const MetricSeries METRIC_SERIES_HOST_MEMORY_USAGE = MetricSeries._(
      2, _omitEnumNames ? '' : 'METRIC_SERIES_HOST_MEMORY_USAGE');

  /// Host root-filesystem utilization as a percentage. Requires device_uid.
  static const MetricSeries METRIC_SERIES_HOST_FILESYSTEM_UTILIZATION =
      MetricSeries._(
          3, _omitEnumNames ? '' : 'METRIC_SERIES_HOST_FILESYSTEM_UTILIZATION');

  /// Phone battery level as a percentage. Requires device_uid.
  static const MetricSeries METRIC_SERIES_PHONE_BATTERY_LEVEL = MetricSeries._(
      4, _omitEnumNames ? '' : 'METRIC_SERIES_PHONE_BATTERY_LEVEL');

  /// Watch battery level as a percentage. Requires device_uid.
  static const MetricSeries METRIC_SERIES_WATCH_BATTERY_LEVEL = MetricSeries._(
      5, _omitEnumNames ? '' : 'METRIC_SERIES_WATCH_BATTERY_LEVEL');

  /// Watch heart rate in beats per minute. Requires device_uid.
  static const MetricSeries METRIC_SERIES_WATCH_HEART_RATE =
      MetricSeries._(6, _omitEnumNames ? '' : 'METRIC_SERIES_WATCH_HEART_RATE');

  /// Watch local-day step count. Requires device_uid.
  static const MetricSeries METRIC_SERIES_WATCH_STEPS =
      MetricSeries._(7, _omitEnumNames ? '' : 'METRIC_SERIES_WATCH_STEPS');

  /// Accessory battery level as a percentage. Requires device_uid and accessory.
  static const MetricSeries METRIC_SERIES_ACCESSORY_BATTERY_LEVEL =
      MetricSeries._(
          8, _omitEnumNames ? '' : 'METRIC_SERIES_ACCESSORY_BATTERY_LEVEL');

  /// Coding-agent budget remaining as a percentage. Requires agent_kind.
  static const MetricSeries METRIC_SERIES_AGENT_BUDGET_REMAINING =
      MetricSeries._(
          9, _omitEnumNames ? '' : 'METRIC_SERIES_AGENT_BUDGET_REMAINING');

  static const $core.List<MetricSeries> values = <MetricSeries>[
    METRIC_SERIES_UNSPECIFIED,
    METRIC_SERIES_HOST_CPU_UTILIZATION,
    METRIC_SERIES_HOST_MEMORY_USAGE,
    METRIC_SERIES_HOST_FILESYSTEM_UTILIZATION,
    METRIC_SERIES_PHONE_BATTERY_LEVEL,
    METRIC_SERIES_WATCH_BATTERY_LEVEL,
    METRIC_SERIES_WATCH_HEART_RATE,
    METRIC_SERIES_WATCH_STEPS,
    METRIC_SERIES_ACCESSORY_BATTERY_LEVEL,
    METRIC_SERIES_AGENT_BUDGET_REMAINING,
  ];

  static final $core.List<MetricSeries?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 9);
  static MetricSeries? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const MetricSeries._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
