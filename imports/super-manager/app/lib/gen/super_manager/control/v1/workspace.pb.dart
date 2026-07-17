// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/workspace.proto.

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

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// Workspace is an approved project directory on the agent host.
class Workspace extends $pb.GeneratedMessage {
  factory Workspace({
    $core.String? uid,
    $core.String? displayName,
    $core.String? path,
    $core.String? activeExecutionUid,
    $0.Timestamp? createTime,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (displayName != null) result.displayName = displayName;
    if (path != null) result.path = path;
    if (activeExecutionUid != null)
      result.activeExecutionUid = activeExecutionUid;
    if (createTime != null) result.createTime = createTime;
    return result;
  }

  Workspace._();

  factory Workspace.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Workspace.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Workspace',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..aOS(3, _omitFieldNames ? '' : 'path')
    ..aOS(4, _omitFieldNames ? '' : 'activeExecutionUid')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'createTime',
        subBuilder: $0.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Workspace clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Workspace copyWith(void Function(Workspace) updates) =>
      super.copyWith((message) => updates(message as Workspace)) as Workspace;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Workspace create() => Workspace._();
  @$core.override
  Workspace createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Workspace getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Workspace>(create);
  static Workspace? _defaultInstance;

  /// uid is the server-assigned UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  /// display_name is the human-readable project label.
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);

  /// path is the canonical absolute path within an allowed workspace root.
  @$pb.TagNumber(3)
  $core.String get path => $_getSZ(2);
  @$pb.TagNumber(3)
  set path($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPath() => $_has(2);
  @$pb.TagNumber(3)
  void clearPath() => $_clearField(3);

  /// active_execution_uid is set while this workspace holds a structured writer lease.
  @$pb.TagNumber(4)
  $core.String get activeExecutionUid => $_getSZ(3);
  @$pb.TagNumber(4)
  set activeExecutionUid($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasActiveExecutionUid() => $_has(3);
  @$pb.TagNumber(4)
  void clearActiveExecutionUid() => $_clearField(4);

  /// create_time is the registration time.
  @$pb.TagNumber(5)
  $0.Timestamp get createTime => $_getN(4);
  @$pb.TagNumber(5)
  set createTime($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasCreateTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearCreateTime() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureCreateTime() => $_ensure(4);
}

/// CreateWorkspaceRequest registers an approved server-local project directory.
class CreateWorkspaceRequest extends $pb.GeneratedMessage {
  factory CreateWorkspaceRequest({
    $core.String? displayName,
    $core.String? path,
  }) {
    final result = create();
    if (displayName != null) result.displayName = displayName;
    if (path != null) result.path = path;
    return result;
  }

  CreateWorkspaceRequest._();

  factory CreateWorkspaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateWorkspaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateWorkspaceRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'displayName')
    ..aOS(2, _omitFieldNames ? '' : 'path')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWorkspaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWorkspaceRequest copyWith(
          void Function(CreateWorkspaceRequest) updates) =>
      super.copyWith((message) => updates(message as CreateWorkspaceRequest))
          as CreateWorkspaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateWorkspaceRequest create() => CreateWorkspaceRequest._();
  @$core.override
  CreateWorkspaceRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CreateWorkspaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateWorkspaceRequest>(create);
  static CreateWorkspaceRequest? _defaultInstance;

  /// display_name is the human-readable project label.
  @$pb.TagNumber(1)
  $core.String get displayName => $_getSZ(0);
  @$pb.TagNumber(1)
  set displayName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDisplayName() => $_has(0);
  @$pb.TagNumber(1)
  void clearDisplayName() => $_clearField(1);

  /// path is an absolute path contained by a configured allowed root.
  @$pb.TagNumber(2)
  $core.String get path => $_getSZ(1);
  @$pb.TagNumber(2)
  set path($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPath() => $_has(1);
  @$pb.TagNumber(2)
  void clearPath() => $_clearField(2);
}

/// CreateWorkspaceResponse returns the registered workspace.
class CreateWorkspaceResponse extends $pb.GeneratedMessage {
  factory CreateWorkspaceResponse({
    Workspace? workspace,
  }) {
    final result = create();
    if (workspace != null) result.workspace = workspace;
    return result;
  }

  CreateWorkspaceResponse._();

  factory CreateWorkspaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateWorkspaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateWorkspaceResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<Workspace>(1, _omitFieldNames ? '' : 'workspace',
        subBuilder: Workspace.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWorkspaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWorkspaceResponse copyWith(
          void Function(CreateWorkspaceResponse) updates) =>
      super.copyWith((message) => updates(message as CreateWorkspaceResponse))
          as CreateWorkspaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateWorkspaceResponse create() => CreateWorkspaceResponse._();
  @$core.override
  CreateWorkspaceResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CreateWorkspaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateWorkspaceResponse>(create);
  static CreateWorkspaceResponse? _defaultInstance;

  /// workspace is the newly registered workspace.
  @$pb.TagNumber(1)
  Workspace get workspace => $_getN(0);
  @$pb.TagNumber(1)
  set workspace(Workspace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspace() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspace() => $_clearField(1);
  @$pb.TagNumber(1)
  Workspace ensureWorkspace() => $_ensure(0);
}

/// GetWorkspaceRequest requests one registered workspace.
class GetWorkspaceRequest extends $pb.GeneratedMessage {
  factory GetWorkspaceRequest({
    $core.String? uid,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    return result;
  }

  GetWorkspaceRequest._();

  factory GetWorkspaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetWorkspaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetWorkspaceRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceRequest copyWith(void Function(GetWorkspaceRequest) updates) =>
      super.copyWith((message) => updates(message as GetWorkspaceRequest))
          as GetWorkspaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetWorkspaceRequest create() => GetWorkspaceRequest._();
  @$core.override
  GetWorkspaceRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetWorkspaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetWorkspaceRequest>(create);
  static GetWorkspaceRequest? _defaultInstance;

  /// uid is the workspace UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);
}

/// GetWorkspaceResponse returns one registered workspace.
class GetWorkspaceResponse extends $pb.GeneratedMessage {
  factory GetWorkspaceResponse({
    Workspace? workspace,
  }) {
    final result = create();
    if (workspace != null) result.workspace = workspace;
    return result;
  }

  GetWorkspaceResponse._();

  factory GetWorkspaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetWorkspaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetWorkspaceResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<Workspace>(1, _omitFieldNames ? '' : 'workspace',
        subBuilder: Workspace.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetWorkspaceResponse copyWith(void Function(GetWorkspaceResponse) updates) =>
      super.copyWith((message) => updates(message as GetWorkspaceResponse))
          as GetWorkspaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetWorkspaceResponse create() => GetWorkspaceResponse._();
  @$core.override
  GetWorkspaceResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetWorkspaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetWorkspaceResponse>(create);
  static GetWorkspaceResponse? _defaultInstance;

  /// workspace is the requested workspace.
  @$pb.TagNumber(1)
  Workspace get workspace => $_getN(0);
  @$pb.TagNumber(1)
  set workspace(Workspace value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspace() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspace() => $_clearField(1);
  @$pb.TagNumber(1)
  Workspace ensureWorkspace() => $_ensure(0);
}

/// ListWorkspacesRequest requests registered workspaces.
class ListWorkspacesRequest extends $pb.GeneratedMessage {
  factory ListWorkspacesRequest({
    $core.int? pageSize,
    $core.String? pageToken,
  }) {
    final result = create();
    if (pageSize != null) result.pageSize = pageSize;
    if (pageToken != null) result.pageToken = pageToken;
    return result;
  }

  ListWorkspacesRequest._();

  factory ListWorkspacesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListWorkspacesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListWorkspacesRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aI(1, _omitFieldNames ? '' : 'pageSize')
    ..aOS(2, _omitFieldNames ? '' : 'pageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspacesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspacesRequest copyWith(
          void Function(ListWorkspacesRequest) updates) =>
      super.copyWith((message) => updates(message as ListWorkspacesRequest))
          as ListWorkspacesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListWorkspacesRequest create() => ListWorkspacesRequest._();
  @$core.override
  ListWorkspacesRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListWorkspacesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListWorkspacesRequest>(create);
  static ListWorkspacesRequest? _defaultInstance;

  /// page_size is the maximum number of workspaces to return.
  @$pb.TagNumber(1)
  $core.int get pageSize => $_getIZ(0);
  @$pb.TagNumber(1)
  set pageSize($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPageSize() => $_has(0);
  @$pb.TagNumber(1)
  void clearPageSize() => $_clearField(1);

  /// page_token is an opaque continuation token returned by the server.
  @$pb.TagNumber(2)
  $core.String get pageToken => $_getSZ(1);
  @$pb.TagNumber(2)
  set pageToken($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasPageToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearPageToken() => $_clearField(2);
}

/// ListWorkspacesResponse returns registered workspaces.
class ListWorkspacesResponse extends $pb.GeneratedMessage {
  factory ListWorkspacesResponse({
    $core.Iterable<Workspace>? workspaces,
    $core.String? nextPageToken,
  }) {
    final result = create();
    if (workspaces != null) result.workspaces.addAll(workspaces);
    if (nextPageToken != null) result.nextPageToken = nextPageToken;
    return result;
  }

  ListWorkspacesResponse._();

  factory ListWorkspacesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListWorkspacesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListWorkspacesResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..pPM<Workspace>(1, _omitFieldNames ? '' : 'workspaces',
        subBuilder: Workspace.create)
    ..aOS(2, _omitFieldNames ? '' : 'nextPageToken')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspacesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListWorkspacesResponse copyWith(
          void Function(ListWorkspacesResponse) updates) =>
      super.copyWith((message) => updates(message as ListWorkspacesResponse))
          as ListWorkspacesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListWorkspacesResponse create() => ListWorkspacesResponse._();
  @$core.override
  ListWorkspacesResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListWorkspacesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListWorkspacesResponse>(create);
  static ListWorkspacesResponse? _defaultInstance;

  /// workspaces contains the current page.
  @$pb.TagNumber(1)
  $pb.PbList<Workspace> get workspaces => $_getList(0);

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

/// DeleteWorkspaceRequest deletes one registered workspace record without deleting files.
class DeleteWorkspaceRequest extends $pb.GeneratedMessage {
  factory DeleteWorkspaceRequest({
    $core.String? uid,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    return result;
  }

  DeleteWorkspaceRequest._();

  factory DeleteWorkspaceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteWorkspaceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWorkspaceRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceRequest copyWith(
          void Function(DeleteWorkspaceRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteWorkspaceRequest))
          as DeleteWorkspaceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceRequest create() => DeleteWorkspaceRequest._();
  @$core.override
  DeleteWorkspaceRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWorkspaceRequest>(create);
  static DeleteWorkspaceRequest? _defaultInstance;

  /// uid is the workspace UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);
}

/// DeleteWorkspaceResponse confirms that the workspace record was deleted.
class DeleteWorkspaceResponse extends $pb.GeneratedMessage {
  factory DeleteWorkspaceResponse() => create();

  DeleteWorkspaceResponse._();

  factory DeleteWorkspaceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteWorkspaceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteWorkspaceResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteWorkspaceResponse copyWith(
          void Function(DeleteWorkspaceResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteWorkspaceResponse))
          as DeleteWorkspaceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceResponse create() => DeleteWorkspaceResponse._();
  @$core.override
  DeleteWorkspaceResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static DeleteWorkspaceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteWorkspaceResponse>(create);
  static DeleteWorkspaceResponse? _defaultInstance;
}

/// WorkspaceService manages approved server-local project directories.
class WorkspaceServiceApi {
  final $pb.RpcClient _client;

  WorkspaceServiceApi(this._client);

  /// CreateWorkspace registers a path contained by an allowed workspace root.
  $async.Future<CreateWorkspaceResponse> createWorkspace(
          $pb.ClientContext? ctx, CreateWorkspaceRequest request) =>
      _client.invoke<CreateWorkspaceResponse>(ctx, 'WorkspaceService',
          'CreateWorkspace', request, CreateWorkspaceResponse());

  /// GetWorkspace returns one registered workspace.
  $async.Future<GetWorkspaceResponse> getWorkspace(
          $pb.ClientContext? ctx, GetWorkspaceRequest request) =>
      _client.invoke<GetWorkspaceResponse>(ctx, 'WorkspaceService',
          'GetWorkspace', request, GetWorkspaceResponse());

  /// ListWorkspaces returns registered workspaces.
  $async.Future<ListWorkspacesResponse> listWorkspaces(
          $pb.ClientContext? ctx, ListWorkspacesRequest request) =>
      _client.invoke<ListWorkspacesResponse>(ctx, 'WorkspaceService',
          'ListWorkspaces', request, ListWorkspacesResponse());

  /// DeleteWorkspace removes only the registration record.
  $async.Future<DeleteWorkspaceResponse> deleteWorkspace(
          $pb.ClientContext? ctx, DeleteWorkspaceRequest request) =>
      _client.invoke<DeleteWorkspaceResponse>(ctx, 'WorkspaceService',
          'DeleteWorkspace', request, DeleteWorkspaceResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
