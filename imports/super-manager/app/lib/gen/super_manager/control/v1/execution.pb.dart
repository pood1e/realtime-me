// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/execution.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../../google/protobuf/timestamp.pb.dart' as $0;
import 'execution.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'execution.pbenum.dart';

/// Execution is one native provider invocation owned by a thread.
class Execution extends $pb.GeneratedMessage {
  factory Execution({
    $core.String? uid,
    $core.String? threadUid,
    $core.String? runId,
    ExecutionState? state,
    $0.Timestamp? startTime,
    $0.Timestamp? endTime,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (threadUid != null) result.threadUid = threadUid;
    if (runId != null) result.runId = runId;
    if (state != null) result.state = state;
    if (startTime != null) result.startTime = startTime;
    if (endTime != null) result.endTime = endTime;
    return result;
  }

  Execution._();

  factory Execution.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Execution.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Execution',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aOS(2, _omitFieldNames ? '' : 'threadUid')
    ..aOS(3, _omitFieldNames ? '' : 'runId')
    ..e<ExecutionState>(4, _omitFieldNames ? '' : 'state', $pb.PbFieldType.OE,
        defaultOrMaker: ExecutionState.EXECUTION_STATE_UNSPECIFIED,
        valueOf: ExecutionState.valueOf,
        enumValues: ExecutionState.values)
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'startTime',
        subBuilder: $0.Timestamp.create)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'endTime',
        subBuilder: $0.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Execution clone() => Execution()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Execution copyWith(void Function(Execution) updates) =>
      super.copyWith((message) => updates(message as Execution)) as Execution;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Execution create() => Execution._();
  @$core.override
  Execution createEmptyInstance() => create();
  static $pb.PbList<Execution> createRepeated() => $pb.PbList<Execution>();
  @$core.pragma('dart2js:noInline')
  static Execution getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Execution>(create);
  static Execution? _defaultInstance;

  /// uid is the server-assigned UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  /// thread_uid identifies the owning thread.
  @$pb.TagNumber(2)
  $core.String get threadUid => $_getSZ(1);
  @$pb.TagNumber(2)
  set threadUid($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasThreadUid() => $_has(1);
  @$pb.TagNumber(2)
  void clearThreadUid() => $_clearField(2);

  /// run_id is the caller-supplied AG-UI idempotency key.
  @$pb.TagNumber(3)
  $core.String get runId => $_getSZ(2);
  @$pb.TagNumber(3)
  set runId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRunId() => $_has(2);
  @$pb.TagNumber(3)
  void clearRunId() => $_clearField(3);

  /// state is the current native provider execution state.
  @$pb.TagNumber(4)
  ExecutionState get state => $_getN(3);
  @$pb.TagNumber(4)
  set state(ExecutionState value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasState() => $_has(3);
  @$pb.TagNumber(4)
  void clearState() => $_clearField(4);

  /// start_time is the provider execution start time.
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

  /// end_time is the terminal time when the execution has finished.
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
}

/// GetExecutionRequest requests one execution.
class GetExecutionRequest extends $pb.GeneratedMessage {
  factory GetExecutionRequest({
    $core.String? uid,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    return result;
  }

  GetExecutionRequest._();

  factory GetExecutionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetExecutionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetExecutionRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetExecutionRequest clone() => GetExecutionRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetExecutionRequest copyWith(void Function(GetExecutionRequest) updates) =>
      super.copyWith((message) => updates(message as GetExecutionRequest))
          as GetExecutionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetExecutionRequest create() => GetExecutionRequest._();
  @$core.override
  GetExecutionRequest createEmptyInstance() => create();
  static $pb.PbList<GetExecutionRequest> createRepeated() =>
      $pb.PbList<GetExecutionRequest>();
  @$core.pragma('dart2js:noInline')
  static GetExecutionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetExecutionRequest>(create);
  static GetExecutionRequest? _defaultInstance;

  /// uid is the execution UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);
}

/// GetExecutionResponse returns one execution.
class GetExecutionResponse extends $pb.GeneratedMessage {
  factory GetExecutionResponse({
    Execution? execution,
  }) {
    final result = create();
    if (execution != null) result.execution = execution;
    return result;
  }

  GetExecutionResponse._();

  factory GetExecutionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetExecutionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetExecutionResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<Execution>(1, _omitFieldNames ? '' : 'execution',
        subBuilder: Execution.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetExecutionResponse clone() =>
      GetExecutionResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetExecutionResponse copyWith(void Function(GetExecutionResponse) updates) =>
      super.copyWith((message) => updates(message as GetExecutionResponse))
          as GetExecutionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetExecutionResponse create() => GetExecutionResponse._();
  @$core.override
  GetExecutionResponse createEmptyInstance() => create();
  static $pb.PbList<GetExecutionResponse> createRepeated() =>
      $pb.PbList<GetExecutionResponse>();
  @$core.pragma('dart2js:noInline')
  static GetExecutionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetExecutionResponse>(create);
  static GetExecutionResponse? _defaultInstance;

  /// execution is the requested execution.
  @$pb.TagNumber(1)
  Execution get execution => $_getN(0);
  @$pb.TagNumber(1)
  set execution(Execution value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasExecution() => $_has(0);
  @$pb.TagNumber(1)
  void clearExecution() => $_clearField(1);
  @$pb.TagNumber(1)
  Execution ensureExecution() => $_ensure(0);
}

/// ListExecutionsRequest requests recent executions in one thread.
class ListExecutionsRequest extends $pb.GeneratedMessage {
  factory ListExecutionsRequest({
    $core.String? threadUid,
    $core.int? pageSize,
    $core.String? pageToken,
  }) {
    final result = create();
    if (threadUid != null) result.threadUid = threadUid;
    if (pageSize != null) result.pageSize = pageSize;
    if (pageToken != null) result.pageToken = pageToken;
    return result;
  }

  ListExecutionsRequest._();

  factory ListExecutionsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListExecutionsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListExecutionsRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'threadUid')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'pageSize', $pb.PbFieldType.O3)
    ..aOS(3, _omitFieldNames ? '' : 'pageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListExecutionsRequest clone() =>
      ListExecutionsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListExecutionsRequest copyWith(
          void Function(ListExecutionsRequest) updates) =>
      super.copyWith((message) => updates(message as ListExecutionsRequest))
          as ListExecutionsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListExecutionsRequest create() => ListExecutionsRequest._();
  @$core.override
  ListExecutionsRequest createEmptyInstance() => create();
  static $pb.PbList<ListExecutionsRequest> createRepeated() =>
      $pb.PbList<ListExecutionsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListExecutionsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListExecutionsRequest>(create);
  static ListExecutionsRequest? _defaultInstance;

  /// thread_uid scopes the list to one thread.
  @$pb.TagNumber(1)
  $core.String get threadUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set threadUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasThreadUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearThreadUid() => $_clearField(1);

  /// page_size is the maximum number of executions to return.
  @$pb.TagNumber(2)
  $core.int get pageSize => $_getIZ(1);
  @$pb.TagNumber(2)
  set pageSize($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPageSize() => $_has(1);
  @$pb.TagNumber(2)
  void clearPageSize() => $_clearField(2);

  /// page_token is an opaque continuation token returned by the server.
  @$pb.TagNumber(3)
  $core.String get pageToken => $_getSZ(2);
  @$pb.TagNumber(3)
  set pageToken($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPageToken() => $_has(2);
  @$pb.TagNumber(3)
  void clearPageToken() => $_clearField(3);
}

/// ListExecutionsResponse returns recent executions in one thread.
class ListExecutionsResponse extends $pb.GeneratedMessage {
  factory ListExecutionsResponse({
    $core.Iterable<Execution>? executions,
    $core.String? nextPageToken,
  }) {
    final result = create();
    if (executions != null) result.executions.addAll(executions);
    if (nextPageToken != null) result.nextPageToken = nextPageToken;
    return result;
  }

  ListExecutionsResponse._();

  factory ListExecutionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListExecutionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListExecutionsResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..pc<Execution>(1, _omitFieldNames ? '' : 'executions', $pb.PbFieldType.PM,
        subBuilder: Execution.create)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListExecutionsResponse clone() =>
      ListExecutionsResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListExecutionsResponse copyWith(
          void Function(ListExecutionsResponse) updates) =>
      super.copyWith((message) => updates(message as ListExecutionsResponse))
          as ListExecutionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListExecutionsResponse create() => ListExecutionsResponse._();
  @$core.override
  ListExecutionsResponse createEmptyInstance() => create();
  static $pb.PbList<ListExecutionsResponse> createRepeated() =>
      $pb.PbList<ListExecutionsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListExecutionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListExecutionsResponse>(create);
  static ListExecutionsResponse? _defaultInstance;

  /// executions contains the current page.
  @$pb.TagNumber(1)
  $pb.PbList<Execution> get executions => $_getList(0);

  /// next_page_token requests the next page and is empty on the last page.
  @$pb.TagNumber(2)
  $core.String get nextPageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set nextPageToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNextPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearNextPageToken() => $_clearField(2);
}

/// CancelExecutionRequest cancels one active provider execution.
class CancelExecutionRequest extends $pb.GeneratedMessage {
  factory CancelExecutionRequest({
    $core.String? uid,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    return result;
  }

  CancelExecutionRequest._();

  factory CancelExecutionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CancelExecutionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelExecutionRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelExecutionRequest clone() =>
      CancelExecutionRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelExecutionRequest copyWith(
          void Function(CancelExecutionRequest) updates) =>
      super.copyWith((message) => updates(message as CancelExecutionRequest))
          as CancelExecutionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CancelExecutionRequest create() => CancelExecutionRequest._();
  @$core.override
  CancelExecutionRequest createEmptyInstance() => create();
  static $pb.PbList<CancelExecutionRequest> createRepeated() =>
      $pb.PbList<CancelExecutionRequest>();
  @$core.pragma('dart2js:noInline')
  static CancelExecutionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelExecutionRequest>(create);
  static CancelExecutionRequest? _defaultInstance;

  /// uid is the execution UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);
}

/// CancelExecutionResponse returns the projected canceled execution.
class CancelExecutionResponse extends $pb.GeneratedMessage {
  factory CancelExecutionResponse({
    Execution? execution,
  }) {
    final result = create();
    if (execution != null) result.execution = execution;
    return result;
  }

  CancelExecutionResponse._();

  factory CancelExecutionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CancelExecutionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CancelExecutionResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<Execution>(1, _omitFieldNames ? '' : 'execution',
        subBuilder: Execution.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelExecutionResponse clone() =>
      CancelExecutionResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CancelExecutionResponse copyWith(
          void Function(CancelExecutionResponse) updates) =>
      super.copyWith((message) => updates(message as CancelExecutionResponse))
          as CancelExecutionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CancelExecutionResponse create() => CancelExecutionResponse._();
  @$core.override
  CancelExecutionResponse createEmptyInstance() => create();
  static $pb.PbList<CancelExecutionResponse> createRepeated() =>
      $pb.PbList<CancelExecutionResponse>();
  @$core.pragma('dart2js:noInline')
  static CancelExecutionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CancelExecutionResponse>(create);
  static CancelExecutionResponse? _defaultInstance;

  /// execution is the execution after the cancel request was accepted.
  @$pb.TagNumber(1)
  Execution get execution => $_getN(0);
  @$pb.TagNumber(1)
  set execution(Execution value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasExecution() => $_has(0);
  @$pb.TagNumber(1)
  void clearExecution() => $_clearField(1);
  @$pb.TagNumber(1)
  Execution ensureExecution() => $_ensure(0);
}

/// SteerExecutionRequest submits an instruction to a steer-capable active execution.
class SteerExecutionRequest extends $pb.GeneratedMessage {
  factory SteerExecutionRequest({
    $core.String? uid,
    $core.String? instruction,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (instruction != null) result.instruction = instruction;
    return result;
  }

  SteerExecutionRequest._();

  factory SteerExecutionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SteerExecutionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SteerExecutionRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aOS(2, _omitFieldNames ? '' : 'instruction')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SteerExecutionRequest clone() =>
      SteerExecutionRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SteerExecutionRequest copyWith(
          void Function(SteerExecutionRequest) updates) =>
      super.copyWith((message) => updates(message as SteerExecutionRequest))
          as SteerExecutionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SteerExecutionRequest create() => SteerExecutionRequest._();
  @$core.override
  SteerExecutionRequest createEmptyInstance() => create();
  static $pb.PbList<SteerExecutionRequest> createRepeated() =>
      $pb.PbList<SteerExecutionRequest>();
  @$core.pragma('dart2js:noInline')
  static SteerExecutionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SteerExecutionRequest>(create);
  static SteerExecutionRequest? _defaultInstance;

  /// uid is the execution UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  /// instruction is the bounded steering text.
  @$pb.TagNumber(2)
  $core.String get instruction => $_getSZ(1);
  @$pb.TagNumber(2)
  set instruction($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInstruction() => $_has(1);
  @$pb.TagNumber(2)
  void clearInstruction() => $_clearField(2);
}

/// SteerExecutionResponse confirms that the instruction was accepted.
class SteerExecutionResponse extends $pb.GeneratedMessage {
  factory SteerExecutionResponse({
    Execution? execution,
  }) {
    final result = create();
    if (execution != null) result.execution = execution;
    return result;
  }

  SteerExecutionResponse._();

  factory SteerExecutionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SteerExecutionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SteerExecutionResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<Execution>(1, _omitFieldNames ? '' : 'execution',
        subBuilder: Execution.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SteerExecutionResponse clone() =>
      SteerExecutionResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SteerExecutionResponse copyWith(
          void Function(SteerExecutionResponse) updates) =>
      super.copyWith((message) => updates(message as SteerExecutionResponse))
          as SteerExecutionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SteerExecutionResponse create() => SteerExecutionResponse._();
  @$core.override
  SteerExecutionResponse createEmptyInstance() => create();
  static $pb.PbList<SteerExecutionResponse> createRepeated() =>
      $pb.PbList<SteerExecutionResponse>();
  @$core.pragma('dart2js:noInline')
  static SteerExecutionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SteerExecutionResponse>(create);
  static SteerExecutionResponse? _defaultInstance;

  /// execution is the still-active execution.
  @$pb.TagNumber(1)
  Execution get execution => $_getN(0);
  @$pb.TagNumber(1)
  set execution(Execution value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasExecution() => $_has(0);
  @$pb.TagNumber(1)
  void clearExecution() => $_clearField(1);
  @$pb.TagNumber(1)
  Execution ensureExecution() => $_ensure(0);
}

/// ExecutionService observes and controls native provider executions.
class ExecutionServiceApi {
  final $pb.RpcClient _client;

  ExecutionServiceApi(this._client);

  /// GetExecution returns one provider execution.
  $async.Future<GetExecutionResponse> getExecution(
          $pb.ClientContext? ctx, GetExecutionRequest request) =>
      _client.invoke<GetExecutionResponse>(ctx, 'ExecutionService',
          'GetExecution', request, GetExecutionResponse());

  /// ListExecutions returns recent executions in one thread.
  $async.Future<ListExecutionsResponse> listExecutions(
          $pb.ClientContext? ctx, ListExecutionsRequest request) =>
      _client.invoke<ListExecutionsResponse>(ctx, 'ExecutionService',
          'ListExecutions', request, ListExecutionsResponse());

  /// CancelExecution cancels one active execution.
  $async.Future<CancelExecutionResponse> cancelExecution(
          $pb.ClientContext? ctx, CancelExecutionRequest request) =>
      _client.invoke<CancelExecutionResponse>(ctx, 'ExecutionService',
          'CancelExecution', request, CancelExecutionResponse());

  /// SteerExecution submits an instruction when the runtime declares support.
  $async.Future<SteerExecutionResponse> steerExecution(
          $pb.ClientContext? ctx, SteerExecutionRequest request) =>
      _client.invoke<SteerExecutionResponse>(ctx, 'ExecutionService',
          'SteerExecution', request, SteerExecutionResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
