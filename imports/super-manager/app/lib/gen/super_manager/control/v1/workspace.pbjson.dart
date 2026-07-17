// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/workspace.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports
// ignore_for_file: unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import 'package:protobuf/well_known_types/google/protobuf/timestamp.pbjson.dart'
    as $0;

@$core.Deprecated('Use workspaceDescriptor instead')
const Workspace$json = {
  '1': 'Workspace',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'displayName'},
    {'1': 'path', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'path'},
    {
      '1': 'active_execution_uid',
      '3': 4,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'activeExecutionUid'
    },
    {
      '1': 'create_time',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createTime'
    },
  ],
};

/// Descriptor for `Workspace`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List workspaceDescriptor = $convert.base64Decode(
    'CglXb3Jrc3BhY2USGgoDdWlkGAEgASgJQgi6SAVyA7ABAVIDdWlkEi0KDGRpc3BsYXlfbmFtZR'
    'gCIAEoCUIKukgHcgUQARiAAVILZGlzcGxheU5hbWUSIgoEcGF0aBgDIAEoCUIOukgLcgkQARiA'
    'IDICXi9SBHBhdGgSPQoUYWN0aXZlX2V4ZWN1dGlvbl91aWQYBCABKAlCC7pICNgBAXIDsAEBUh'
    'JhY3RpdmVFeGVjdXRpb25VaWQSOwoLY3JlYXRlX3RpbWUYBSABKAsyGi5nb29nbGUucHJvdG9i'
    'dWYuVGltZXN0YW1wUgpjcmVhdGVUaW1l');

@$core.Deprecated('Use createWorkspaceRequestDescriptor instead')
const CreateWorkspaceRequest$json = {
  '1': 'CreateWorkspaceRequest',
  '2': [
    {'1': 'display_name', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'displayName'},
    {'1': 'path', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'path'},
  ],
};

/// Descriptor for `CreateWorkspaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWorkspaceRequestDescriptor = $convert.base64Decode(
    'ChZDcmVhdGVXb3Jrc3BhY2VSZXF1ZXN0Ei0KDGRpc3BsYXlfbmFtZRgBIAEoCUIKukgHcgUQAR'
    'iAAVILZGlzcGxheU5hbWUSIgoEcGF0aBgCIAEoCUIOukgLcgkQARiAIDICXi9SBHBhdGg=');

@$core.Deprecated('Use createWorkspaceResponseDescriptor instead')
const CreateWorkspaceResponse$json = {
  '1': 'CreateWorkspaceResponse',
  '2': [
    {
      '1': 'workspace',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.control.v1.Workspace',
      '8': {},
      '10': 'workspace'
    },
  ],
};

/// Descriptor for `CreateWorkspaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWorkspaceResponseDescriptor =
    $convert.base64Decode(
        'ChdDcmVhdGVXb3Jrc3BhY2VSZXNwb25zZRJJCgl3b3Jrc3BhY2UYASABKAsyIy5zdXBlcl9tYW'
        '5hZ2VyLmNvbnRyb2wudjEuV29ya3NwYWNlQga6SAPIAQFSCXdvcmtzcGFjZQ==');

@$core.Deprecated('Use getWorkspaceRequestDescriptor instead')
const GetWorkspaceRequest$json = {
  '1': 'GetWorkspaceRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
  ],
};

/// Descriptor for `GetWorkspaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWorkspaceRequestDescriptor =
    $convert.base64Decode(
        'ChNHZXRXb3Jrc3BhY2VSZXF1ZXN0EhoKA3VpZBgBIAEoCUIIukgFcgOwAQFSA3VpZA==');

@$core.Deprecated('Use getWorkspaceResponseDescriptor instead')
const GetWorkspaceResponse$json = {
  '1': 'GetWorkspaceResponse',
  '2': [
    {
      '1': 'workspace',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.control.v1.Workspace',
      '8': {},
      '10': 'workspace'
    },
  ],
};

/// Descriptor for `GetWorkspaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getWorkspaceResponseDescriptor = $convert.base64Decode(
    'ChRHZXRXb3Jrc3BhY2VSZXNwb25zZRJJCgl3b3Jrc3BhY2UYASABKAsyIy5zdXBlcl9tYW5hZ2'
    'VyLmNvbnRyb2wudjEuV29ya3NwYWNlQga6SAPIAQFSCXdvcmtzcGFjZQ==');

@$core.Deprecated('Use listWorkspacesRequestDescriptor instead')
const ListWorkspacesRequest$json = {
  '1': 'ListWorkspacesRequest',
  '2': [
    {'1': 'page_size', '3': 1, '4': 1, '5': 5, '8': {}, '10': 'pageSize'},
    {'1': 'page_token', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'pageToken'},
  ],
};

/// Descriptor for `ListWorkspacesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWorkspacesRequestDescriptor = $convert.base64Decode(
    'ChVMaXN0V29ya3NwYWNlc1JlcXVlc3QSJgoJcGFnZV9zaXplGAEgASgFQgm6SAYaBBhkKABSCH'
    'BhZ2VTaXplEicKCnBhZ2VfdG9rZW4YAiABKAlCCLpIBXIDGIACUglwYWdlVG9rZW4=');

@$core.Deprecated('Use listWorkspacesResponseDescriptor instead')
const ListWorkspacesResponse$json = {
  '1': 'ListWorkspacesResponse',
  '2': [
    {
      '1': 'workspaces',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.super_manager.control.v1.Workspace',
      '10': 'workspaces'
    },
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListWorkspacesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listWorkspacesResponseDescriptor = $convert.base64Decode(
    'ChZMaXN0V29ya3NwYWNlc1Jlc3BvbnNlEkMKCndvcmtzcGFjZXMYASADKAsyIy5zdXBlcl9tYW'
    '5hZ2VyLmNvbnRyb2wudjEuV29ya3NwYWNlUgp3b3Jrc3BhY2VzEiYKD25leHRfcGFnZV90b2tl'
    'bhgCIAEoCVINbmV4dFBhZ2VUb2tlbg==');

@$core.Deprecated('Use deleteWorkspaceRequestDescriptor instead')
const DeleteWorkspaceRequest$json = {
  '1': 'DeleteWorkspaceRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
  ],
};

/// Descriptor for `DeleteWorkspaceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWorkspaceRequestDescriptor =
    $convert.base64Decode(
        'ChZEZWxldGVXb3Jrc3BhY2VSZXF1ZXN0EhoKA3VpZBgBIAEoCUIIukgFcgOwAQFSA3VpZA==');

@$core.Deprecated('Use deleteWorkspaceResponseDescriptor instead')
const DeleteWorkspaceResponse$json = {
  '1': 'DeleteWorkspaceResponse',
};

/// Descriptor for `DeleteWorkspaceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteWorkspaceResponseDescriptor =
    $convert.base64Decode('ChdEZWxldGVXb3Jrc3BhY2VSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> WorkspaceServiceBase$json = {
  '1': 'WorkspaceService',
  '2': [
    {
      '1': 'CreateWorkspace',
      '2': '.super_manager.control.v1.CreateWorkspaceRequest',
      '3': '.super_manager.control.v1.CreateWorkspaceResponse'
    },
    {
      '1': 'GetWorkspace',
      '2': '.super_manager.control.v1.GetWorkspaceRequest',
      '3': '.super_manager.control.v1.GetWorkspaceResponse'
    },
    {
      '1': 'ListWorkspaces',
      '2': '.super_manager.control.v1.ListWorkspacesRequest',
      '3': '.super_manager.control.v1.ListWorkspacesResponse'
    },
    {
      '1': 'DeleteWorkspace',
      '2': '.super_manager.control.v1.DeleteWorkspaceRequest',
      '3': '.super_manager.control.v1.DeleteWorkspaceResponse'
    },
  ],
};

@$core.Deprecated('Use workspaceServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    WorkspaceServiceBase$messageJson = {
  '.super_manager.control.v1.CreateWorkspaceRequest':
      CreateWorkspaceRequest$json,
  '.super_manager.control.v1.CreateWorkspaceResponse':
      CreateWorkspaceResponse$json,
  '.super_manager.control.v1.Workspace': Workspace$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.super_manager.control.v1.GetWorkspaceRequest': GetWorkspaceRequest$json,
  '.super_manager.control.v1.GetWorkspaceResponse': GetWorkspaceResponse$json,
  '.super_manager.control.v1.ListWorkspacesRequest': ListWorkspacesRequest$json,
  '.super_manager.control.v1.ListWorkspacesResponse':
      ListWorkspacesResponse$json,
  '.super_manager.control.v1.DeleteWorkspaceRequest':
      DeleteWorkspaceRequest$json,
  '.super_manager.control.v1.DeleteWorkspaceResponse':
      DeleteWorkspaceResponse$json,
};

/// Descriptor for `WorkspaceService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List workspaceServiceDescriptor = $convert.base64Decode(
    'ChBXb3Jrc3BhY2VTZXJ2aWNlEnYKD0NyZWF0ZVdvcmtzcGFjZRIwLnN1cGVyX21hbmFnZXIuY2'
    '9udHJvbC52MS5DcmVhdGVXb3Jrc3BhY2VSZXF1ZXN0GjEuc3VwZXJfbWFuYWdlci5jb250cm9s'
    'LnYxLkNyZWF0ZVdvcmtzcGFjZVJlc3BvbnNlEm0KDEdldFdvcmtzcGFjZRItLnN1cGVyX21hbm'
    'FnZXIuY29udHJvbC52MS5HZXRXb3Jrc3BhY2VSZXF1ZXN0Gi4uc3VwZXJfbWFuYWdlci5jb250'
    'cm9sLnYxLkdldFdvcmtzcGFjZVJlc3BvbnNlEnMKDkxpc3RXb3Jrc3BhY2VzEi8uc3VwZXJfbW'
    'FuYWdlci5jb250cm9sLnYxLkxpc3RXb3Jrc3BhY2VzUmVxdWVzdBowLnN1cGVyX21hbmFnZXIu'
    'Y29udHJvbC52MS5MaXN0V29ya3NwYWNlc1Jlc3BvbnNlEnYKD0RlbGV0ZVdvcmtzcGFjZRIwLn'
    'N1cGVyX21hbmFnZXIuY29udHJvbC52MS5EZWxldGVXb3Jrc3BhY2VSZXF1ZXN0GjEuc3VwZXJf'
    'bWFuYWdlci5jb250cm9sLnYxLkRlbGV0ZVdvcmtzcGFjZVJlc3BvbnNl');
