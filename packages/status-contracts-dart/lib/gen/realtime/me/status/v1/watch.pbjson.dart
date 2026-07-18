// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/watch.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use chargeStateDescriptor instead')
const ChargeState$json = {
  '1': 'ChargeState',
  '2': [
    {'1': 'CHARGE_STATE_UNSPECIFIED', '2': 0},
    {'1': 'CHARGE_STATE_NOT_CHARGING', '2': 1},
    {'1': 'CHARGE_STATE_CHARGING', '2': 2},
  ],
};

/// Descriptor for `ChargeState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List chargeStateDescriptor = $convert.base64Decode(
    'CgtDaGFyZ2VTdGF0ZRIcChhDSEFSR0VfU1RBVEVfVU5TUEVDSUZJRUQQABIdChlDSEFSR0VfU1'
    'RBVEVfTk9UX0NIQVJHSU5HEAESGQoVQ0hBUkdFX1NUQVRFX0NIQVJHSU5HEAI=');

@$core.Deprecated('Use watchSnapshotDescriptor instead')
const WatchSnapshot$json = {
  '1': 'WatchSnapshot',
  '2': [
    {'1': 'snapshot_id', '3': 1, '4': 1, '5': 9, '10': 'snapshotId'},
    {
      '1': 'record_time',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'recordTime'
    },
    {
      '1': 'heart_rate',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.HeartRateSample',
      '10': 'heartRate'
    },
    {
      '1': 'activity_totals',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.ActivityTotals',
      '10': 'activityTotals'
    },
    {
      '1': 'watch_state',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.WatchState',
      '10': 'watchState'
    },
    {
      '1': 'device_info',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.DeviceInfo',
      '10': 'deviceInfo'
    },
  ],
};

/// Descriptor for `WatchSnapshot`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List watchSnapshotDescriptor = $convert.base64Decode(
    'Cg1XYXRjaFNuYXBzaG90Eh8KC3NuYXBzaG90X2lkGAEgASgJUgpzbmFwc2hvdElkEjsKC3JlY2'
    '9yZF90aW1lGAIgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcFIKcmVjb3JkVGltZRJF'
    'CgpoZWFydF9yYXRlGAMgASgLMiYucmVhbHRpbWUubWUuc3RhdHVzLnYxLkhlYXJ0UmF0ZVNhbX'
    'BsZVIJaGVhcnRSYXRlEk4KD2FjdGl2aXR5X3RvdGFscxgEIAEoCzIlLnJlYWx0aW1lLm1lLnN0'
    'YXR1cy52MS5BY3Rpdml0eVRvdGFsc1IOYWN0aXZpdHlUb3RhbHMSQgoLd2F0Y2hfc3RhdGUYBS'
    'ABKAsyIS5yZWFsdGltZS5tZS5zdGF0dXMudjEuV2F0Y2hTdGF0ZVIKd2F0Y2hTdGF0ZRJCCgtk'
    'ZXZpY2VfaW5mbxgGIAEoCzIhLnJlYWx0aW1lLm1lLnN0YXR1cy52MS5EZXZpY2VJbmZvUgpkZX'
    'ZpY2VJbmZv');

@$core.Deprecated('Use deviceInfoDescriptor instead')
const DeviceInfo$json = {
  '1': 'DeviceInfo',
  '2': [
    {'1': 'display_name', '3': 1, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'model', '3': 2, '4': 1, '5': 9, '10': 'model'},
  ],
};

/// Descriptor for `DeviceInfo`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deviceInfoDescriptor = $convert.base64Decode(
    'CgpEZXZpY2VJbmZvEiEKDGRpc3BsYXlfbmFtZRgBIAEoCVILZGlzcGxheU5hbWUSFAoFbW9kZW'
    'wYAiABKAlSBW1vZGVs');

@$core.Deprecated('Use heartRateSampleDescriptor instead')
const HeartRateSample$json = {
  '1': 'HeartRateSample',
  '2': [
    {'1': 'beats_per_minute', '3': 1, '4': 1, '5': 5, '10': 'beatsPerMinute'},
    {
      '1': 'sample_time',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'sampleTime'
    },
  ],
};

/// Descriptor for `HeartRateSample`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List heartRateSampleDescriptor = $convert.base64Decode(
    'Cg9IZWFydFJhdGVTYW1wbGUSKAoQYmVhdHNfcGVyX21pbnV0ZRgBIAEoBVIOYmVhdHNQZXJNaW'
    '51dGUSOwoLc2FtcGxlX3RpbWUYAiABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgpz'
    'YW1wbGVUaW1l');

@$core.Deprecated('Use activityTotalsDescriptor instead')
const ActivityTotals$json = {
  '1': 'ActivityTotals',
  '2': [
    {'1': 'steps', '3': 1, '4': 1, '5': 5, '10': 'steps'},
    {
      '1': 'sample_time',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'sampleTime'
    },
  ],
  '9': [
    {'1': 2, '2': 3},
  ],
  '10': ['calories'],
};

/// Descriptor for `ActivityTotals`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List activityTotalsDescriptor = $convert.base64Decode(
    'Cg5BY3Rpdml0eVRvdGFscxIUCgVzdGVwcxgBIAEoBVIFc3RlcHMSOwoLc2FtcGxlX3RpbWUYAy'
    'ABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgpzYW1wbGVUaW1lSgQIAhADUghjYWxv'
    'cmllcw==');

@$core.Deprecated('Use watchStateDescriptor instead')
const WatchState$json = {
  '1': 'WatchState',
  '2': [
    {'1': 'battery_percent', '3': 1, '4': 1, '5': 5, '10': 'batteryPercent'},
    {
      '1': 'charge_state',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.ChargeState',
      '10': 'chargeState'
    },
    {
      '1': 'sample_time',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'sampleTime'
    },
  ],
  '9': [
    {'1': 3, '2': 4},
  ],
  '10': ['wrist_state'],
};

/// Descriptor for `WatchState`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List watchStateDescriptor = $convert.base64Decode(
    'CgpXYXRjaFN0YXRlEicKD2JhdHRlcnlfcGVyY2VudBgBIAEoBVIOYmF0dGVyeVBlcmNlbnQSRQ'
    'oMY2hhcmdlX3N0YXRlGAIgASgOMiIucmVhbHRpbWUubWUuc3RhdHVzLnYxLkNoYXJnZVN0YXRl'
    'UgtjaGFyZ2VTdGF0ZRI7CgtzYW1wbGVfdGltZRgEIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW'
    '1lc3RhbXBSCnNhbXBsZVRpbWVKBAgDEARSC3dyaXN0X3N0YXRl');

@$core.Deprecated('Use reportWatchSnapshotRequestDescriptor instead')
const ReportWatchSnapshotRequest$json = {
  '1': 'ReportWatchSnapshotRequest',
  '2': [
    {
      '1': 'watch_snapshot',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.WatchSnapshot',
      '10': 'watchSnapshot'
    },
  ],
};

/// Descriptor for `ReportWatchSnapshotRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reportWatchSnapshotRequestDescriptor =
    $convert.base64Decode(
        'ChpSZXBvcnRXYXRjaFNuYXBzaG90UmVxdWVzdBJLCg53YXRjaF9zbmFwc2hvdBgBIAEoCzIkLn'
        'JlYWx0aW1lLm1lLnN0YXR1cy52MS5XYXRjaFNuYXBzaG90Ug13YXRjaFNuYXBzaG90');
