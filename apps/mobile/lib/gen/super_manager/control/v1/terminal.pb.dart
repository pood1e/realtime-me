// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/terminal.proto.

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
import 'terminal.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'terminal.pbenum.dart';

/// TerminalSession describes one shell created and owned by Super Manager.
class TerminalSession extends $pb.GeneratedMessage {
  factory TerminalSession({
    $core.String? uid,
    $core.String? workspaceUid,
    $core.String? displayName,
    $core.String? cwd,
    TerminalSessionState? state,
    $0.Timestamp? createTime,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (workspaceUid != null) result.workspaceUid = workspaceUid;
    if (displayName != null) result.displayName = displayName;
    if (cwd != null) result.cwd = cwd;
    if (state != null) result.state = state;
    if (createTime != null) result.createTime = createTime;
    return result;
  }

  TerminalSession._();

  factory TerminalSession.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalSession.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalSession',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aOS(2, _omitFieldNames ? '' : 'workspaceUid')
    ..aOS(3, _omitFieldNames ? '' : 'displayName')
    ..aOS(4, _omitFieldNames ? '' : 'cwd')
    ..e<TerminalSessionState>(
        5, _omitFieldNames ? '' : 'state', $pb.PbFieldType.OE,
        defaultOrMaker: TerminalSessionState.TERMINAL_SESSION_STATE_UNSPECIFIED,
        valueOf: TerminalSessionState.valueOf,
        enumValues: TerminalSessionState.values)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'createTime',
        subBuilder: $0.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalSession clone() => TerminalSession()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalSession copyWith(void Function(TerminalSession) updates) =>
      super.copyWith((message) => updates(message as TerminalSession))
          as TerminalSession;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalSession create() => TerminalSession._();
  @$core.override
  TerminalSession createEmptyInstance() => create();
  static $pb.PbList<TerminalSession> createRepeated() =>
      $pb.PbList<TerminalSession>();
  @$core.pragma('dart2js:noInline')
  static TerminalSession getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalSession>(create);
  static TerminalSession? _defaultInstance;

  /// uid is the server-assigned UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  /// workspace_uid identifies the workspace used as the initial directory.
  @$pb.TagNumber(2)
  $core.String get workspaceUid => $_getSZ(1);
  @$pb.TagNumber(2)
  set workspaceUid($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWorkspaceUid() => $_has(1);
  @$pb.TagNumber(2)
  void clearWorkspaceUid() => $_clearField(2);

  /// display_name is the human-readable shell label.
  @$pb.TagNumber(3)
  $core.String get displayName => $_getSZ(2);
  @$pb.TagNumber(3)
  set displayName($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDisplayName() => $_has(2);
  @$pb.TagNumber(3)
  void clearDisplayName() => $_clearField(3);

  /// cwd is the canonical initial working directory.
  @$pb.TagNumber(4)
  $core.String get cwd => $_getSZ(3);
  @$pb.TagNumber(4)
  set cwd($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCwd() => $_has(3);
  @$pb.TagNumber(4)
  void clearCwd() => $_clearField(4);

  /// state is the projected shell lifecycle state.
  @$pb.TagNumber(5)
  TerminalSessionState get state => $_getN(4);
  @$pb.TagNumber(5)
  set state(TerminalSessionState value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasState() => $_has(4);
  @$pb.TagNumber(5)
  void clearState() => $_clearField(5);

  /// create_time is the shell creation time.
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
}

/// CreateTerminalSessionRequest creates a tmux-backed shell in one workspace.
class CreateTerminalSessionRequest extends $pb.GeneratedMessage {
  factory CreateTerminalSessionRequest({
    $core.String? workspaceUid,
    $core.String? displayName,
    $core.int? columns,
    $core.int? rows,
  }) {
    final result = create();
    if (workspaceUid != null) result.workspaceUid = workspaceUid;
    if (displayName != null) result.displayName = displayName;
    if (columns != null) result.columns = columns;
    if (rows != null) result.rows = rows;
    return result;
  }

  CreateTerminalSessionRequest._();

  factory CreateTerminalSessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateTerminalSessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateTerminalSessionRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceUid')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'columns', $pb.PbFieldType.OU3)
    ..a<$core.int>(4, _omitFieldNames ? '' : 'rows', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTerminalSessionRequest clone() =>
      CreateTerminalSessionRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTerminalSessionRequest copyWith(
          void Function(CreateTerminalSessionRequest) updates) =>
      super.copyWith(
              (message) => updates(message as CreateTerminalSessionRequest))
          as CreateTerminalSessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateTerminalSessionRequest create() =>
      CreateTerminalSessionRequest._();
  @$core.override
  CreateTerminalSessionRequest createEmptyInstance() => create();
  static $pb.PbList<CreateTerminalSessionRequest> createRepeated() =>
      $pb.PbList<CreateTerminalSessionRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateTerminalSessionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateTerminalSessionRequest>(create);
  static CreateTerminalSessionRequest? _defaultInstance;

  /// workspace_uid identifies the registered workspace.
  @$pb.TagNumber(1)
  $core.String get workspaceUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceUid() => $_clearField(1);

  /// display_name is the human-readable shell label.
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);

  /// columns is the initial terminal width.
  @$pb.TagNumber(3)
  $core.int get columns => $_getIZ(2);
  @$pb.TagNumber(3)
  set columns($core.int value) => $_setUnsignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasColumns() => $_has(2);
  @$pb.TagNumber(3)
  void clearColumns() => $_clearField(3);

  /// rows is the initial terminal height.
  @$pb.TagNumber(4)
  $core.int get rows => $_getIZ(3);
  @$pb.TagNumber(4)
  set rows($core.int value) => $_setUnsignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRows() => $_has(3);
  @$pb.TagNumber(4)
  void clearRows() => $_clearField(4);
}

/// CreateTerminalSessionResponse returns the new terminal session.
class CreateTerminalSessionResponse extends $pb.GeneratedMessage {
  factory CreateTerminalSessionResponse({
    TerminalSession? terminalSession,
  }) {
    final result = create();
    if (terminalSession != null) result.terminalSession = terminalSession;
    return result;
  }

  CreateTerminalSessionResponse._();

  factory CreateTerminalSessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateTerminalSessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateTerminalSessionResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<TerminalSession>(1, _omitFieldNames ? '' : 'terminalSession',
        subBuilder: TerminalSession.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTerminalSessionResponse clone() =>
      CreateTerminalSessionResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateTerminalSessionResponse copyWith(
          void Function(CreateTerminalSessionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as CreateTerminalSessionResponse))
          as CreateTerminalSessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateTerminalSessionResponse create() =>
      CreateTerminalSessionResponse._();
  @$core.override
  CreateTerminalSessionResponse createEmptyInstance() => create();
  static $pb.PbList<CreateTerminalSessionResponse> createRepeated() =>
      $pb.PbList<CreateTerminalSessionResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateTerminalSessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateTerminalSessionResponse>(create);
  static CreateTerminalSessionResponse? _defaultInstance;

  /// terminal_session is the newly created terminal session.
  @$pb.TagNumber(1)
  TerminalSession get terminalSession => $_getN(0);
  @$pb.TagNumber(1)
  set terminalSession(TerminalSession value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasTerminalSession() => $_has(0);
  @$pb.TagNumber(1)
  void clearTerminalSession() => $_clearField(1);
  @$pb.TagNumber(1)
  TerminalSession ensureTerminalSession() => $_ensure(0);
}

/// GetTerminalSessionRequest requests one terminal session.
class GetTerminalSessionRequest extends $pb.GeneratedMessage {
  factory GetTerminalSessionRequest({
    $core.String? uid,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    return result;
  }

  GetTerminalSessionRequest._();

  factory GetTerminalSessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetTerminalSessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTerminalSessionRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTerminalSessionRequest clone() =>
      GetTerminalSessionRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTerminalSessionRequest copyWith(
          void Function(GetTerminalSessionRequest) updates) =>
      super.copyWith((message) => updates(message as GetTerminalSessionRequest))
          as GetTerminalSessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetTerminalSessionRequest create() => GetTerminalSessionRequest._();
  @$core.override
  GetTerminalSessionRequest createEmptyInstance() => create();
  static $pb.PbList<GetTerminalSessionRequest> createRepeated() =>
      $pb.PbList<GetTerminalSessionRequest>();
  @$core.pragma('dart2js:noInline')
  static GetTerminalSessionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTerminalSessionRequest>(create);
  static GetTerminalSessionRequest? _defaultInstance;

  /// uid is the terminal session UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);
}

/// GetTerminalSessionResponse returns one terminal session.
class GetTerminalSessionResponse extends $pb.GeneratedMessage {
  factory GetTerminalSessionResponse({
    TerminalSession? terminalSession,
  }) {
    final result = create();
    if (terminalSession != null) result.terminalSession = terminalSession;
    return result;
  }

  GetTerminalSessionResponse._();

  factory GetTerminalSessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetTerminalSessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetTerminalSessionResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<TerminalSession>(1, _omitFieldNames ? '' : 'terminalSession',
        subBuilder: TerminalSession.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTerminalSessionResponse clone() =>
      GetTerminalSessionResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetTerminalSessionResponse copyWith(
          void Function(GetTerminalSessionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as GetTerminalSessionResponse))
          as GetTerminalSessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetTerminalSessionResponse create() => GetTerminalSessionResponse._();
  @$core.override
  GetTerminalSessionResponse createEmptyInstance() => create();
  static $pb.PbList<GetTerminalSessionResponse> createRepeated() =>
      $pb.PbList<GetTerminalSessionResponse>();
  @$core.pragma('dart2js:noInline')
  static GetTerminalSessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetTerminalSessionResponse>(create);
  static GetTerminalSessionResponse? _defaultInstance;

  /// terminal_session is the requested terminal session.
  @$pb.TagNumber(1)
  TerminalSession get terminalSession => $_getN(0);
  @$pb.TagNumber(1)
  set terminalSession(TerminalSession value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasTerminalSession() => $_has(0);
  @$pb.TagNumber(1)
  void clearTerminalSession() => $_clearField(1);
  @$pb.TagNumber(1)
  TerminalSession ensureTerminalSession() => $_ensure(0);
}

/// ListTerminalSessionsRequest requests terminal sessions in one workspace.
class ListTerminalSessionsRequest extends $pb.GeneratedMessage {
  factory ListTerminalSessionsRequest({
    $core.String? workspaceUid,
  }) {
    final result = create();
    if (workspaceUid != null) result.workspaceUid = workspaceUid;
    return result;
  }

  ListTerminalSessionsRequest._();

  factory ListTerminalSessionsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListTerminalSessionsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTerminalSessionsRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'workspaceUid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTerminalSessionsRequest clone() =>
      ListTerminalSessionsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTerminalSessionsRequest copyWith(
          void Function(ListTerminalSessionsRequest) updates) =>
      super.copyWith(
              (message) => updates(message as ListTerminalSessionsRequest))
          as ListTerminalSessionsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListTerminalSessionsRequest create() =>
      ListTerminalSessionsRequest._();
  @$core.override
  ListTerminalSessionsRequest createEmptyInstance() => create();
  static $pb.PbList<ListTerminalSessionsRequest> createRepeated() =>
      $pb.PbList<ListTerminalSessionsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListTerminalSessionsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListTerminalSessionsRequest>(create);
  static ListTerminalSessionsRequest? _defaultInstance;

  /// workspace_uid scopes the list to one workspace.
  @$pb.TagNumber(1)
  $core.String get workspaceUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set workspaceUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasWorkspaceUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearWorkspaceUid() => $_clearField(1);
}

/// ListTerminalSessionsResponse returns terminal sessions in one workspace.
class ListTerminalSessionsResponse extends $pb.GeneratedMessage {
  factory ListTerminalSessionsResponse({
    $core.Iterable<TerminalSession>? terminalSessions,
  }) {
    final result = create();
    if (terminalSessions != null)
      result.terminalSessions.addAll(terminalSessions);
    return result;
  }

  ListTerminalSessionsResponse._();

  factory ListTerminalSessionsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListTerminalSessionsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListTerminalSessionsResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..pc<TerminalSession>(
        1, _omitFieldNames ? '' : 'terminalSessions', $pb.PbFieldType.PM,
        subBuilder: TerminalSession.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTerminalSessionsResponse clone() =>
      ListTerminalSessionsResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListTerminalSessionsResponse copyWith(
          void Function(ListTerminalSessionsResponse) updates) =>
      super.copyWith(
              (message) => updates(message as ListTerminalSessionsResponse))
          as ListTerminalSessionsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListTerminalSessionsResponse create() =>
      ListTerminalSessionsResponse._();
  @$core.override
  ListTerminalSessionsResponse createEmptyInstance() => create();
  static $pb.PbList<ListTerminalSessionsResponse> createRepeated() =>
      $pb.PbList<ListTerminalSessionsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListTerminalSessionsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListTerminalSessionsResponse>(create);
  static ListTerminalSessionsResponse? _defaultInstance;

  /// terminal_sessions contains the bounded workspace collection.
  @$pb.TagNumber(1)
  $pb.PbList<TerminalSession> get terminalSessions => $_getList(0);
}

/// DeleteTerminalSessionRequest closes and deletes one terminal session.
class DeleteTerminalSessionRequest extends $pb.GeneratedMessage {
  factory DeleteTerminalSessionRequest({
    $core.String? uid,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    return result;
  }

  DeleteTerminalSessionRequest._();

  factory DeleteTerminalSessionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteTerminalSessionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteTerminalSessionRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTerminalSessionRequest clone() =>
      DeleteTerminalSessionRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTerminalSessionRequest copyWith(
          void Function(DeleteTerminalSessionRequest) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteTerminalSessionRequest))
          as DeleteTerminalSessionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteTerminalSessionRequest create() =>
      DeleteTerminalSessionRequest._();
  @$core.override
  DeleteTerminalSessionRequest createEmptyInstance() => create();
  static $pb.PbList<DeleteTerminalSessionRequest> createRepeated() =>
      $pb.PbList<DeleteTerminalSessionRequest>();
  @$core.pragma('dart2js:noInline')
  static DeleteTerminalSessionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteTerminalSessionRequest>(create);
  static DeleteTerminalSessionRequest? _defaultInstance;

  /// uid is the terminal session UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);
}

/// DeleteTerminalSessionResponse confirms that the terminal session was closed.
class DeleteTerminalSessionResponse extends $pb.GeneratedMessage {
  factory DeleteTerminalSessionResponse() => create();

  DeleteTerminalSessionResponse._();

  factory DeleteTerminalSessionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteTerminalSessionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteTerminalSessionResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTerminalSessionResponse clone() =>
      DeleteTerminalSessionResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteTerminalSessionResponse copyWith(
          void Function(DeleteTerminalSessionResponse) updates) =>
      super.copyWith(
              (message) => updates(message as DeleteTerminalSessionResponse))
          as DeleteTerminalSessionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteTerminalSessionResponse create() =>
      DeleteTerminalSessionResponse._();
  @$core.override
  DeleteTerminalSessionResponse createEmptyInstance() => create();
  static $pb.PbList<DeleteTerminalSessionResponse> createRepeated() =>
      $pb.PbList<DeleteTerminalSessionResponse>();
  @$core.pragma('dart2js:noInline')
  static DeleteTerminalSessionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteTerminalSessionResponse>(create);
  static DeleteTerminalSessionResponse? _defaultInstance;
}

/// TerminalService manages tmux-backed raw shell sessions.
class TerminalServiceApi {
  final $pb.RpcClient _client;

  TerminalServiceApi(this._client);

  /// CreateTerminalSession creates a raw shell owned by Super Manager.
  $async.Future<CreateTerminalSessionResponse> createTerminalSession(
          $pb.ClientContext? ctx, CreateTerminalSessionRequest request) =>
      _client.invoke<CreateTerminalSessionResponse>(ctx, 'TerminalService',
          'CreateTerminalSession', request, CreateTerminalSessionResponse());

  /// GetTerminalSession returns one raw shell session.
  $async.Future<GetTerminalSessionResponse> getTerminalSession(
          $pb.ClientContext? ctx, GetTerminalSessionRequest request) =>
      _client.invoke<GetTerminalSessionResponse>(ctx, 'TerminalService',
          'GetTerminalSession', request, GetTerminalSessionResponse());

  /// ListTerminalSessions returns raw shell sessions in one workspace.
  $async.Future<ListTerminalSessionsResponse> listTerminalSessions(
          $pb.ClientContext? ctx, ListTerminalSessionsRequest request) =>
      _client.invoke<ListTerminalSessionsResponse>(ctx, 'TerminalService',
          'ListTerminalSessions', request, ListTerminalSessionsResponse());

  /// DeleteTerminalSession closes the shell and removes its record.
  $async.Future<DeleteTerminalSessionResponse> deleteTerminalSession(
          $pb.ClientContext? ctx, DeleteTerminalSessionRequest request) =>
      _client.invoke<DeleteTerminalSessionResponse>(ctx, 'TerminalService',
          'DeleteTerminalSession', request, DeleteTerminalSessionResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
