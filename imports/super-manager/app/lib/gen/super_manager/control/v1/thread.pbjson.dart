// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/thread.proto.

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

@$core.Deprecated('Use threadStateDescriptor instead')
const ThreadState$json = {
  '1': 'ThreadState',
  '2': [
    {'1': 'THREAD_STATE_UNSPECIFIED', '2': 0},
    {'1': 'THREAD_STATE_IDLE', '2': 1},
    {'1': 'THREAD_STATE_RUNNING', '2': 2},
    {'1': 'THREAD_STATE_INPUT_REQUIRED', '2': 3},
    {'1': 'THREAD_STATE_LOST', '2': 4},
  ],
};

/// Descriptor for `ThreadState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List threadStateDescriptor = $convert.base64Decode(
    'CgtUaHJlYWRTdGF0ZRIcChhUSFJFQURfU1RBVEVfVU5TUEVDSUZJRUQQABIVChFUSFJFQURfU1'
    'RBVEVfSURMRRABEhgKFFRIUkVBRF9TVEFURV9SVU5OSU5HEAISHwobVEhSRUFEX1NUQVRFX0lO'
    'UFVUX1JFUVVJUkVEEAMSFQoRVEhSRUFEX1NUQVRFX0xPU1QQBA==');

@$core.Deprecated('Use threadDescriptor instead')
const Thread$json = {
  '1': 'Thread',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
    {
      '1': 'workspace_uid',
      '3': 2,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'workspaceUid'
    },
    {'1': 'runtime_uid', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'runtimeUid'},
    {'1': 'display_name', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'displayName'},
    {
      '1': 'state',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.super_manager.control.v1.ThreadState',
      '8': {},
      '10': 'state'
    },
    {
      '1': 'create_time',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createTime'
    },
    {
      '1': 'update_time',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updateTime'
    },
  ],
};

/// Descriptor for `Thread`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List threadDescriptor = $convert.base64Decode(
    'CgZUaHJlYWQSGgoDdWlkGAEgASgJQgi6SAVyA7ABAVIDdWlkEi0KDXdvcmtzcGFjZV91aWQYAi'
    'ABKAlCCLpIBXIDsAEBUgx3b3Jrc3BhY2VVaWQSKQoLcnVudGltZV91aWQYAyABKAlCCLpIBXID'
    'sAEBUgpydW50aW1lVWlkEi0KDGRpc3BsYXlfbmFtZRgEIAEoCUIKukgHcgUQARiAAVILZGlzcG'
    'xheU5hbWUSRQoFc3RhdGUYBSABKA4yJS5zdXBlcl9tYW5hZ2VyLmNvbnRyb2wudjEuVGhyZWFk'
    'U3RhdGVCCLpIBYIBAhABUgVzdGF0ZRI7CgtjcmVhdGVfdGltZRgGIAEoCzIaLmdvb2dsZS5wcm'
    '90b2J1Zi5UaW1lc3RhbXBSCmNyZWF0ZVRpbWUSOwoLdXBkYXRlX3RpbWUYByABKAsyGi5nb29n'
    'bGUucHJvdG9idWYuVGltZXN0YW1wUgp1cGRhdGVUaW1l');

@$core.Deprecated('Use createThreadRequestDescriptor instead')
const CreateThreadRequest$json = {
  '1': 'CreateThreadRequest',
  '2': [
    {
      '1': 'workspace_uid',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'workspaceUid'
    },
    {'1': 'runtime_uid', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'runtimeUid'},
    {'1': 'display_name', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'displayName'},
  ],
};

/// Descriptor for `CreateThreadRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createThreadRequestDescriptor = $convert.base64Decode(
    'ChNDcmVhdGVUaHJlYWRSZXF1ZXN0Ei0KDXdvcmtzcGFjZV91aWQYASABKAlCCLpIBXIDsAEBUg'
    'x3b3Jrc3BhY2VVaWQSKQoLcnVudGltZV91aWQYAiABKAlCCLpIBXIDsAEBUgpydW50aW1lVWlk'
    'Ei0KDGRpc3BsYXlfbmFtZRgDIAEoCUIKukgHcgUQARiAAVILZGlzcGxheU5hbWU=');

@$core.Deprecated('Use createThreadResponseDescriptor instead')
const CreateThreadResponse$json = {
  '1': 'CreateThreadResponse',
  '2': [
    {
      '1': 'thread',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.control.v1.Thread',
      '8': {},
      '10': 'thread'
    },
  ],
};

/// Descriptor for `CreateThreadResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createThreadResponseDescriptor = $convert.base64Decode(
    'ChRDcmVhdGVUaHJlYWRSZXNwb25zZRJACgZ0aHJlYWQYASABKAsyIC5zdXBlcl9tYW5hZ2VyLm'
    'NvbnRyb2wudjEuVGhyZWFkQga6SAPIAQFSBnRocmVhZA==');

@$core.Deprecated('Use getThreadRequestDescriptor instead')
const GetThreadRequest$json = {
  '1': 'GetThreadRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
  ],
};

/// Descriptor for `GetThreadRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getThreadRequestDescriptor = $convert.base64Decode(
    'ChBHZXRUaHJlYWRSZXF1ZXN0EhoKA3VpZBgBIAEoCUIIukgFcgOwAQFSA3VpZA==');

@$core.Deprecated('Use getThreadResponseDescriptor instead')
const GetThreadResponse$json = {
  '1': 'GetThreadResponse',
  '2': [
    {
      '1': 'thread',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.control.v1.Thread',
      '8': {},
      '10': 'thread'
    },
  ],
};

/// Descriptor for `GetThreadResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getThreadResponseDescriptor = $convert.base64Decode(
    'ChFHZXRUaHJlYWRSZXNwb25zZRJACgZ0aHJlYWQYASABKAsyIC5zdXBlcl9tYW5hZ2VyLmNvbn'
    'Ryb2wudjEuVGhyZWFkQga6SAPIAQFSBnRocmVhZA==');

@$core.Deprecated('Use listThreadsRequestDescriptor instead')
const ListThreadsRequest$json = {
  '1': 'ListThreadsRequest',
  '2': [
    {
      '1': 'workspace_uid',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'workspaceUid'
    },
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '8': {}, '10': 'pageSize'},
    {'1': 'page_token', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'pageToken'},
  ],
};

/// Descriptor for `ListThreadsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listThreadsRequestDescriptor = $convert.base64Decode(
    'ChJMaXN0VGhyZWFkc1JlcXVlc3QSLQoNd29ya3NwYWNlX3VpZBgBIAEoCUIIukgFcgOwAQFSDH'
    'dvcmtzcGFjZVVpZBImCglwYWdlX3NpemUYAiABKAVCCbpIBhoEGGQoAFIIcGFnZVNpemUSJwoK'
    'cGFnZV90b2tlbhgDIAEoCUIIukgFcgMYgAJSCXBhZ2VUb2tlbg==');

@$core.Deprecated('Use listThreadsResponseDescriptor instead')
const ListThreadsResponse$json = {
  '1': 'ListThreadsResponse',
  '2': [
    {
      '1': 'threads',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.super_manager.control.v1.Thread',
      '10': 'threads'
    },
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListThreadsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listThreadsResponseDescriptor = $convert.base64Decode(
    'ChNMaXN0VGhyZWFkc1Jlc3BvbnNlEjoKB3RocmVhZHMYASADKAsyIC5zdXBlcl9tYW5hZ2VyLm'
    'NvbnRyb2wudjEuVGhyZWFkUgd0aHJlYWRzEiYKD25leHRfcGFnZV90b2tlbhgCIAEoCVINbmV4'
    'dFBhZ2VUb2tlbg==');

@$core.Deprecated('Use deleteThreadRequestDescriptor instead')
const DeleteThreadRequest$json = {
  '1': 'DeleteThreadRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
  ],
};

/// Descriptor for `DeleteThreadRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteThreadRequestDescriptor =
    $convert.base64Decode(
        'ChNEZWxldGVUaHJlYWRSZXF1ZXN0EhoKA3VpZBgBIAEoCUIIukgFcgOwAQFSA3VpZA==');

@$core.Deprecated('Use deleteThreadResponseDescriptor instead')
const DeleteThreadResponse$json = {
  '1': 'DeleteThreadResponse',
};

/// Descriptor for `DeleteThreadResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteThreadResponseDescriptor =
    $convert.base64Decode('ChREZWxldGVUaHJlYWRSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> ThreadServiceBase$json = {
  '1': 'ThreadService',
  '2': [
    {
      '1': 'CreateThread',
      '2': '.super_manager.control.v1.CreateThreadRequest',
      '3': '.super_manager.control.v1.CreateThreadResponse'
    },
    {
      '1': 'GetThread',
      '2': '.super_manager.control.v1.GetThreadRequest',
      '3': '.super_manager.control.v1.GetThreadResponse'
    },
    {
      '1': 'ListThreads',
      '2': '.super_manager.control.v1.ListThreadsRequest',
      '3': '.super_manager.control.v1.ListThreadsResponse'
    },
    {
      '1': 'DeleteThread',
      '2': '.super_manager.control.v1.DeleteThreadRequest',
      '3': '.super_manager.control.v1.DeleteThreadResponse'
    },
  ],
};

@$core.Deprecated('Use threadServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ThreadServiceBase$messageJson = {
  '.super_manager.control.v1.CreateThreadRequest': CreateThreadRequest$json,
  '.super_manager.control.v1.CreateThreadResponse': CreateThreadResponse$json,
  '.super_manager.control.v1.Thread': Thread$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.super_manager.control.v1.GetThreadRequest': GetThreadRequest$json,
  '.super_manager.control.v1.GetThreadResponse': GetThreadResponse$json,
  '.super_manager.control.v1.ListThreadsRequest': ListThreadsRequest$json,
  '.super_manager.control.v1.ListThreadsResponse': ListThreadsResponse$json,
  '.super_manager.control.v1.DeleteThreadRequest': DeleteThreadRequest$json,
  '.super_manager.control.v1.DeleteThreadResponse': DeleteThreadResponse$json,
};

/// Descriptor for `ThreadService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List threadServiceDescriptor = $convert.base64Decode(
    'Cg1UaHJlYWRTZXJ2aWNlEm0KDENyZWF0ZVRocmVhZBItLnN1cGVyX21hbmFnZXIuY29udHJvbC'
    '52MS5DcmVhdGVUaHJlYWRSZXF1ZXN0Gi4uc3VwZXJfbWFuYWdlci5jb250cm9sLnYxLkNyZWF0'
    'ZVRocmVhZFJlc3BvbnNlEmQKCUdldFRocmVhZBIqLnN1cGVyX21hbmFnZXIuY29udHJvbC52MS'
    '5HZXRUaHJlYWRSZXF1ZXN0Gisuc3VwZXJfbWFuYWdlci5jb250cm9sLnYxLkdldFRocmVhZFJl'
    'c3BvbnNlEmoKC0xpc3RUaHJlYWRzEiwuc3VwZXJfbWFuYWdlci5jb250cm9sLnYxLkxpc3RUaH'
    'JlYWRzUmVxdWVzdBotLnN1cGVyX21hbmFnZXIuY29udHJvbC52MS5MaXN0VGhyZWFkc1Jlc3Bv'
    'bnNlEm0KDERlbGV0ZVRocmVhZBItLnN1cGVyX21hbmFnZXIuY29udHJvbC52MS5EZWxldGVUaH'
    'JlYWRSZXF1ZXN0Gi4uc3VwZXJfbWFuYWdlci5jb250cm9sLnYxLkRlbGV0ZVRocmVhZFJlc3Bv'
    'bnNl');
