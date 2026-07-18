// This is a generated file - do not edit.
//
// Generated from realtime/me/site/v1/projects.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import '../../../../google/protobuf/timestamp.pb.dart' as $0;
import 'projects.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'projects.pbenum.dart';

/// ListProjectsRequest is the request for the curated projects.
class ListProjectsRequest extends $pb.GeneratedMessage {
  factory ListProjectsRequest() => create();

  ListProjectsRequest._();

  factory ListProjectsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListProjectsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListProjectsRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'realtime.me.site.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProjectsRequest clone() => ListProjectsRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProjectsRequest copyWith(void Function(ListProjectsRequest) updates) =>
      super.copyWith((message) => updates(message as ListProjectsRequest))
          as ListProjectsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListProjectsRequest create() => ListProjectsRequest._();
  @$core.override
  ListProjectsRequest createEmptyInstance() => create();
  static $pb.PbList<ListProjectsRequest> createRepeated() =>
      $pb.PbList<ListProjectsRequest>();
  @$core.pragma('dart2js:noInline')
  static ListProjectsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListProjectsRequest>(create);
  static ListProjectsRequest? _defaultInstance;
}

/// ListProjectsResponse carries the curated projects in display order.
class ListProjectsResponse extends $pb.GeneratedMessage {
  factory ListProjectsResponse({
    $core.Iterable<Project>? projects,
  }) {
    final result = create();
    if (projects != null) result.projects.addAll(projects);
    return result;
  }

  ListProjectsResponse._();

  factory ListProjectsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListProjectsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListProjectsResponse',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'realtime.me.site.v1'),
      createEmptyInstance: create)
    ..pc<Project>(1, _omitFieldNames ? '' : 'projects', $pb.PbFieldType.PM,
        subBuilder: Project.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProjectsResponse clone() =>
      ListProjectsResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListProjectsResponse copyWith(void Function(ListProjectsResponse) updates) =>
      super.copyWith((message) => updates(message as ListProjectsResponse))
          as ListProjectsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListProjectsResponse create() => ListProjectsResponse._();
  @$core.override
  ListProjectsResponse createEmptyInstance() => create();
  static $pb.PbList<ListProjectsResponse> createRepeated() =>
      $pb.PbList<ListProjectsResponse>();
  @$core.pragma('dart2js:noInline')
  static ListProjectsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListProjectsResponse>(create);
  static ListProjectsResponse? _defaultInstance;

  /// projects are the curated GitHub projects in display order.
  @$pb.TagNumber(1)
  $pb.PbList<Project> get projects => $_getList(0);
}

/// Project is one curated GitHub repository presented as a personal project.
class Project extends $pb.GeneratedMessage {
  factory Project({
    $core.String? uid,
    $core.String? displayName,
    $core.String? description,
    $core.String? summary,
    ProjectVisibility? visibility,
    $core.String? primaryLanguage,
    $core.Iterable<$core.String>? topics,
    $core.int? starCount,
    $core.String? repositoryUrl,
    $core.String? homepageUrl,
    $0.Timestamp? lastPushTime,
    $0.Timestamp? createTime,
    $core.bool? archived,
    $core.Iterable<LanguageShare>? languages,
    $core.Iterable<$core.int>? commitActivity,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (displayName != null) result.displayName = displayName;
    if (description != null) result.description = description;
    if (summary != null) result.summary = summary;
    if (visibility != null) result.visibility = visibility;
    if (primaryLanguage != null) result.primaryLanguage = primaryLanguage;
    if (topics != null) result.topics.addAll(topics);
    if (starCount != null) result.starCount = starCount;
    if (repositoryUrl != null) result.repositoryUrl = repositoryUrl;
    if (homepageUrl != null) result.homepageUrl = homepageUrl;
    if (lastPushTime != null) result.lastPushTime = lastPushTime;
    if (createTime != null) result.createTime = createTime;
    if (archived != null) result.archived = archived;
    if (languages != null) result.languages.addAll(languages);
    if (commitActivity != null) result.commitActivity.addAll(commitActivity);
    return result;
  }

  Project._();

  factory Project.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Project.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Project',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'realtime.me.site.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..aOS(3, _omitFieldNames ? '' : 'description')
    ..aOS(4, _omitFieldNames ? '' : 'summary')
    ..e<ProjectVisibility>(
        5, _omitFieldNames ? '' : 'visibility', $pb.PbFieldType.OE,
        defaultOrMaker: ProjectVisibility.PROJECT_VISIBILITY_UNSPECIFIED,
        valueOf: ProjectVisibility.valueOf,
        enumValues: ProjectVisibility.values)
    ..aOS(6, _omitFieldNames ? '' : 'primaryLanguage')
    ..pPS(7, _omitFieldNames ? '' : 'topics')
    ..a<$core.int>(8, _omitFieldNames ? '' : 'starCount', $pb.PbFieldType.O3)
    ..aOS(9, _omitFieldNames ? '' : 'repositoryUrl')
    ..aOS(10, _omitFieldNames ? '' : 'homepageUrl')
    ..aOM<$0.Timestamp>(11, _omitFieldNames ? '' : 'lastPushTime',
        subBuilder: $0.Timestamp.create)
    ..aOM<$0.Timestamp>(12, _omitFieldNames ? '' : 'createTime',
        subBuilder: $0.Timestamp.create)
    ..aOB(13, _omitFieldNames ? '' : 'archived')
    ..pc<LanguageShare>(
        14, _omitFieldNames ? '' : 'languages', $pb.PbFieldType.PM,
        subBuilder: LanguageShare.create)
    ..p<$core.int>(
        15, _omitFieldNames ? '' : 'commitActivity', $pb.PbFieldType.K3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Project clone() => Project()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Project copyWith(void Function(Project) updates) =>
      super.copyWith((message) => updates(message as Project)) as Project;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Project create() => Project._();
  @$core.override
  Project createEmptyInstance() => create();
  static $pb.PbList<Project> createRepeated() => $pb.PbList<Project>();
  @$core.pragma('dart2js:noInline')
  static Project getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Project>(create);
  static Project? _defaultInstance;

  /// uid is the system-assigned opaque identifier for this project.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  /// display_name is the shown project title; it defaults to the repository name.
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);

  /// description is the repository's own short description from GitHub.
  @$pb.TagNumber(3)
  $core.String get description => $_getSZ(2);
  @$pb.TagNumber(3)
  set description($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasDescription() => $_has(2);
  @$pb.TagNumber(3)
  void clearDescription() => $_clearField(3);

  /// summary is the owner's own blurb for this project, shown on the card in place
  /// of the repository's GitHub description. It is written by hand and is empty
  /// when none has been written.
  @$pb.TagNumber(4)
  $core.String get summary => $_getSZ(3);
  @$pb.TagNumber(4)
  set summary($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSummary() => $_has(3);
  @$pb.TagNumber(4)
  void clearSummary() => $_clearField(4);

  /// visibility indicates whether the source repository is public or private.
  @$pb.TagNumber(5)
  ProjectVisibility get visibility => $_getN(4);
  @$pb.TagNumber(5)
  set visibility(ProjectVisibility value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasVisibility() => $_has(4);
  @$pb.TagNumber(5)
  void clearVisibility() => $_clearField(5);

  /// primary_language is the repository's main programming language.
  @$pb.TagNumber(6)
  $core.String get primaryLanguage => $_getSZ(5);
  @$pb.TagNumber(6)
  set primaryLanguage($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPrimaryLanguage() => $_has(5);
  @$pb.TagNumber(6)
  void clearPrimaryLanguage() => $_clearField(6);

  /// topics are the repository topics used as tags.
  @$pb.TagNumber(7)
  $pb.PbList<$core.String> get topics => $_getList(6);

  /// star_count is the repository's stargazer count.
  @$pb.TagNumber(8)
  $core.int get starCount => $_getIZ(7);
  @$pb.TagNumber(8)
  set starCount($core.int value) => $_setSignedInt32(7, value);
  @$pb.TagNumber(8)
  $core.bool hasStarCount() => $_has(7);
  @$pb.TagNumber(8)
  void clearStarCount() => $_clearField(8);

  /// repository_url is the GitHub page for the repository. It is omitted for
  /// private projects on the public surface.
  @$pb.TagNumber(9)
  $core.String get repositoryUrl => $_getSZ(8);
  @$pb.TagNumber(9)
  set repositoryUrl($core.String value) => $_setString(8, value);
  @$pb.TagNumber(9)
  $core.bool hasRepositoryUrl() => $_has(8);
  @$pb.TagNumber(9)
  void clearRepositoryUrl() => $_clearField(9);

  /// homepage_url is the project's homepage or live demo, if any.
  @$pb.TagNumber(10)
  $core.String get homepageUrl => $_getSZ(9);
  @$pb.TagNumber(10)
  set homepageUrl($core.String value) => $_setString(9, value);
  @$pb.TagNumber(10)
  $core.bool hasHomepageUrl() => $_has(9);
  @$pb.TagNumber(10)
  void clearHomepageUrl() => $_clearField(10);

  /// last_push_time is the most recent push to the repository.
  @$pb.TagNumber(11)
  $0.Timestamp get lastPushTime => $_getN(10);
  @$pb.TagNumber(11)
  set lastPushTime($0.Timestamp value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasLastPushTime() => $_has(10);
  @$pb.TagNumber(11)
  void clearLastPushTime() => $_clearField(11);
  @$pb.TagNumber(11)
  $0.Timestamp ensureLastPushTime() => $_ensure(10);

  /// create_time is when the repository was created on GitHub.
  @$pb.TagNumber(12)
  $0.Timestamp get createTime => $_getN(11);
  @$pb.TagNumber(12)
  set createTime($0.Timestamp value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasCreateTime() => $_has(11);
  @$pb.TagNumber(12)
  void clearCreateTime() => $_clearField(12);
  @$pb.TagNumber(12)
  $0.Timestamp ensureCreateTime() => $_ensure(11);

  /// archived is true when the source repository is archived (read-only) on GitHub.
  @$pb.TagNumber(13)
  $core.bool get archived => $_getBF(12);
  @$pb.TagNumber(13)
  set archived($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(13)
  $core.bool hasArchived() => $_has(12);
  @$pb.TagNumber(13)
  void clearArchived() => $_clearField(13);

  /// languages is the repository's language breakdown by source bytes, ordered
  /// from the largest share to the smallest.
  @$pb.TagNumber(14)
  $pb.PbList<LanguageShare> get languages => $_getList(13);

  /// commit_activity is the weekly commit counts for roughly the last year,
  /// oldest week first, used to draw a contribution sparkline.
  @$pb.TagNumber(15)
  $pb.PbList<$core.int> get commitActivity => $_getList(14);
}

/// LanguageShare is one programming language's share of a repository, measured
/// in source bytes.
class LanguageShare extends $pb.GeneratedMessage {
  factory LanguageShare({
    $core.String? name,
    $fixnum.Int64? bytes,
  }) {
    final result = create();
    if (name != null) result.name = name;
    if (bytes != null) result.bytes = bytes;
    return result;
  }

  LanguageShare._();

  factory LanguageShare.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LanguageShare.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LanguageShare',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'realtime.me.site.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'name')
    ..aInt64(2, _omitFieldNames ? '' : 'bytes')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LanguageShare clone() => LanguageShare()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LanguageShare copyWith(void Function(LanguageShare) updates) =>
      super.copyWith((message) => updates(message as LanguageShare))
          as LanguageShare;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LanguageShare create() => LanguageShare._();
  @$core.override
  LanguageShare createEmptyInstance() => create();
  static $pb.PbList<LanguageShare> createRepeated() =>
      $pb.PbList<LanguageShare>();
  @$core.pragma('dart2js:noInline')
  static LanguageShare getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LanguageShare>(create);
  static LanguageShare? _defaultInstance;

  /// name is the language name, for example "TypeScript".
  @$pb.TagNumber(1)
  $core.String get name => $_getSZ(0);
  @$pb.TagNumber(1)
  set name($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasName() => $_has(0);
  @$pb.TagNumber(1)
  void clearName() => $_clearField(1);

  /// bytes is the amount of source attributed to this language, in bytes.
  @$pb.TagNumber(2)
  $fixnum.Int64 get bytes => $_getI64(1);
  @$pb.TagNumber(2)
  set bytes($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBytes() => $_has(1);
  @$pb.TagNumber(2)
  void clearBytes() => $_clearField(2);
}

/// ProjectsService serves the curated GitHub projects shown on the projects page.
/// The projects are collected once and stored as data; the gateway never calls
/// GitHub to answer this.
class ProjectsServiceApi {
  final $pb.RpcClient _client;

  ProjectsServiceApi(this._client);

  /// ListProjects returns the curated projects in display order.
  $async.Future<ListProjectsResponse> listProjects(
          $pb.ClientContext? ctx, ListProjectsRequest request) =>
      _client.invoke<ListProjectsResponse>(ctx, 'ProjectsService',
          'ListProjects', request, ListProjectsResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
