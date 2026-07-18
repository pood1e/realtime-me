// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/ingest.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../../../../google/protobuf/timestamp.pbjson.dart' as $2;
import 'status_types.pbjson.dart' as $0;
import 'watch.pbjson.dart' as $1;

@$core.Deprecated('Use scrapeJobDescriptor instead')
const ScrapeJob$json = {
  '1': 'ScrapeJob',
  '2': [
    {'1': 'SCRAPE_JOB_UNSPECIFIED', '2': 0},
    {'1': 'SCRAPE_JOB_NODE_EXPORTER', '2': 1},
    {'1': 'SCRAPE_JOB_VM_NODE_EXPORTER', '2': 2},
    {'1': 'SCRAPE_JOB_DEVICE_EXPORTER', '2': 3},
    {'1': 'SCRAPE_JOB_AGENT_EXPORTER', '2': 4},
  ],
};

/// Descriptor for `ScrapeJob`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List scrapeJobDescriptor = $convert.base64Decode(
    'CglTY3JhcGVKb2ISGgoWU0NSQVBFX0pPQl9VTlNQRUNJRklFRBAAEhwKGFNDUkFQRV9KT0JfTk'
    '9ERV9FWFBPUlRFUhABEh8KG1NDUkFQRV9KT0JfVk1fTk9ERV9FWFBPUlRFUhACEh4KGlNDUkFQ'
    'RV9KT0JfREVWSUNFX0VYUE9SVEVSEAMSHQoZU0NSQVBFX0pPQl9BR0VOVF9FWFBPUlRFUhAE');

@$core.Deprecated('Use enrollDeviceRequestDescriptor instead')
const EnrollDeviceRequest$json = {
  '1': 'EnrollDeviceRequest',
  '2': [
    {
      '1': 'kind',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.DeviceKind',
      '10': 'kind'
    },
    {
      '1': 'role',
      '3': 2,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.DeviceRole',
      '10': 'role'
    },
    {'1': 'display_name', '3': 3, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'model', '3': 4, '4': 1, '5': 9, '10': 'model'},
  ],
};

/// Descriptor for `EnrollDeviceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List enrollDeviceRequestDescriptor = $convert.base64Decode(
    'ChNFbnJvbGxEZXZpY2VSZXF1ZXN0EjUKBGtpbmQYASABKA4yIS5yZWFsdGltZS5tZS5zdGF0dX'
    'MudjEuRGV2aWNlS2luZFIEa2luZBI1CgRyb2xlGAIgASgOMiEucmVhbHRpbWUubWUuc3RhdHVz'
    'LnYxLkRldmljZVJvbGVSBHJvbGUSIQoMZGlzcGxheV9uYW1lGAMgASgJUgtkaXNwbGF5TmFtZR'
    'IUCgVtb2RlbBgEIAEoCVIFbW9kZWw=');

@$core.Deprecated('Use enrollDeviceResponseDescriptor instead')
const EnrollDeviceResponse$json = {
  '1': 'EnrollDeviceResponse',
  '2': [
    {'1': 'device_uid', '3': 1, '4': 1, '5': 9, '10': 'deviceUid'},
  ],
};

/// Descriptor for `EnrollDeviceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List enrollDeviceResponseDescriptor = $convert.base64Decode(
    'ChRFbnJvbGxEZXZpY2VSZXNwb25zZRIdCgpkZXZpY2VfdWlkGAEgASgJUglkZXZpY2VVaWQ=');

@$core.Deprecated('Use reportMobileStatusRequestDescriptor instead')
const ReportMobileStatusRequest$json = {
  '1': 'ReportMobileStatusRequest',
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
      '1': 'switch_presence',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.status.v1.SwitchPresence',
      '10': 'switchPresence'
    },
  ],
};

/// Descriptor for `ReportMobileStatusRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reportMobileStatusRequestDescriptor = $convert.base64Decode(
    'ChlSZXBvcnRNb2JpbGVTdGF0dXNSZXF1ZXN0Eh0KCmRldmljZV91aWQYASABKAlSCWRldmljZV'
    'VpZBIhCgxkaXNwbGF5X25hbWUYAiABKAlSC2Rpc3BsYXlOYW1lEhQKBW1vZGVsGAMgASgJUgVt'
    'b2RlbBI3CgVwaG9uZRgEIAEoCzIhLnJlYWx0aW1lLm1lLnN0YXR1cy52MS5QaG9uZVN0YXRlUg'
    'VwaG9uZRI6CgV3YXRjaBgFIAEoCzIkLnJlYWx0aW1lLm1lLnN0YXR1cy52MS5XYXRjaFNuYXBz'
    'aG90UgV3YXRjaBJOCg9zd2l0Y2hfcHJlc2VuY2UYBiABKAsyJS5yZWFsdGltZS5tZS5zdGF0dX'
    'MudjEuU3dpdGNoUHJlc2VuY2VSDnN3aXRjaFByZXNlbmNl');

@$core.Deprecated('Use reportMobileStatusResponseDescriptor instead')
const ReportMobileStatusResponse$json = {
  '1': 'ReportMobileStatusResponse',
};

/// Descriptor for `ReportMobileStatusResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List reportMobileStatusResponseDescriptor =
    $convert.base64Decode('ChpSZXBvcnRNb2JpbGVTdGF0dXNSZXNwb25zZQ==');

@$core.Deprecated('Use registerScrapeTargetsRequestDescriptor instead')
const RegisterScrapeTargetsRequest$json = {
  '1': 'RegisterScrapeTargetsRequest',
  '2': [
    {
      '1': 'targets',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.status.v1.ScrapeTarget',
      '10': 'targets'
    },
    {'1': 'device_uid', '3': 2, '4': 1, '5': 9, '10': 'deviceUid'},
  ],
};

/// Descriptor for `RegisterScrapeTargetsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List registerScrapeTargetsRequestDescriptor =
    $convert.base64Decode(
        'ChxSZWdpc3RlclNjcmFwZVRhcmdldHNSZXF1ZXN0Ej0KB3RhcmdldHMYASADKAsyIy5yZWFsdG'
        'ltZS5tZS5zdGF0dXMudjEuU2NyYXBlVGFyZ2V0Ugd0YXJnZXRzEh0KCmRldmljZV91aWQYAiAB'
        'KAlSCWRldmljZVVpZA==');

@$core.Deprecated('Use registerScrapeTargetsResponseDescriptor instead')
const RegisterScrapeTargetsResponse$json = {
  '1': 'RegisterScrapeTargetsResponse',
};

/// Descriptor for `RegisterScrapeTargetsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List registerScrapeTargetsResponseDescriptor =
    $convert.base64Decode('Ch1SZWdpc3RlclNjcmFwZVRhcmdldHNSZXNwb25zZQ==');

@$core.Deprecated('Use scrapeTargetDescriptor instead')
const ScrapeTarget$json = {
  '1': 'ScrapeTarget',
  '2': [
    {
      '1': 'job',
      '3': 1,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.status.v1.ScrapeJob',
      '10': 'job'
    },
    {'1': 'target', '3': 2, '4': 1, '5': 9, '10': 'target'},
  ],
  '9': [
    {'1': 3, '2': 4},
    {'1': 4, '2': 5},
    {'1': 5, '2': 6},
    {'1': 6, '2': 7},
    {'1': 7, '2': 8},
  ],
  '10': ['device_uid', 'display_name', 'model', 'kind', 'role'],
};

/// Descriptor for `ScrapeTarget`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List scrapeTargetDescriptor = $convert.base64Decode(
    'CgxTY3JhcGVUYXJnZXQSMgoDam9iGAEgASgOMiAucmVhbHRpbWUubWUuc3RhdHVzLnYxLlNjcm'
    'FwZUpvYlIDam9iEhYKBnRhcmdldBgCIAEoCVIGdGFyZ2V0SgQIAxAESgQIBBAFSgQIBRAGSgQI'
    'BhAHSgQIBxAIUgpkZXZpY2VfdWlkUgxkaXNwbGF5X25hbWVSBW1vZGVsUgRraW5kUgRyb2xl');

const $core.Map<$core.String, $core.dynamic> EnrollmentServiceBase$json = {
  '1': 'EnrollmentService',
  '2': [
    {
      '1': 'EnrollDevice',
      '2': '.realtime.me.status.v1.EnrollDeviceRequest',
      '3': '.realtime.me.status.v1.EnrollDeviceResponse'
    },
  ],
};

@$core.Deprecated('Use enrollmentServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    EnrollmentServiceBase$messageJson = {
  '.realtime.me.status.v1.EnrollDeviceRequest': EnrollDeviceRequest$json,
  '.realtime.me.status.v1.EnrollDeviceResponse': EnrollDeviceResponse$json,
};

/// Descriptor for `EnrollmentService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List enrollmentServiceDescriptor = $convert.base64Decode(
    'ChFFbnJvbGxtZW50U2VydmljZRJnCgxFbnJvbGxEZXZpY2USKi5yZWFsdGltZS5tZS5zdGF0dX'
    'MudjEuRW5yb2xsRGV2aWNlUmVxdWVzdBorLnJlYWx0aW1lLm1lLnN0YXR1cy52MS5FbnJvbGxE'
    'ZXZpY2VSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> IngestServiceBase$json = {
  '1': 'IngestService',
  '2': [
    {
      '1': 'ReportMobileStatus',
      '2': '.realtime.me.status.v1.ReportMobileStatusRequest',
      '3': '.realtime.me.status.v1.ReportMobileStatusResponse'
    },
    {
      '1': 'RegisterScrapeTargets',
      '2': '.realtime.me.status.v1.RegisterScrapeTargetsRequest',
      '3': '.realtime.me.status.v1.RegisterScrapeTargetsResponse'
    },
  ],
};

@$core.Deprecated('Use ingestServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    IngestServiceBase$messageJson = {
  '.realtime.me.status.v1.ReportMobileStatusRequest':
      ReportMobileStatusRequest$json,
  '.realtime.me.status.v1.PhoneState': $0.PhoneState$json,
  '.realtime.me.status.v1.Accessory': $0.Accessory$json,
  '.realtime.me.status.v1.WatchSnapshot': $1.WatchSnapshot$json,
  '.google.protobuf.Timestamp': $2.Timestamp$json,
  '.realtime.me.status.v1.HeartRateSample': $1.HeartRateSample$json,
  '.realtime.me.status.v1.ActivityTotals': $1.ActivityTotals$json,
  '.realtime.me.status.v1.WatchState': $1.WatchState$json,
  '.realtime.me.status.v1.DeviceInfo': $1.DeviceInfo$json,
  '.realtime.me.status.v1.SwitchPresence': $0.SwitchPresence$json,
  '.realtime.me.status.v1.ReportMobileStatusResponse':
      ReportMobileStatusResponse$json,
  '.realtime.me.status.v1.RegisterScrapeTargetsRequest':
      RegisterScrapeTargetsRequest$json,
  '.realtime.me.status.v1.ScrapeTarget': ScrapeTarget$json,
  '.realtime.me.status.v1.RegisterScrapeTargetsResponse':
      RegisterScrapeTargetsResponse$json,
};

/// Descriptor for `IngestService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List ingestServiceDescriptor = $convert.base64Decode(
    'Cg1Jbmdlc3RTZXJ2aWNlEnkKElJlcG9ydE1vYmlsZVN0YXR1cxIwLnJlYWx0aW1lLm1lLnN0YX'
    'R1cy52MS5SZXBvcnRNb2JpbGVTdGF0dXNSZXF1ZXN0GjEucmVhbHRpbWUubWUuc3RhdHVzLnYx'
    'LlJlcG9ydE1vYmlsZVN0YXR1c1Jlc3BvbnNlEoIBChVSZWdpc3RlclNjcmFwZVRhcmdldHMSMy'
    '5yZWFsdGltZS5tZS5zdGF0dXMudjEuUmVnaXN0ZXJTY3JhcGVUYXJnZXRzUmVxdWVzdBo0LnJl'
    'YWx0aW1lLm1lLnN0YXR1cy52MS5SZWdpc3RlclNjcmFwZVRhcmdldHNSZXNwb25zZQ==');
