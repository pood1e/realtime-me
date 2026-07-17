// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/device.proto.

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
      '6': '.super_manager.control.v1.DeviceStatus',
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
    'EoCUIKukgHcgUQARiAAVILZGlzcGxheU5hbWUSSAoGc3RhdHVzGAMgASgOMiYuc3VwZXJfbWFu'
    'YWdlci5jb250cm9sLnYxLkRldmljZVN0YXR1c0IIukgFggECEAFSBnN0YXR1cxI3ChJjZXJ0aW'
    'ZpY2F0ZV9zZXJpYWwYBCABKAlCCLpIBXIDGIABUhFjZXJ0aWZpY2F0ZVNlcmlhbBI7CgtjcmVh'
    'dGVfdGltZRgFIAEoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCmNyZWF0ZVRpbWUSOw'
    'oLZXhwaXJlX3RpbWUYBiABKAsyGi5nb29nbGUucHJvdG9idWYuVGltZXN0YW1wUgpleHBpcmVU'
    'aW1l');

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
      '6': '.super_manager.control.v1.Device',
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
    'ChJQYWlyRGV2aWNlUmVzcG9uc2USQAoGZGV2aWNlGAEgASgLMiAuc3VwZXJfbWFuYWdlci5jb2'
    '50cm9sLnYxLkRldmljZUIGukgDyAEBUgZkZXZpY2USLAoNZGV2aWNlX3BrY3MxMhgCIAEoDEIH'
    'ukgEegIQAVIMZGV2aWNlUGtjczEyEjAKD3BrY3MxMl9wYXNzd29yZBgDIAEoCUIHukgEcgIQEF'
    'IOcGtjczEyUGFzc3dvcmQSKgoMZGV2aWNlX3Rva2VuGAQgASgJQge6SARyAhAgUgtkZXZpY2VU'
    'b2tlbhI1ChJjYV9jZXJ0aWZpY2F0ZV9wZW0YBSABKAxCB7pIBHoCEAFSEGNhQ2VydGlmaWNhdG'
    'VQZW0=');

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
      '6': '.super_manager.control.v1.Device',
      '10': 'devices'
    },
  ],
};

/// Descriptor for `ListDevicesResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listDevicesResponseDescriptor = $convert.base64Decode(
    'ChNMaXN0RGV2aWNlc1Jlc3BvbnNlEjoKB2RldmljZXMYASADKAsyIC5zdXBlcl9tYW5hZ2VyLm'
    'NvbnRyb2wudjEuRGV2aWNlUgdkZXZpY2Vz');

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
      '2': '.super_manager.control.v1.PairDeviceRequest',
      '3': '.super_manager.control.v1.PairDeviceResponse'
    },
  ],
};

@$core.Deprecated('Use pairingServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    PairingServiceBase$messageJson = {
  '.super_manager.control.v1.PairDeviceRequest': PairDeviceRequest$json,
  '.super_manager.control.v1.PairDeviceResponse': PairDeviceResponse$json,
  '.super_manager.control.v1.Device': Device$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
};

/// Descriptor for `PairingService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List pairingServiceDescriptor = $convert.base64Decode(
    'Cg5QYWlyaW5nU2VydmljZRJnCgpQYWlyRGV2aWNlEisuc3VwZXJfbWFuYWdlci5jb250cm9sLn'
    'YxLlBhaXJEZXZpY2VSZXF1ZXN0Giwuc3VwZXJfbWFuYWdlci5jb250cm9sLnYxLlBhaXJEZXZp'
    'Y2VSZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> DeviceServiceBase$json = {
  '1': 'DeviceService',
  '2': [
    {
      '1': 'ListDevices',
      '2': '.super_manager.control.v1.ListDevicesRequest',
      '3': '.super_manager.control.v1.ListDevicesResponse'
    },
    {
      '1': 'DeleteDevice',
      '2': '.super_manager.control.v1.DeleteDeviceRequest',
      '3': '.super_manager.control.v1.DeleteDeviceResponse'
    },
  ],
};

@$core.Deprecated('Use deviceServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    DeviceServiceBase$messageJson = {
  '.super_manager.control.v1.ListDevicesRequest': ListDevicesRequest$json,
  '.super_manager.control.v1.ListDevicesResponse': ListDevicesResponse$json,
  '.super_manager.control.v1.Device': Device$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.super_manager.control.v1.DeleteDeviceRequest': DeleteDeviceRequest$json,
  '.super_manager.control.v1.DeleteDeviceResponse': DeleteDeviceResponse$json,
};

/// Descriptor for `DeviceService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List deviceServiceDescriptor = $convert.base64Decode(
    'Cg1EZXZpY2VTZXJ2aWNlEmoKC0xpc3REZXZpY2VzEiwuc3VwZXJfbWFuYWdlci5jb250cm9sLn'
    'YxLkxpc3REZXZpY2VzUmVxdWVzdBotLnN1cGVyX21hbmFnZXIuY29udHJvbC52MS5MaXN0RGV2'
    'aWNlc1Jlc3BvbnNlEm0KDERlbGV0ZURldmljZRItLnN1cGVyX21hbmFnZXIuY29udHJvbC52MS'
    '5EZWxldGVEZXZpY2VSZXF1ZXN0Gi4uc3VwZXJfbWFuYWdlci5jb250cm9sLnYxLkRlbGV0ZURl'
    'dmljZVJlc3BvbnNl');
