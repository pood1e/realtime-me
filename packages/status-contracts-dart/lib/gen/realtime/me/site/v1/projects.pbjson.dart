// This is a generated file - do not edit.
//
// Generated from realtime/me/site/v1/projects.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../../../../google/protobuf/timestamp.pbjson.dart' as $0;

@$core.Deprecated('Use projectVisibilityDescriptor instead')
const ProjectVisibility$json = {
  '1': 'ProjectVisibility',
  '2': [
    {'1': 'PROJECT_VISIBILITY_UNSPECIFIED', '2': 0},
    {'1': 'PROJECT_VISIBILITY_PUBLIC', '2': 1},
    {'1': 'PROJECT_VISIBILITY_PRIVATE', '2': 2},
  ],
};

/// Descriptor for `ProjectVisibility`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List projectVisibilityDescriptor = $convert.base64Decode(
    'ChFQcm9qZWN0VmlzaWJpbGl0eRIiCh5QUk9KRUNUX1ZJU0lCSUxJVFlfVU5TUEVDSUZJRUQQAB'
    'IdChlQUk9KRUNUX1ZJU0lCSUxJVFlfUFVCTElDEAESHgoaUFJPSkVDVF9WSVNJQklMSVRZX1BS'
    'SVZBVEUQAg==');

@$core.Deprecated('Use listProjectsRequestDescriptor instead')
const ListProjectsRequest$json = {
  '1': 'ListProjectsRequest',
};

/// Descriptor for `ListProjectsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listProjectsRequestDescriptor =
    $convert.base64Decode('ChNMaXN0UHJvamVjdHNSZXF1ZXN0');

@$core.Deprecated('Use listProjectsResponseDescriptor instead')
const ListProjectsResponse$json = {
  '1': 'ListProjectsResponse',
  '2': [
    {
      '1': 'projects',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.site.v1.Project',
      '10': 'projects'
    },
  ],
};

/// Descriptor for `ListProjectsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listProjectsResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0UHJvamVjdHNSZXNwb25zZRI4Cghwcm9qZWN0cxgBIAMoCzIcLnJlYWx0aW1lLm1lLn'
    'NpdGUudjEuUHJvamVjdFIIcHJvamVjdHM=');

@$core.Deprecated('Use projectDescriptor instead')
const Project$json = {
  '1': 'Project',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '10': 'uid'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'description', '3': 3, '4': 1, '5': 9, '10': 'description'},
    {'1': 'summary', '3': 4, '4': 1, '5': 9, '10': 'summary'},
    {
      '1': 'visibility',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.site.v1.ProjectVisibility',
      '10': 'visibility'
    },
    {'1': 'primary_language', '3': 6, '4': 1, '5': 9, '10': 'primaryLanguage'},
    {'1': 'topics', '3': 7, '4': 3, '5': 9, '10': 'topics'},
    {'1': 'star_count', '3': 8, '4': 1, '5': 5, '10': 'starCount'},
    {'1': 'repository_url', '3': 9, '4': 1, '5': 9, '10': 'repositoryUrl'},
    {'1': 'homepage_url', '3': 10, '4': 1, '5': 9, '10': 'homepageUrl'},
    {
      '1': 'last_push_time',
      '3': 11,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastPushTime'
    },
    {
      '1': 'create_time',
      '3': 12,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createTime'
    },
    {'1': 'archived', '3': 13, '4': 1, '5': 8, '10': 'archived'},
    {
      '1': 'languages',
      '3': 14,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.site.v1.LanguageShare',
      '10': 'languages'
    },
    {'1': 'commit_activity', '3': 15, '4': 3, '5': 5, '10': 'commitActivity'},
  ],
};

/// Descriptor for `Project`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List projectDescriptor = $convert.base64Decode(
    'CgdQcm9qZWN0EhAKA3VpZBgBIAEoCVIDdWlkEiEKDGRpc3BsYXlfbmFtZRgCIAEoCVILZGlzcG'
    'xheU5hbWUSIAoLZGVzY3JpcHRpb24YAyABKAlSC2Rlc2NyaXB0aW9uEhgKB3N1bW1hcnkYBCAB'
    'KAlSB3N1bW1hcnkSRgoKdmlzaWJpbGl0eRgFIAEoDjImLnJlYWx0aW1lLm1lLnNpdGUudjEuUH'
    'JvamVjdFZpc2liaWxpdHlSCnZpc2liaWxpdHkSKQoQcHJpbWFyeV9sYW5ndWFnZRgGIAEoCVIP'
    'cHJpbWFyeUxhbmd1YWdlEhYKBnRvcGljcxgHIAMoCVIGdG9waWNzEh0KCnN0YXJfY291bnQYCC'
    'ABKAVSCXN0YXJDb3VudBIlCg5yZXBvc2l0b3J5X3VybBgJIAEoCVINcmVwb3NpdG9yeVVybBIh'
    'Cgxob21lcGFnZV91cmwYCiABKAlSC2hvbWVwYWdlVXJsEkAKDmxhc3RfcHVzaF90aW1lGAsgAS'
    'gLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIMbGFzdFB1c2hUaW1lEjsKC2NyZWF0ZV90'
    'aW1lGAwgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIKY3JlYXRlVGltZRIaCghhcm'
    'NoaXZlZBgNIAEoCFIIYXJjaGl2ZWQSQAoJbGFuZ3VhZ2VzGA4gAygLMiIucmVhbHRpbWUubWUu'
    'c2l0ZS52MS5MYW5ndWFnZVNoYXJlUglsYW5ndWFnZXMSJwoPY29tbWl0X2FjdGl2aXR5GA8gAy'
    'gFUg5jb21taXRBY3Rpdml0eQ==');

@$core.Deprecated('Use languageShareDescriptor instead')
const LanguageShare$json = {
  '1': 'LanguageShare',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'bytes', '3': 2, '4': 1, '5': 3, '10': 'bytes'},
  ],
};

/// Descriptor for `LanguageShare`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List languageShareDescriptor = $convert.base64Decode(
    'Cg1MYW5ndWFnZVNoYXJlEhIKBG5hbWUYASABKAlSBG5hbWUSFAoFYnl0ZXMYAiABKANSBWJ5dG'
    'Vz');

const $core.Map<$core.String, $core.dynamic> ProjectsServiceBase$json = {
  '1': 'ProjectsService',
  '2': [
    {
      '1': 'ListProjects',
      '2': '.realtime.me.site.v1.ListProjectsRequest',
      '3': '.realtime.me.site.v1.ListProjectsResponse'
    },
  ],
};

@$core.Deprecated('Use projectsServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ProjectsServiceBase$messageJson = {
  '.realtime.me.site.v1.ListProjectsRequest': ListProjectsRequest$json,
  '.realtime.me.site.v1.ListProjectsResponse': ListProjectsResponse$json,
  '.realtime.me.site.v1.Project': Project$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.realtime.me.site.v1.LanguageShare': LanguageShare$json,
};

/// Descriptor for `ProjectsService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List projectsServiceDescriptor = $convert.base64Decode(
    'Cg9Qcm9qZWN0c1NlcnZpY2USYwoMTGlzdFByb2plY3RzEigucmVhbHRpbWUubWUuc2l0ZS52MS'
    '5MaXN0UHJvamVjdHNSZXF1ZXN0GikucmVhbHRpbWUubWUuc2l0ZS52MS5MaXN0UHJvamVjdHNS'
    'ZXNwb25zZQ==');
