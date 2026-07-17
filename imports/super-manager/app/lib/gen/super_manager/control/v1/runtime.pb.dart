// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/runtime.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $0;

import 'runtime.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'runtime.pbenum.dart';

/// Runtime describes one fixed installed coding-agent CLI.
class Runtime extends $pb.GeneratedMessage {
  factory Runtime({
    $core.String? uid,
    RuntimeKind? kind,
    $core.String? displayName,
    $core.String? version,
    RuntimeAvailability? availability,
    $core.Iterable<RuntimeCapability>? capabilities,
    $core.String? diagnostic,
    $0.Timestamp? updateTime,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (kind != null) result.kind = kind;
    if (displayName != null) result.displayName = displayName;
    if (version != null) result.version = version;
    if (availability != null) result.availability = availability;
    if (capabilities != null) result.capabilities.addAll(capabilities);
    if (diagnostic != null) result.diagnostic = diagnostic;
    if (updateTime != null) result.updateTime = updateTime;
    return result;
  }

  Runtime._();

  factory Runtime.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Runtime.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Runtime',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aE<RuntimeKind>(2, _omitFieldNames ? '' : 'kind',
        enumValues: RuntimeKind.values)
    ..aOS(3, _omitFieldNames ? '' : 'displayName')
    ..aOS(4, _omitFieldNames ? '' : 'version')
    ..aE<RuntimeAvailability>(5, _omitFieldNames ? '' : 'availability',
        enumValues: RuntimeAvailability.values)
    ..pc<RuntimeCapability>(
        6, _omitFieldNames ? '' : 'capabilities', $pb.PbFieldType.KE,
        valueOf: RuntimeCapability.valueOf,
        enumValues: RuntimeCapability.values,
        defaultEnumValue: RuntimeCapability.RUNTIME_CAPABILITY_UNSPECIFIED)
    ..aOS(7, _omitFieldNames ? '' : 'diagnostic')
    ..aOM<$0.Timestamp>(8, _omitFieldNames ? '' : 'updateTime',
        subBuilder: $0.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Runtime clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Runtime copyWith(void Function(Runtime) updates) =>
      super.copyWith((message) => updates(message as Runtime)) as Runtime;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Runtime create() => Runtime._();
  @$core.override
  Runtime createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Runtime getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Runtime>(create);
  static Runtime? _defaultInstance;

  /// uid is the server-assigned UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  /// kind identifies the CLI implementation.
  @$pb.TagNumber(2)
  RuntimeKind get kind => $_getN(1);
  @$pb.TagNumber(2)
  set kind(RuntimeKind value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasKind() => $_has(1);
  @$pb.TagNumber(2)
  void clearKind() => $_clearField(2);

  /// display_name is the human-readable runtime label.
  @$pb.TagNumber(3)
  $core.String get displayName => $_getSZ(2);
  @$pb.TagNumber(3)
  set displayName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDisplayName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDisplayName() => $_clearField(3);

  /// version is the detected CLI version.
  @$pb.TagNumber(4)
  $core.String get version => $_getSZ(3);
  @$pb.TagNumber(4)
  set version($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasVersion() => $_has(3);
  @$pb.TagNumber(4)
  void clearVersion() => $_clearField(4);

  /// availability is the latest startup probe result.
  @$pb.TagNumber(5)
  RuntimeAvailability get availability => $_getN(4);
  @$pb.TagNumber(5)
  set availability(RuntimeAvailability value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasAvailability() => $_has(4);
  @$pb.TagNumber(5)
  void clearAvailability() => $_clearField(5);

  /// capabilities contains only behavior verified for this installed version.
  @$pb.TagNumber(6)
  $pb.PbList<RuntimeCapability> get capabilities => $_getList(5);

  /// diagnostic is a bounded, non-secret explanation when unavailable.
  @$pb.TagNumber(7)
  $core.String get diagnostic => $_getSZ(6);
  @$pb.TagNumber(7)
  set diagnostic($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasDiagnostic() => $_has(6);
  @$pb.TagNumber(7)
  void clearDiagnostic() => $_clearField(7);

  /// update_time is the time of the latest runtime probe.
  @$pb.TagNumber(8)
  $0.Timestamp get updateTime => $_getN(7);
  @$pb.TagNumber(8)
  set updateTime($0.Timestamp value) => $_setField(8, value);
  @$pb.TagNumber(8)
  $core.bool hasUpdateTime() => $_has(7);
  @$pb.TagNumber(8)
  void clearUpdateTime() => $_clearField(8);
  @$pb.TagNumber(8)
  $0.Timestamp ensureUpdateTime() => $_ensure(7);
}

/// QuotaSnapshot is the latest account-level quota observation for a runtime.
class QuotaSnapshot extends $pb.GeneratedMessage {
  factory QuotaSnapshot({
    $core.String? runtimeUid,
    QuotaFreshness? freshness,
    $core.double? usedRatio,
    $0.Timestamp? resetTime,
    $0.Timestamp? observeTime,
    $core.String? source,
  }) {
    final result = create();
    if (runtimeUid != null) result.runtimeUid = runtimeUid;
    if (freshness != null) result.freshness = freshness;
    if (usedRatio != null) result.usedRatio = usedRatio;
    if (resetTime != null) result.resetTime = resetTime;
    if (observeTime != null) result.observeTime = observeTime;
    if (source != null) result.source = source;
    return result;
  }

  QuotaSnapshot._();

  factory QuotaSnapshot.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory QuotaSnapshot.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'QuotaSnapshot',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'runtimeUid')
    ..aE<QuotaFreshness>(2, _omitFieldNames ? '' : 'freshness',
        enumValues: QuotaFreshness.values)
    ..aD(3, _omitFieldNames ? '' : 'usedRatio')
    ..aOM<$0.Timestamp>(4, _omitFieldNames ? '' : 'resetTime',
        subBuilder: $0.Timestamp.create)
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'observeTime',
        subBuilder: $0.Timestamp.create)
    ..aOS(6, _omitFieldNames ? '' : 'source')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  QuotaSnapshot clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  QuotaSnapshot copyWith(void Function(QuotaSnapshot) updates) =>
      super.copyWith((message) => updates(message as QuotaSnapshot))
          as QuotaSnapshot;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static QuotaSnapshot create() => QuotaSnapshot._();
  @$core.override
  QuotaSnapshot createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static QuotaSnapshot getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<QuotaSnapshot>(create);
  static QuotaSnapshot? _defaultInstance;

  /// runtime_uid identifies the observed runtime.
  @$pb.TagNumber(1)
  $core.String get runtimeUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set runtimeUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRuntimeUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearRuntimeUid() => $_clearField(1);

  /// freshness describes whether this snapshot is suitable for display.
  @$pb.TagNumber(2)
  QuotaFreshness get freshness => $_getN(1);
  @$pb.TagNumber(2)
  set freshness(QuotaFreshness value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasFreshness() => $_has(1);
  @$pb.TagNumber(2)
  void clearFreshness() => $_clearField(2);

  /// used_ratio is the observed fraction used, in the inclusive range 0 to 1.
  @$pb.TagNumber(3)
  $core.double get usedRatio => $_getN(2);
  @$pb.TagNumber(3)
  set usedRatio($core.double value) => $_setDouble(2, value);
  @$pb.TagNumber(3)
  $core.bool hasUsedRatio() => $_has(2);
  @$pb.TagNumber(3)
  void clearUsedRatio() => $_clearField(3);

  /// reset_time is the provider-reported reset instant when available.
  @$pb.TagNumber(4)
  $0.Timestamp get resetTime => $_getN(3);
  @$pb.TagNumber(4)
  set resetTime($0.Timestamp value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasResetTime() => $_has(3);
  @$pb.TagNumber(4)
  void clearResetTime() => $_clearField(4);
  @$pb.TagNumber(4)
  $0.Timestamp ensureResetTime() => $_ensure(3);

  /// observe_time is the time this value was observed.
  @$pb.TagNumber(5)
  $0.Timestamp get observeTime => $_getN(4);
  @$pb.TagNumber(5)
  set observeTime($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasObserveTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearObserveTime() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureObserveTime() => $_ensure(4);

  /// source is the bounded structured source name.
  @$pb.TagNumber(6)
  $core.String get source => $_getSZ(5);
  @$pb.TagNumber(6)
  set source($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasSource() => $_has(5);
  @$pb.TagNumber(6)
  void clearSource() => $_clearField(6);
}

/// GetRuntimeRequest requests one runtime.
class GetRuntimeRequest extends $pb.GeneratedMessage {
  factory GetRuntimeRequest({
    $core.String? uid,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    return result;
  }

  GetRuntimeRequest._();

  factory GetRuntimeRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetRuntimeRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRuntimeRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRuntimeRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRuntimeRequest copyWith(void Function(GetRuntimeRequest) updates) =>
      super.copyWith((message) => updates(message as GetRuntimeRequest))
          as GetRuntimeRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetRuntimeRequest create() => GetRuntimeRequest._();
  @$core.override
  GetRuntimeRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetRuntimeRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRuntimeRequest>(create);
  static GetRuntimeRequest? _defaultInstance;

  /// uid is the runtime UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);
}

/// GetRuntimeResponse returns one runtime.
class GetRuntimeResponse extends $pb.GeneratedMessage {
  factory GetRuntimeResponse({
    Runtime? runtime,
  }) {
    final result = create();
    if (runtime != null) result.runtime = runtime;
    return result;
  }

  GetRuntimeResponse._();

  factory GetRuntimeResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetRuntimeResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRuntimeResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<Runtime>(1, _omitFieldNames ? '' : 'runtime',
        subBuilder: Runtime.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRuntimeResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRuntimeResponse copyWith(void Function(GetRuntimeResponse) updates) =>
      super.copyWith((message) => updates(message as GetRuntimeResponse))
          as GetRuntimeResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetRuntimeResponse create() => GetRuntimeResponse._();
  @$core.override
  GetRuntimeResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetRuntimeResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRuntimeResponse>(create);
  static GetRuntimeResponse? _defaultInstance;

  /// runtime is the requested runtime.
  @$pb.TagNumber(1)
  Runtime get runtime => $_getN(0);
  @$pb.TagNumber(1)
  set runtime(Runtime value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasRuntime() => $_has(0);
  @$pb.TagNumber(1)
  void clearRuntime() => $_clearField(1);
  @$pb.TagNumber(1)
  Runtime ensureRuntime() => $_ensure(0);
}

/// ListRuntimesRequest requests all configured runtimes.
class ListRuntimesRequest extends $pb.GeneratedMessage {
  factory ListRuntimesRequest() => create();

  ListRuntimesRequest._();

  factory ListRuntimesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListRuntimesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListRuntimesRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRuntimesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRuntimesRequest copyWith(void Function(ListRuntimesRequest) updates) =>
      super.copyWith((message) => updates(message as ListRuntimesRequest))
          as ListRuntimesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListRuntimesRequest create() => ListRuntimesRequest._();
  @$core.override
  ListRuntimesRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListRuntimesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListRuntimesRequest>(create);
  static ListRuntimesRequest? _defaultInstance;
}

/// ListRuntimesResponse returns all configured runtimes.
class ListRuntimesResponse extends $pb.GeneratedMessage {
  factory ListRuntimesResponse({
    $core.Iterable<Runtime>? runtimes,
  }) {
    final result = create();
    if (runtimes != null) result.runtimes.addAll(runtimes);
    return result;
  }

  ListRuntimesResponse._();

  factory ListRuntimesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListRuntimesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListRuntimesResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..pPM<Runtime>(1, _omitFieldNames ? '' : 'runtimes',
        subBuilder: Runtime.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRuntimesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListRuntimesResponse copyWith(void Function(ListRuntimesResponse) updates) =>
      super.copyWith((message) => updates(message as ListRuntimesResponse))
          as ListRuntimesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListRuntimesResponse create() => ListRuntimesResponse._();
  @$core.override
  ListRuntimesResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListRuntimesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListRuntimesResponse>(create);
  static ListRuntimesResponse? _defaultInstance;

  /// runtimes is the fixed runtime collection.
  @$pb.TagNumber(1)
  $pb.PbList<Runtime> get runtimes => $_getList(0);
}

/// GetRuntimeQuotaRequest requests the latest quota snapshot for one runtime.
class GetRuntimeQuotaRequest extends $pb.GeneratedMessage {
  factory GetRuntimeQuotaRequest({
    $core.String? runtimeUid,
  }) {
    final result = create();
    if (runtimeUid != null) result.runtimeUid = runtimeUid;
    return result;
  }

  GetRuntimeQuotaRequest._();

  factory GetRuntimeQuotaRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetRuntimeQuotaRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRuntimeQuotaRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'runtimeUid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRuntimeQuotaRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRuntimeQuotaRequest copyWith(
          void Function(GetRuntimeQuotaRequest) updates) =>
      super.copyWith((message) => updates(message as GetRuntimeQuotaRequest))
          as GetRuntimeQuotaRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetRuntimeQuotaRequest create() => GetRuntimeQuotaRequest._();
  @$core.override
  GetRuntimeQuotaRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetRuntimeQuotaRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRuntimeQuotaRequest>(create);
  static GetRuntimeQuotaRequest? _defaultInstance;

  /// runtime_uid is the runtime UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get runtimeUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set runtimeUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRuntimeUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearRuntimeUid() => $_clearField(1);
}

/// GetRuntimeQuotaResponse returns the latest quota snapshot.
class GetRuntimeQuotaResponse extends $pb.GeneratedMessage {
  factory GetRuntimeQuotaResponse({
    QuotaSnapshot? quotaSnapshot,
  }) {
    final result = create();
    if (quotaSnapshot != null) result.quotaSnapshot = quotaSnapshot;
    return result;
  }

  GetRuntimeQuotaResponse._();

  factory GetRuntimeQuotaResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetRuntimeQuotaResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetRuntimeQuotaResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<QuotaSnapshot>(1, _omitFieldNames ? '' : 'quotaSnapshot',
        subBuilder: QuotaSnapshot.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRuntimeQuotaResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetRuntimeQuotaResponse copyWith(
          void Function(GetRuntimeQuotaResponse) updates) =>
      super.copyWith((message) => updates(message as GetRuntimeQuotaResponse))
          as GetRuntimeQuotaResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetRuntimeQuotaResponse create() => GetRuntimeQuotaResponse._();
  @$core.override
  GetRuntimeQuotaResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetRuntimeQuotaResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetRuntimeQuotaResponse>(create);
  static GetRuntimeQuotaResponse? _defaultInstance;

  /// quota_snapshot is the latest bounded observation.
  @$pb.TagNumber(1)
  QuotaSnapshot get quotaSnapshot => $_getN(0);
  @$pb.TagNumber(1)
  set quotaSnapshot(QuotaSnapshot value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasQuotaSnapshot() => $_has(0);
  @$pb.TagNumber(1)
  void clearQuotaSnapshot() => $_clearField(1);
  @$pb.TagNumber(1)
  QuotaSnapshot ensureQuotaSnapshot() => $_ensure(0);
}

/// RuntimeService exposes installed runtime status and account telemetry.
class RuntimeServiceApi {
  final $pb.RpcClient _client;

  RuntimeServiceApi(this._client);

  /// GetRuntime returns one installed runtime.
  $async.Future<GetRuntimeResponse> getRuntime(
          $pb.ClientContext? ctx, GetRuntimeRequest request) =>
      _client.invoke<GetRuntimeResponse>(
          ctx, 'RuntimeService', 'GetRuntime', request, GetRuntimeResponse());

  /// ListRuntimes returns all installed runtime slots.
  $async.Future<ListRuntimesResponse> listRuntimes(
          $pb.ClientContext? ctx, ListRuntimesRequest request) =>
      _client.invoke<ListRuntimesResponse>(ctx, 'RuntimeService',
          'ListRuntimes', request, ListRuntimesResponse());

  /// GetRuntimeQuota returns the latest account-level quota observation.
  $async.Future<GetRuntimeQuotaResponse> getRuntimeQuota(
          $pb.ClientContext? ctx, GetRuntimeQuotaRequest request) =>
      _client.invoke<GetRuntimeQuotaResponse>(ctx, 'RuntimeService',
          'GetRuntimeQuota', request, GetRuntimeQuotaResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
