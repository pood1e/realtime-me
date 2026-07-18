// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/ingest.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import 'ingest.pbenum.dart';
import 'status_types.pb.dart' as $0;
import 'watch.pb.dart' as $1;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'ingest.pbenum.dart';

/// EnrollDeviceRequest describes the device seeking a gateway identity.
class EnrollDeviceRequest extends $pb.GeneratedMessage {
  factory EnrollDeviceRequest({
    $0.DeviceKind? kind,
    $0.DeviceRole? role,
    $core.String? displayName,
    $core.String? model,
  }) {
    final result = create();
    if (kind != null) result.kind = kind;
    if (role != null) result.role = role;
    if (displayName != null) result.displayName = displayName;
    if (model != null) result.model = model;
    return result;
  }

  EnrollDeviceRequest._();

  factory EnrollDeviceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory EnrollDeviceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EnrollDeviceRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..e<$0.DeviceKind>(1, _omitFieldNames ? '' : 'kind', $pb.PbFieldType.OE,
        defaultOrMaker: $0.DeviceKind.DEVICE_KIND_UNSPECIFIED,
        valueOf: $0.DeviceKind.valueOf,
        enumValues: $0.DeviceKind.values)
    ..e<$0.DeviceRole>(2, _omitFieldNames ? '' : 'role', $pb.PbFieldType.OE,
        defaultOrMaker: $0.DeviceRole.DEVICE_ROLE_UNSPECIFIED,
        valueOf: $0.DeviceRole.valueOf,
        enumValues: $0.DeviceRole.values)
    ..aOS(3, _omitFieldNames ? '' : 'displayName')
    ..aOS(4, _omitFieldNames ? '' : 'model')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnrollDeviceRequest clone() => EnrollDeviceRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnrollDeviceRequest copyWith(void Function(EnrollDeviceRequest) updates) =>
      super.copyWith((message) => updates(message as EnrollDeviceRequest))
          as EnrollDeviceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EnrollDeviceRequest create() => EnrollDeviceRequest._();
  @$core.override
  EnrollDeviceRequest createEmptyInstance() => create();
  static $pb.PbList<EnrollDeviceRequest> createRepeated() =>
      $pb.PbList<EnrollDeviceRequest>();
  @$core.pragma('dart2js:noInline')
  static EnrollDeviceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EnrollDeviceRequest>(create);
  static EnrollDeviceRequest? _defaultInstance;

  /// kind is the broad device category being enrolled.
  @$pb.TagNumber(1)
  $0.DeviceKind get kind => $_getN(0);
  @$pb.TagNumber(1)
  set kind($0.DeviceKind value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasKind() => $_has(0);
  @$pb.TagNumber(1)
  void clearKind() => $_clearField(1);

  /// role is the device's operational role.
  @$pb.TagNumber(2)
  $0.DeviceRole get role => $_getN(1);
  @$pb.TagNumber(2)
  set role($0.DeviceRole value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasRole() => $_has(1);
  @$pb.TagNumber(2)
  void clearRole() => $_clearField(2);

  /// display_name is the human-readable device label.
  @$pb.TagNumber(3)
  $core.String get displayName => $_getSZ(2);
  @$pb.TagNumber(3)
  set displayName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDisplayName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDisplayName() => $_clearField(3);

  /// model is the hardware model reported by the operating system.
  @$pb.TagNumber(4)
  $core.String get model => $_getSZ(3);
  @$pb.TagNumber(4)
  set model($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasModel() => $_has(3);
  @$pb.TagNumber(4)
  void clearModel() => $_clearField(4);
}

/// EnrollDeviceResponse returns the gateway-owned device identity.
class EnrollDeviceResponse extends $pb.GeneratedMessage {
  factory EnrollDeviceResponse({
    $core.String? deviceUid,
  }) {
    final result = create();
    if (deviceUid != null) result.deviceUid = deviceUid;
    return result;
  }

  EnrollDeviceResponse._();

  factory EnrollDeviceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory EnrollDeviceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EnrollDeviceResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'deviceUid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnrollDeviceResponse clone() =>
      EnrollDeviceResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EnrollDeviceResponse copyWith(void Function(EnrollDeviceResponse) updates) =>
      super.copyWith((message) => updates(message as EnrollDeviceResponse))
          as EnrollDeviceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EnrollDeviceResponse create() => EnrollDeviceResponse._();
  @$core.override
  EnrollDeviceResponse createEmptyInstance() => create();
  static $pb.PbList<EnrollDeviceResponse> createRepeated() =>
      $pb.PbList<EnrollDeviceResponse>();
  @$core.pragma('dart2js:noInline')
  static EnrollDeviceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EnrollDeviceResponse>(create);
  static EnrollDeviceResponse? _defaultInstance;

  /// device_uid is the gateway-assigned opaque device identifier.
  @$pb.TagNumber(1)
  $core.String get deviceUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set deviceUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDeviceUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearDeviceUid() => $_clearField(1);
}

/// ReportMobileStatusRequest carries the phone's own state and the paired watch
/// snapshot, keyed by the gateway-assigned device uid.
class ReportMobileStatusRequest extends $pb.GeneratedMessage {
  factory ReportMobileStatusRequest({
    $core.String? deviceUid,
    $core.String? displayName,
    $core.String? model,
    $0.PhoneState? phone,
    $1.WatchSnapshot? watch,
    $0.SwitchPresence? switchPresence,
  }) {
    final result = create();
    if (deviceUid != null) result.deviceUid = deviceUid;
    if (displayName != null) result.displayName = displayName;
    if (model != null) result.model = model;
    if (phone != null) result.phone = phone;
    if (watch != null) result.watch = watch;
    if (switchPresence != null) result.switchPresence = switchPresence;
    return result;
  }

  ReportMobileStatusRequest._();

  factory ReportMobileStatusRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ReportMobileStatusRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReportMobileStatusRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'deviceUid')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..aOS(3, _omitFieldNames ? '' : 'model')
    ..aOM<$0.PhoneState>(4, _omitFieldNames ? '' : 'phone',
        subBuilder: $0.PhoneState.create)
    ..aOM<$1.WatchSnapshot>(5, _omitFieldNames ? '' : 'watch',
        subBuilder: $1.WatchSnapshot.create)
    ..aOM<$0.SwitchPresence>(6, _omitFieldNames ? '' : 'switchPresence',
        subBuilder: $0.SwitchPresence.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReportMobileStatusRequest clone() =>
      ReportMobileStatusRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReportMobileStatusRequest copyWith(
          void Function(ReportMobileStatusRequest) updates) =>
      super.copyWith((message) => updates(message as ReportMobileStatusRequest))
          as ReportMobileStatusRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ReportMobileStatusRequest create() => ReportMobileStatusRequest._();
  @$core.override
  ReportMobileStatusRequest createEmptyInstance() => create();
  static $pb.PbList<ReportMobileStatusRequest> createRepeated() =>
      $pb.PbList<ReportMobileStatusRequest>();
  @$core.pragma('dart2js:noInline')
  static ReportMobileStatusRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReportMobileStatusRequest>(create);
  static ReportMobileStatusRequest? _defaultInstance;

  /// device_uid is the enrolled phone identity.
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

  /// watch is the latest paired-watch snapshot, reusing the Data Layer contract.
  @$pb.TagNumber(5)
  $1.WatchSnapshot get watch => $_getN(4);
  @$pb.TagNumber(5)
  set watch($1.WatchSnapshot value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasWatch() => $_has(4);
  @$pb.TagNumber(5)
  void clearWatch() => $_clearField(5);
  @$pb.TagNumber(5)
  $1.WatchSnapshot ensureWatch() => $_ensure(4);

  /// switch_presence is the owner's Nintendo Switch play presence, if available.
  @$pb.TagNumber(6)
  $0.SwitchPresence get switchPresence => $_getN(5);
  @$pb.TagNumber(6)
  set switchPresence($0.SwitchPresence value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasSwitchPresence() => $_has(5);
  @$pb.TagNumber(6)
  void clearSwitchPresence() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.SwitchPresence ensureSwitchPresence() => $_ensure(5);
}

/// ReportMobileStatusResponse is the acknowledgement of a mobile report.
class ReportMobileStatusResponse extends $pb.GeneratedMessage {
  factory ReportMobileStatusResponse() => create();

  ReportMobileStatusResponse._();

  factory ReportMobileStatusResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ReportMobileStatusResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ReportMobileStatusResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReportMobileStatusResponse clone() =>
      ReportMobileStatusResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ReportMobileStatusResponse copyWith(
          void Function(ReportMobileStatusResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ReportMobileStatusResponse))
          as ReportMobileStatusResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ReportMobileStatusResponse create() => ReportMobileStatusResponse._();
  @$core.override
  ReportMobileStatusResponse createEmptyInstance() => create();
  static $pb.PbList<ReportMobileStatusResponse> createRepeated() =>
      $pb.PbList<ReportMobileStatusResponse>();
  @$core.pragma('dart2js:noInline')
  static ReportMobileStatusResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ReportMobileStatusResponse>(create);
  static ReportMobileStatusResponse? _defaultInstance;
}

/// RegisterScrapeTargetsRequest declares the complete set of scrape targets for
/// one enrolled device. Declaring an empty set removes the device from discovery,
/// so a decommissioned host stops being scraped instead of failing forever.
class RegisterScrapeTargetsRequest extends $pb.GeneratedMessage {
  factory RegisterScrapeTargetsRequest({
    $core.Iterable<ScrapeTarget>? targets,
    $core.String? deviceUid,
  }) {
    final result = create();
    if (targets != null) result.targets.addAll(targets);
    if (deviceUid != null) result.deviceUid = deviceUid;
    return result;
  }

  RegisterScrapeTargetsRequest._();

  factory RegisterScrapeTargetsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RegisterScrapeTargetsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RegisterScrapeTargetsRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..pc<ScrapeTarget>(1, _omitFieldNames ? '' : 'targets', $pb.PbFieldType.PM,
        subBuilder: ScrapeTarget.create)
    ..aOS(2, _omitFieldNames ? '' : 'deviceUid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterScrapeTargetsRequest clone() =>
      RegisterScrapeTargetsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterScrapeTargetsRequest copyWith(
          void Function(RegisterScrapeTargetsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as RegisterScrapeTargetsRequest))
          as RegisterScrapeTargetsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RegisterScrapeTargetsRequest create() =>
      RegisterScrapeTargetsRequest._();
  @$core.override
  RegisterScrapeTargetsRequest createEmptyInstance() => create();
  static $pb.PbList<RegisterScrapeTargetsRequest> createRepeated() =>
      $pb.PbList<RegisterScrapeTargetsRequest>();
  @$core.pragma('dart2js:noInline')
  static RegisterScrapeTargetsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RegisterScrapeTargetsRequest>(create);
  static RegisterScrapeTargetsRequest? _defaultInstance;

  /// targets is the complete set of endpoints Prometheus should scrape for the device.
  @$pb.TagNumber(1)
  $pb.PbList<ScrapeTarget> get targets => $_getList(0);

  /// device_uid is the enrolled device the targets belong to.
  @$pb.TagNumber(2)
  $core.String get deviceUid => $_getSZ(1);
  @$pb.TagNumber(2)
  set deviceUid($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDeviceUid() => $_has(1);
  @$pb.TagNumber(2)
  void clearDeviceUid() => $_clearField(2);
}

/// RegisterScrapeTargetsResponse is the acknowledgement of a registration.
class RegisterScrapeTargetsResponse extends $pb.GeneratedMessage {
  factory RegisterScrapeTargetsResponse() => create();

  RegisterScrapeTargetsResponse._();

  factory RegisterScrapeTargetsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RegisterScrapeTargetsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RegisterScrapeTargetsResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterScrapeTargetsResponse clone() =>
      RegisterScrapeTargetsResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RegisterScrapeTargetsResponse copyWith(
          void Function(RegisterScrapeTargetsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as RegisterScrapeTargetsResponse))
          as RegisterScrapeTargetsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RegisterScrapeTargetsResponse create() =>
      RegisterScrapeTargetsResponse._();
  @$core.override
  RegisterScrapeTargetsResponse createEmptyInstance() => create();
  static $pb.PbList<RegisterScrapeTargetsResponse> createRepeated() =>
      $pb.PbList<RegisterScrapeTargetsResponse>();
  @$core.pragma('dart2js:noInline')
  static RegisterScrapeTargetsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RegisterScrapeTargetsResponse>(create);
  static RegisterScrapeTargetsResponse? _defaultInstance;
}

/// ScrapeTarget is one endpoint Prometheus scrapes. The device it belongs to, and
/// every label describing that device, come from the enrollment the gateway owns;
/// a caller cannot assert them here.
class ScrapeTarget extends $pb.GeneratedMessage {
  factory ScrapeTarget({
    ScrapeJob? job,
    $core.String? target,
  }) {
    final result = create();
    if (job != null) result.job = job;
    if (target != null) result.target = target;
    return result;
  }

  ScrapeTarget._();

  factory ScrapeTarget.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ScrapeTarget.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ScrapeTarget',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.status.v1'),
      createEmptyInstance: create)
    ..e<ScrapeJob>(1, _omitFieldNames ? '' : 'job', $pb.PbFieldType.OE,
        defaultOrMaker: ScrapeJob.SCRAPE_JOB_UNSPECIFIED,
        valueOf: ScrapeJob.valueOf,
        enumValues: ScrapeJob.values)
    ..aOS(2, _omitFieldNames ? '' : 'target')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ScrapeTarget clone() => ScrapeTarget()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ScrapeTarget copyWith(void Function(ScrapeTarget) updates) =>
      super.copyWith((message) => updates(message as ScrapeTarget))
          as ScrapeTarget;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ScrapeTarget create() => ScrapeTarget._();
  @$core.override
  ScrapeTarget createEmptyInstance() => create();
  static $pb.PbList<ScrapeTarget> createRepeated() =>
      $pb.PbList<ScrapeTarget>();
  @$core.pragma('dart2js:noInline')
  static ScrapeTarget getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ScrapeTarget>(create);
  static ScrapeTarget? _defaultInstance;

  /// job is the Prometheus job the target belongs to.
  @$pb.TagNumber(1)
  ScrapeJob get job => $_getN(0);
  @$pb.TagNumber(1)
  set job(ScrapeJob value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasJob() => $_has(0);
  @$pb.TagNumber(1)
  void clearJob() => $_clearField(1);

  /// target is the host:port endpoint Prometheus scrapes.
  @$pb.TagNumber(2)
  $core.String get target => $_getSZ(1);
  @$pb.TagNumber(2)
  set target($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTarget() => $_has(1);
  @$pb.TagNumber(2)
  void clearTarget() => $_clearField(2);
}

/// EnrollmentService mints the opaque device identifiers the gateway owns.
class EnrollmentServiceApi {
  final $pb.RpcClient _client;

  EnrollmentServiceApi(this._client);

  /// EnrollDevice registers a device and returns its gateway-assigned uid. The
  /// caller persists the uid and presents it on every subsequent report; it must
  /// never construct the identifier itself.
  $async.Future<EnrollDeviceResponse> enrollDevice(
          $pb.ClientContext? ctx, EnrollDeviceRequest request) =>
      _client.invoke<EnrollDeviceResponse>(ctx, 'EnrollmentService',
          'EnrollDevice', request, EnrollDeviceResponse());
}

/// IngestService accepts the phone's push and the operator's scrape-target
/// registration. Hosts, virtual machines, and coding agents are never pushed:
/// they run read-only exporters that Prometheus discovers and scrapes. Only the
/// phone pushes, because it cannot be scraped.
class IngestServiceApi {
  final $pb.RpcClient _client;

  IngestServiceApi(this._client);

  /// ReportMobileStatus records the latest phone-and-watch status.
  $async.Future<ReportMobileStatusResponse> reportMobileStatus(
          $pb.ClientContext? ctx, ReportMobileStatusRequest request) =>
      _client.invoke<ReportMobileStatusResponse>(ctx, 'IngestService',
          'ReportMobileStatus', request, ReportMobileStatusResponse());

  /// RegisterScrapeTargets advertises Prometheus scrape targets for discovery.
  $async.Future<RegisterScrapeTargetsResponse> registerScrapeTargets(
          $pb.ClientContext? ctx, RegisterScrapeTargetsRequest request) =>
      _client.invoke<RegisterScrapeTargetsResponse>(ctx, 'IngestService',
          'RegisterScrapeTargets', request, RegisterScrapeTargetsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
