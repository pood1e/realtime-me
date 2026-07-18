// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/status.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// GitHubSyncState is the state of GitHub status synchronization.
class GithubSyncState extends $pb.ProtobufEnum {
  /// State is not known.
  static const GithubSyncState GITHUB_SYNC_STATE_UNSPECIFIED =
      GithubSyncState._(
          0, _omitEnumNames ? '' : 'GITHUB_SYNC_STATE_UNSPECIFIED');

  /// Synchronization is disabled because no token is configured.
  static const GithubSyncState GITHUB_SYNC_STATE_DISABLED =
      GithubSyncState._(1, _omitEnumNames ? '' : 'GITHUB_SYNC_STATE_DISABLED');

  /// A synchronization attempt is in progress.
  static const GithubSyncState GITHUB_SYNC_STATE_PENDING =
      GithubSyncState._(2, _omitEnumNames ? '' : 'GITHUB_SYNC_STATE_PENDING');

  /// The last synchronization succeeded.
  static const GithubSyncState GITHUB_SYNC_STATE_OK =
      GithubSyncState._(3, _omitEnumNames ? '' : 'GITHUB_SYNC_STATE_OK');

  /// The last synchronization failed.
  static const GithubSyncState GITHUB_SYNC_STATE_ERROR =
      GithubSyncState._(4, _omitEnumNames ? '' : 'GITHUB_SYNC_STATE_ERROR');

  static const $core.List<GithubSyncState> values = <GithubSyncState>[
    GITHUB_SYNC_STATE_UNSPECIFIED,
    GITHUB_SYNC_STATE_DISABLED,
    GITHUB_SYNC_STATE_PENDING,
    GITHUB_SYNC_STATE_OK,
    GITHUB_SYNC_STATE_ERROR,
  ];

  static final $core.List<GithubSyncState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static GithubSyncState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const GithubSyncState._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
