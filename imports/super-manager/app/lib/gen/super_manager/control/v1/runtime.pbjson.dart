// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/runtime.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../../../google/protobuf/timestamp.pbjson.dart' as $0;

@$core.Deprecated('Use runtimeKindDescriptor instead')
const RuntimeKind$json = {
  '1': 'RuntimeKind',
  '2': [
    {'1': 'RUNTIME_KIND_UNSPECIFIED', '2': 0},
    {'1': 'RUNTIME_KIND_CODEX', '2': 1},
    {'1': 'RUNTIME_KIND_CLAUDE_CODE', '2': 2},
  ],
};

/// Descriptor for `RuntimeKind`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List runtimeKindDescriptor = $convert.base64Decode(
    'CgtSdW50aW1lS2luZBIcChhSVU5USU1FX0tJTkRfVU5TUEVDSUZJRUQQABIWChJSVU5USU1FX0'
    'tJTkRfQ09ERVgQARIcChhSVU5USU1FX0tJTkRfQ0xBVURFX0NPREUQAg==');

@$core.Deprecated('Use runtimeAvailabilityDescriptor instead')
const RuntimeAvailability$json = {
  '1': 'RuntimeAvailability',
  '2': [
    {'1': 'RUNTIME_AVAILABILITY_UNSPECIFIED', '2': 0},
    {'1': 'RUNTIME_AVAILABILITY_AVAILABLE', '2': 1},
    {'1': 'RUNTIME_AVAILABILITY_NOT_INSTALLED', '2': 2},
    {'1': 'RUNTIME_AVAILABILITY_NOT_AUTHENTICATED', '2': 3},
    {'1': 'RUNTIME_AVAILABILITY_INCOMPATIBLE', '2': 4},
    {'1': 'RUNTIME_AVAILABILITY_UNHEALTHY', '2': 5},
  ],
};

/// Descriptor for `RuntimeAvailability`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List runtimeAvailabilityDescriptor = $convert.base64Decode(
    'ChNSdW50aW1lQXZhaWxhYmlsaXR5EiQKIFJVTlRJTUVfQVZBSUxBQklMSVRZX1VOU1BFQ0lGSU'
    'VEEAASIgoeUlVOVElNRV9BVkFJTEFCSUxJVFlfQVZBSUxBQkxFEAESJgoiUlVOVElNRV9BVkFJ'
    'TEFCSUxJVFlfTk9UX0lOU1RBTExFRBACEioKJlJVTlRJTUVfQVZBSUxBQklMSVRZX05PVF9BVV'
    'RIRU5USUNBVEVEEAMSJQohUlVOVElNRV9BVkFJTEFCSUxJVFlfSU5DT01QQVRJQkxFEAQSIgoe'
    'UlVOVElNRV9BVkFJTEFCSUxJVFlfVU5IRUFMVEhZEAU=');

@$core.Deprecated('Use runtimeCapabilityDescriptor instead')
const RuntimeCapability$json = {
  '1': 'RuntimeCapability',
  '2': [
    {'1': 'RUNTIME_CAPABILITY_UNSPECIFIED', '2': 0},
    {'1': 'RUNTIME_CAPABILITY_TEXT_STREAMING', '2': 1},
    {'1': 'RUNTIME_CAPABILITY_TOOL_STREAMING', '2': 2},
    {'1': 'RUNTIME_CAPABILITY_STRUCTURED_QUESTIONS', '2': 3},
    {'1': 'RUNTIME_CAPABILITY_CANCEL', '2': 4},
    {'1': 'RUNTIME_CAPABILITY_STEER', '2': 5},
    {'1': 'RUNTIME_CAPABILITY_QUOTA', '2': 6},
    {'1': 'RUNTIME_CAPABILITY_REASONING_SUMMARIES', '2': 7},
  ],
};

/// Descriptor for `RuntimeCapability`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List runtimeCapabilityDescriptor = $convert.base64Decode(
    'ChFSdW50aW1lQ2FwYWJpbGl0eRIiCh5SVU5USU1FX0NBUEFCSUxJVFlfVU5TUEVDSUZJRUQQAB'
    'IlCiFSVU5USU1FX0NBUEFCSUxJVFlfVEVYVF9TVFJFQU1JTkcQARIlCiFSVU5USU1FX0NBUEFC'
    'SUxJVFlfVE9PTF9TVFJFQU1JTkcQAhIrCidSVU5USU1FX0NBUEFCSUxJVFlfU1RSVUNUVVJFRF'
    '9RVUVTVElPTlMQAxIdChlSVU5USU1FX0NBUEFCSUxJVFlfQ0FOQ0VMEAQSHAoYUlVOVElNRV9D'
    'QVBBQklMSVRZX1NURUVSEAUSHAoYUlVOVElNRV9DQVBBQklMSVRZX1FVT1RBEAYSKgomUlVOVE'
    'lNRV9DQVBBQklMSVRZX1JFQVNPTklOR19TVU1NQVJJRVMQBw==');

@$core.Deprecated('Use quotaFreshnessDescriptor instead')
const QuotaFreshness$json = {
  '1': 'QuotaFreshness',
  '2': [
    {'1': 'QUOTA_FRESHNESS_UNSPECIFIED', '2': 0},
    {'1': 'QUOTA_FRESHNESS_FRESH', '2': 1},
    {'1': 'QUOTA_FRESHNESS_STALE', '2': 2},
    {'1': 'QUOTA_FRESHNESS_UNAVAILABLE', '2': 3},
  ],
};

/// Descriptor for `QuotaFreshness`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List quotaFreshnessDescriptor = $convert.base64Decode(
    'Cg5RdW90YUZyZXNobmVzcxIfChtRVU9UQV9GUkVTSE5FU1NfVU5TUEVDSUZJRUQQABIZChVRVU'
    '9UQV9GUkVTSE5FU1NfRlJFU0gQARIZChVRVU9UQV9GUkVTSE5FU1NfU1RBTEUQAhIfChtRVU9U'
    'QV9GUkVTSE5FU1NfVU5BVkFJTEFCTEUQAw==');

@$core.Deprecated('Use runtimeDescriptor instead')
const Runtime$json = {
  '1': 'Runtime',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
    {
      '1': 'kind',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.super_manager.control.v1.RuntimeKind',
      '8': {},
      '10': 'kind'
    },
    {'1': 'display_name', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'displayName'},
    {'1': 'version', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'version'},
    {
      '1': 'availability',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.super_manager.control.v1.RuntimeAvailability',
      '8': {},
      '10': 'availability'
    },
    {
      '1': 'capabilities',
      '3': 6,
      '4': 3,
      '5': 14,
      '6': '.super_manager.control.v1.RuntimeCapability',
      '8': {},
      '10': 'capabilities'
    },
    {'1': 'diagnostic', '3': 7, '4': 1, '5': 9, '8': {}, '10': 'diagnostic'},
    {
      '1': 'update_time',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updateTime'
    },
  ],
};

/// Descriptor for `Runtime`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List runtimeDescriptor = $convert.base64Decode(
    'CgdSdW50aW1lEhoKA3VpZBgBIAEoCUIIukgFcgOwAQFSA3VpZBJDCgRraW5kGAIgASgOMiUuc3'
    'VwZXJfbWFuYWdlci5jb250cm9sLnYxLlJ1bnRpbWVLaW5kQgi6SAWCAQIQAVIEa2luZBIsCgxk'
    'aXNwbGF5X25hbWUYAyABKAlCCbpIBnIEEAEYQFILZGlzcGxheU5hbWUSIQoHdmVyc2lvbhgEIA'
    'EoCUIHukgEcgIYQFIHdmVyc2lvbhJbCgxhdmFpbGFiaWxpdHkYBSABKA4yLS5zdXBlcl9tYW5h'
    'Z2VyLmNvbnRyb2wudjEuUnVudGltZUF2YWlsYWJpbGl0eUIIukgFggECEAFSDGF2YWlsYWJpbG'
    'l0eRJgCgxjYXBhYmlsaXRpZXMYBiADKA4yKy5zdXBlcl9tYW5hZ2VyLmNvbnRyb2wudjEuUnVu'
    'dGltZUNhcGFiaWxpdHlCD7pIDJIBCRgBIgWCAQIQAVIMY2FwYWJpbGl0aWVzEigKCmRpYWdub3'
    'N0aWMYByABKAlCCLpIBXIDGIAEUgpkaWFnbm9zdGljEjsKC3VwZGF0ZV90aW1lGAggASgLMhou'
    'Z29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIKdXBkYXRlVGltZQ==');

@$core.Deprecated('Use quotaSnapshotDescriptor instead')
const QuotaSnapshot$json = {
  '1': 'QuotaSnapshot',
  '2': [
    {'1': 'runtime_uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'runtimeUid'},
    {
      '1': 'freshness',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.super_manager.control.v1.QuotaFreshness',
      '8': {},
      '10': 'freshness'
    },
    {
      '1': 'used_ratio',
      '3': 3,
      '4': 1,
      '5': 1,
      '8': {},
      '9': 0,
      '10': 'usedRatio',
      '17': true
    },
    {
      '1': 'reset_time',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'resetTime'
    },
    {
      '1': 'observe_time',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'observeTime'
    },
    {'1': 'source', '3': 6, '4': 1, '5': 9, '8': {}, '10': 'source'},
  ],
  '8': [
    {'1': '_used_ratio'},
  ],
};

/// Descriptor for `QuotaSnapshot`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List quotaSnapshotDescriptor = $convert.base64Decode(
    'Cg1RdW90YVNuYXBzaG90EikKC3J1bnRpbWVfdWlkGAEgASgJQgi6SAVyA7ABAVIKcnVudGltZV'
    'VpZBJQCglmcmVzaG5lc3MYAiABKA4yKC5zdXBlcl9tYW5hZ2VyLmNvbnRyb2wudjEuUXVvdGFG'
    'cmVzaG5lc3NCCLpIBYIBAhABUglmcmVzaG5lc3MSOwoKdXNlZF9yYXRpbxgDIAEoAUIXukgUEh'
    'IZAAAAAAAA8D8pAAAAAAAAAABIAFIJdXNlZFJhdGlviAEBEjkKCnJlc2V0X3RpbWUYBCABKAsy'
    'Gi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUglyZXNldFRpbWUSPQoMb2JzZXJ2ZV90aW1lGA'
    'UgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFILb2JzZXJ2ZVRpbWUSHwoGc291cmNl'
    'GAYgASgJQge6SARyAhhAUgZzb3VyY2VCDQoLX3VzZWRfcmF0aW8=');

@$core.Deprecated('Use getRuntimeRequestDescriptor instead')
const GetRuntimeRequest$json = {
  '1': 'GetRuntimeRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
  ],
};

/// Descriptor for `GetRuntimeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRuntimeRequestDescriptor = $convert.base64Decode(
    'ChFHZXRSdW50aW1lUmVxdWVzdBIaCgN1aWQYASABKAlCCLpIBXIDsAEBUgN1aWQ=');

@$core.Deprecated('Use getRuntimeResponseDescriptor instead')
const GetRuntimeResponse$json = {
  '1': 'GetRuntimeResponse',
  '2': [
    {
      '1': 'runtime',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.control.v1.Runtime',
      '8': {},
      '10': 'runtime'
    },
  ],
};

/// Descriptor for `GetRuntimeResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRuntimeResponseDescriptor = $convert.base64Decode(
    'ChJHZXRSdW50aW1lUmVzcG9uc2USQwoHcnVudGltZRgBIAEoCzIhLnN1cGVyX21hbmFnZXIuY2'
    '9udHJvbC52MS5SdW50aW1lQga6SAPIAQFSB3J1bnRpbWU=');

@$core.Deprecated('Use listRuntimesRequestDescriptor instead')
const ListRuntimesRequest$json = {
  '1': 'ListRuntimesRequest',
};

/// Descriptor for `ListRuntimesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listRuntimesRequestDescriptor =
    $convert.base64Decode('ChNMaXN0UnVudGltZXNSZXF1ZXN0');

@$core.Deprecated('Use listRuntimesResponseDescriptor instead')
const ListRuntimesResponse$json = {
  '1': 'ListRuntimesResponse',
  '2': [
    {
      '1': 'runtimes',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.super_manager.control.v1.Runtime',
      '10': 'runtimes'
    },
  ],
};

/// Descriptor for `ListRuntimesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listRuntimesResponseDescriptor = $convert.base64Decode(
    'ChRMaXN0UnVudGltZXNSZXNwb25zZRI9CghydW50aW1lcxgBIAMoCzIhLnN1cGVyX21hbmFnZX'
    'IuY29udHJvbC52MS5SdW50aW1lUghydW50aW1lcw==');

@$core.Deprecated('Use getRuntimeQuotaRequestDescriptor instead')
const GetRuntimeQuotaRequest$json = {
  '1': 'GetRuntimeQuotaRequest',
  '2': [
    {'1': 'runtime_uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'runtimeUid'},
  ],
};

/// Descriptor for `GetRuntimeQuotaRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRuntimeQuotaRequestDescriptor =
    $convert.base64Decode(
        'ChZHZXRSdW50aW1lUXVvdGFSZXF1ZXN0EikKC3J1bnRpbWVfdWlkGAEgASgJQgi6SAVyA7ABAV'
        'IKcnVudGltZVVpZA==');

@$core.Deprecated('Use getRuntimeQuotaResponseDescriptor instead')
const GetRuntimeQuotaResponse$json = {
  '1': 'GetRuntimeQuotaResponse',
  '2': [
    {
      '1': 'quota_snapshot',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.control.v1.QuotaSnapshot',
      '8': {},
      '10': 'quotaSnapshot'
    },
  ],
};

/// Descriptor for `GetRuntimeQuotaResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getRuntimeQuotaResponseDescriptor = $convert.base64Decode(
    'ChdHZXRSdW50aW1lUXVvdGFSZXNwb25zZRJWCg5xdW90YV9zbmFwc2hvdBgBIAEoCzInLnN1cG'
    'VyX21hbmFnZXIuY29udHJvbC52MS5RdW90YVNuYXBzaG90Qga6SAPIAQFSDXF1b3RhU25hcHNo'
    'b3Q=');

const $core.Map<$core.String, $core.dynamic> RuntimeServiceBase$json = {
  '1': 'RuntimeService',
  '2': [
    {
      '1': 'GetRuntime',
      '2': '.super_manager.control.v1.GetRuntimeRequest',
      '3': '.super_manager.control.v1.GetRuntimeResponse'
    },
    {
      '1': 'ListRuntimes',
      '2': '.super_manager.control.v1.ListRuntimesRequest',
      '3': '.super_manager.control.v1.ListRuntimesResponse'
    },
    {
      '1': 'GetRuntimeQuota',
      '2': '.super_manager.control.v1.GetRuntimeQuotaRequest',
      '3': '.super_manager.control.v1.GetRuntimeQuotaResponse'
    },
  ],
};

@$core.Deprecated('Use runtimeServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    RuntimeServiceBase$messageJson = {
  '.super_manager.control.v1.GetRuntimeRequest': GetRuntimeRequest$json,
  '.super_manager.control.v1.GetRuntimeResponse': GetRuntimeResponse$json,
  '.super_manager.control.v1.Runtime': Runtime$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.super_manager.control.v1.ListRuntimesRequest': ListRuntimesRequest$json,
  '.super_manager.control.v1.ListRuntimesResponse': ListRuntimesResponse$json,
  '.super_manager.control.v1.GetRuntimeQuotaRequest':
      GetRuntimeQuotaRequest$json,
  '.super_manager.control.v1.GetRuntimeQuotaResponse':
      GetRuntimeQuotaResponse$json,
  '.super_manager.control.v1.QuotaSnapshot': QuotaSnapshot$json,
};

/// Descriptor for `RuntimeService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List runtimeServiceDescriptor = $convert.base64Decode(
    'Cg5SdW50aW1lU2VydmljZRJnCgpHZXRSdW50aW1lEisuc3VwZXJfbWFuYWdlci5jb250cm9sLn'
    'YxLkdldFJ1bnRpbWVSZXF1ZXN0Giwuc3VwZXJfbWFuYWdlci5jb250cm9sLnYxLkdldFJ1bnRp'
    'bWVSZXNwb25zZRJtCgxMaXN0UnVudGltZXMSLS5zdXBlcl9tYW5hZ2VyLmNvbnRyb2wudjEuTG'
    'lzdFJ1bnRpbWVzUmVxdWVzdBouLnN1cGVyX21hbmFnZXIuY29udHJvbC52MS5MaXN0UnVudGlt'
    'ZXNSZXNwb25zZRJ2Cg9HZXRSdW50aW1lUXVvdGESMC5zdXBlcl9tYW5hZ2VyLmNvbnRyb2wudj'
    'EuR2V0UnVudGltZVF1b3RhUmVxdWVzdBoxLnN1cGVyX21hbmFnZXIuY29udHJvbC52MS5HZXRS'
    'dW50aW1lUXVvdGFSZXNwb25zZQ==');
