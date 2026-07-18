// This is a generated file - do not edit.
//
// Generated from realtime/me/manager/control/v1/device.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

import '../../../../../google/protobuf/timestamp.pbjson.dart' as $0;

@$core.Deprecated('Use deviceStatusDescriptor instead')
const DeviceStatus$json = {
  '1': 'DeviceStatus',
  '2': [
    {'1': 'DEVICE_STATUS_UNSPECIFIED', '2': 0},
    {'1': 'DEVICE_STATUS_ACTIVE', '2': 1},
    {'1': 'DEVICE_STATUS_REVOKED', '2': 2},
  ],
};

/// Descriptor for `DeviceStatus`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List deviceStatusDescriptor = $convert.base64Decode(
    'CgxEZXZpY2VTdGF0dXMSHQoZREVWSUNFX1NUQVRVU19VTlNQRUNJRklFRBAAEhgKFERFVklDRV'
    '9TVEFUVVNfQUNUSVZFEAESGQoVREVWSUNFX1NUQVRVU19SRVZPS0VEEAI=');

@$core.Deprecated('Use deviceDescriptor instead')
const Device$json = {
  '1': 'Device',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'displayName'},
    {
      '1': 'status',
      '3': 3,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.manager.control.v1.DeviceStatus',
      '8': {},
      '10': 'status'
    },
    {
      '1': 'certificate_serial',
      '3': 4,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'certificateSerial'
    },
    {
      '1': 'create_time',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createTime'
    },
    {
      '1': 'expire_time',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'expireTime'
    },
  ],
};

/// Descriptor for `Device`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deviceDescriptor = $convert.base64Decode(
    'CgZEZXZpY2USGgoDdWlkGAEgASgJQgi6SAVyA7ABAVIDdWlkEi0KDGRpc3BsYXlfbmFtZRgCIA'
    'EoCUIKukgHcgUQARiAAVILZGlzcGxheU5hbWUSTgoGc3RhdHVzGAMgASgOMiwucmVhbHRpbWUu'
    'bWUubWFuYWdlci5jb250cm9sLnYxLkRldmljZVN0YXR1c0IIukgFggECEAFSBnN0YXR1cxI3Ch'
    'JjZXJ0aWZpY2F0ZV9zZXJpYWwYBCABKAlCCLpIBXIDGIABUhFjZXJ0aWZpY2F0ZVNlcmlhbBI7'
    'CgtjcmVhdGVfdGltZRgFIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCmNyZWF0ZV'
    'RpbWUSOwoLZXhwaXJlX3RpbWUYBiABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgpl'
    'eHBpcmVUaW1l');

@$core.Deprecated('Use pairDeviceRequestDescriptor instead')
const PairDeviceRequest$json = {
  '1': 'PairDeviceRequest',
  '2': [
    {
      '1': 'pairing_secret',
      '3': 1,
      '4': 1,
      '5': 12,
      '8': {},
      '10': 'pairingSecret'
    },
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'displayName'},
  ],
};

/// Descriptor for `PairDeviceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List pairDeviceRequestDescriptor = $convert.base64Decode(
    'ChFQYWlyRGV2aWNlUmVxdWVzdBIuCg5wYWlyaW5nX3NlY3JldBgBIAEoDEIHukgEegJoIFINcG'
    'FpcmluZ1NlY3JldBItCgxkaXNwbGF5X25hbWUYAiABKAlCCrpIB3IFEAEYgAFSC2Rpc3BsYXlO'
    'YW1l');

@$core.Deprecated('Use pairDeviceResponseDescriptor instead')
const PairDeviceResponse$json = {
  '1': 'PairDeviceResponse',
  '2': [
    {
      '1': 'device',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.manager.control.v1.Device',
      '8': {},
      '10': 'device'
    },
    {
      '1': 'device_pkcs12',
      '3': 2,
      '4': 1,
      '5': 12,
      '8': {},
      '10': 'devicePkcs12'
    },
    {
      '1': 'pkcs12_password',
      '3': 3,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'pkcs12Password'
    },
    {'1': 'device_token', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'deviceToken'},
    {
      '1': 'ca_certificate_pem',
      '3': 5,
      '4': 1,
      '5': 12,
      '8': {},
      '10': 'caCertificatePem'
    },
  ],
};

/// Descriptor for `PairDeviceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List pairDeviceResponseDescriptor = $convert.base64Decode(
    'ChJQYWlyRGV2aWNlUmVzcG9uc2USRgoGZGV2aWNlGAEgASgLMiYucmVhbHRpbWUubWUubWFuYW'
    'dlci5jb250cm9sLnYxLkRldmljZUIGukgDyAEBUgZkZXZpY2USLAoNZGV2aWNlX3BrY3MxMhgC'
    'IAEoDEIHukgEegIQAVIMZGV2aWNlUGtjczEyEjAKD3BrY3MxMl9wYXNzd29yZBgDIAEoCUIHuk'
    'gEcgIQEFIOcGtjczEyUGFzc3dvcmQSKgoMZGV2aWNlX3Rva2VuGAQgASgJQge6SARyAhAgUgtk'
    'ZXZpY2VUb2tlbhI1ChJjYV9jZXJ0aWZpY2F0ZV9wZW0YBSABKAxCB7pIBHoCEAFSEGNhQ2VydG'
    'lmaWNhdGVQZW0=');

@$core.Deprecated('Use listDevicesRequestDescriptor instead')
const ListDevicesRequest$json = {
  '1': 'ListDevicesRequest',
};

/// Descriptor for `ListDevicesRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDevicesRequestDescriptor =
    $convert.base64Decode('ChJMaXN0RGV2aWNlc1JlcXVlc3Q=');

@$core.Deprecated('Use listDevicesResponseDescriptor instead')
const ListDevicesResponse$json = {
  '1': 'ListDevicesResponse',
  '2': [
    {
      '1': 'devices',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.manager.control.v1.Device',
      '10': 'devices'
    },
  ],
};

/// Descriptor for `ListDevicesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDevicesResponseDescriptor = $convert.base64Decode(
    'ChNMaXN0RGV2aWNlc1Jlc3BvbnNlEkAKB2RldmljZXMYASADKAsyJi5yZWFsdGltZS5tZS5tYW'
    '5hZ2VyLmNvbnRyb2wudjEuRGV2aWNlUgdkZXZpY2Vz');

@$core.Deprecated('Use deleteDeviceRequestDescriptor instead')
const DeleteDeviceRequest$json = {
  '1': 'DeleteDeviceRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
  ],
};

/// Descriptor for `DeleteDeviceRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteDeviceRequestDescriptor =
    $convert.base64Decode(
        'ChNEZWxldGVEZXZpY2VSZXF1ZXN0EhoKA3VpZBgBIAEoCUIIukgFcgOwAQFSA3VpZA==');

@$core.Deprecated('Use deleteDeviceResponseDescriptor instead')
const DeleteDeviceResponse$json = {
  '1': 'DeleteDeviceResponse',
};

/// Descriptor for `DeleteDeviceResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteDeviceResponseDescriptor =
    $convert.base64Decode('ChREZWxldGVEZXZpY2VSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> PairingServiceBase$json = {
  '1': 'PairingService',
  '2': [
    {
      '1': 'PairDevice',
      '2': '.realtime.me.manager.control.v1.PairDeviceRequest',
      '3': '.realtime.me.manager.control.v1.PairDeviceResponse'
    },
  ],
};

@$core.Deprecated('Use pairingServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    PairingServiceBase$messageJson = {
  '.realtime.me.manager.control.v1.PairDeviceRequest': PairDeviceRequest$json,
  '.realtime.me.manager.control.v1.PairDeviceResponse': PairDeviceResponse$json,
  '.realtime.me.manager.control.v1.Device': Device$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
};

/// Descriptor for `PairingService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List pairingServiceDescriptor = $convert.base64Decode(
    'Cg5QYWlyaW5nU2VydmljZRJzCgpQYWlyRGV2aWNlEjEucmVhbHRpbWUubWUubWFuYWdlci5jb2'
    '50cm9sLnYxLlBhaXJEZXZpY2VSZXF1ZXN0GjIucmVhbHRpbWUubWUubWFuYWdlci5jb250cm9s'
    'LnYxLlBhaXJEZXZpY2VSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> DeviceServiceBase$json = {
  '1': 'DeviceService',
  '2': [
    {
      '1': 'ListDevices',
      '2': '.realtime.me.manager.control.v1.ListDevicesRequest',
      '3': '.realtime.me.manager.control.v1.ListDevicesResponse'
    },
    {
      '1': 'DeleteDevice',
      '2': '.realtime.me.manager.control.v1.DeleteDeviceRequest',
      '3': '.realtime.me.manager.control.v1.DeleteDeviceResponse'
    },
  ],
};

@$core.Deprecated('Use deviceServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    DeviceServiceBase$messageJson = {
  '.realtime.me.manager.control.v1.ListDevicesRequest': ListDevicesRequest$json,
  '.realtime.me.manager.control.v1.ListDevicesResponse':
      ListDevicesResponse$json,
  '.realtime.me.manager.control.v1.Device': Device$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.realtime.me.manager.control.v1.DeleteDeviceRequest':
      DeleteDeviceRequest$json,
  '.realtime.me.manager.control.v1.DeleteDeviceResponse':
      DeleteDeviceResponse$json,
};

/// Descriptor for `DeviceService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List deviceServiceDescriptor = $convert.base64Decode(
    'Cg1EZXZpY2VTZXJ2aWNlEnYKC0xpc3REZXZpY2VzEjIucmVhbHRpbWUubWUubWFuYWdlci5jb2'
    '50cm9sLnYxLkxpc3REZXZpY2VzUmVxdWVzdBozLnJlYWx0aW1lLm1lLm1hbmFnZXIuY29udHJv'
    'bC52MS5MaXN0RGV2aWNlc1Jlc3BvbnNlEnkKDERlbGV0ZURldmljZRIzLnJlYWx0aW1lLm1lLm'
    '1hbmFnZXIuY29udHJvbC52MS5EZWxldGVEZXZpY2VSZXF1ZXN0GjQucmVhbHRpbWUubWUubWFu'
    'YWdlci5jb250cm9sLnYxLkRlbGV0ZURldmljZVJlc3BvbnNl');
