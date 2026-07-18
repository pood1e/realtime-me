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

import '../../../../google/protobuf/timestamp.pb.dart' as $0;
import 'watch.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'watch.pbenum.dart';

/// WatchSnapshot is the latest health and device snapshot produced by the watch.
class WatchSnapshot extends $pb.GeneratedMessage {
  factory WatchSnapshot({
    $core.String? snapshotId,
    $0.Timestamp? recordTime,
    HeartRateSample? heartRate,
    ActivityTotals? activityTotals,
    WatchState? watchState,
    DeviceInfo? deviceInfo,
  }) {
    final result = create();
    if (snapshotId != null) result.snapshotId = snapshotId;
    if (recordTime != null) result.recordTime = recordTime;
    if (heartRate != null) result.heartRate = heartRate;
    if (activityTotals != null) result.activityTotals = activityTotals;
    if (watchState != null) result.watchState = watchState;
    if (deviceInfo != null) result.deviceInfo = deviceInfo;
    return result;
  }

  WatchSnapshot._();

  factory WatchSnapshot.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory WatchSnapshot.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WatchSnapshot',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'snapshotId')
    ..aOM<$0.Timestamp>(2, _omitFieldNames ? '' : 'recordTime',
        subBuilder: $0.Timestamp.create)
    ..aOM<HeartRateSample>(3, _omitFieldNames ? '' : 'heartRate',
        subBuilder: HeartRateSample.create)
    ..aOM<ActivityTotals>(4, _omitFieldNames ? '' : 'activityTotals',
        subBuilder: ActivityTotals.create)
    ..aOM<WatchState>(5, _omitFieldNames ? '' : 'watchState',
        subBuilder: WatchState.create)
    ..aOM<DeviceInfo>(6, _omitFieldNames ? '' : 'deviceInfo',
        subBuilder: DeviceInfo.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WatchSnapshot clone() => WatchSnapshot()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WatchSnapshot copyWith(void Function(WatchSnapshot) updates) =>
      super.copyWith((message) => updates(message as WatchSnapshot))
          as WatchSnapshot;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WatchSnapshot create() => WatchSnapshot._();
  @$core.override
  WatchSnapshot createEmptyInstance() => create();
  static $pb.PbList<WatchSnapshot> createRepeated() =>
      $pb.PbList<WatchSnapshot>();
  @$core.pragma('dart2js:noInline')
  static WatchSnapshot getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WatchSnapshot>(create);
  static WatchSnapshot? _defaultInstance;

  /// snapshot_id is a watch-generated UUID for de-duplicating Data Layer deliveries.
  @$pb.TagNumber(1)
  $core.String get snapshotId => $_getSZ(0);
  @$pb.TagNumber(1)
  set snapshotId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSnapshotId() => $_has(0);
  @$pb.TagNumber(1)
  void clearSnapshotId() => $_clearField(1);

  /// record_time is the wall-clock time when the snapshot was assembled.
  @$pb.TagNumber(2)
  $0.Timestamp get recordTime => $_getN(1);
  @$pb.TagNumber(2)
  set recordTime($0.Timestamp value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasRecordTime() => $_has(1);
  @$pb.TagNumber(2)
  void clearRecordTime() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Timestamp ensureRecordTime() => $_ensure(1);

  /// heart_rate is the latest heart-rate sensor sample, if available.
  @$pb.TagNumber(3)
  HeartRateSample get heartRate => $_getN(2);
  @$pb.TagNumber(3)
  set heartRate(HeartRateSample value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasHeartRate() => $_has(2);
  @$pb.TagNumber(3)
  void clearHeartRate() => $_clearField(3);
  @$pb.TagNumber(3)
  HeartRateSample ensureHeartRate() => $_ensure(2);

  /// activity_totals contains step counters reported by the watch.
  @$pb.TagNumber(4)
  ActivityTotals get activityTotals => $_getN(3);
  @$pb.TagNumber(4)
  set activityTotals(ActivityTotals value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasActivityTotals() => $_has(3);
  @$pb.TagNumber(4)
  void clearActivityTotals() => $_clearField(4);
  @$pb.TagNumber(4)
  ActivityTotals ensureActivityTotals() => $_ensure(3);

  /// watch_state contains battery and worn-state information.
  @$pb.TagNumber(5)
  WatchState get watchState => $_getN(4);
  @$pb.TagNumber(5)
  set watchState(WatchState value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasWatchState() => $_has(4);
  @$pb.TagNumber(5)
  void clearWatchState() => $_clearField(5);
  @$pb.TagNumber(5)
  WatchState ensureWatchState() => $_ensure(4);

  /// device_info contains non-sensitive watch identity shown on status surfaces.
  @$pb.TagNumber(6)
  DeviceInfo get deviceInfo => $_getN(5);
  @$pb.TagNumber(6)
  set deviceInfo(DeviceInfo value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasDeviceInfo() => $_has(5);
  @$pb.TagNumber(6)
  void clearDeviceInfo() => $_clearField(6);
  @$pb.TagNumber(6)
  DeviceInfo ensureDeviceInfo() => $_ensure(5);
}

/// DeviceInfo contains non-sensitive device identity.
class DeviceInfo extends $pb.GeneratedMessage {
  factory DeviceInfo({
    $core.String? displayName,
    $core.String? model,
  }) {
    final result = create();
    if (displayName != null) result.displayName = displayName;
    if (model != null) result.model = model;
    return result;
  }

  DeviceInfo._();

  factory DeviceInfo.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeviceInfo.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeviceInfo',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'displayName')
    ..aOS(2, _omitFieldNames ? '' : 'model')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeviceInfo clone() => DeviceInfo()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeviceInfo copyWith(void Function(DeviceInfo) updates) =>
      super.copyWith((message) => updates(message as DeviceInfo)) as DeviceInfo;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeviceInfo create() => DeviceInfo._();
  @$core.override
  DeviceInfo createEmptyInstance() => create();
  static $pb.PbList<DeviceInfo> createRepeated() => $pb.PbList<DeviceInfo>();
  @$core.pragma('dart2js:noInline')
  static DeviceInfo getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeviceInfo>(create);
  static DeviceInfo? _defaultInstance;

  /// display_name is the human-readable device label.
  @$pb.TagNumber(1)
  $core.String get displayName => $_getSZ(0);
  @$pb.TagNumber(1)
  set displayName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDisplayName() => $_has(0);
  @$pb.TagNumber(1)
  void clearDisplayName() => $_clearField(1);

  /// model is the hardware model reported by the operating system.
  @$pb.TagNumber(2)
  $core.String get model => $_getSZ(1);
  @$pb.TagNumber(2)
  set model($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasModel() => $_has(1);
  @$pb.TagNumber(2)
  void clearModel() => $_clearField(2);
}

/// HeartRateSample is one heart-rate observation in beats per minute.
class HeartRateSample extends $pb.GeneratedMessage {
  factory HeartRateSample({
    $core.int? beatsPerMinute,
    $0.Timestamp? sampleTime,
  }) {
    final result = create();
    if (beatsPerMinute != null) result.beatsPerMinute = beatsPerMinute;
    if (sampleTime != null) result.sampleTime = sampleTime;
    return result;
  }

  HeartRateSample._();

  factory HeartRateSample.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory HeartRateSample.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'HeartRateSample',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..a<$core.int>(
        1, _omitFieldNames ? '' : 'beatsPerMinute', $pb.PbFieldType.O3)
    ..aOM<$0.Timestamp>(2, _omitFieldNames ? '' : 'sampleTime',
        subBuilder: $0.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HeartRateSample clone() => HeartRateSample()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  HeartRateSample copyWith(void Function(HeartRateSample) updates) =>
      super.copyWith((message) => updates(message as HeartRateSample))
          as HeartRateSample;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static HeartRateSample create() => HeartRateSample._();
  @$core.override
  HeartRateSample createEmptyInstance() => create();
  static $pb.PbList<HeartRateSample> createRepeated() =>
      $pb.PbList<HeartRateSample>();
  @$core.pragma('dart2js:noInline')
  static HeartRateSample getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<HeartRateSample>(create);
  static HeartRateSample? _defaultInstance;

  /// beats_per_minute is the measured heart rate.
  @$pb.TagNumber(1)
  $core.int get beatsPerMinute => $_getIZ(0);
  @$pb.TagNumber(1)
  set beatsPerMinute($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasBeatsPerMinute() => $_has(0);
  @$pb.TagNumber(1)
  void clearBeatsPerMinute() => $_clearField(1);

  /// sample_time is the source sample time.
  @$pb.TagNumber(2)
  $0.Timestamp get sampleTime => $_getN(1);
  @$pb.TagNumber(2)
  set sampleTime($0.Timestamp value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasSampleTime() => $_has(1);
  @$pb.TagNumber(2)
  void clearSampleTime() => $_clearField(2);
  @$pb.TagNumber(2)
  $0.Timestamp ensureSampleTime() => $_ensure(1);
}

/// ActivityTotals contains step counters using the watch's local day boundary.
class ActivityTotals extends $pb.GeneratedMessage {
  factory ActivityTotals({
    $core.int? steps,
    $0.Timestamp? sampleTime,
  }) {
    final result = create();
    if (steps != null) result.steps = steps;
    if (sampleTime != null) result.sampleTime = sampleTime;
    return result;
  }

  ActivityTotals._();

  factory ActivityTotals.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ActivityTotals.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ActivityTotals',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'steps', $pb.PbFieldType.O3)
    ..aOM<$0.Timestamp>(3, _omitFieldNames ? '' : 'sampleTime',
        subBuilder: $0.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ActivityTotals clone() => ActivityTotals()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ActivityTotals copyWith(void Function(ActivityTotals) updates) =>
      super.copyWith((message) => updates(message as ActivityTotals))
          as ActivityTotals;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ActivityTotals create() => ActivityTotals._();
  @$core.override
  ActivityTotals createEmptyInstance() => create();
  static $pb.PbList<ActivityTotals> createRepeated() =>
      $pb.PbList<ActivityTotals>();
  @$core.pragma('dart2js:noInline')
  static ActivityTotals getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ActivityTotals>(create);
  static ActivityTotals? _defaultInstance;

  /// steps is the daily step count.
  @$pb.TagNumber(1)
  $core.int get steps => $_getIZ(0);
  @$pb.TagNumber(1)
  set steps($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSteps() => $_has(0);
  @$pb.TagNumber(1)
  void clearSteps() => $_clearField(1);

  /// sample_time is the latest source sample time for these counters.
  @$pb.TagNumber(3)
  $0.Timestamp get sampleTime => $_getN(1);
  @$pb.TagNumber(3)
  set sampleTime($0.Timestamp value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasSampleTime() => $_has(1);
  @$pb.TagNumber(3)
  void clearSampleTime() => $_clearField(3);
  @$pb.TagNumber(3)
  $0.Timestamp ensureSampleTime() => $_ensure(1);
}

/// WatchState contains low-frequency device state shown in GitHub status.
class WatchState extends $pb.GeneratedMessage {
  factory WatchState({
    $core.int? batteryPercent,
    ChargeState? chargeState,
    $0.Timestamp? sampleTime,
  }) {
    final result = create();
    if (batteryPercent != null) result.batteryPercent = batteryPercent;
    if (chargeState != null) result.chargeState = chargeState;
    if (sampleTime != null) result.sampleTime = sampleTime;
    return result;
  }

  WatchState._();

  factory WatchState.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory WatchState.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WatchState',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..a<$core.int>(
        1, _omitFieldNames ? '' : 'batteryPercent', $pb.PbFieldType.O3)
    ..e<ChargeState>(
        2, _omitFieldNames ? '' : 'chargeState', $pb.PbFieldType.OE,
        defaultOrMaker: ChargeState.CHARGE_STATE_UNSPECIFIED,
        valueOf: ChargeState.valueOf,
        enumValues: ChargeState.values)
    ..aOM<$0.Timestamp>(4, _omitFieldNames ? '' : 'sampleTime',
        subBuilder: $0.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WatchState clone() => WatchState()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WatchState copyWith(void Function(WatchState) updates) =>
      super.copyWith((message) => updates(message as WatchState)) as WatchState;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WatchState create() => WatchState._();
  @$core.override
  WatchState createEmptyInstance() => create();
  static $pb.PbList<WatchState> createRepeated() => $pb.PbList<WatchState>();
  @$core.pragma('dart2js:noInline')
  static WatchState getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WatchState>(create);
  static WatchState? _defaultInstance;

  /// battery_percent is the watch battery level from 0 to 100.
  @$pb.TagNumber(1)
  $core.int get batteryPercent => $_getIZ(0);
  @$pb.TagNumber(1)
  set batteryPercent($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasBatteryPercent() => $_has(0);
  @$pb.TagNumber(1)
  void clearBatteryPercent() => $_clearField(1);

  /// charge_state is the current charging state.
  @$pb.TagNumber(2)
  ChargeState get chargeState => $_getN(1);
  @$pb.TagNumber(2)
  set chargeState(ChargeState value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasChargeState() => $_has(1);
  @$pb.TagNumber(2)
  void clearChargeState() => $_clearField(2);

  /// sample_time is the state sample time.
  @$pb.TagNumber(4)
  $0.Timestamp get sampleTime => $_getN(2);
  @$pb.TagNumber(4)
  set sampleTime($0.Timestamp value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasSampleTime() => $_has(2);
  @$pb.TagNumber(4)
  void clearSampleTime() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Timestamp ensureSampleTime() => $_ensure(2);
}

/// ReportWatchSnapshotRequest is the Data Layer payload sent from watch to phone.
class ReportWatchSnapshotRequest extends $pb.GeneratedMessage {
  factory ReportWatchSnapshotRequest({
    WatchSnapshot? watchSnapshot,
  }) {
    final result = create();
    if (watchSnapshot != null) result.watchSnapshot = watchSnapshot;
    return result;
  }

  ReportWatchSnapshotRequest._();

  factory ReportWatchSnapshotRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ReportWatchSnapshotRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReportWatchSnapshotRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOM<WatchSnapshot>(1, _omitFieldNames ? '' : 'watchSnapshot',
        subBuilder: WatchSnapshot.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReportWatchSnapshotRequest clone() =>
      ReportWatchSnapshotRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReportWatchSnapshotRequest copyWith(
          void Function(ReportWatchSnapshotRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ReportWatchSnapshotRequest))
          as ReportWatchSnapshotRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ReportWatchSnapshotRequest create() => ReportWatchSnapshotRequest._();
  @$core.override
  ReportWatchSnapshotRequest createEmptyInstance() => create();
  static $pb.PbList<ReportWatchSnapshotRequest> createRepeated() =>
      $pb.PbList<ReportWatchSnapshotRequest>();
  @$core.pragma('dart2js:noInline')
  static ReportWatchSnapshotRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReportWatchSnapshotRequest>(create);
  static ReportWatchSnapshotRequest? _defaultInstance;

  /// watch_snapshot is the latest snapshot to publish.
  @$pb.TagNumber(1)
  WatchSnapshot get watchSnapshot => $_getN(0);
  @$pb.TagNumber(1)
  set watchSnapshot(WatchSnapshot value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWatchSnapshot() => $_has(0);
  @$pb.TagNumber(1)
  void clearWatchSnapshot() => $_clearField(1);
  @$pb.TagNumber(1)
  WatchSnapshot ensureWatchSnapshot() => $_ensure(0);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
