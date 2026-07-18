// This is a generated file - do not edit.
//
// Generated from realtime/me/manager/control/v1/thread.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

import '../../../../../google/protobuf/timestamp.pb.dart' as $0;
import 'thread.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'thread.pbenum.dart';

/// Thread is a durable structured conversation in one workspace and runtime.
class Thread extends $pb.GeneratedMessage {
  factory Thread({
    $core.String? uid,
    $core.String? workspaceUid,
    $core.String? runtimeUid,
    $core.String? displayName,
    ThreadState? state,
    $0.Timestamp? createTime,
    $0.Timestamp? updateTime,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (workspaceUid != null) result.workspaceUid = workspaceUid;
    if (runtimeUid != null) result.runtimeUid = runtimeUid;
    if (displayName != null) result.displayName = displayName;
    if (state != null) result.state = state;
    if (createTime != null) result.createTime = createTime;
    if (updateTime != null) result.updateTime = updateTime;
    return result;
  }

  Thread._();

  factory Thread.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Thread.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Thread',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aOS(2, _omitFieldNames ? '' : 'workspaceUid')
    ..aOS(3, _omitFieldNames ? '' : 'runtimeUid')
    ..aOS(4, _omitFieldNames ? '' : 'displayName')
    ..e<ThreadState>(5, _omitFieldNames ? '' : 'state', $pb.PbFieldType.OE,
        defaultOrMaker: ThreadState.THREAD_STATE_UNSPECIFIED,
        valueOf: ThreadState.valueOf,
        enumValues: ThreadState.values)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'createTime',
        subBuilder: $0.Timestamp.create)
    ..aOM<$0.Timestamp>(7, _omitFieldNames ? '' : 'updateTime',
        subBuilder: $0.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Thread clone() => Thread()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Thread copyWith(void Function(Thread) updates) =>
      super.copyWith((message) => updates(message as Thread)) as Thread;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Thread create() => Thread._();
  @$core.override
  Thread createEmptyInstance() => create();
  static $pb.PbList<Thread> createRepeated() => $pb.PbList<Thread>();
  @$core.pragma('dart2js:noInline')
  static Thread getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Thread>(create);
  static Thread? _defaultInstance;

  /// uid is the server-assigned UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  /// workspace_uid identifies the workspace that owns this thread.
  @$pb.TagNumber(2)
  $core.String get workspaceUid => $_getSZ(1);
  @$pb.TagNumber(2)
  set workspaceUid($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWorkspaceUid() => $_has(1);
  @$pb.TagNumber(2)
  void clearWorkspaceUid() => $_clearField(2);

  /// runtime_uid identifies the runtime that executes this thread.
  @$pb.TagNumber(3)
  $core.String get runtimeUid => $_getSZ(2);
  @$pb.TagNumber(3)
  set runtimeUid($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRuntimeUid() => $_has(2);
  @$pb.TagNumber(3)
  void clearRuntimeUid() => $_clearField(3);

  /// display_name is the human-readable conversation label.
  @$pb.TagNumber(4)
  $core.String get displayName => $_getSZ(3);
  @$pb.TagNumber(4)
  set displayName($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDisplayName() => $_has(3);
  @$pb.TagNumber(4)
  void clearDisplayName() => $_clearField(4);

  /// state is the current projected conversation state.
  @$pb.TagNumber(5)
  ThreadState get state => $_getN(4);
  @$pb.TagNumber(5)
  set state(ThreadState value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasState() => $_has(4);
  @$pb.TagNumber(5)
  void clearState() => $_clearField(5);

  /// create_time is the thread creation time.
  @$pb.TagNumber(6)
  $0.Timestamp get createTime => $_getN(5);
  @$pb.TagNumber(6)
  set createTime($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasCreateTime() => $_has(5);
  @$pb.TagNumber(6)
  void clearCreateTime() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureCreateTime() => $_ensure(5);

  /// update_time is the latest state or event update time.
  @$pb.TagNumber(7)
  $0.Timestamp get updateTime => $_getN(6);
  @$pb.TagNumber(7)
  set updateTime($0.Timestamp value) => $_setField(7, value);
  @$pb.TagNumber(7)
  $core.bool hasUpdateTime() => $_has(6);
  @$pb.TagNumber(7)
  void clearUpdateTime() => $_clearField(7);
  @$pb.TagNumber(7)
  $0.Timestamp ensureUpdateTime() => $_ensure(6);
}

/// CreateThreadRequest creates a structured conversation.
class CreateThreadRequest extends $pb.GeneratedMessage {
  factory CreateThreadRequest({
    $core.String? workspaceUid,
    $core.String? runtimeUid,
    $core.String? displayName,
  }) {
    final result = create();
    if (workspaceUid != null) result.workspaceUid = workspaceUid;
    if (runtimeUid != null) result.runtimeUid = runtimeUid;
    if (displayName != null) result.displayName = displayName;
    return result;
  }

  CreateThreadRequest._();

  factory CreateThreadRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateThreadRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateThreadRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceUid')
    ..aOS(2, _omitFieldNames ? '' : 'runtimeUid')
    ..aOS(3, _omitFieldNames ? '' : 'displayName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateThreadRequest clone() => CreateThreadRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateThreadRequest copyWith(void Function(CreateThreadRequest) updates) =>
      super.copyWith((message) => updates(message as CreateThreadRequest))
          as CreateThreadRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateThreadRequest create() => CreateThreadRequest._();
  @$core.override
  CreateThreadRequest createEmptyInstance() => create();
  static $pb.PbList<CreateThreadRequest> createRepeated() =>
      $pb.PbList<CreateThreadRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateThreadRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateThreadRequest>(create);
  static CreateThreadRequest? _defaultInstance;

  /// workspace_uid identifies the registered workspace.
  @$pb.TagNumber(1)
  $core.String get workspaceUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceUid() => $_clearField(1);

  /// runtime_uid identifies an available runtime.
  @$pb.TagNumber(2)
  $core.String get runtimeUid => $_getSZ(1);
  @$pb.TagNumber(2)
  set runtimeUid($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRuntimeUid() => $_has(1);
  @$pb.TagNumber(2)
  void clearRuntimeUid() => $_clearField(2);

  /// display_name is the human-readable conversation label.
  @$pb.TagNumber(3)
  $core.String get displayName => $_getSZ(2);
  @$pb.TagNumber(3)
  set displayName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDisplayName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDisplayName() => $_clearField(3);
}

/// CreateThreadResponse returns the new thread.
class CreateThreadResponse extends $pb.GeneratedMessage {
  factory CreateThreadResponse({
    Thread? thread,
  }) {
    final result = create();
    if (thread != null) result.thread = thread;
    return result;
  }

  CreateThreadResponse._();

  factory CreateThreadResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateThreadResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateThreadResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<Thread>(1, _omitFieldNames ? '' : 'thread', subBuilder: Thread.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateThreadResponse clone() =>
      CreateThreadResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateThreadResponse copyWith(void Function(CreateThreadResponse) updates) =>
      super.copyWith((message) => updates(message as CreateThreadResponse))
          as CreateThreadResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateThreadResponse create() => CreateThreadResponse._();
  @$core.override
  CreateThreadResponse createEmptyInstance() => create();
  static $pb.PbList<CreateThreadResponse> createRepeated() =>
      $pb.PbList<CreateThreadResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateThreadResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateThreadResponse>(create);
  static CreateThreadResponse? _defaultInstance;

  /// thread is the newly created thread.
  @$pb.TagNumber(1)
  Thread get thread => $_getN(0);
  @$pb.TagNumber(1)
  set thread(Thread value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasThread() => $_has(0);
  @$pb.TagNumber(1)
  void clearThread() => $_clearField(1);
  @$pb.TagNumber(1)
  Thread ensureThread() => $_ensure(0);
}

/// GetThreadRequest requests one thread.
class GetThreadRequest extends $pb.GeneratedMessage {
  factory GetThreadRequest({
    $core.String? uid,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    return result;
  }

  GetThreadRequest._();

  factory GetThreadRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetThreadRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetThreadRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetThreadRequest clone() => GetThreadRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetThreadRequest copyWith(void Function(GetThreadRequest) updates) =>
      super.copyWith((message) => updates(message as GetThreadRequest))
          as GetThreadRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetThreadRequest create() => GetThreadRequest._();
  @$core.override
  GetThreadRequest createEmptyInstance() => create();
  static $pb.PbList<GetThreadRequest> createRepeated() =>
      $pb.PbList<GetThreadRequest>();
  @$core.pragma('dart2js:noInline')
  static GetThreadRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetThreadRequest>(create);
  static GetThreadRequest? _defaultInstance;

  /// uid is the thread UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);
}

/// GetThreadResponse returns one thread.
class GetThreadResponse extends $pb.GeneratedMessage {
  factory GetThreadResponse({
    Thread? thread,
  }) {
    final result = create();
    if (thread != null) result.thread = thread;
    return result;
  }

  GetThreadResponse._();

  factory GetThreadResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetThreadResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetThreadResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<Thread>(1, _omitFieldNames ? '' : 'thread', subBuilder: Thread.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetThreadResponse clone() => GetThreadResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetThreadResponse copyWith(void Function(GetThreadResponse) updates) =>
      super.copyWith((message) => updates(message as GetThreadResponse))
          as GetThreadResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetThreadResponse create() => GetThreadResponse._();
  @$core.override
  GetThreadResponse createEmptyInstance() => create();
  static $pb.PbList<GetThreadResponse> createRepeated() =>
      $pb.PbList<GetThreadResponse>();
  @$core.pragma('dart2js:noInline')
  static GetThreadResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetThreadResponse>(create);
  static GetThreadResponse? _defaultInstance;

  /// thread is the requested thread.
  @$pb.TagNumber(1)
  Thread get thread => $_getN(0);
  @$pb.TagNumber(1)
  set thread(Thread value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasThread() => $_has(0);
  @$pb.TagNumber(1)
  void clearThread() => $_clearField(1);
  @$pb.TagNumber(1)
  Thread ensureThread() => $_ensure(0);
}

/// ListThreadsRequest requests threads in one workspace.
class ListThreadsRequest extends $pb.GeneratedMessage {
  factory ListThreadsRequest({
    $core.String? workspaceUid,
    $core.int? pageSize,
    $core.String? pageToken,
  }) {
    final result = create();
    if (workspaceUid != null) result.workspaceUid = workspaceUid;
    if (pageSize != null) result.pageSize = pageSize;
    if (pageToken != null) result.pageToken = pageToken;
    return result;
  }

  ListThreadsRequest._();

  factory ListThreadsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListThreadsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListThreadsRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceUid')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'pageSize', $pb.PbFieldType.O3)
    ..aOS(3, _omitFieldNames ? '' : 'pageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListThreadsRequest clone() => ListThreadsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListThreadsRequest copyWith(void Function(ListThreadsRequest) updates) =>
      super.copyWith((message) => updates(message as ListThreadsRequest))
          as ListThreadsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListThreadsRequest create() => ListThreadsRequest._();
  @$core.override
  ListThreadsRequest createEmptyInstance() => create();
  static $pb.PbList<ListThreadsRequest> createRepeated() =>
      $pb.PbList<ListThreadsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListThreadsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListThreadsRequest>(create);
  static ListThreadsRequest? _defaultInstance;

  /// workspace_uid scopes the list to one workspace.
  @$pb.TagNumber(1)
  $core.String get workspaceUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceUid() => $_clearField(1);

  /// page_size is the maximum number of threads to return.
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

/// ListThreadsResponse returns threads in one workspace.
class ListThreadsResponse extends $pb.GeneratedMessage {
  factory ListThreadsResponse({
    $core.Iterable<Thread>? threads,
    $core.String? nextPageToken,
  }) {
    final result = create();
    if (threads != null) result.threads.addAll(threads);
    if (nextPageToken != null) result.nextPageToken = nextPageToken;
    return result;
  }

  ListThreadsResponse._();

  factory ListThreadsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListThreadsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListThreadsResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.manager.control.v1'),
      createEmptyInstance: create)
    ..pc<Thread>(1, _omitFieldNames ? '' : 'threads', $pb.PbFieldType.PM,
        subBuilder: Thread.create)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListThreadsResponse clone() => ListThreadsResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListThreadsResponse copyWith(void Function(ListThreadsResponse) updates) =>
      super.copyWith((message) => updates(message as ListThreadsResponse))
          as ListThreadsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListThreadsResponse create() => ListThreadsResponse._();
  @$core.override
  ListThreadsResponse createEmptyInstance() => create();
  static $pb.PbList<ListThreadsResponse> createRepeated() =>
      $pb.PbList<ListThreadsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListThreadsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListThreadsResponse>(create);
  static ListThreadsResponse? _defaultInstance;

  /// threads contains the current page.
  @$pb.TagNumber(1)
  $pb.PbList<Thread> get threads => $_getList(0);

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

/// DeleteThreadRequest deletes one thread and its semantic event history.
class DeleteThreadRequest extends $pb.GeneratedMessage {
  factory DeleteThreadRequest({
    $core.String? uid,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    return result;
  }

  DeleteThreadRequest._();

  factory DeleteThreadRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteThreadRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteThreadRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteThreadRequest clone() => DeleteThreadRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteThreadRequest copyWith(void Function(DeleteThreadRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteThreadRequest))
          as DeleteThreadRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteThreadRequest create() => DeleteThreadRequest._();
  @$core.override
  DeleteThreadRequest createEmptyInstance() => create();
  static $pb.PbList<DeleteThreadRequest> createRepeated() =>
      $pb.PbList<DeleteThreadRequest>();
  @$core.pragma('dart2js:noInline')
  static DeleteThreadRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteThreadRequest>(create);
  static DeleteThreadRequest? _defaultInstance;

  /// uid is the thread UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);
}

/// DeleteThreadResponse confirms that the thread was deleted.
class DeleteThreadResponse extends $pb.GeneratedMessage {
  factory DeleteThreadResponse() => create();

  DeleteThreadResponse._();

  factory DeleteThreadResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteThreadResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteThreadResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'realtime.me.manager.control.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteThreadResponse clone() =>
      DeleteThreadResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteThreadResponse copyWith(void Function(DeleteThreadResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteThreadResponse))
          as DeleteThreadResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteThreadResponse create() => DeleteThreadResponse._();
  @$core.override
  DeleteThreadResponse createEmptyInstance() => create();
  static $pb.PbList<DeleteThreadResponse> createRepeated() =>
      $pb.PbList<DeleteThreadResponse>();
  @$core.pragma('dart2js:noInline')
  static DeleteThreadResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteThreadResponse>(create);
  static DeleteThreadResponse? _defaultInstance;
}

/// ThreadService manages durable structured conversations.
class ThreadServiceApi {
  final $pb.RpcClient _client;

  ThreadServiceApi(this._client);

  /// CreateThread creates a conversation for one workspace and runtime.
  $async.Future<CreateThreadResponse> createThread(
          $pb.ClientContext? ctx, CreateThreadRequest request) =>
      _client.invoke<CreateThreadResponse>(ctx, 'ThreadService', 'CreateThread',
          request, CreateThreadResponse());

  /// GetThread returns one conversation.
  $async.Future<GetThreadResponse> getThread(
          $pb.ClientContext? ctx, GetThreadRequest request) =>
      _client.invoke<GetThreadResponse>(
          ctx, 'ThreadService', 'GetThread', request, GetThreadResponse());

  /// ListThreads returns conversations in one workspace.
  $async.Future<ListThreadsResponse> listThreads(
          $pb.ClientContext? ctx, ListThreadsRequest request) =>
      _client.invoke<ListThreadsResponse>(
          ctx, 'ThreadService', 'ListThreads', request, ListThreadsResponse());

  /// DeleteThread deletes a conversation and its semantic history.
  $async.Future<DeleteThreadResponse> deleteThread(
          $pb.ClientContext? ctx, DeleteThreadRequest request) =>
      _client.invoke<DeleteThreadResponse>(ctx, 'ThreadService', 'DeleteThread',
          request, DeleteThreadResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
