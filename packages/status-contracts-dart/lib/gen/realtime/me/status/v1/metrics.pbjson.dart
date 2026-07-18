// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/metrics.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../../../../google/protobuf/duration.pbjson.dart' as $1;
import '../../../../google/protobuf/timestamp.pbjson.dart' as $0;

@$core.Deprecated('Use metricSeriesDescriptor instead')
const MetricSeries$json = {
  '1': 'MetricSeries',
  '2': [
    {'1': 'METRIC_SERIES_UNSPECIFIED', '2': 0},
    {'1': 'METRIC_SERIES_HOST_CPU_UTILIZATION', '2': 1},
    {'1': 'METRIC_SERIES_HOST_MEMORY_USAGE', '2': 2},
    {'1': 'METRIC_SERIES_HOST_FILESYSTEM_UTILIZATION', '2': 3},
    {'1': 'METRIC_SERIES_PHONE_BATTERY_LEVEL', '2': 4},
    {'1': 'METRIC_SERIES_WATCH_BATTERY_LEVEL', '2': 5},
    {'1': 'METRIC_SERIES_WATCH_HEART_RATE', '2': 6},
    {'1': 'METRIC_SERIES_WATCH_STEPS', '2': 7},
    {'1': 'METRIC_SERIES_ACCESSORY_BATTERY_LEVEL', '2': 8},
    {'1': 'METRIC_SERIES_AGENT_BUDGET_REMAINING', '2': 9},
  ],
};

/// Descriptor for `MetricSeries`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List metricSeriesDescriptor = $convert.base64Decode(
    'CgxNZXRyaWNTZXJpZXMSHQoZTUVUUklDX1NFUklFU19VTlNQRUNJRklFRBAAEiYKIk1FVFJJQ1'
    '9TRVJJRVNfSE9TVF9DUFVfVVRJTElaQVRJT04QARIjCh9NRVRSSUNfU0VSSUVTX0hPU1RfTUVN'
    'T1JZX1VTQUdFEAISLQopTUVUUklDX1NFUklFU19IT1NUX0ZJTEVTWVNURU1fVVRJTElaQVRJT0'
    '4QAxIlCiFNRVRSSUNfU0VSSUVTX1BIT05FX0JBVFRFUllfTEVWRUwQBBIlCiFNRVRSSUNfU0VS'
    'SUVTX1dBVENIX0JBVFRFUllfTEVWRUwQBRIiCh5NRVRSSUNfU0VSSUVTX1dBVENIX0hFQVJUX1'
    'JBVEUQBhIdChlNRVRSSUNfU0VSSUVTX1dBVENIX1NURVBTEAcSKQolTUVUUklDX1NFUklFU19B'
    'Q0NFU1NPUllfQkFUVEVSWV9MRVZFTBAIEigKJE1FVFJJQ19TRVJJRVNfQUdFTlRfQlVER0VUX1'
    'JFTUFJTklORxAJ');

@$core.Deprecated('Use accessorySelectorDescriptor instead')
const AccessorySelector$json = {
  '1': 'AccessorySelector',
  '2': [
    {'1': 'kind', '3': 1, '4': 1, '5': 9, '10': 'kind'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '10': 'displayName'},
  ],
};

/// Descriptor for `AccessorySelector`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List accessorySelectorDescriptor = $convert.base64Decode(
    'ChFBY2Nlc3NvcnlTZWxlY3RvchISCgRraW5kGAEgASgJUgRraW5kEiEKDGRpc3BsYXlfbmFtZR'
    'gCIAEoCVILZGlzcGxheU5hbWU=');

@$core.Deprecated('Use getMetricRangeRequestDescriptor instead')
const GetMetricRangeRequest$json = {
  '1': 'GetMetricRangeRequest',
  '2': [
    {
      '1': 'series',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.MetricSeries',
      '10': 'series'
    },
    {'1': 'device_uid', '3': 2, '4': 1, '5': 9, '10': 'deviceUid'},
    {'1': 'agent_kind', '3': 3, '4': 1, '5': 9, '10': 'agentKind'},
    {
      '1': 'accessory',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.AccessorySelector',
      '10': 'accessory'
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
    {
      '1': 'step',
      '3': 7,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Duration',
      '10': 'step'
    },
  ],
};

/// Descriptor for `GetMetricRangeRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMetricRangeRequestDescriptor = $convert.base64Decode(
    'ChVHZXRNZXRyaWNSYW5nZVJlcXVlc3QSOwoGc2VyaWVzGAEgASgOMiMucmVhbHRpbWUubWUuc3'
    'RhdHVzLnYxLk1ldHJpY1Nlcmllc1IGc2VyaWVzEh0KCmRldmljZV91aWQYAiABKAlSCWRldmlj'
    'ZVVpZBIdCgphZ2VudF9raW5kGAMgASgJUglhZ2VudEtpbmQSRgoJYWNjZXNzb3J5GAQgASgLMi'
    'gucmVhbHRpbWUubWUuc3RhdHVzLnYxLkFjY2Vzc29yeVNlbGVjdG9yUglhY2Nlc3NvcnkSOQoK'
    'c3RhcnRfdGltZRgFIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCXN0YXJ0VGltZR'
    'I1CghlbmRfdGltZRgGIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSB2VuZFRpbWUS'
    'LQoEc3RlcBgHIAEoCzIZLmdvb2dsZS5wcm90b2J1Zi5EdXJhdGlvblIEc3RlcA==');

@$core.Deprecated('Use metricPointDescriptor instead')
const MetricPoint$json = {
  '1': 'MetricPoint',
  '2': [
    {
      '1': 'time',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'time'
    },
    {'1': 'value', '3': 2, '4': 1, '5': 1, '10': 'value'},
  ],
};

/// Descriptor for `MetricPoint`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List metricPointDescriptor = $convert.base64Decode(
    'CgtNZXRyaWNQb2ludBIuCgR0aW1lGAEgASgLMhouZ29vZ2xlLnByb3RvYnVmLlRpbWVzdGFtcF'
    'IEdGltZRIUCgV2YWx1ZRgCIAEoAVIFdmFsdWU=');

@$core.Deprecated('Use getMetricRangeResponseDescriptor instead')
const GetMetricRangeResponse$json = {
  '1': 'GetMetricRangeResponse',
  '2': [
    {
      '1': 'points',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.MetricPoint',
      '10': 'points'
    },
  ],
};

/// Descriptor for `GetMetricRangeResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getMetricRangeResponseDescriptor =
    $convert.base64Decode(
        'ChZHZXRNZXRyaWNSYW5nZVJlc3BvbnNlEjoKBnBvaW50cxgBIAMoCzIiLnJlYWx0aW1lLm1lLn'
        'N0YXR1cy52MS5NZXRyaWNQb2ludFIGcG9pbnRz');

const $core.Map<$core.String, $core.dynamic> MetricsServiceBase$json = {
  '1': 'MetricsService',
  '2': [
    {
      '1': 'GetMetricRange',
      '2': '.realtime.me.status.v1.GetMetricRangeRequest',
      '3': '.realtime.me.status.v1.GetMetricRangeResponse'
    },
  ],
};

@$core.Deprecated('Use metricsServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    MetricsServiceBase$messageJson = {
  '.realtime.me.status.v1.GetMetricRangeRequest': GetMetricRangeRequest$json,
  '.realtime.me.status.v1.AccessorySelector': AccessorySelector$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.google.protobuf.Duration': $1.Duration$json,
  '.realtime.me.status.v1.GetMetricRangeResponse': GetMetricRangeResponse$json,
  '.realtime.me.status.v1.MetricPoint': MetricPoint$json,
};

/// Descriptor for `MetricsService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List metricsServiceDescriptor = $convert.base64Decode(
    'Cg5NZXRyaWNzU2VydmljZRJtCg5HZXRNZXRyaWNSYW5nZRIsLnJlYWx0aW1lLm1lLnN0YXR1cy'
    '52MS5HZXRNZXRyaWNSYW5nZVJlcXVlc3QaLS5yZWFsdGltZS5tZS5zdGF0dXMudjEuR2V0TWV0'
    'cmljUmFuZ2VSZXNwb25zZQ==');
