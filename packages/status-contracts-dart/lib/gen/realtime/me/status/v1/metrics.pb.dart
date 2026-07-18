// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/metrics.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../../../google/protobuf/duration.pb.dart' as $1;
import '../../../../google/protobuf/timestamp.pb.dart' as $0;
import 'metrics.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'metrics.pbenum.dart';

/// AccessorySelector identifies one accessory attached to a device.
class AccessorySelector extends $pb.GeneratedMessage {
  factory AccessorySelector({
    $core.String? kind,
    $core.String? displayName,
  }) {
    final result = create();
    if (kind != null) result.kind = kind;
    if (displayName != null) result.displayName = displayName;
    return result;
  }

  AccessorySelector._();

  factory AccessorySelector.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory AccessorySelector.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'AccessorySelector',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'kind')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccessorySelector clone() => AccessorySelector()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  AccessorySelector copyWith(void Function(AccessorySelector) updates) =>
      super.copyWith((message) => updates(message as AccessorySelector))
          as AccessorySelector;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AccessorySelector create() => AccessorySelector._();
  @$core.override
  AccessorySelector createEmptyInstance() => create();
  static $pb.PbList<AccessorySelector> createRepeated() =>
      $pb.PbList<AccessorySelector>();
  @$core.pragma('dart2js:noInline')
  static AccessorySelector getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<AccessorySelector>(create);
  static AccessorySelector? _defaultInstance;

  /// kind is the accessory category key, such as "bluetooth_audio".
  @$pb.TagNumber(1)
  $core.String get kind => $_getSZ(0);
  @$pb.TagNumber(1)
  set kind($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasKind() => $_has(0);
  @$pb.TagNumber(1)
  void clearKind() => $_clearField(1);

  /// display_name is the accessory label, as reported on the status surface.
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);
}

/// GetMetricRangeRequest asks for one series over a time range. Which selector
/// fields apply is fixed by the series; unrelated fields are ignored.
class GetMetricRangeRequest extends $pb.GeneratedMessage {
  factory GetMetricRangeRequest({
    MetricSeries? series,
    $core.String? deviceUid,
    $core.String? agentKind,
    AccessorySelector? accessory,
    $0.Timestamp? startTime,
    $0.Timestamp? endTime,
    $1.Duration? step,
  }) {
    final result = create();
    if (series != null) result.series = series;
    if (deviceUid != null) result.deviceUid = deviceUid;
    if (agentKind != null) result.agentKind = agentKind;
    if (accessory != null) result.accessory = accessory;
    if (startTime != null) result.startTime = startTime;
    if (endTime != null) result.endTime = endTime;
    if (step != null) result.step = step;
    return result;
  }

  GetMetricRangeRequest._();

  factory GetMetricRangeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetMetricRangeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMetricRangeRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..e<MetricSeries>(1, _omitFieldNames ? '' : 'series', $pb.PbFieldType.OE,
        defaultOrMaker: MetricSeries.METRIC_SERIES_UNSPECIFIED,
        valueOf: MetricSeries.valueOf,
        enumValues: MetricSeries.values)
    ..aOS(2, _omitFieldNames ? '' : 'deviceUid')
    ..aOS(3, _omitFieldNames ? '' : 'agentKind')
    ..aOM<AccessorySelector>(4, _omitFieldNames ? '' : 'accessory',
        subBuilder: AccessorySelector.create)
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'startTime',
        subBuilder: $0.Timestamp.create)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'endTime',
        subBuilder: $0.Timestamp.create)
    ..aOM<$1.Duration>(7, _omitFieldNames ? '' : 'step',
        subBuilder: $1.Duration.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMetricRangeRequest clone() =>
      GetMetricRangeRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMetricRangeRequest copyWith(
          void Function(GetMetricRangeRequest) updates) =>
      super.copyWith((message) => updates(message as GetMetricRangeRequest))
          as GetMetricRangeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetMetricRangeRequest create() => GetMetricRangeRequest._();
  @$core.override
  GetMetricRangeRequest createEmptyInstance() => create();
  static $pb.PbList<GetMetricRangeRequest> createRepeated() =>
      $pb.PbList<GetMetricRangeRequest>();
  @$core.pragma('dart2js:noInline')
  static GetMetricRangeRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMetricRangeRequest>(create);
  static GetMetricRangeRequest? _defaultInstance;

  /// series is the series to sample.
  @$pb.TagNumber(1)
  MetricSeries get series => $_getN(0);
  @$pb.TagNumber(1)
  set series(MetricSeries value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasSeries() => $_has(0);
  @$pb.TagNumber(1)
  void clearSeries() => $_clearField(1);

  /// device_uid is the device the series belongs to.
  @$pb.TagNumber(2)
  $core.String get deviceUid => $_getSZ(1);
  @$pb.TagNumber(2)
  set deviceUid($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDeviceUid() => $_has(1);
  @$pb.TagNumber(2)
  void clearDeviceUid() => $_clearField(2);

  /// agent_kind is the agent kind, such as "codex" or "claude".
  @$pb.TagNumber(3)
  $core.String get agentKind => $_getSZ(2);
  @$pb.TagNumber(3)
  set agentKind($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAgentKind() => $_has(2);
  @$pb.TagNumber(3)
  void clearAgentKind() => $_clearField(3);

  /// accessory identifies the accessory for accessory-scoped series.
  @$pb.TagNumber(4)
  AccessorySelector get accessory => $_getN(3);
  @$pb.TagNumber(4)
  set accessory(AccessorySelector value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasAccessory() => $_has(3);
  @$pb.TagNumber(4)
  void clearAccessory() => $_clearField(4);
  @$pb.TagNumber(4)
  AccessorySelector ensureAccessory() => $_ensure(3);

  /// start_time is the inclusive start of the range.
  @$pb.TagNumber(5)
  $0.Timestamp get startTime => $_getN(4);
  @$pb.TagNumber(5)
  set startTime($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasStartTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearStartTime() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureStartTime() => $_ensure(4);

  /// end_time is the inclusive end of the range.
  @$pb.TagNumber(6)
  $0.Timestamp get endTime => $_getN(5);
  @$pb.TagNumber(6)
  set endTime($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasEndTime() => $_has(5);
  @$pb.TagNumber(6)
  void clearEndTime() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureEndTime() => $_ensure(5);

  /// step is the sampling interval between points.
  @$pb.TagNumber(7)
  $1.Duration get step => $_getN(6);
  @$pb.TagNumber(7)
  set step($1.Duration value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasStep() => $_has(6);
  @$pb.TagNumber(7)
  void clearStep() => $_clearField(7);
  @$pb.TagNumber(7)
  $1.Duration ensureStep() => $_ensure(6);
}

/// MetricPoint is one sample in a time series.
class MetricPoint extends $pb.GeneratedMessage {
  factory MetricPoint({
    $0.Timestamp? time,
    $core.double? value,
  }) {
    final result = create();
    if (time != null) result.time = time;
    if (value != null) result.value = value;
    return result;
  }

  MetricPoint._();

  factory MetricPoint.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory MetricPoint.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MetricPoint',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOM<$0.Timestamp>(1, _omitFieldNames ? '' : 'time',
        subBuilder: $0.Timestamp.create)
    ..a<$core.double>(2, _omitFieldNames ? '' : 'value', $pb.PbFieldType.OD)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MetricPoint clone() => MetricPoint()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MetricPoint copyWith(void Function(MetricPoint) updates) =>
      super.copyWith((message) => updates(message as MetricPoint))
          as MetricPoint;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MetricPoint create() => MetricPoint._();
  @$core.override
  MetricPoint createEmptyInstance() => create();
  static $pb.PbList<MetricPoint> createRepeated() => $pb.PbList<MetricPoint>();
  @$core.pragma('dart2js:noInline')
  static MetricPoint getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MetricPoint>(create);
  static MetricPoint? _defaultInstance;

  /// time is when the sample was taken.
  @$pb.TagNumber(1)
  $0.Timestamp get time => $_getN(0);
  @$pb.TagNumber(1)
  set time($0.Timestamp value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasTime() => $_has(0);
  @$pb.TagNumber(1)
  void clearTime() => $_clearField(1);
  @$pb.TagNumber(1)
  $0.Timestamp ensureTime() => $_ensure(0);

  /// value is the sample value, already scaled to the series' natural unit.
  @$pb.TagNumber(2)
  $core.double get value => $_getN(1);
  @$pb.TagNumber(2)
  set value($core.double value) => $_setDouble(1, value);
  @$pb.TagNumber(2)
  $core.bool hasValue() => $_has(1);
  @$pb.TagNumber(2)
  void clearValue() => $_clearField(2);
}

/// GetMetricRangeResponse carries the sampled series.
class GetMetricRangeResponse extends $pb.GeneratedMessage {
  factory GetMetricRangeResponse({
    $core.Iterable<MetricPoint>? points,
  }) {
    final result = create();
    if (points != null) result.points.addAll(points);
    return result;
  }

  GetMetricRangeResponse._();

  factory GetMetricRangeResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetMetricRangeResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetMetricRangeResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..pc<MetricPoint>(1, _omitFieldNames ? '' : 'points', $pb.PbFieldType.PM,
        subBuilder: MetricPoint.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMetricRangeResponse clone() =>
      GetMetricRangeResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetMetricRangeResponse copyWith(
          void Function(GetMetricRangeResponse) updates) =>
      super.copyWith((message) => updates(message as GetMetricRangeResponse))
          as GetMetricRangeResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetMetricRangeResponse create() => GetMetricRangeResponse._();
  @$core.override
  GetMetricRangeResponse createEmptyInstance() => create();
  static $pb.PbList<GetMetricRangeResponse> createRepeated() =>
      $pb.PbList<GetMetricRangeResponse>();
  @$core.pragma('dart2js:noInline')
  static GetMetricRangeResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetMetricRangeResponse>(create);
  static GetMetricRangeResponse? _defaultInstance;

  /// points are the samples, ordered by time.
  @$pb.TagNumber(1)
  $pb.PbList<MetricPoint> get points => $_getList(0);
}

/// MetricsService serves the time series behind the internal dashboard's charts.
/// Callers name the series they want; the gateway owns every metric name, label,
/// and query expression, so the storage engine stays an implementation detail.
class MetricsServiceApi {
  final $pb.RpcClient _client;

  MetricsServiceApi(this._client);

  /// GetMetricRange returns one series sampled over a time range.
  $async.Future<GetMetricRangeResponse> getMetricRange(
          $pb.ClientContext? ctx, GetMetricRangeRequest request) =>
      _client.invoke<GetMetricRangeResponse>(ctx, 'MetricsService',
          'GetMetricRange', request, GetMetricRangeResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
