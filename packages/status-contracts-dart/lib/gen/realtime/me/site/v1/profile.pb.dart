// This is a generated file - do not edit.
//
// Generated from realtime/me/site/v1/profile.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// GetProfileRequest is the request for the owner's identity.
class GetProfileRequest extends $pb.GeneratedMessage {
  factory GetProfileRequest() => create();

  GetProfileRequest._();

  factory GetProfileRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetProfileRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetProfileRequest',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'realtime.me.site.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetProfileRequest clone() => GetProfileRequest()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetProfileRequest copyWith(void Function(GetProfileRequest) updates) =>
      super.copyWith((message) => updates(message as GetProfileRequest))
          as GetProfileRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetProfileRequest create() => GetProfileRequest._();
  @$core.override
  GetProfileRequest createEmptyInstance() => create();
  static $pb.PbList<GetProfileRequest> createRepeated() =>
      $pb.PbList<GetProfileRequest>();
  @$core.pragma('dart2js:noInline')
  static GetProfileRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetProfileRequest>(create);
  static GetProfileRequest? _defaultInstance;
}

/// GetProfileResponse carries the owner's identity.
class GetProfileResponse extends $pb.GeneratedMessage {
  factory GetProfileResponse({
    Profile? profile,
  }) {
    final result = create();
    if (profile != null) result.profile = profile;
    return result;
  }

  GetProfileResponse._();

  factory GetProfileResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetProfileResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetProfileResponse',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'realtime.me.site.v1'),
      createEmptyInstance: create)
    ..aOM<Profile>(1, _omitFieldNames ? '' : 'profile',
        subBuilder: Profile.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetProfileResponse clone() => GetProfileResponse()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetProfileResponse copyWith(void Function(GetProfileResponse) updates) =>
      super.copyWith((message) => updates(message as GetProfileResponse))
          as GetProfileResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetProfileResponse create() => GetProfileResponse._();
  @$core.override
  GetProfileResponse createEmptyInstance() => create();
  static $pb.PbList<GetProfileResponse> createRepeated() =>
      $pb.PbList<GetProfileResponse>();
  @$core.pragma('dart2js:noInline')
  static GetProfileResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetProfileResponse>(create);
  static GetProfileResponse? _defaultInstance;

  /// profile is the site owner's public identity.
  @$pb.TagNumber(1)
  Profile get profile => $_getN(0);
  @$pb.TagNumber(1)
  set profile(Profile value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasProfile() => $_has(0);
  @$pb.TagNumber(1)
  void clearProfile() => $_clearField(1);
  @$pb.TagNumber(1)
  Profile ensureProfile() => $_ensure(0);
}

/// Profile is the site owner's public identity: who the page belongs to, and how
/// to reach them. It says nothing about what the owner has built — that is a
/// project, and ProjectsService serves those.
class Profile extends $pb.GeneratedMessage {
  factory Profile({
    $core.String? displayName,
    $core.String? avatarUrl,
    $core.String? githubLogin,
    $core.Iterable<ProfileLink>? links,
  }) {
    final result = create();
    if (displayName != null) result.displayName = displayName;
    if (avatarUrl != null) result.avatarUrl = avatarUrl;
    if (githubLogin != null) result.githubLogin = githubLogin;
    if (links != null) result.links.addAll(links);
    return result;
  }

  Profile._();

  factory Profile.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Profile.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Profile',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'realtime.me.site.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'displayName')
    ..aOS(4, _omitFieldNames ? '' : 'avatarUrl')
    ..aOS(6, _omitFieldNames ? '' : 'githubLogin')
    ..pc<ProfileLink>(7, _omitFieldNames ? '' : 'links', $pb.PbFieldType.PM,
        subBuilder: ProfileLink.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Profile clone() => Profile()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Profile copyWith(void Function(Profile) updates) =>
      super.copyWith((message) => updates(message as Profile)) as Profile;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Profile create() => Profile._();
  @$core.override
  Profile createEmptyInstance() => create();
  static $pb.PbList<Profile> createRepeated() => $pb.PbList<Profile>();
  @$core.pragma('dart2js:noInline')
  static Profile getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Profile>(create);
  static Profile? _defaultInstance;

  /// display_name is the human-readable name shown in the topbar.
  @$pb.TagNumber(1)
  $core.String get displayName => $_getSZ(0);
  @$pb.TagNumber(1)
  set displayName($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDisplayName() => $_has(0);
  @$pb.TagNumber(1)
  void clearDisplayName() => $_clearField(1);

  /// avatar_url is the URL of the profile picture.
  @$pb.TagNumber(4)
  $core.String get avatarUrl => $_getSZ(1);
  @$pb.TagNumber(4)
  set avatarUrl($core.String value) => $_setString(1, value);
  @$pb.TagNumber(4)
  $core.bool hasAvatarUrl() => $_has(1);
  @$pb.TagNumber(4)
  void clearAvatarUrl() => $_clearField(4);

  /// github_login is the owner's GitHub username.
  @$pb.TagNumber(6)
  $core.String get githubLogin => $_getSZ(2);
  @$pb.TagNumber(6)
  set githubLogin($core.String value) => $_setString(2, value);
  @$pb.TagNumber(6)
  $core.bool hasGithubLogin() => $_has(2);
  @$pb.TagNumber(6)
  void clearGithubLogin() => $_clearField(6);

  /// links are the owner's public contact links.
  @$pb.TagNumber(7)
  $pb.PbList<ProfileLink> get links => $_getList(3);
}

/// ProfileLink is one public contact link, such as an email address or a social
/// handle.
class ProfileLink extends $pb.GeneratedMessage {
  factory ProfileLink({
    $core.String? label,
    $core.String? uri,
    $core.String? platform,
  }) {
    final result = create();
    if (label != null) result.label = label;
    if (uri != null) result.uri = uri;
    if (platform != null) result.platform = platform;
    return result;
  }

  ProfileLink._();

  factory ProfileLink.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ProfileLink.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ProfileLink',
      package:
          const $pb.PackageName(_omitMessageNames ? '' : 'realtime.me.site.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'label')
    ..aOS(2, _omitFieldNames ? '' : 'uri')
    ..aOS(3, _omitFieldNames ? '' : 'platform')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileLink clone() => ProfileLink()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ProfileLink copyWith(void Function(ProfileLink) updates) =>
      super.copyWith((message) => updates(message as ProfileLink))
          as ProfileLink;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ProfileLink create() => ProfileLink._();
  @$core.override
  ProfileLink createEmptyInstance() => create();
  static $pb.PbList<ProfileLink> createRepeated() => $pb.PbList<ProfileLink>();
  @$core.pragma('dart2js:noInline')
  static ProfileLink getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ProfileLink>(create);
  static ProfileLink? _defaultInstance;

  /// label is the human-readable link text, used as the icon's accessible name.
  @$pb.TagNumber(1)
  $core.String get label => $_getSZ(0);
  @$pb.TagNumber(1)
  set label($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasLabel() => $_has(0);
  @$pb.TagNumber(1)
  void clearLabel() => $_clearField(1);

  /// uri is the link target.
  @$pb.TagNumber(2)
  $core.String get uri => $_getSZ(1);
  @$pb.TagNumber(2)
  set uri($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasUri() => $_has(1);
  @$pb.TagNumber(2)
  void clearUri() => $_clearField(2);

  /// platform is a lowercase platform key used to pick an icon, such as
  /// "github", "email", or "website". Unknown values fall back to a generic icon.
  @$pb.TagNumber(3)
  $core.String get platform => $_getSZ(2);
  @$pb.TagNumber(3)
  set platform($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPlatform() => $_has(2);
  @$pb.TagNumber(3)
  void clearPlatform() => $_clearField(3);
}

/// ProfileService serves the site owner's identity. The page carries it on every
/// screen — the name and avatar in the topbar, the contact links beside them — so
/// it is fetched on its own and owes nothing to whatever a given page is showing.
class ProfileServiceApi {
  final $pb.RpcClient _client;

  ProfileServiceApi(this._client);

  /// GetProfile returns the site owner's public identity.
  $async.Future<GetProfileResponse> getProfile(
          $pb.ClientContext? ctx, GetProfileRequest request) =>
      _client.invoke<GetProfileResponse>(
          ctx, 'ProfileService', 'GetProfile', request, GetProfileResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
