// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/status_types.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use deviceKindDescriptor instead')
const DeviceKind$json = {
  '1': 'DeviceKind',
  '2': [
    {'1': 'DEVICE_KIND_UNSPECIFIED', '2': 0},
    {'1': 'DEVICE_KIND_HOST', '2': 1},
    {'1': 'DEVICE_KIND_VIRTUAL_MACHINE', '2': 2},
    {'1': 'DEVICE_KIND_PHONE', '2': 3},
    {'1': 'DEVICE_KIND_WATCH', '2': 4},
  ],
};

/// Descriptor for `DeviceKind`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List deviceKindDescriptor = $convert.base64Decode(
    'CgpEZXZpY2VLaW5kEhsKF0RFVklDRV9LSU5EX1VOU1BFQ0lGSUVEEAASFAoQREVWSUNFX0tJTk'
    'RfSE9TVBABEh8KG0RFVklDRV9LSU5EX1ZJUlRVQUxfTUFDSElORRACEhUKEURFVklDRV9LSU5E'
    'X1BIT05FEAMSFQoRREVWSUNFX0tJTkRfV0FUQ0gQBA==');

@$core.Deprecated('Use deviceRoleDescriptor instead')
const DeviceRole$json = {
  '1': 'DeviceRole',
  '2': [
    {'1': 'DEVICE_ROLE_UNSPECIFIED', '2': 0},
    {'1': 'DEVICE_ROLE_SERVER', '2': 1},
    {'1': 'DEVICE_ROLE_DESKTOP', '2': 2},
    {'1': 'DEVICE_ROLE_VM', '2': 3},
  ],
};

/// Descriptor for `DeviceRole`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List deviceRoleDescriptor = $convert.base64Decode(
    'CgpEZXZpY2VSb2xlEhsKF0RFVklDRV9ST0xFX1VOU1BFQ0lGSUVEEAASFgoSREVWSUNFX1JPTE'
    'VfU0VSVkVSEAESFwoTREVWSUNFX1JPTEVfREVTS1RPUBACEhIKDkRFVklDRV9ST0xFX1ZNEAM=');

@$core.Deprecated('Use onlineStateDescriptor instead')
const OnlineState$json = {
  '1': 'OnlineState',
  '2': [
    {'1': 'ONLINE_STATE_UNSPECIFIED', '2': 0},
    {'1': 'ONLINE_STATE_ONLINE', '2': 1},
    {'1': 'ONLINE_STATE_OFFLINE', '2': 2},
  ],
};

/// Descriptor for `OnlineState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List onlineStateDescriptor = $convert.base64Decode(
    'CgtPbmxpbmVTdGF0ZRIcChhPTkxJTkVfU1RBVEVfVU5TUEVDSUZJRUQQABIXChNPTkxJTkVfU1'
    'RBVEVfT05MSU5FEAESGAoUT05MSU5FX1NUQVRFX09GRkxJTkUQAg==');

@$core.Deprecated('Use networkStateDescriptor instead')
const NetworkState$json = {
  '1': 'NetworkState',
  '2': [
    {'1': 'NETWORK_STATE_UNSPECIFIED', '2': 0},
    {'1': 'NETWORK_STATE_OFFLINE', '2': 1},
    {'1': 'NETWORK_STATE_WIFI', '2': 2},
    {'1': 'NETWORK_STATE_CELLULAR', '2': 3},
    {'1': 'NETWORK_STATE_VPN', '2': 4},
    {'1': 'NETWORK_STATE_ONLINE', '2': 5},
  ],
};

/// Descriptor for `NetworkState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List networkStateDescriptor = $convert.base64Decode(
    'CgxOZXR3b3JrU3RhdGUSHQoZTkVUV09SS19TVEFURV9VTlNQRUNJRklFRBAAEhkKFU5FVFdPUk'
    'tfU1RBVEVfT0ZGTElORRABEhYKEk5FVFdPUktfU1RBVEVfV0lGSRACEhoKFk5FVFdPUktfU1RB'
    'VEVfQ0VMTFVMQVIQAxIVChFORVRXT1JLX1NUQVRFX1ZQThAEEhgKFE5FVFdPUktfU1RBVEVfT0'
    '5MSU5FEAU=');

@$core.Deprecated('Use agentStateDescriptor instead')
const AgentState$json = {
  '1': 'AgentState',
  '2': [
    {'1': 'AGENT_STATE_UNSPECIFIED', '2': 0},
    {'1': 'AGENT_STATE_IDLE', '2': 1},
    {'1': 'AGENT_STATE_RUNNING', '2': 2},
    {'1': 'AGENT_STATE_FAILED', '2': 3},
  ],
};

/// Descriptor for `AgentState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List agentStateDescriptor = $convert.base64Decode(
    'CgpBZ2VudFN0YXRlEhsKF0FHRU5UX1NUQVRFX1VOU1BFQ0lGSUVEEAASFAoQQUdFTlRfU1RBVE'
    'VfSURMRRABEhcKE0FHRU5UX1NUQVRFX1JVTk5JTkcQAhIWChJBR0VOVF9TVEFURV9GQUlMRUQQ'
    'Aw==');

@$core.Deprecated('Use accessoryDescriptor instead')
const Accessory$json = {
  '1': 'Accessory',
  '2': [
    {'1': 'kind', '3': 1, '4': 1, '5': 9, '10': 'kind'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'model', '3': 3, '4': 1, '5': 9, '10': 'model'},
    {
      '1': 'battery_percent',
      '3': 4,
      '4': 1,
      '5': 5,
      '9': 0,
      '10': 'batteryPercent',
      '17': true
    },
  ],
  '8': [
    {'1': '_battery_percent'},
  ],
};

/// Descriptor for `Accessory`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accessoryDescriptor = $convert.base64Decode(
    'CglBY2Nlc3NvcnkSEgoEa2luZBgBIAEoCVIEa2luZBIhCgxkaXNwbGF5X25hbWUYAiABKAlSC2'
    'Rpc3BsYXlOYW1lEhQKBW1vZGVsGAMgASgJUgVtb2RlbBIsCg9iYXR0ZXJ5X3BlcmNlbnQYBCAB'
    'KAVIAFIOYmF0dGVyeVBlcmNlbnSIAQFCEgoQX2JhdHRlcnlfcGVyY2VudA==');

@$core.Deprecated('Use mediaStatusDescriptor instead')
const MediaStatus$json = {
  '1': 'MediaStatus',
  '2': [
    {'1': 'title', '3': 1, '4': 1, '5': 9, '10': 'title'},
    {'1': 'artist', '3': 2, '4': 1, '5': 9, '10': 'artist'},
  ],
  '9': [
    {'1': 3, '2': 4},
  ],
  '10': ['cover_url'],
};

/// Descriptor for `MediaStatus`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List mediaStatusDescriptor = $convert.base64Decode(
    'CgtNZWRpYVN0YXR1cxIUCgV0aXRsZRgBIAEoCVIFdGl0bGUSFgoGYXJ0aXN0GAIgASgJUgZhcn'
    'Rpc3RKBAgDEARSCWNvdmVyX3VybA==');

@$core.Deprecated('Use switchPresenceDescriptor instead')
const SwitchPresence$json = {
  '1': 'SwitchPresence',
  '2': [
    {
      '1': 'state',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.OnlineState',
      '10': 'state'
    },
    {'1': 'game_name', '3': 2, '4': 1, '5': 9, '10': 'gameName'},
    {'1': 'title_id', '3': 3, '4': 1, '5': 9, '10': 'titleId'},
    {'1': 'image_uri', '3': 4, '4': 1, '5': 9, '10': 'imageUri'},
    {
      '1': 'presence_update_time',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'presenceUpdateTime'
    },
    {
      '1': 'logout_time',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'logoutTime'
    },
    {
      '1': 'fetch_time',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'fetchTime'
    },
  ],
};

/// Descriptor for `SwitchPresence`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List switchPresenceDescriptor = $convert.base64Decode(
    'Cg5Td2l0Y2hQcmVzZW5jZRI4CgVzdGF0ZRgBIAEoDjIiLnJlYWx0aW1lLm1lLnN0YXR1cy52MS'
    '5PbmxpbmVTdGF0ZVIFc3RhdGUSGwoJZ2FtZV9uYW1lGAIgASgJUghnYW1lTmFtZRIZCgh0aXRs'
    'ZV9pZBgDIAEoCVIHdGl0bGVJZBIbCglpbWFnZV91cmkYBCABKAlSCGltYWdlVXJpEkwKFHByZX'
    'NlbmNlX3VwZGF0ZV90aW1lGAUgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIScHJl'
    'c2VuY2VVcGRhdGVUaW1lEjsKC2xvZ291dF90aW1lGAYgASgLMhouZ29vZ2xlLnByb3RvYnVmLl'
    'RpbWVzdGFtcFIKbG9nb3V0VGltZRI5CgpmZXRjaF90aW1lGAcgASgLMhouZ29vZ2xlLnByb3Rv'
    'YnVmLlRpbWVzdGFtcFIJZmV0Y2hUaW1l');

@$core.Deprecated('Use metricSampleDescriptor instead')
const MetricSample$json = {
  '1': 'MetricSample',
  '2': [
    {'1': 'name', '3': 1, '4': 1, '5': 9, '10': 'name'},
    {'1': 'unit', '3': 2, '4': 1, '5': 9, '10': 'unit'},
    {'1': 'value', '3': 3, '4': 1, '5': 1, '10': 'value'},
    {
      '1': 'attributes',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.MetricSample.AttributesEntry',
      '10': 'attributes'
    },
  ],
  '3': [MetricSample_AttributesEntry$json],
};

@$core.Deprecated('Use metricSampleDescriptor instead')
const MetricSample_AttributesEntry$json = {
  '1': 'AttributesEntry',
  '2': [
    {'1': 'key', '3': 1, '4': 1, '5': 9, '10': 'key'},
    {'1': 'value', '3': 2, '4': 1, '5': 9, '10': 'value'},
  ],
  '7': {'7': true},
};

/// Descriptor for `MetricSample`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List metricSampleDescriptor = $convert.base64Decode(
    'CgxNZXRyaWNTYW1wbGUSEgoEbmFtZRgBIAEoCVIEbmFtZRISCgR1bml0GAIgASgJUgR1bml0Eh'
    'QKBXZhbHVlGAMgASgBUgV2YWx1ZRJTCgphdHRyaWJ1dGVzGAQgAygLMjMucmVhbHRpbWUubWUu'
    'c3RhdHVzLnYxLk1ldHJpY1NhbXBsZS5BdHRyaWJ1dGVzRW50cnlSCmF0dHJpYnV0ZXMaPQoPQX'
    'R0cmlidXRlc0VudHJ5EhAKA2tleRgBIAEoCVIDa2V5EhQKBXZhbHVlGAIgASgJUgV2YWx1ZToC'
    'OAE=');

@$core.Deprecated('Use phoneStateDescriptor instead')
const PhoneState$json = {
  '1': 'PhoneState',
  '2': [
    {
      '1': 'battery_percent',
      '3': 1,
      '4': 1,
      '5': 5,
      '9': 0,
      '10': 'batteryPercent',
      '17': true
    },
    {
      '1': 'charge_state',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.ChargeState',
      '10': 'chargeState'
    },
    {
      '1': 'network',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.NetworkState',
      '10': 'network'
    },
    {
      '1': 'accessories',
      '3': 4,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.Accessory',
      '10': 'accessories'
    },
  ],
  '8': [
    {'1': '_battery_percent'},
  ],
};

/// Descriptor for `PhoneState`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List phoneStateDescriptor = $convert.base64Decode(
    'CgpQaG9uZVN0YXRlEiwKD2JhdHRlcnlfcGVyY2VudBgBIAEoBUgAUg5iYXR0ZXJ5UGVyY2VudI'
    'gBARJFCgxjaGFyZ2Vfc3RhdGUYAiABKA4yIi5yZWFsdGltZS5tZS5zdGF0dXMudjEuQ2hhcmdl'
    'U3RhdGVSC2NoYXJnZVN0YXRlEj0KB25ldHdvcmsYAyABKA4yIy5yZWFsdGltZS5tZS5zdGF0dX'
    'MudjEuTmV0d29ya1N0YXRlUgduZXR3b3JrEkIKC2FjY2Vzc29yaWVzGAQgAygLMiAucmVhbHRp'
    'bWUubWUuc3RhdHVzLnYxLkFjY2Vzc29yeVILYWNjZXNzb3JpZXNCEgoQX2JhdHRlcnlfcGVyY2'
    'VudA==');
