// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/status.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../../../../google/protobuf/timestamp.pbjson.dart' as $1;
import 'status_types.pbjson.dart' as $0;
import 'watch.pbjson.dart' as $2;

@$core.Deprecated('Use githubSyncStateDescriptor instead')
const GithubSyncState$json = {
  '1': 'GithubSyncState',
  '2': [
    {'1': 'GITHUB_SYNC_STATE_UNSPECIFIED', '2': 0},
    {'1': 'GITHUB_SYNC_STATE_DISABLED', '2': 1},
    {'1': 'GITHUB_SYNC_STATE_PENDING', '2': 2},
    {'1': 'GITHUB_SYNC_STATE_OK', '2': 3},
    {'1': 'GITHUB_SYNC_STATE_ERROR', '2': 4},
  ],
};

/// Descriptor for `GithubSyncState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List githubSyncStateDescriptor = $convert.base64Decode(
    'Cg9HaXRodWJTeW5jU3RhdGUSIQodR0lUSFVCX1NZTkNfU1RBVEVfVU5TUEVDSUZJRUQQABIeCh'
    'pHSVRIVUJfU1lOQ19TVEFURV9ESVNBQkxFRBABEh0KGUdJVEhVQl9TWU5DX1NUQVRFX1BFTkRJ'
    'TkcQAhIYChRHSVRIVUJfU1lOQ19TVEFURV9PSxADEhsKF0dJVEhVQl9TWU5DX1NUQVRFX0VSUk'
    '9SEAQ=');

@$core.Deprecated('Use getPublicStatusRequestDescriptor instead')
const GetPublicStatusRequest$json = {
  '1': 'GetPublicStatusRequest',
};

/// Descriptor for `GetPublicStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPublicStatusRequestDescriptor =
    $convert.base64Decode('ChZHZXRQdWJsaWNTdGF0dXNSZXF1ZXN0');

@$core.Deprecated('Use getPublicStatusResponseDescriptor instead')
const GetPublicStatusResponse$json = {
  '1': 'GetPublicStatusResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.PublicStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `GetPublicStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getPublicStatusResponseDescriptor =
    $convert.base64Decode(
        'ChdHZXRQdWJsaWNTdGF0dXNSZXNwb25zZRI7CgZzdGF0dXMYASABKAsyIy5yZWFsdGltZS5tZS'
        '5zdGF0dXMudjEuUHVibGljU3RhdHVzUgZzdGF0dXM=');

@$core.Deprecated('Use getInternalStatusRequestDescriptor instead')
const GetInternalStatusRequest$json = {
  '1': 'GetInternalStatusRequest',
};

/// Descriptor for `GetInternalStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getInternalStatusRequestDescriptor =
    $convert.base64Decode('ChhHZXRJbnRlcm5hbFN0YXR1c1JlcXVlc3Q=');

@$core.Deprecated('Use getInternalStatusResponseDescriptor instead')
const GetInternalStatusResponse$json = {
  '1': 'GetInternalStatusResponse',
  '2': [
    {
      '1': 'status',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.InternalStatus',
      '10': 'status'
    },
  ],
};

/// Descriptor for `GetInternalStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getInternalStatusResponseDescriptor =
    $convert.base64Decode(
        'ChlHZXRJbnRlcm5hbFN0YXR1c1Jlc3BvbnNlEj0KBnN0YXR1cxgBIAEoCzIlLnJlYWx0aW1lLm'
        '1lLnN0YXR1cy52MS5JbnRlcm5hbFN0YXR1c1IGc3RhdHVz');

@$core.Deprecated('Use deviceStateDescriptor instead')
const DeviceState$json = {
  '1': 'DeviceState',
  '2': [
    {'1': 'device_uid', '3': 1, '4': 1, '5': 9, '10': 'deviceUid'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'model', '3': 3, '4': 1, '5': 9, '10': 'model'},
    {
      '1': 'kind',
      '3': 4,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.DeviceKind',
      '10': 'kind'
    },
    {
      '1': 'role',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.DeviceRole',
      '10': 'role'
    },
    {
      '1': 'state',
      '3': 6,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.OnlineState',
      '10': 'state'
    },
    {
      '1': 'metrics',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.MetricSample',
      '10': 'metrics'
    },
    {
      '1': 'media',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.MediaStatus',
      '10': 'media'
    },
    {
      '1': 'accessories',
      '3': 9,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.Accessory',
      '10': 'accessories'
    },
    {
      '1': 'update_time',
      '3': 11,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updateTime'
    },
  ],
  '9': [
    {'1': 10, '2': 11},
  ],
  '10': ['children'],
};

/// Descriptor for `DeviceState`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deviceStateDescriptor = $convert.base64Decode(
    'CgtEZXZpY2VTdGF0ZRIdCgpkZXZpY2VfdWlkGAEgASgJUglkZXZpY2VVaWQSIQoMZGlzcGxheV'
    '9uYW1lGAIgASgJUgtkaXNwbGF5TmFtZRIUCgVtb2RlbBgDIAEoCVIFbW9kZWwSNQoEa2luZBgE'
    'IAEoDjIhLnJlYWx0aW1lLm1lLnN0YXR1cy52MS5EZXZpY2VLaW5kUgRraW5kEjUKBHJvbGUYBS'
    'ABKA4yIS5yZWFsdGltZS5tZS5zdGF0dXMudjEuRGV2aWNlUm9sZVIEcm9sZRI4CgVzdGF0ZRgG'
    'IAEoDjIiLnJlYWx0aW1lLm1lLnN0YXR1cy52MS5PbmxpbmVTdGF0ZVIFc3RhdGUSPQoHbWV0cm'
    'ljcxgHIAMoCzIjLnJlYWx0aW1lLm1lLnN0YXR1cy52MS5NZXRyaWNTYW1wbGVSB21ldHJpY3MS'
    'OAoFbWVkaWEYCCABKAsyIi5yZWFsdGltZS5tZS5zdGF0dXMudjEuTWVkaWFTdGF0dXNSBW1lZG'
    'lhEkIKC2FjY2Vzc29yaWVzGAkgAygLMiAucmVhbHRpbWUubWUuc3RhdHVzLnYxLkFjY2Vzc29y'
    'eVILYWNjZXNzb3JpZXMSOwoLdXBkYXRlX3RpbWUYCyABKAsyGi5nb29nbGUucHJvdG9idWYuVG'
    'ltZXN0YW1wUgp1cGRhdGVUaW1lSgQIChALUghjaGlsZHJlbg==');

@$core.Deprecated('Use mobileStateDescriptor instead')
const MobileState$json = {
  '1': 'MobileState',
  '2': [
    {'1': 'device_uid', '3': 1, '4': 1, '5': 9, '10': 'deviceUid'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'model', '3': 3, '4': 1, '5': 9, '10': 'model'},
    {
      '1': 'phone',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.PhoneState',
      '10': 'phone'
    },
    {
      '1': 'watch',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.WatchSnapshot',
      '10': 'watch'
    },
    {
      '1': 'update_time',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updateTime'
    },
    {
      '1': 'switch_presence',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.SwitchPresence',
      '10': 'switchPresence'
    },
  ],
};

/// Descriptor for `MobileState`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mobileStateDescriptor = $convert.base64Decode(
    'CgtNb2JpbGVTdGF0ZRIdCgpkZXZpY2VfdWlkGAEgASgJUglkZXZpY2VVaWQSIQoMZGlzcGxheV'
    '9uYW1lGAIgASgJUgtkaXNwbGF5TmFtZRIUCgVtb2RlbBgDIAEoCVIFbW9kZWwSNwoFcGhvbmUY'
    'BCABKAsyIS5yZWFsdGltZS5tZS5zdGF0dXMudjEuUGhvbmVTdGF0ZVIFcGhvbmUSOgoFd2F0Y2'
    'gYBSABKAsyJC5yZWFsdGltZS5tZS5zdGF0dXMudjEuV2F0Y2hTbmFwc2hvdFIFd2F0Y2gSOwoL'
    'dXBkYXRlX3RpbWUYBiABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgp1cGRhdGVUaW'
    '1lEk4KD3N3aXRjaF9wcmVzZW5jZRgHIAEoCzIlLnJlYWx0aW1lLm1lLnN0YXR1cy52MS5Td2l0'
    'Y2hQcmVzZW5jZVIOc3dpdGNoUHJlc2VuY2U=');

@$core.Deprecated('Use subagentDescriptor instead')
const Subagent$json = {
  '1': 'Subagent',
  '2': [
    {'1': 'model', '3': 1, '4': 1, '5': 9, '10': 'model'},
  ],
};

/// Descriptor for `Subagent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List subagentDescriptor =
    $convert.base64Decode('CghTdWJhZ2VudBIUCgVtb2RlbBgBIAEoCVIFbW9kZWw=');

@$core.Deprecated('Use agentDescriptor instead')
const Agent$json = {
  '1': 'Agent',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '10': 'uid'},
    {'1': 'kind', '3': 2, '4': 1, '5': 9, '10': 'kind'},
    {'1': 'device_uid', '3': 3, '4': 1, '5': 9, '10': 'deviceUid'},
    {'1': 'display_name', '3': 4, '4': 1, '5': 9, '10': 'displayName'},
    {
      '1': 'state',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.AgentState',
      '10': 'state'
    },
    {
      '1': 'budget_remaining_percent',
      '3': 6,
      '4': 1,
      '5': 5,
      '9': 0,
      '10': 'budgetRemainingPercent',
      '17': true
    },
    {
      '1': 'update_time',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updateTime'
    },
    {'1': 'model', '3': 8, '4': 1, '5': 9, '10': 'model'},
    {
      '1': 'subagents',
      '3': 10,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.Subagent',
      '10': 'subagents'
    },
  ],
  '8': [
    {'1': '_budget_remaining_percent'},
  ],
  '9': [
    {'1': 9, '2': 10},
  ],
  '10': ['subagent_count'],
};

/// Descriptor for `Agent`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List agentDescriptor = $convert.base64Decode(
    'CgVBZ2VudBIQCgN1aWQYASABKAlSA3VpZBISCgRraW5kGAIgASgJUgRraW5kEh0KCmRldmljZV'
    '91aWQYAyABKAlSCWRldmljZVVpZBIhCgxkaXNwbGF5X25hbWUYBCABKAlSC2Rpc3BsYXlOYW1l'
    'EjcKBXN0YXRlGAUgASgOMiEucmVhbHRpbWUubWUuc3RhdHVzLnYxLkFnZW50U3RhdGVSBXN0YX'
    'RlEj0KGGJ1ZGdldF9yZW1haW5pbmdfcGVyY2VudBgGIAEoBUgAUhZidWRnZXRSZW1haW5pbmdQ'
    'ZXJjZW50iAEBEjsKC3VwZGF0ZV90aW1lGAcgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdG'
    'FtcFIKdXBkYXRlVGltZRIUCgVtb2RlbBgIIAEoCVIFbW9kZWwSPQoJc3ViYWdlbnRzGAogAygL'
    'Mh8ucmVhbHRpbWUubWUuc3RhdHVzLnYxLlN1YmFnZW50UglzdWJhZ2VudHNCGwoZX2J1ZGdldF'
    '9yZW1haW5pbmdfcGVyY2VudEoECAkQClIOc3ViYWdlbnRfY291bnQ=');

@$core.Deprecated('Use githubStatusDescriptor instead')
const GithubStatus$json = {
  '1': 'GithubStatus',
  '2': [
    {'1': 'enabled', '3': 1, '4': 1, '5': 8, '10': 'enabled'},
    {
      '1': 'state',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.GithubSyncState',
      '10': 'state'
    },
    {'1': 'emoji', '3': 3, '4': 1, '5': 9, '10': 'emoji'},
    {'1': 'message', '3': 4, '4': 1, '5': 9, '10': 'message'},
    {
      '1': 'update_time',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updateTime'
    },
  ],
};

/// Descriptor for `GithubStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List githubStatusDescriptor = $convert.base64Decode(
    'CgxHaXRodWJTdGF0dXMSGAoHZW5hYmxlZBgBIAEoCFIHZW5hYmxlZBI8CgVzdGF0ZRgCIAEoDj'
    'ImLnJlYWx0aW1lLm1lLnN0YXR1cy52MS5HaXRodWJTeW5jU3RhdGVSBXN0YXRlEhQKBWVtb2pp'
    'GAMgASgJUgVlbW9qaRIYCgdtZXNzYWdlGAQgASgJUgdtZXNzYWdlEjsKC3VwZGF0ZV90aW1lGA'
    'UgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIKdXBkYXRlVGltZQ==');

@$core.Deprecated('Use githubSyncDetailDescriptor instead')
const GithubSyncDetail$json = {
  '1': 'GithubSyncDetail',
  '2': [
    {'1': 'configured', '3': 1, '4': 1, '5': 8, '10': 'configured'},
    {
      '1': 'state',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.GithubSyncState',
      '10': 'state'
    },
    {'1': 'emoji', '3': 3, '4': 1, '5': 9, '10': 'emoji'},
    {'1': 'message', '3': 4, '4': 1, '5': 9, '10': 'message'},
    {
      '1': 'last_success_time',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastSuccessTime'
    },
    {
      '1': 'last_error_time',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastErrorTime'
    },
    {'1': 'last_error', '3': 7, '4': 1, '5': 9, '10': 'lastError'},
    {
      '1': 'last_attempt_time',
      '3': 8,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'lastAttemptTime'
    },
    {'1': 'last_signature', '3': 9, '4': 1, '5': 9, '10': 'lastSignature'},
  ],
};

/// Descriptor for `GithubSyncDetail`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List githubSyncDetailDescriptor = $convert.base64Decode(
    'ChBHaXRodWJTeW5jRGV0YWlsEh4KCmNvbmZpZ3VyZWQYASABKAhSCmNvbmZpZ3VyZWQSPAoFc3'
    'RhdGUYAiABKA4yJi5yZWFsdGltZS5tZS5zdGF0dXMudjEuR2l0aHViU3luY1N0YXRlUgVzdGF0'
    'ZRIUCgVlbW9qaRgDIAEoCVIFZW1vamkSGAoHbWVzc2FnZRgEIAEoCVIHbWVzc2FnZRJGChFsYX'
    'N0X3N1Y2Nlc3NfdGltZRgFIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSD2xhc3RT'
    'dWNjZXNzVGltZRJCCg9sYXN0X2Vycm9yX3RpbWUYBiABKAsyGi5nb29nbGUucHJvdG9idWYuVG'
    'ltZXN0YW1wUg1sYXN0RXJyb3JUaW1lEh0KCmxhc3RfZXJyb3IYByABKAlSCWxhc3RFcnJvchJG'
    'ChFsYXN0X2F0dGVtcHRfdGltZRgIIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSD2'
    'xhc3RBdHRlbXB0VGltZRIlCg5sYXN0X3NpZ25hdHVyZRgJIAEoCVINbGFzdFNpZ25hdHVyZQ==');

@$core.Deprecated('Use publicStatusDescriptor instead')
const PublicStatus$json = {
  '1': 'PublicStatus',
  '2': [
    {
      '1': 'server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.DeviceState',
      '10': 'server'
    },
    {
      '1': 'devices',
      '3': 3,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.DeviceState',
      '10': 'devices'
    },
    {
      '1': 'agents',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.Agent',
      '10': 'agents'
    },
    {
      '1': 'github',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.GithubStatus',
      '10': 'github'
    },
    {
      '1': 'update_time',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updateTime'
    },
    {
      '1': 'mobiles',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.MobileState',
      '10': 'mobiles'
    },
  ],
  '9': [
    {'1': 2, '2': 3},
  ],
  '10': ['mobile'],
};

/// Descriptor for `PublicStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List publicStatusDescriptor = $convert.base64Decode(
    'CgxQdWJsaWNTdGF0dXMSOgoGc2VydmVyGAEgASgLMiIucmVhbHRpbWUubWUuc3RhdHVzLnYxLk'
    'RldmljZVN0YXRlUgZzZXJ2ZXISPAoHZGV2aWNlcxgDIAMoCzIiLnJlYWx0aW1lLm1lLnN0YXR1'
    'cy52MS5EZXZpY2VTdGF0ZVIHZGV2aWNlcxI0CgZhZ2VudHMYBCADKAsyHC5yZWFsdGltZS5tZS'
    '5zdGF0dXMudjEuQWdlbnRSBmFnZW50cxI7CgZnaXRodWIYBSABKAsyIy5yZWFsdGltZS5tZS5z'
    'dGF0dXMudjEuR2l0aHViU3RhdHVzUgZnaXRodWISOwoLdXBkYXRlX3RpbWUYBiABKAsyGi5nb2'
    '9nbGUucHJvdG9idWYuVGltZXN0YW1wUgp1cGRhdGVUaW1lEjwKB21vYmlsZXMYByADKAsyIi5y'
    'ZWFsdGltZS5tZS5zdGF0dXMudjEuTW9iaWxlU3RhdGVSB21vYmlsZXNKBAgCEANSBm1vYmlsZQ'
    '==');

@$core.Deprecated('Use internalStatusDescriptor instead')
const InternalStatus$json = {
  '1': 'InternalStatus',
  '2': [
    {
      '1': 'server',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.DeviceState',
      '10': 'server'
    },
    {
      '1': 'devices',
      '3': 3,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.DeviceState',
      '10': 'devices'
    },
    {
      '1': 'agents',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.Agent',
      '10': 'agents'
    },
    {
      '1': 'github',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.GithubSyncDetail',
      '10': 'github'
    },
    {
      '1': 'update_time',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'updateTime'
    },
    {
      '1': 'mobiles',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.MobileState',
      '10': 'mobiles'
    },
  ],
  '9': [
    {'1': 2, '2': 3},
  ],
  '10': ['mobile'],
};

/// Descriptor for `InternalStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List internalStatusDescriptor = $convert.base64Decode(
    'Cg5JbnRlcm5hbFN0YXR1cxI6CgZzZXJ2ZXIYASABKAsyIi5yZWFsdGltZS5tZS5zdGF0dXMudj'
    'EuRGV2aWNlU3RhdGVSBnNlcnZlchI8CgdkZXZpY2VzGAMgAygLMiIucmVhbHRpbWUubWUuc3Rh'
    'dHVzLnYxLkRldmljZVN0YXRlUgdkZXZpY2VzEjQKBmFnZW50cxgEIAMoCzIcLnJlYWx0aW1lLm'
    '1lLnN0YXR1cy52MS5BZ2VudFIGYWdlbnRzEj8KBmdpdGh1YhgFIAEoCzInLnJlYWx0aW1lLm1l'
    'LnN0YXR1cy52MS5HaXRodWJTeW5jRGV0YWlsUgZnaXRodWISOwoLdXBkYXRlX3RpbWUYBiABKA'
    'syGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgp1cGRhdGVUaW1lEjwKB21vYmlsZXMYByAD'
    'KAsyIi5yZWFsdGltZS5tZS5zdGF0dXMudjEuTW9iaWxlU3RhdGVSB21vYmlsZXNKBAgCEANSBm'
    '1vYmlsZQ==');

const $core.Map<$core.String, $core.dynamic> StatusServiceBase$json = {
  '1': 'StatusService',
  '2': [
    {
      '1': 'GetPublicStatus',
      '2': '.realtime.me.status.v1.GetPublicStatusRequest',
      '3': '.realtime.me.status.v1.GetPublicStatusResponse'
    },
    {
      '1': 'GetInternalStatus',
      '2': '.realtime.me.status.v1.GetInternalStatusRequest',
      '3': '.realtime.me.status.v1.GetInternalStatusResponse'
    },
  ],
};

@$core.Deprecated('Use statusServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    StatusServiceBase$messageJson = {
  '.realtime.me.status.v1.GetPublicStatusRequest': GetPublicStatusRequest$json,
  '.realtime.me.status.v1.GetPublicStatusResponse':
      GetPublicStatusResponse$json,
  '.realtime.me.status.v1.PublicStatus': PublicStatus$json,
  '.realtime.me.status.v1.DeviceState': DeviceState$json,
  '.realtime.me.status.v1.MetricSample': $0.MetricSample$json,
  '.realtime.me.status.v1.MetricSample.AttributesEntry':
      $0.MetricSample_AttributesEntry$json,
  '.realtime.me.status.v1.MediaStatus': $0.MediaStatus$json,
  '.realtime.me.status.v1.Accessory': $0.Accessory$json,
  '.google.protobuf.Timestamp': $1.Timestamp$json,
  '.realtime.me.status.v1.Agent': Agent$json,
  '.realtime.me.status.v1.Subagent': Subagent$json,
  '.realtime.me.status.v1.GithubStatus': GithubStatus$json,
  '.realtime.me.status.v1.MobileState': MobileState$json,
  '.realtime.me.status.v1.PhoneState': $0.PhoneState$json,
  '.realtime.me.status.v1.WatchSnapshot': $2.WatchSnapshot$json,
  '.realtime.me.status.v1.HeartRateSample': $2.HeartRateSample$json,
  '.realtime.me.status.v1.ActivityTotals': $2.ActivityTotals$json,
  '.realtime.me.status.v1.WatchState': $2.WatchState$json,
  '.realtime.me.status.v1.DeviceInfo': $2.DeviceInfo$json,
  '.realtime.me.status.v1.SwitchPresence': $0.SwitchPresence$json,
  '.realtime.me.status.v1.GetInternalStatusRequest':
      GetInternalStatusRequest$json,
  '.realtime.me.status.v1.GetInternalStatusResponse':
      GetInternalStatusResponse$json,
  '.realtime.me.status.v1.InternalStatus': InternalStatus$json,
  '.realtime.me.status.v1.GithubSyncDetail': GithubSyncDetail$json,
};

/// Descriptor for `StatusService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List statusServiceDescriptor = $convert.base64Decode(
    'Cg1TdGF0dXNTZXJ2aWNlEnAKD0dldFB1YmxpY1N0YXR1cxItLnJlYWx0aW1lLm1lLnN0YXR1cy'
    '52MS5HZXRQdWJsaWNTdGF0dXNSZXF1ZXN0Gi4ucmVhbHRpbWUubWUuc3RhdHVzLnYxLkdldFB1'
    'YmxpY1N0YXR1c1Jlc3BvbnNlEnYKEUdldEludGVybmFsU3RhdHVzEi8ucmVhbHRpbWUubWUuc3'
    'RhdHVzLnYxLkdldEludGVybmFsU3RhdHVzUmVxdWVzdBowLnJlYWx0aW1lLm1lLnN0YXR1cy52'
    'MS5HZXRJbnRlcm5hbFN0YXR1c1Jlc3BvbnNl');
