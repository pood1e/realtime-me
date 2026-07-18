// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/status.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../../../google/protobuf/timestamp.pb.dart' as $1;
import 'status.pbenum.dart';
import 'status_types.pb.dart' as $0;
import 'watch.pb.dart' as $2;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'status.pbenum.dart';

/// GetPublicStatusRequest is the request for the public status document.
class GetPublicStatusRequest extends $pb.GeneratedMessage {
  factory GetPublicStatusRequest() => create();

  GetPublicStatusRequest._();

  factory GetPublicStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetPublicStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetPublicStatusRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetPublicStatusRequest clone() =>
      GetPublicStatusRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetPublicStatusRequest copyWith(
          void Function(GetPublicStatusRequest) updates) =>
      super.copyWith((message) => updates(message as GetPublicStatusRequest))
          as GetPublicStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetPublicStatusRequest create() => GetPublicStatusRequest._();
  @$core.override
  GetPublicStatusRequest createEmptyInstance() => create();
  static $pb.PbList<GetPublicStatusRequest> createRepeated() =>
      $pb.PbList<GetPublicStatusRequest>();
  @$core.pragma('dart2js:noInline')
  static GetPublicStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetPublicStatusRequest>(create);
  static GetPublicStatusRequest? _defaultInstance;
}

/// GetPublicStatusResponse carries the public status document.
class GetPublicStatusResponse extends $pb.GeneratedMessage {
  factory GetPublicStatusResponse({
    PublicStatus? status,
  }) {
    final result = create();
    if (status != null) result.status = status;
    return result;
  }

  GetPublicStatusResponse._();

  factory GetPublicStatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetPublicStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetPublicStatusResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOM<PublicStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: PublicStatus.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetPublicStatusResponse clone() =>
      GetPublicStatusResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetPublicStatusResponse copyWith(
          void Function(GetPublicStatusResponse) updates) =>
      super.copyWith((message) => updates(message as GetPublicStatusResponse))
          as GetPublicStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetPublicStatusResponse create() => GetPublicStatusResponse._();
  @$core.override
  GetPublicStatusResponse createEmptyInstance() => create();
  static $pb.PbList<GetPublicStatusResponse> createRepeated() =>
      $pb.PbList<GetPublicStatusResponse>();
  @$core.pragma('dart2js:noInline')
  static GetPublicStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetPublicStatusResponse>(create);
  static GetPublicStatusResponse? _defaultInstance;

  /// status is the public status document.
  @$pb.TagNumber(1)
  PublicStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(PublicStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  PublicStatus ensureStatus() => $_ensure(0);
}

/// GetInternalStatusRequest is the request for the internal status document.
class GetInternalStatusRequest extends $pb.GeneratedMessage {
  factory GetInternalStatusRequest() => create();

  GetInternalStatusRequest._();

  factory GetInternalStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetInternalStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetInternalStatusRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetInternalStatusRequest clone() =>
      GetInternalStatusRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetInternalStatusRequest copyWith(
          void Function(GetInternalStatusRequest) updates) =>
      super.copyWith((message) => updates(message as GetInternalStatusRequest))
          as GetInternalStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetInternalStatusRequest create() => GetInternalStatusRequest._();
  @$core.override
  GetInternalStatusRequest createEmptyInstance() => create();
  static $pb.PbList<GetInternalStatusRequest> createRepeated() =>
      $pb.PbList<GetInternalStatusRequest>();
  @$core.pragma('dart2js:noInline')
  static GetInternalStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetInternalStatusRequest>(create);
  static GetInternalStatusRequest? _defaultInstance;
}

/// GetInternalStatusResponse carries the internal status document.
class GetInternalStatusResponse extends $pb.GeneratedMessage {
  factory GetInternalStatusResponse({
    InternalStatus? status,
  }) {
    final result = create();
    if (status != null) result.status = status;
    return result;
  }

  GetInternalStatusResponse._();

  factory GetInternalStatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetInternalStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetInternalStatusResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOM<InternalStatus>(1, _omitFieldNames ? '' : 'status',
        subBuilder: InternalStatus.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetInternalStatusResponse clone() =>
      GetInternalStatusResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetInternalStatusResponse copyWith(
          void Function(GetInternalStatusResponse) updates) =>
      super.copyWith((message) => updates(message as GetInternalStatusResponse))
          as GetInternalStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetInternalStatusResponse create() => GetInternalStatusResponse._();
  @$core.override
  GetInternalStatusResponse createEmptyInstance() => create();
  static $pb.PbList<GetInternalStatusResponse> createRepeated() =>
      $pb.PbList<GetInternalStatusResponse>();
  @$core.pragma('dart2js:noInline')
  static GetInternalStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetInternalStatusResponse>(create);
  static GetInternalStatusResponse? _defaultInstance;

  /// status is the internal status document.
  @$pb.TagNumber(1)
  InternalStatus get status => $_getN(0);
  @$pb.TagNumber(1)
  set status(InternalStatus value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasStatus() => $_has(0);
  @$pb.TagNumber(1)
  void clearStatus() => $_clearField(1);
  @$pb.TagNumber(1)
  InternalStatus ensureStatus() => $_ensure(0);
}

/// DeviceState is a device's current status on a read surface.
class DeviceState extends $pb.GeneratedMessage {
  factory DeviceState({
    $core.String? deviceUid,
    $core.String? displayName,
    $core.String? model,
    $0.DeviceKind? kind,
    $0.DeviceRole? role,
    $0.OnlineState? state,
    $core.Iterable<$0.MetricSample>? metrics,
    $0.MediaStatus? media,
    $core.Iterable<$0.Accessory>? accessories,
    $1.Timestamp? updateTime,
  }) {
    final result = create();
    if (deviceUid != null) result.deviceUid = deviceUid;
    if (displayName != null) result.displayName = displayName;
    if (model != null) result.model = model;
    if (kind != null) result.kind = kind;
    if (role != null) result.role = role;
    if (state != null) result.state = state;
    if (metrics != null) result.metrics.addAll(metrics);
    if (media != null) result.media = media;
    if (accessories != null) result.accessories.addAll(accessories);
    if (updateTime != null) result.updateTime = updateTime;
    return result;
  }

  DeviceState._();

  factory DeviceState.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeviceState.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeviceState',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'deviceUid')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..aOS(3, _omitFieldNames ? '' : 'model')
    ..e<$0.DeviceKind>(4, _omitFieldNames ? '' : 'kind', $pb.PbFieldType.OE,
        defaultOrMaker: $0.DeviceKind.DEVICE_KIND_UNSPECIFIED,
        valueOf: $0.DeviceKind.valueOf,
        enumValues: $0.DeviceKind.values)
    ..e<$0.DeviceRole>(5, _omitFieldNames ? '' : 'role', $pb.PbFieldType.OE,
        defaultOrMaker: $0.DeviceRole.DEVICE_ROLE_UNSPECIFIED,
        valueOf: $0.DeviceRole.valueOf,
        enumValues: $0.DeviceRole.values)
    ..e<$0.OnlineState>(6, _omitFieldNames ? '' : 'state', $pb.PbFieldType.OE,
        defaultOrMaker: $0.OnlineState.ONLINE_STATE_UNSPECIFIED,
        valueOf: $0.OnlineState.valueOf,
        enumValues: $0.OnlineState.values)
    ..pc<$0.MetricSample>(
        7, _omitFieldNames ? '' : 'metrics', $pb.PbFieldType.PM,
        subBuilder: $0.MetricSample.create)
    ..aOM<$0.MediaStatus>(8, _omitFieldNames ? '' : 'media',
        subBuilder: $0.MediaStatus.create)
    ..pc<$0.Accessory>(
        9, _omitFieldNames ? '' : 'accessories', $pb.PbFieldType.PM,
        subBuilder: $0.Accessory.create)
    ..aOM<$1.Timestamp>(11, _omitFieldNames ? '' : 'updateTime',
        subBuilder: $1.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeviceState clone() => DeviceState()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeviceState copyWith(void Function(DeviceState) updates) =>
      super.copyWith((message) => updates(message as DeviceState))
          as DeviceState;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeviceState create() => DeviceState._();
  @$core.override
  DeviceState createEmptyInstance() => create();
  static $pb.PbList<DeviceState> createRepeated() => $pb.PbList<DeviceState>();
  @$core.pragma('dart2js:noInline')
  static DeviceState getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeviceState>(create);
  static DeviceState? _defaultInstance;

  /// device_uid is the gateway-assigned device identity.
  @$pb.TagNumber(1)
  $core.String get deviceUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set deviceUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDeviceUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearDeviceUid() => $_clearField(1);

  /// display_name is the human-readable device label.
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);

  /// model is the device hardware model.
  @$pb.TagNumber(3)
  $core.String get model => $_getSZ(2);
  @$pb.TagNumber(3)
  set model($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasModel() => $_has(2);
  @$pb.TagNumber(3)
  void clearModel() => $_clearField(3);

  /// kind is the device category.
  @$pb.TagNumber(4)
  $0.DeviceKind get kind => $_getN(3);
  @$pb.TagNumber(4)
  set kind($0.DeviceKind value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasKind() => $_has(3);
  @$pb.TagNumber(4)
  void clearKind() => $_clearField(4);

  /// role is the device's operational role.
  @$pb.TagNumber(5)
  $0.DeviceRole get role => $_getN(4);
  @$pb.TagNumber(5)
  set role($0.DeviceRole value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasRole() => $_has(4);
  @$pb.TagNumber(5)
  void clearRole() => $_clearField(5);

  /// state is the device's reachability.
  @$pb.TagNumber(6)
  $0.OnlineState get state => $_getN(5);
  @$pb.TagNumber(6)
  set state($0.OnlineState value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasState() => $_has(5);
  @$pb.TagNumber(6)
  void clearState() => $_clearField(6);

  /// metrics are the device's latest OpenTelemetry-named metrics.
  @$pb.TagNumber(7)
  $pb.PbList<$0.MetricSample> get metrics => $_getList(6);

  /// media is the currently playing media, if any.
  @$pb.TagNumber(8)
  $0.MediaStatus get media => $_getN(7);
  @$pb.TagNumber(8)
  set media($0.MediaStatus value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasMedia() => $_has(7);
  @$pb.TagNumber(8)
  void clearMedia() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.MediaStatus ensureMedia() => $_ensure(7);

  /// accessories are connected peripherals.
  @$pb.TagNumber(9)
  $pb.PbList<$0.Accessory> get accessories => $_getList(8);

  /// update_time is when this device's status was last refreshed.
  @$pb.TagNumber(11)
  $1.Timestamp get updateTime => $_getN(9);
  @$pb.TagNumber(11)
  set updateTime($1.Timestamp value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasUpdateTime() => $_has(9);
  @$pb.TagNumber(11)
  void clearUpdateTime() => $_clearField(11);
  @$pb.TagNumber(11)
  $1.Timestamp ensureUpdateTime() => $_ensure(9);
}

/// MobileState is the phone-and-watch pair's current status.
class MobileState extends $pb.GeneratedMessage {
  factory MobileState({
    $core.String? deviceUid,
    $core.String? displayName,
    $core.String? model,
    $0.PhoneState? phone,
    $2.WatchSnapshot? watch,
    $1.Timestamp? updateTime,
    $0.SwitchPresence? switchPresence,
  }) {
    final result = create();
    if (deviceUid != null) result.deviceUid = deviceUid;
    if (displayName != null) result.displayName = displayName;
    if (model != null) result.model = model;
    if (phone != null) result.phone = phone;
    if (watch != null) result.watch = watch;
    if (updateTime != null) result.updateTime = updateTime;
    if (switchPresence != null) result.switchPresence = switchPresence;
    return result;
  }

  MobileState._();

  factory MobileState.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory MobileState.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MobileState',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'deviceUid')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..aOS(3, _omitFieldNames ? '' : 'model')
    ..aOM<$0.PhoneState>(4, _omitFieldNames ? '' : 'phone',
        subBuilder: $0.PhoneState.create)
    ..aOM<$2.WatchSnapshot>(5, _omitFieldNames ? '' : 'watch',
        subBuilder: $2.WatchSnapshot.create)
    ..aOM<$1.Timestamp>(6, _omitFieldNames ? '' : 'updateTime',
        subBuilder: $1.Timestamp.create)
    ..aOM<$0.SwitchPresence>(7, _omitFieldNames ? '' : 'switchPresence',
        subBuilder: $0.SwitchPresence.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MobileState clone() => MobileState()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MobileState copyWith(void Function(MobileState) updates) =>
      super.copyWith((message) => updates(message as MobileState))
          as MobileState;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MobileState create() => MobileState._();
  @$core.override
  MobileState createEmptyInstance() => create();
  static $pb.PbList<MobileState> createRepeated() => $pb.PbList<MobileState>();
  @$core.pragma('dart2js:noInline')
  static MobileState getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MobileState>(create);
  static MobileState? _defaultInstance;

  /// device_uid is the gateway-assigned phone identity.
  @$pb.TagNumber(1)
  $core.String get deviceUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set deviceUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDeviceUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearDeviceUid() => $_clearField(1);

  /// display_name is the human-readable phone label.
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);

  /// model is the phone hardware model.
  @$pb.TagNumber(3)
  $core.String get model => $_getSZ(2);
  @$pb.TagNumber(3)
  set model($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasModel() => $_has(2);
  @$pb.TagNumber(3)
  void clearModel() => $_clearField(3);

  /// phone is the phone's own device state.
  @$pb.TagNumber(4)
  $0.PhoneState get phone => $_getN(3);
  @$pb.TagNumber(4)
  set phone($0.PhoneState value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasPhone() => $_has(3);
  @$pb.TagNumber(4)
  void clearPhone() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.PhoneState ensurePhone() => $_ensure(3);

  /// watch is the latest paired-watch snapshot.
  @$pb.TagNumber(5)
  $2.WatchSnapshot get watch => $_getN(4);
  @$pb.TagNumber(5)
  set watch($2.WatchSnapshot value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasWatch() => $_has(4);
  @$pb.TagNumber(5)
  void clearWatch() => $_clearField(5);
  @$pb.TagNumber(5)
  $2.WatchSnapshot ensureWatch() => $_ensure(4);

  /// update_time is when this status was last refreshed.
  @$pb.TagNumber(6)
  $1.Timestamp get updateTime => $_getN(5);
  @$pb.TagNumber(6)
  set updateTime($1.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdateTime() => $_has(5);
  @$pb.TagNumber(6)
  void clearUpdateTime() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Timestamp ensureUpdateTime() => $_ensure(5);

  /// switch_presence is the owner's Nintendo Switch play presence, if the phone
  /// can read the local Nintendo Switch Online app state.
  @$pb.TagNumber(7)
  $0.SwitchPresence get switchPresence => $_getN(6);
  @$pb.TagNumber(7)
  set switchPresence($0.SwitchPresence value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasSwitchPresence() => $_has(6);
  @$pb.TagNumber(7)
  void clearSwitchPresence() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.SwitchPresence ensureSwitchPresence() => $_ensure(6);
}

/// Subagent is one worker a coding agent has running right now. It carries the
/// model and nothing else: what a sub-agent was asked to do is never collected.
class Subagent extends $pb.GeneratedMessage {
  factory Subagent({
    $core.String? model,
  }) {
    final result = create();
    if (model != null) result.model = model;
    return result;
  }

  Subagent._();

  factory Subagent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Subagent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Subagent',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'model')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Subagent clone() => Subagent()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Subagent copyWith(void Function(Subagent) updates) =>
      super.copyWith((message) => updates(message as Subagent)) as Subagent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Subagent create() => Subagent._();
  @$core.override
  Subagent createEmptyInstance() => create();
  static $pb.PbList<Subagent> createRepeated() => $pb.PbList<Subagent>();
  @$core.pragma('dart2js:noInline')
  static Subagent getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Subagent>(create);
  static Subagent? _defaultInstance;

  /// model is the model the sub-agent runs, such as "claude-opus-4-8".
  @$pb.TagNumber(1)
  $core.String get model => $_getSZ(0);
  @$pb.TagNumber(1)
  set model($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasModel() => $_has(0);
  @$pb.TagNumber(1)
  void clearModel() => $_clearField(1);
}

/// Agent is a coding agent's current status on a read surface. It never exposes
/// the underlying task text or any internal session identifier.
class Agent extends $pb.GeneratedMessage {
  factory Agent({
    $core.String? uid,
    $core.String? kind,
    $core.String? deviceUid,
    $core.String? displayName,
    $0.AgentState? state,
    $core.int? budgetRemainingPercent,
    $1.Timestamp? updateTime,
    $core.String? model,
    $core.Iterable<Subagent>? subagents,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (kind != null) result.kind = kind;
    if (deviceUid != null) result.deviceUid = deviceUid;
    if (displayName != null) result.displayName = displayName;
    if (state != null) result.state = state;
    if (budgetRemainingPercent != null)
      result.budgetRemainingPercent = budgetRemainingPercent;
    if (updateTime != null) result.updateTime = updateTime;
    if (model != null) result.model = model;
    if (subagents != null) result.subagents.addAll(subagents);
    return result;
  }

  Agent._();

  factory Agent.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Agent.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Agent',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aOS(2, _omitFieldNames ? '' : 'kind')
    ..aOS(3, _omitFieldNames ? '' : 'deviceUid')
    ..aOS(4, _omitFieldNames ? '' : 'displayName')
    ..e<$0.AgentState>(5, _omitFieldNames ? '' : 'state', $pb.PbFieldType.OE,
        defaultOrMaker: $0.AgentState.AGENT_STATE_UNSPECIFIED,
        valueOf: $0.AgentState.valueOf,
        enumValues: $0.AgentState.values)
    ..a<$core.int>(
        6, _omitFieldNames ? '' : 'budgetRemainingPercent', $pb.PbFieldType.O3)
    ..aOM<$1.Timestamp>(7, _omitFieldNames ? '' : 'updateTime',
        subBuilder: $1.Timestamp.create)
    ..aOS(8, _omitFieldNames ? '' : 'model')
    ..pc<Subagent>(10, _omitFieldNames ? '' : 'subagents', $pb.PbFieldType.PM,
        subBuilder: Subagent.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Agent clone() => Agent()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Agent copyWith(void Function(Agent) updates) =>
      super.copyWith((message) => updates(message as Agent)) as Agent;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Agent create() => Agent._();
  @$core.override
  Agent createEmptyInstance() => create();
  static $pb.PbList<Agent> createRepeated() => $pb.PbList<Agent>();
  @$core.pragma('dart2js:noInline')
  static Agent getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Agent>(create);
  static Agent? _defaultInstance;

  /// uid is the gateway-assigned opaque agent identity.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  /// kind is the agent kind, such as "codex" or "claude".
  @$pb.TagNumber(2)
  $core.String get kind => $_getSZ(1);
  @$pb.TagNumber(2)
  set kind($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasKind() => $_has(1);
  @$pb.TagNumber(2)
  void clearKind() => $_clearField(2);

  /// device_uid is the host the agent runs on.
  @$pb.TagNumber(3)
  $core.String get deviceUid => $_getSZ(2);
  @$pb.TagNumber(3)
  set deviceUid($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDeviceUid() => $_has(2);
  @$pb.TagNumber(3)
  void clearDeviceUid() => $_clearField(3);

  /// display_name is the human-readable host label.
  @$pb.TagNumber(4)
  $core.String get displayName => $_getSZ(3);
  @$pb.TagNumber(4)
  set displayName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDisplayName() => $_has(3);
  @$pb.TagNumber(4)
  void clearDisplayName() => $_clearField(4);

  /// state is the agent's run state.
  @$pb.TagNumber(5)
  $0.AgentState get state => $_getN(4);
  @$pb.TagNumber(5)
  set state($0.AgentState value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasState() => $_has(4);
  @$pb.TagNumber(5)
  void clearState() => $_clearField(5);

  /// budget_remaining_percent is the remaining budget from 0 to 100, if known.
  @$pb.TagNumber(6)
  $core.int get budgetRemainingPercent => $_getIZ(5);
  @$pb.TagNumber(6)
  set budgetRemainingPercent($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasBudgetRemainingPercent() => $_has(5);
  @$pb.TagNumber(6)
  void clearBudgetRemainingPercent() => $_clearField(6);

  /// update_time is when this agent's status was last refreshed.
  @$pb.TagNumber(7)
  $1.Timestamp get updateTime => $_getN(6);
  @$pb.TagNumber(7)
  set updateTime($1.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasUpdateTime() => $_has(6);
  @$pb.TagNumber(7)
  void clearUpdateTime() => $_clearField(7);
  @$pb.TagNumber(7)
  $1.Timestamp ensureUpdateTime() => $_ensure(6);

  /// model is the model the agent is currently running, such as "claude-opus-4-8".
  @$pb.TagNumber(8)
  $core.String get model => $_getSZ(7);
  @$pb.TagNumber(8)
  set model($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasModel() => $_has(7);
  @$pb.TagNumber(8)
  void clearModel() => $_clearField(8);

  /// subagents are the workers the agent has running right now, one per worker.
  @$pb.TagNumber(10)
  $pb.PbList<Subagent> get subagents => $_getList(8);
}

/// GithubStatus is the public view of GitHub status synchronization.
class GithubStatus extends $pb.GeneratedMessage {
  factory GithubStatus({
    $core.bool? enabled,
    GithubSyncState? state,
    $core.String? emoji,
    $core.String? message,
    $1.Timestamp? updateTime,
  }) {
    final result = create();
    if (enabled != null) result.enabled = enabled;
    if (state != null) result.state = state;
    if (emoji != null) result.emoji = emoji;
    if (message != null) result.message = message;
    if (updateTime != null) result.updateTime = updateTime;
    return result;
  }

  GithubStatus._();

  factory GithubStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GithubStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GithubStatus',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'enabled')
    ..e<GithubSyncState>(2, _omitFieldNames ? '' : 'state', $pb.PbFieldType.OE,
        defaultOrMaker: GithubSyncState.GITHUB_SYNC_STATE_UNSPECIFIED,
        valueOf: GithubSyncState.valueOf,
        enumValues: GithubSyncState.values)
    ..aOS(3, _omitFieldNames ? '' : 'emoji')
    ..aOS(4, _omitFieldNames ? '' : 'message')
    ..aOM<$1.Timestamp>(5, _omitFieldNames ? '' : 'updateTime',
        subBuilder: $1.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GithubStatus clone() => GithubStatus()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GithubStatus copyWith(void Function(GithubStatus) updates) =>
      super.copyWith((message) => updates(message as GithubStatus))
          as GithubStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GithubStatus create() => GithubStatus._();
  @$core.override
  GithubStatus createEmptyInstance() => create();
  static $pb.PbList<GithubStatus> createRepeated() =>
      $pb.PbList<GithubStatus>();
  @$core.pragma('dart2js:noInline')
  static GithubStatus getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GithubStatus>(create);
  static GithubStatus? _defaultInstance;

  /// enabled is whether GitHub synchronization is configured.
  @$pb.TagNumber(1)
  $core.bool get enabled => $_getBF(0);
  @$pb.TagNumber(1)
  set enabled($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEnabled() => $_has(0);
  @$pb.TagNumber(1)
  void clearEnabled() => $_clearField(1);

  /// state is the current synchronization state.
  @$pb.TagNumber(2)
  GithubSyncState get state => $_getN(1);
  @$pb.TagNumber(2)
  set state(GithubSyncState value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasState() => $_has(1);
  @$pb.TagNumber(2)
  void clearState() => $_clearField(2);

  /// emoji is the emoji shown on the current GitHub status.
  @$pb.TagNumber(3)
  $core.String get emoji => $_getSZ(2);
  @$pb.TagNumber(3)
  set emoji($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEmoji() => $_has(2);
  @$pb.TagNumber(3)
  void clearEmoji() => $_clearField(3);

  /// message is the message shown on the current GitHub status.
  @$pb.TagNumber(4)
  $core.String get message => $_getSZ(3);
  @$pb.TagNumber(4)
  set message($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMessage() => $_has(3);
  @$pb.TagNumber(4)
  void clearMessage() => $_clearField(4);

  /// update_time is when the GitHub status last synchronized successfully.
  @$pb.TagNumber(5)
  $1.Timestamp get updateTime => $_getN(4);
  @$pb.TagNumber(5)
  set updateTime($1.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasUpdateTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearUpdateTime() => $_clearField(5);
  @$pb.TagNumber(5)
  $1.Timestamp ensureUpdateTime() => $_ensure(4);
}

/// GithubSyncDetail is the authenticated view including error diagnostics.
class GithubSyncDetail extends $pb.GeneratedMessage {
  factory GithubSyncDetail({
    $core.bool? configured,
    GithubSyncState? state,
    $core.String? emoji,
    $core.String? message,
    $1.Timestamp? lastSuccessTime,
    $1.Timestamp? lastErrorTime,
    $core.String? lastError,
    $1.Timestamp? lastAttemptTime,
    $core.String? lastSignature,
  }) {
    final result = create();
    if (configured != null) result.configured = configured;
    if (state != null) result.state = state;
    if (emoji != null) result.emoji = emoji;
    if (message != null) result.message = message;
    if (lastSuccessTime != null) result.lastSuccessTime = lastSuccessTime;
    if (lastErrorTime != null) result.lastErrorTime = lastErrorTime;
    if (lastError != null) result.lastError = lastError;
    if (lastAttemptTime != null) result.lastAttemptTime = lastAttemptTime;
    if (lastSignature != null) result.lastSignature = lastSignature;
    return result;
  }

  GithubSyncDetail._();

  factory GithubSyncDetail.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GithubSyncDetail.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GithubSyncDetail',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'configured')
    ..e<GithubSyncState>(2, _omitFieldNames ? '' : 'state', $pb.PbFieldType.OE,
        defaultOrMaker: GithubSyncState.GITHUB_SYNC_STATE_UNSPECIFIED,
        valueOf: GithubSyncState.valueOf,
        enumValues: GithubSyncState.values)
    ..aOS(3, _omitFieldNames ? '' : 'emoji')
    ..aOS(4, _omitFieldNames ? '' : 'message')
    ..aOM<$1.Timestamp>(5, _omitFieldNames ? '' : 'lastSuccessTime',
        subBuilder: $1.Timestamp.create)
    ..aOM<$1.Timestamp>(6, _omitFieldNames ? '' : 'lastErrorTime',
        subBuilder: $1.Timestamp.create)
    ..aOS(7, _omitFieldNames ? '' : 'lastError')
    ..aOM<$1.Timestamp>(8, _omitFieldNames ? '' : 'lastAttemptTime',
        subBuilder: $1.Timestamp.create)
    ..aOS(9, _omitFieldNames ? '' : 'lastSignature')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GithubSyncDetail clone() => GithubSyncDetail()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GithubSyncDetail copyWith(void Function(GithubSyncDetail) updates) =>
      super.copyWith((message) => updates(message as GithubSyncDetail))
          as GithubSyncDetail;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GithubSyncDetail create() => GithubSyncDetail._();
  @$core.override
  GithubSyncDetail createEmptyInstance() => create();
  static $pb.PbList<GithubSyncDetail> createRepeated() =>
      $pb.PbList<GithubSyncDetail>();
  @$core.pragma('dart2js:noInline')
  static GithubSyncDetail getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GithubSyncDetail>(create);
  static GithubSyncDetail? _defaultInstance;

  /// configured is whether a GitHub token is present.
  @$pb.TagNumber(1)
  $core.bool get configured => $_getBF(0);
  @$pb.TagNumber(1)
  set configured($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasConfigured() => $_has(0);
  @$pb.TagNumber(1)
  void clearConfigured() => $_clearField(1);

  /// state is the current synchronization state.
  @$pb.TagNumber(2)
  GithubSyncState get state => $_getN(1);
  @$pb.TagNumber(2)
  set state(GithubSyncState value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasState() => $_has(1);
  @$pb.TagNumber(2)
  void clearState() => $_clearField(2);

  /// emoji is the emoji on the current GitHub status.
  @$pb.TagNumber(3)
  $core.String get emoji => $_getSZ(2);
  @$pb.TagNumber(3)
  set emoji($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEmoji() => $_has(2);
  @$pb.TagNumber(3)
  void clearEmoji() => $_clearField(3);

  /// message is the message on the current GitHub status.
  @$pb.TagNumber(4)
  $core.String get message => $_getSZ(3);
  @$pb.TagNumber(4)
  set message($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMessage() => $_has(3);
  @$pb.TagNumber(4)
  void clearMessage() => $_clearField(4);

  /// last_success_time is when synchronization last succeeded.
  @$pb.TagNumber(5)
  $1.Timestamp get lastSuccessTime => $_getN(4);
  @$pb.TagNumber(5)
  set lastSuccessTime($1.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasLastSuccessTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearLastSuccessTime() => $_clearField(5);
  @$pb.TagNumber(5)
  $1.Timestamp ensureLastSuccessTime() => $_ensure(4);

  /// last_error_time is when synchronization last failed.
  @$pb.TagNumber(6)
  $1.Timestamp get lastErrorTime => $_getN(5);
  @$pb.TagNumber(6)
  set lastErrorTime($1.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasLastErrorTime() => $_has(5);
  @$pb.TagNumber(6)
  void clearLastErrorTime() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Timestamp ensureLastErrorTime() => $_ensure(5);

  /// last_error is the most recent synchronization error.
  @$pb.TagNumber(7)
  $core.String get lastError => $_getSZ(6);
  @$pb.TagNumber(7)
  set lastError($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasLastError() => $_has(6);
  @$pb.TagNumber(7)
  void clearLastError() => $_clearField(7);

  /// last_attempt_time is when synchronization was last attempted; it drives the
  /// minimum-interval rate limit.
  @$pb.TagNumber(8)
  $1.Timestamp get lastAttemptTime => $_getN(7);
  @$pb.TagNumber(8)
  set lastAttemptTime($1.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasLastAttemptTime() => $_has(7);
  @$pb.TagNumber(8)
  void clearLastAttemptTime() => $_clearField(8);
  @$pb.TagNumber(8)
  $1.Timestamp ensureLastAttemptTime() => $_ensure(7);

  /// last_signature is a fingerprint of the last published status used to skip
  /// redundant updates.
  @$pb.TagNumber(9)
  $core.String get lastSignature => $_getSZ(8);
  @$pb.TagNumber(9)
  set lastSignature($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasLastSignature() => $_has(8);
  @$pb.TagNumber(9)
  void clearLastSignature() => $_clearField(9);
}

/// PublicStatus is the unauthenticated status document.
class PublicStatus extends $pb.GeneratedMessage {
  factory PublicStatus({
    DeviceState? server,
    $core.Iterable<DeviceState>? devices,
    $core.Iterable<Agent>? agents,
    GithubStatus? github,
    $1.Timestamp? updateTime,
    $core.Iterable<MobileState>? mobiles,
  }) {
    final result = create();
    if (server != null) result.server = server;
    if (devices != null) result.devices.addAll(devices);
    if (agents != null) result.agents.addAll(agents);
    if (github != null) result.github = github;
    if (updateTime != null) result.updateTime = updateTime;
    if (mobiles != null) result.mobiles.addAll(mobiles);
    return result;
  }

  PublicStatus._();

  factory PublicStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PublicStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PublicStatus',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOM<DeviceState>(1, _omitFieldNames ? '' : 'server',
        subBuilder: DeviceState.create)
    ..pc<DeviceState>(3, _omitFieldNames ? '' : 'devices', $pb.PbFieldType.PM,
        subBuilder: DeviceState.create)
    ..pc<Agent>(4, _omitFieldNames ? '' : 'agents', $pb.PbFieldType.PM,
        subBuilder: Agent.create)
    ..aOM<GithubStatus>(5, _omitFieldNames ? '' : 'github',
        subBuilder: GithubStatus.create)
    ..aOM<$1.Timestamp>(6, _omitFieldNames ? '' : 'updateTime',
        subBuilder: $1.Timestamp.create)
    ..pc<MobileState>(7, _omitFieldNames ? '' : 'mobiles', $pb.PbFieldType.PM,
        subBuilder: MobileState.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PublicStatus clone() => PublicStatus()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PublicStatus copyWith(void Function(PublicStatus) updates) =>
      super.copyWith((message) => updates(message as PublicStatus))
          as PublicStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PublicStatus create() => PublicStatus._();
  @$core.override
  PublicStatus createEmptyInstance() => create();
  static $pb.PbList<PublicStatus> createRepeated() =>
      $pb.PbList<PublicStatus>();
  @$core.pragma('dart2js:noInline')
  static PublicStatus getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PublicStatus>(create);
  static PublicStatus? _defaultInstance;

  /// server is the always-on server's status.
  @$pb.TagNumber(1)
  DeviceState get server => $_getN(0);
  @$pb.TagNumber(1)
  set server(DeviceState value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearServer() => $_clearField(1);
  @$pb.TagNumber(1)
  DeviceState ensureServer() => $_ensure(0);

  /// devices are the non-server devices' statuses.
  @$pb.TagNumber(3)
  $pb.PbList<DeviceState> get devices => $_getList(1);

  /// agents are the currently visible coding agents.
  @$pb.TagNumber(4)
  $pb.PbList<Agent> get agents => $_getList(2);

  /// github is the public GitHub synchronization status.
  @$pb.TagNumber(5)
  GithubStatus get github => $_getN(3);
  @$pb.TagNumber(5)
  set github(GithubStatus value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasGithub() => $_has(3);
  @$pb.TagNumber(5)
  void clearGithub() => $_clearField(5);
  @$pb.TagNumber(5)
  GithubStatus ensureGithub() => $_ensure(3);

  /// update_time is when this document was assembled.
  @$pb.TagNumber(6)
  $1.Timestamp get updateTime => $_getN(4);
  @$pb.TagNumber(6)
  set updateTime($1.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdateTime() => $_has(4);
  @$pb.TagNumber(6)
  void clearUpdateTime() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Timestamp ensureUpdateTime() => $_ensure(4);

  /// mobiles are the independently reporting phone-and-watch pairs.
  @$pb.TagNumber(7)
  $pb.PbList<MobileState> get mobiles => $_getList(5);
}

/// InternalStatus is the authenticated status document with sync diagnostics.
class InternalStatus extends $pb.GeneratedMessage {
  factory InternalStatus({
    DeviceState? server,
    $core.Iterable<DeviceState>? devices,
    $core.Iterable<Agent>? agents,
    GithubSyncDetail? github,
    $1.Timestamp? updateTime,
    $core.Iterable<MobileState>? mobiles,
  }) {
    final result = create();
    if (server != null) result.server = server;
    if (devices != null) result.devices.addAll(devices);
    if (agents != null) result.agents.addAll(agents);
    if (github != null) result.github = github;
    if (updateTime != null) result.updateTime = updateTime;
    if (mobiles != null) result.mobiles.addAll(mobiles);
    return result;
  }

  InternalStatus._();

  factory InternalStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory InternalStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InternalStatus',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOM<DeviceState>(1, _omitFieldNames ? '' : 'server',
        subBuilder: DeviceState.create)
    ..pc<DeviceState>(3, _omitFieldNames ? '' : 'devices', $pb.PbFieldType.PM,
        subBuilder: DeviceState.create)
    ..pc<Agent>(4, _omitFieldNames ? '' : 'agents', $pb.PbFieldType.PM,
        subBuilder: Agent.create)
    ..aOM<GithubSyncDetail>(5, _omitFieldNames ? '' : 'github',
        subBuilder: GithubSyncDetail.create)
    ..aOM<$1.Timestamp>(6, _omitFieldNames ? '' : 'updateTime',
        subBuilder: $1.Timestamp.create)
    ..pc<MobileState>(7, _omitFieldNames ? '' : 'mobiles', $pb.PbFieldType.PM,
        subBuilder: MobileState.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InternalStatus clone() => InternalStatus()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InternalStatus copyWith(void Function(InternalStatus) updates) =>
      super.copyWith((message) => updates(message as InternalStatus))
          as InternalStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static InternalStatus create() => InternalStatus._();
  @$core.override
  InternalStatus createEmptyInstance() => create();
  static $pb.PbList<InternalStatus> createRepeated() =>
      $pb.PbList<InternalStatus>();
  @$core.pragma('dart2js:noInline')
  static InternalStatus getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InternalStatus>(create);
  static InternalStatus? _defaultInstance;

  /// server is the always-on server's status.
  @$pb.TagNumber(1)
  DeviceState get server => $_getN(0);
  @$pb.TagNumber(1)
  set server(DeviceState value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasServer() => $_has(0);
  @$pb.TagNumber(1)
  void clearServer() => $_clearField(1);
  @$pb.TagNumber(1)
  DeviceState ensureServer() => $_ensure(0);

  /// devices are the non-server devices' statuses.
  @$pb.TagNumber(3)
  $pb.PbList<DeviceState> get devices => $_getList(1);

  /// agents are the currently visible coding agents.
  @$pb.TagNumber(4)
  $pb.PbList<Agent> get agents => $_getList(2);

  /// github is the detailed GitHub synchronization status.
  @$pb.TagNumber(5)
  GithubSyncDetail get github => $_getN(3);
  @$pb.TagNumber(5)
  set github(GithubSyncDetail value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasGithub() => $_has(3);
  @$pb.TagNumber(5)
  void clearGithub() => $_clearField(5);
  @$pb.TagNumber(5)
  GithubSyncDetail ensureGithub() => $_ensure(3);

  /// update_time is when this document was assembled.
  @$pb.TagNumber(6)
  $1.Timestamp get updateTime => $_getN(4);
  @$pb.TagNumber(6)
  set updateTime($1.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasUpdateTime() => $_has(4);
  @$pb.TagNumber(6)
  void clearUpdateTime() => $_clearField(6);
  @$pb.TagNumber(6)
  $1.Timestamp ensureUpdateTime() => $_ensure(4);

  /// mobiles are the independently reporting phone-and-watch pairs.
  @$pb.TagNumber(7)
  $pb.PbList<MobileState> get mobiles => $_getList(5);
}

/// StatusService serves the aggregated status surfaces.
class StatusServiceApi {
  final $pb.RpcClient _client;

  StatusServiceApi(this._client);

  /// GetPublicStatus returns the unauthenticated status document.
  $async.Future<GetPublicStatusResponse> getPublicStatus(
          $pb.ClientContext? ctx, GetPublicStatusRequest request) =>
      _client.invoke<GetPublicStatusResponse>(ctx, 'StatusService',
          'GetPublicStatus', request, GetPublicStatusResponse());

  /// GetInternalStatus returns the authenticated status document with sync
  /// diagnostics.
  $async.Future<GetInternalStatusResponse> getInternalStatus(
          $pb.ClientContext? ctx, GetInternalStatusRequest request) =>
      _client.invoke<GetInternalStatusResponse>(ctx, 'StatusService',
          'GetInternalStatus', request, GetInternalStatusResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
