// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/execution.proto.

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

@$core.Deprecated('Use executionStateDescriptor instead')
const ExecutionState$json = {
  '1': 'ExecutionState',
  '2': [
    {'1': 'EXECUTION_STATE_UNSPECIFIED', '2': 0},
    {'1': 'EXECUTION_STATE_RUNNING', '2': 1},
    {'1': 'EXECUTION_STATE_INPUT_REQUIRED', '2': 2},
    {'1': 'EXECUTION_STATE_SUCCEEDED', '2': 3},
    {'1': 'EXECUTION_STATE_FAILED', '2': 4},
    {'1': 'EXECUTION_STATE_CANCELED', '2': 5},
    {'1': 'EXECUTION_STATE_LOST', '2': 6},
  ],
};

/// Descriptor for `ExecutionState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List executionStateDescriptor = $convert.base64Decode(
    'Cg5FeGVjdXRpb25TdGF0ZRIfChtFWEVDVVRJT05fU1RBVEVfVU5TUEVDSUZJRUQQABIbChdFWE'
    'VDVVRJT05fU1RBVEVfUlVOTklORxABEiIKHkVYRUNVVElPTl9TVEFURV9JTlBVVF9SRVFVSVJF'
    'RBACEh0KGUVYRUNVVElPTl9TVEFURV9TVUNDRUVERUQQAxIaChZFWEVDVVRJT05fU1RBVEVfRk'
    'FJTEVEEAQSHAoYRVhFQ1VUSU9OX1NUQVRFX0NBTkNFTEVEEAUSGAoURVhFQ1VUSU9OX1NUQVRF'
    'X0xPU1QQBg==');

@$core.Deprecated('Use executionDescriptor instead')
const Execution$json = {
  '1': 'Execution',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
    {'1': 'thread_uid', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'threadUid'},
    {'1': 'run_id', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'runId'},
    {
      '1': 'state',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.super_manager.control.v1.ExecutionState',
      '8': {},
      '10': 'state'
    },
    {
      '1': 'start_time',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'startTime'
    },
    {
      '1': 'end_time',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'endTime'
    },
  ],
};

/// Descriptor for `Execution`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List executionDescriptor = $convert.base64Decode(
    'CglFeGVjdXRpb24SGgoDdWlkGAEgASgJQgi6SAVyA7ABAVIDdWlkEicKCnRocmVhZF91aWQYAi'
    'ABKAlCCLpIBXIDsAEBUgl0aHJlYWRVaWQSIQoGcnVuX2lkGAMgASgJQgq6SAdyBRABGIABUgVy'
    'dW5JZBJICgVzdGF0ZRgEIAEoDjIoLnN1cGVyX21hbmFnZXIuY29udHJvbC52MS5FeGVjdXRpb2'
    '5TdGF0ZUIIukgFggECEAFSBXN0YXRlEjkKCnN0YXJ0X3RpbWUYBSABKAsyGi5nb29nbGUucHJv'
    'dG9idWYuVGltZXN0YW1wUglzdGFydFRpbWUSNQoIZW5kX3RpbWUYBiABKAsyGi5nb29nbGUucH'
    'JvdG9idWYuVGltZXN0YW1wUgdlbmRUaW1l');

@$core.Deprecated('Use getExecutionRequestDescriptor instead')
const GetExecutionRequest$json = {
  '1': 'GetExecutionRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
  ],
};

/// Descriptor for `GetExecutionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getExecutionRequestDescriptor =
    $convert.base64Decode(
        'ChNHZXRFeGVjdXRpb25SZXF1ZXN0EhoKA3VpZBgBIAEoCUIIukgFcgOwAQFSA3VpZA==');

@$core.Deprecated('Use getExecutionResponseDescriptor instead')
const GetExecutionResponse$json = {
  '1': 'GetExecutionResponse',
  '2': [
    {
      '1': 'execution',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.control.v1.Execution',
      '8': {},
      '10': 'execution'
    },
  ],
};

/// Descriptor for `GetExecutionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getExecutionResponseDescriptor = $convert.base64Decode(
    'ChRHZXRFeGVjdXRpb25SZXNwb25zZRJJCglleGVjdXRpb24YASABKAsyIy5zdXBlcl9tYW5hZ2'
    'VyLmNvbnRyb2wudjEuRXhlY3V0aW9uQga6SAPIAQFSCWV4ZWN1dGlvbg==');

@$core.Deprecated('Use listExecutionsRequestDescriptor instead')
const ListExecutionsRequest$json = {
  '1': 'ListExecutionsRequest',
  '2': [
    {'1': 'thread_uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'threadUid'},
    {'1': 'page_size', '3': 2, '4': 1, '5': 5, '8': {}, '10': 'pageSize'},
    {'1': 'page_token', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'pageToken'},
  ],
};

/// Descriptor for `ListExecutionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listExecutionsRequestDescriptor = $convert.base64Decode(
    'ChVMaXN0RXhlY3V0aW9uc1JlcXVlc3QSJwoKdGhyZWFkX3VpZBgBIAEoCUIIukgFcgOwAQFSCX'
    'RocmVhZFVpZBImCglwYWdlX3NpemUYAiABKAVCCbpIBhoEGGQoAFIIcGFnZVNpemUSJwoKcGFn'
    'ZV90b2tlbhgDIAEoCUIIukgFcgMYgAJSCXBhZ2VUb2tlbg==');

@$core.Deprecated('Use listExecutionsResponseDescriptor instead')
const ListExecutionsResponse$json = {
  '1': 'ListExecutionsResponse',
  '2': [
    {
      '1': 'executions',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.super_manager.control.v1.Execution',
      '10': 'executions'
    },
    {'1': 'next_page_token', '3': 2, '4': 1, '5': 9, '10': 'nextPageToken'},
  ],
};

/// Descriptor for `ListExecutionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listExecutionsResponseDescriptor = $convert.base64Decode(
    'ChZMaXN0RXhlY3V0aW9uc1Jlc3BvbnNlEkMKCmV4ZWN1dGlvbnMYASADKAsyIy5zdXBlcl9tYW'
    '5hZ2VyLmNvbnRyb2wudjEuRXhlY3V0aW9uUgpleGVjdXRpb25zEiYKD25leHRfcGFnZV90b2tl'
    'bhgCIAEoCVINbmV4dFBhZ2VUb2tlbg==');

@$core.Deprecated('Use cancelExecutionRequestDescriptor instead')
const CancelExecutionRequest$json = {
  '1': 'CancelExecutionRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
  ],
};

/// Descriptor for `CancelExecutionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelExecutionRequestDescriptor =
    $convert.base64Decode(
        'ChZDYW5jZWxFeGVjdXRpb25SZXF1ZXN0EhoKA3VpZBgBIAEoCUIIukgFcgOwAQFSA3VpZA==');

@$core.Deprecated('Use cancelExecutionResponseDescriptor instead')
const CancelExecutionResponse$json = {
  '1': 'CancelExecutionResponse',
  '2': [
    {
      '1': 'execution',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.control.v1.Execution',
      '8': {},
      '10': 'execution'
    },
  ],
};

/// Descriptor for `CancelExecutionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List cancelExecutionResponseDescriptor =
    $convert.base64Decode(
        'ChdDYW5jZWxFeGVjdXRpb25SZXNwb25zZRJJCglleGVjdXRpb24YASABKAsyIy5zdXBlcl9tYW'
        '5hZ2VyLmNvbnRyb2wudjEuRXhlY3V0aW9uQga6SAPIAQFSCWV4ZWN1dGlvbg==');

@$core.Deprecated('Use steerExecutionRequestDescriptor instead')
const SteerExecutionRequest$json = {
  '1': 'SteerExecutionRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
    {'1': 'instruction', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'instruction'},
  ],
};

/// Descriptor for `SteerExecutionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List steerExecutionRequestDescriptor = $convert.base64Decode(
    'ChVTdGVlckV4ZWN1dGlvblJlcXVlc3QSGgoDdWlkGAEgASgJQgi6SAVyA7ABAVIDdWlkEi0KC2'
    'luc3RydWN0aW9uGAIgASgJQgu6SAhyBhABKICACFILaW5zdHJ1Y3Rpb24=');

@$core.Deprecated('Use steerExecutionResponseDescriptor instead')
const SteerExecutionResponse$json = {
  '1': 'SteerExecutionResponse',
  '2': [
    {
      '1': 'execution',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.control.v1.Execution',
      '8': {},
      '10': 'execution'
    },
  ],
};

/// Descriptor for `SteerExecutionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List steerExecutionResponseDescriptor =
    $convert.base64Decode(
        'ChZTdGVlckV4ZWN1dGlvblJlc3BvbnNlEkkKCWV4ZWN1dGlvbhgBIAEoCzIjLnN1cGVyX21hbm'
        'FnZXIuY29udHJvbC52MS5FeGVjdXRpb25CBrpIA8gBAVIJZXhlY3V0aW9u');

const $core.Map<$core.String, $core.dynamic> ExecutionServiceBase$json = {
  '1': 'ExecutionService',
  '2': [
    {
      '1': 'GetExecution',
      '2': '.super_manager.control.v1.GetExecutionRequest',
      '3': '.super_manager.control.v1.GetExecutionResponse'
    },
    {
      '1': 'ListExecutions',
      '2': '.super_manager.control.v1.ListExecutionsRequest',
      '3': '.super_manager.control.v1.ListExecutionsResponse'
    },
    {
      '1': 'CancelExecution',
      '2': '.super_manager.control.v1.CancelExecutionRequest',
      '3': '.super_manager.control.v1.CancelExecutionResponse'
    },
    {
      '1': 'SteerExecution',
      '2': '.super_manager.control.v1.SteerExecutionRequest',
      '3': '.super_manager.control.v1.SteerExecutionResponse'
    },
  ],
};

@$core.Deprecated('Use executionServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ExecutionServiceBase$messageJson = {
  '.super_manager.control.v1.GetExecutionRequest': GetExecutionRequest$json,
  '.super_manager.control.v1.GetExecutionResponse': GetExecutionResponse$json,
  '.super_manager.control.v1.Execution': Execution$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.super_manager.control.v1.ListExecutionsRequest': ListExecutionsRequest$json,
  '.super_manager.control.v1.ListExecutionsResponse':
      ListExecutionsResponse$json,
  '.super_manager.control.v1.CancelExecutionRequest':
      CancelExecutionRequest$json,
  '.super_manager.control.v1.CancelExecutionResponse':
      CancelExecutionResponse$json,
  '.super_manager.control.v1.SteerExecutionRequest': SteerExecutionRequest$json,
  '.super_manager.control.v1.SteerExecutionResponse':
      SteerExecutionResponse$json,
};

/// Descriptor for `ExecutionService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List executionServiceDescriptor = $convert.base64Decode(
    'ChBFeGVjdXRpb25TZXJ2aWNlEm0KDEdldEV4ZWN1dGlvbhItLnN1cGVyX21hbmFnZXIuY29udH'
    'JvbC52MS5HZXRFeGVjdXRpb25SZXF1ZXN0Gi4uc3VwZXJfbWFuYWdlci5jb250cm9sLnYxLkdl'
    'dEV4ZWN1dGlvblJlc3BvbnNlEnMKDkxpc3RFeGVjdXRpb25zEi8uc3VwZXJfbWFuYWdlci5jb2'
    '50cm9sLnYxLkxpc3RFeGVjdXRpb25zUmVxdWVzdBowLnN1cGVyX21hbmFnZXIuY29udHJvbC52'
    'MS5MaXN0RXhlY3V0aW9uc1Jlc3BvbnNlEnYKD0NhbmNlbEV4ZWN1dGlvbhIwLnN1cGVyX21hbm'
    'FnZXIuY29udHJvbC52MS5DYW5jZWxFeGVjdXRpb25SZXF1ZXN0GjEuc3VwZXJfbWFuYWdlci5j'
    'b250cm9sLnYxLkNhbmNlbEV4ZWN1dGlvblJlc3BvbnNlEnMKDlN0ZWVyRXhlY3V0aW9uEi8uc3'
    'VwZXJfbWFuYWdlci5jb250cm9sLnYxLlN0ZWVyRXhlY3V0aW9uUmVxdWVzdBowLnN1cGVyX21h'
    'bmFnZXIuY29udHJvbC52MS5TdGVlckV4ZWN1dGlvblJlc3BvbnNl');
