// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/status_types.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../../../google/protobuf/timestamp.pb.dart' as $0;
import 'status_types.pbenum.dart';
import 'watch.pbenum.dart' as $1;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'status_types.pbenum.dart';

/// Accessory is a connected peripheral, such as a Bluetooth audio device.
class Accessory extends $pb.GeneratedMessage {
  factory Accessory({
    $core.String? kind,
    $core.String? displayName,
    $core.String? model,
    $core.int? batteryPercent,
  }) {
    final result = create();
    if (kind != null) result.kind = kind;
    if (displayName != null) result.displayName = displayName;
    if (model != null) result.model = model;
    if (batteryPercent != null) result.batteryPercent = batteryPercent;
    return result;
  }

  Accessory._();

  factory Accessory.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Accessory.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Accessory',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'kind')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..aOS(3, _omitFieldNames ? '' : 'model')
    ..a<$core.int>(
        4, _omitFieldNames ? '' : 'batteryPercent', $pb.PbFieldType.O3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Accessory clone() => Accessory()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Accessory copyWith(void Function(Accessory) updates) =>
      super.copyWith((message) => updates(message as Accessory)) as Accessory;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Accessory create() => Accessory._();
  @$core.override
  Accessory createEmptyInstance() => create();
  static $pb.PbList<Accessory> createRepeated() => $pb.PbList<Accessory>();
  @$core.pragma('dart2js:noInline')
  static Accessory getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Accessory>(create);
  static Accessory? _defaultInstance;

  /// kind is a lowercase accessory category key, such as "bluetooth_audio".
  @$pb.TagNumber(1)
  $core.String get kind => $_getSZ(0);
  @$pb.TagNumber(1)
  set kind($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasKind() => $_has(0);
  @$pb.TagNumber(1)
  void clearKind() => $_clearField(1);

  /// display_name is the human-readable accessory label.
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);

  /// model is the accessory hardware model, if known.
  @$pb.TagNumber(3)
  $core.String get model => $_getSZ(2);
  @$pb.TagNumber(3)
  set model($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasModel() => $_has(2);
  @$pb.TagNumber(3)
  void clearModel() => $_clearField(3);

  /// battery_percent is the accessory battery level from 0 to 100, if reported.
  @$pb.TagNumber(4)
  $core.int get batteryPercent => $_getIZ(3);
  @$pb.TagNumber(4)
  set batteryPercent($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasBatteryPercent() => $_has(3);
  @$pb.TagNumber(4)
  void clearBatteryPercent() => $_clearField(4);
}

/// MediaStatus is the media item a device is currently playing.
class MediaStatus extends $pb.GeneratedMessage {
  factory MediaStatus({
    $core.String? title,
    $core.String? artist,
  }) {
    final result = create();
    if (title != null) result.title = title;
    if (artist != null) result.artist = artist;
    return result;
  }

  MediaStatus._();

  factory MediaStatus.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory MediaStatus.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MediaStatus',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'title')
    ..aOS(2, _omitFieldNames ? '' : 'artist')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MediaStatus clone() => MediaStatus()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MediaStatus copyWith(void Function(MediaStatus) updates) =>
      super.copyWith((message) => updates(message as MediaStatus))
          as MediaStatus;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MediaStatus create() => MediaStatus._();
  @$core.override
  MediaStatus createEmptyInstance() => create();
  static $pb.PbList<MediaStatus> createRepeated() => $pb.PbList<MediaStatus>();
  @$core.pragma('dart2js:noInline')
  static MediaStatus getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MediaStatus>(create);
  static MediaStatus? _defaultInstance;

  /// title is the track or media title.
  @$pb.TagNumber(1)
  $core.String get title => $_getSZ(0);
  @$pb.TagNumber(1)
  set title($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTitle() => $_has(0);
  @$pb.TagNumber(1)
  void clearTitle() => $_clearField(1);

  /// artist is the performing artist, if known.
  @$pb.TagNumber(2)
  $core.String get artist => $_getSZ(1);
  @$pb.TagNumber(2)
  set artist($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasArtist() => $_has(1);
  @$pb.TagNumber(2)
  void clearArtist() => $_clearField(2);
}

/// SwitchPresence is the Nintendo Switch play presence reported by the owner's
/// Android device. It intentionally carries only public presence facts, never
/// Nintendo tokens or account identifiers.
class SwitchPresence extends $pb.GeneratedMessage {
  factory SwitchPresence({
    OnlineState? state,
    $core.String? gameName,
    $core.String? titleId,
    $core.String? imageUri,
    $0.Timestamp? presenceUpdateTime,
    $0.Timestamp? logoutTime,
    $0.Timestamp? fetchTime,
  }) {
    final result = create();
    if (state != null) result.state = state;
    if (gameName != null) result.gameName = gameName;
    if (titleId != null) result.titleId = titleId;
    if (imageUri != null) result.imageUri = imageUri;
    if (presenceUpdateTime != null)
      result.presenceUpdateTime = presenceUpdateTime;
    if (logoutTime != null) result.logoutTime = logoutTime;
    if (fetchTime != null) result.fetchTime = fetchTime;
    return result;
  }

  SwitchPresence._();

  factory SwitchPresence.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SwitchPresence.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SwitchPresence',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..e<OnlineState>(1, _omitFieldNames ? '' : 'state', $pb.PbFieldType.OE,
        defaultOrMaker: OnlineState.ONLINE_STATE_UNSPECIFIED,
        valueOf: OnlineState.valueOf,
        enumValues: OnlineState.values)
    ..aOS(2, _omitFieldNames ? '' : 'gameName')
    ..aOS(3, _omitFieldNames ? '' : 'titleId')
    ..aOS(4, _omitFieldNames ? '' : 'imageUri')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'presenceUpdateTime',
        subBuilder: $0.Timestamp.create)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'logoutTime',
        subBuilder: $0.Timestamp.create)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'fetchTime',
        subBuilder: $0.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SwitchPresence clone() => SwitchPresence()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SwitchPresence copyWith(void Function(SwitchPresence) updates) =>
      super.copyWith((message) => updates(message as SwitchPresence))
          as SwitchPresence;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SwitchPresence create() => SwitchPresence._();
  @$core.override
  SwitchPresence createEmptyInstance() => create();
  static $pb.PbList<SwitchPresence> createRepeated() =>
      $pb.PbList<SwitchPresence>();
  @$core.pragma('dart2js:noInline')
  static SwitchPresence getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SwitchPresence>(create);
  static SwitchPresence? _defaultInstance;

  /// state is whether the Switch presence is currently online.
  @$pb.TagNumber(1)
  OnlineState get state => $_getN(0);
  @$pb.TagNumber(1)
  set state(OnlineState value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasState() => $_has(0);
  @$pb.TagNumber(1)
  void clearState() => $_clearField(1);

  /// game_name is the currently played title, if Nintendo reports one.
  @$pb.TagNumber(2)
  $core.String get gameName => $_getSZ(1);
  @$pb.TagNumber(2)
  set gameName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasGameName() => $_has(1);
  @$pb.TagNumber(2)
  void clearGameName() => $_clearField(2);

  /// title_id is Nintendo's title identifier for the current game, if present.
  @$pb.TagNumber(3)
  $core.String get titleId => $_getSZ(2);
  @$pb.TagNumber(3)
  set titleId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasTitleId() => $_has(2);
  @$pb.TagNumber(3)
  void clearTitleId() => $_clearField(3);

  /// image_uri is the title image Nintendo returns for the current game.
  @$pb.TagNumber(4)
  $core.String get imageUri => $_getSZ(3);
  @$pb.TagNumber(4)
  set imageUri($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasImageUri() => $_has(3);
  @$pb.TagNumber(4)
  void clearImageUri() => $_clearField(4);

  /// presence_update_time is Nintendo's timestamp for the presence record.
  @$pb.TagNumber(5)
  $0.Timestamp get presenceUpdateTime => $_getN(4);
  @$pb.TagNumber(5)
  set presenceUpdateTime($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasPresenceUpdateTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearPresenceUpdateTime() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensurePresenceUpdateTime() => $_ensure(4);

  /// logout_time is Nintendo's last logout timestamp, when reported.
  @$pb.TagNumber(6)
  $0.Timestamp get logoutTime => $_getN(5);
  @$pb.TagNumber(6)
  set logoutTime($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasLogoutTime() => $_has(5);
  @$pb.TagNumber(6)
  void clearLogoutTime() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureLogoutTime() => $_ensure(5);

  /// fetch_time is when the reporting APK fetched this presence.
  @$pb.TagNumber(7)
  $0.Timestamp get fetchTime => $_getN(6);
  @$pb.TagNumber(7)
  set fetchTime($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasFetchTime() => $_has(6);
  @$pb.TagNumber(7)
  void clearFetchTime() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureFetchTime() => $_ensure(6);
}

/// MetricSample is one OpenTelemetry-named numeric sample reported by a host.
class MetricSample extends $pb.GeneratedMessage {
  factory MetricSample({
    $core.String? name,
    $core.String? unit,
    $core.double? value,
    $core.Iterable<$core.MapEntry<$core.String, $core.String>>? attributes,
  }) {
    final result = create();
    if (name != null) result.name = name;
    if (unit != null) result.unit = unit;
    if (value != null) result.value = value;
    if (attributes != null) result.attributes.addEntries(attributes);
    return result;
  }

  MetricSample._();

  factory MetricSample.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory MetricSample.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MetricSample',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aOS(2, _omitFieldNames ? '' : 'unit')
    ..a<$core.double>(3, _omitFieldNames ? '' : 'value', $pb.PbFieldType.OD)
    ..m<$core.String, $core.String>(4, _omitFieldNames ? '' : 'attributes',
        entryClassName: 'MetricSample.AttributesEntry',
        keyFieldType: $pb.PbFieldType.OS,
        valueFieldType: $pb.PbFieldType.OS,
        packageName: const $pb.PackageName('realtime.me.status.v1'))
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MetricSample clone() => MetricSample()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MetricSample copyWith(void Function(MetricSample) updates) =>
      super.copyWith((message) => updates(message as MetricSample))
          as MetricSample;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MetricSample create() => MetricSample._();
  @$core.override
  MetricSample createEmptyInstance() => create();
  static $pb.PbList<MetricSample> createRepeated() =>
      $pb.PbList<MetricSample>();
  @$core.pragma('dart2js:noInline')
  static MetricSample getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MetricSample>(create);
  static MetricSample? _defaultInstance;

  /// name is the OpenTelemetry metric name, such as "system.cpu.utilization".
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// unit is the UCUM unit annotation, such as "By" or "1".
  @$pb.TagNumber(2)
  $core.String get unit => $_getSZ(1);
  @$pb.TagNumber(2)
  set unit($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUnit() => $_has(1);
  @$pb.TagNumber(2)
  void clearUnit() => $_clearField(2);

  /// value is the sample value.
  @$pb.TagNumber(3)
  $core.double get value => $_getN(2);
  @$pb.TagNumber(3)
  set value($core.double value) => $_setDouble(2, value);
  @$pb.TagNumber(3)
  $core.bool hasValue() => $_has(2);
  @$pb.TagNumber(3)
  void clearValue() => $_clearField(3);

  /// attributes are OpenTelemetry attributes attached to the sample.
  @$pb.TagNumber(4)
  $pb.PbMap<$core.String, $core.String> get attributes => $_getMap(3);
}

/// PhoneState is the phone's own device state, excluding the paired watch.
class PhoneState extends $pb.GeneratedMessage {
  factory PhoneState({
    $core.int? batteryPercent,
    $1.ChargeState? chargeState,
    NetworkState? network,
    $core.Iterable<Accessory>? accessories,
  }) {
    final result = create();
    if (batteryPercent != null) result.batteryPercent = batteryPercent;
    if (chargeState != null) result.chargeState = chargeState;
    if (network != null) result.network = network;
    if (accessories != null) result.accessories.addAll(accessories);
    return result;
  }

  PhoneState._();

  factory PhoneState.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PhoneState.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PhoneState',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..a<$core.int>(
        1, _omitFieldNames ? '' : 'batteryPercent', $pb.PbFieldType.O3)
    ..e<$1.ChargeState>(
        2, _omitFieldNames ? '' : 'chargeState', $pb.PbFieldType.OE,
        defaultOrMaker: $1.ChargeState.CHARGE_STATE_UNSPECIFIED,
        valueOf: $1.ChargeState.valueOf,
        enumValues: $1.ChargeState.values)
    ..e<NetworkState>(3, _omitFieldNames ? '' : 'network', $pb.PbFieldType.OE,
        defaultOrMaker: NetworkState.NETWORK_STATE_UNSPECIFIED,
        valueOf: NetworkState.valueOf,
        enumValues: NetworkState.values)
    ..pc<Accessory>(4, _omitFieldNames ? '' : 'accessories', $pb.PbFieldType.PM,
        subBuilder: Accessory.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PhoneState clone() => PhoneState()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PhoneState copyWith(void Function(PhoneState) updates) =>
      super.copyWith((message) => updates(message as PhoneState)) as PhoneState;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PhoneState create() => PhoneState._();
  @$core.override
  PhoneState createEmptyInstance() => create();
  static $pb.PbList<PhoneState> createRepeated() => $pb.PbList<PhoneState>();
  @$core.pragma('dart2js:noInline')
  static PhoneState getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PhoneState>(create);
  static PhoneState? _defaultInstance;

  /// battery_percent is the phone battery level from 0 to 100, if reported.
  @$pb.TagNumber(1)
  $core.int get batteryPercent => $_getIZ(0);
  @$pb.TagNumber(1)
  set batteryPercent($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasBatteryPercent() => $_has(0);
  @$pb.TagNumber(1)
  void clearBatteryPercent() => $_clearField(1);

  /// charge_state is the phone charging state.
  @$pb.TagNumber(2)
  $1.ChargeState get chargeState => $_getN(1);
  @$pb.TagNumber(2)
  set chargeState($1.ChargeState value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasChargeState() => $_has(1);
  @$pb.TagNumber(2)
  void clearChargeState() => $_clearField(2);

  /// network is the phone's current network transport.
  @$pb.TagNumber(3)
  NetworkState get network => $_getN(2);
  @$pb.TagNumber(3)
  set network(NetworkState value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasNetwork() => $_has(2);
  @$pb.TagNumber(3)
  void clearNetwork() => $_clearField(3);

  /// accessories are peripherals connected to the phone.
  @$pb.TagNumber(4)
  $pb.PbList<Accessory> get accessories => $_getList(3);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
