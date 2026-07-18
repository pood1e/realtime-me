// This is a generated file - do not edit.
//
// Generated from realtime/me/site/v1/projects.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// ProjectVisibility describes the source repository's visibility on GitHub.
class ProjectVisibility extends $pb.ProtobufEnum {
  /// Visibility is not known.
  static const ProjectVisibility PROJECT_VISIBILITY_UNSPECIFIED =
      ProjectVisibility._(
          0, _omitEnumNames ? '' : 'PROJECT_VISIBILITY_UNSPECIFIED');

  /// The repository is publicly visible on GitHub.
  static const ProjectVisibility PROJECT_VISIBILITY_PUBLIC =
      ProjectVisibility._(1, _omitEnumNames ? '' : 'PROJECT_VISIBILITY_PUBLIC');

  /// The repository is private on GitHub.
  static const ProjectVisibility PROJECT_VISIBILITY_PRIVATE =
      ProjectVisibility._(
          2, _omitEnumNames ? '' : 'PROJECT_VISIBILITY_PRIVATE');

  static const $core.List<ProjectVisibility> values = <ProjectVisibility>[
    PROJECT_VISIBILITY_UNSPECIFIED,
    PROJECT_VISIBILITY_PUBLIC,
    PROJECT_VISIBILITY_PRIVATE,
  ];

  static final $core.List<ProjectVisibility?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static ProjectVisibility? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const ProjectVisibility._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
