// This is a generated file - do not edit.
//
// Generated from super_manager/terminal/v1/terminal.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use terminalAttachDescriptor instead')
const TerminalAttach$json = {
  '1': 'TerminalAttach',
  '2': [
    {
      '1': 'terminal_session_uid',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'terminalSessionUid'
    },
  ],
};

/// Descriptor for `TerminalAttach`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalAttachDescriptor = $convert.base64Decode(
    'Cg5UZXJtaW5hbEF0dGFjaBI6ChR0ZXJtaW5hbF9zZXNzaW9uX3VpZBgBIAEoCUIIukgFcgOwAQ'
    'FSEnRlcm1pbmFsU2Vzc2lvblVpZA==');

@$core.Deprecated('Use terminalInputDescriptor instead')
const TerminalInput$json = {
  '1': 'TerminalInput',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 12, '8': {}, '10': 'data'},
  ],
};

/// Descriptor for `TerminalInput`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalInputDescriptor = $convert.base64Decode(
    'Cg1UZXJtaW5hbElucHV0Eh8KBGRhdGEYASABKAxCC7pICHoGEAEYgIAEUgRkYXRh');

@$core.Deprecated('Use terminalResizeDescriptor instead')
const TerminalResize$json = {
  '1': 'TerminalResize',
  '2': [
    {'1': 'columns', '3': 1, '4': 1, '5': 13, '8': {}, '10': 'columns'},
    {'1': 'rows', '3': 2, '4': 1, '5': 13, '8': {}, '10': 'rows'},
  ],
};

/// Descriptor for `TerminalResize`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalResizeDescriptor = $convert.base64Decode(
    'Cg5UZXJtaW5hbFJlc2l6ZRIkCgdjb2x1bW5zGAEgASgNQgq6SAcqBRjoBygUUgdjb2x1bW5zEh'
    '4KBHJvd3MYAiABKA1CCrpIByoFGOgHKAVSBHJvd3M=');

@$core.Deprecated('Use terminalDetachDescriptor instead')
const TerminalDetach$json = {
  '1': 'TerminalDetach',
};

/// Descriptor for `TerminalDetach`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalDetachDescriptor =
    $convert.base64Decode('Cg5UZXJtaW5hbERldGFjaA==');

@$core.Deprecated('Use terminalCloseDescriptor instead')
const TerminalClose$json = {
  '1': 'TerminalClose',
};

/// Descriptor for `TerminalClose`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalCloseDescriptor =
    $convert.base64Decode('Cg1UZXJtaW5hbENsb3Nl');

@$core.Deprecated('Use terminalPingDescriptor instead')
const TerminalPing$json = {
  '1': 'TerminalPing',
  '2': [
    {'1': 'nonce', '3': 1, '4': 1, '5': 4, '10': 'nonce'},
  ],
};

/// Descriptor for `TerminalPing`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalPingDescriptor =
    $convert.base64Decode('CgxUZXJtaW5hbFBpbmcSFAoFbm9uY2UYASABKARSBW5vbmNl');

@$core.Deprecated('Use terminalClientFrameDescriptor instead')
const TerminalClientFrame$json = {
  '1': 'TerminalClientFrame',
  '2': [
    {
      '1': 'attach',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalAttach',
      '9': 0,
      '10': 'attach'
    },
    {
      '1': 'input',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalInput',
      '9': 0,
      '10': 'input'
    },
    {
      '1': 'resize',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalResize',
      '9': 0,
      '10': 'resize'
    },
    {
      '1': 'detach',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalDetach',
      '9': 0,
      '10': 'detach'
    },
    {
      '1': 'close',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalClose',
      '9': 0,
      '10': 'close'
    },
    {
      '1': 'ping',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalPing',
      '9': 0,
      '10': 'ping'
    },
  ],
  '8': [
    {'1': 'payload'},
  ],
};

/// Descriptor for `TerminalClientFrame`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalClientFrameDescriptor = $convert.base64Decode(
    'ChNUZXJtaW5hbENsaWVudEZyYW1lEkMKBmF0dGFjaBgBIAEoCzIpLnN1cGVyX21hbmFnZXIudG'
    'VybWluYWwudjEuVGVybWluYWxBdHRhY2hIAFIGYXR0YWNoEkAKBWlucHV0GAIgASgLMiguc3Vw'
    'ZXJfbWFuYWdlci50ZXJtaW5hbC52MS5UZXJtaW5hbElucHV0SABSBWlucHV0EkMKBnJlc2l6ZR'
    'gDIAEoCzIpLnN1cGVyX21hbmFnZXIudGVybWluYWwudjEuVGVybWluYWxSZXNpemVIAFIGcmVz'
    'aXplEkMKBmRldGFjaBgEIAEoCzIpLnN1cGVyX21hbmFnZXIudGVybWluYWwudjEuVGVybWluYW'
    'xEZXRhY2hIAFIGZGV0YWNoEkAKBWNsb3NlGAUgASgLMiguc3VwZXJfbWFuYWdlci50ZXJtaW5h'
    'bC52MS5UZXJtaW5hbENsb3NlSABSBWNsb3NlEj0KBHBpbmcYBiABKAsyJy5zdXBlcl9tYW5hZ2'
    'VyLnRlcm1pbmFsLnYxLlRlcm1pbmFsUGluZ0gAUgRwaW5nQgkKB3BheWxvYWQ=');

@$core.Deprecated('Use terminalAttachedDescriptor instead')
const TerminalAttached$json = {
  '1': 'TerminalAttached',
  '2': [
    {
      '1': 'terminal_session_uid',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'terminalSessionUid'
    },
    {'1': 'columns', '3': 2, '4': 1, '5': 13, '10': 'columns'},
    {'1': 'rows', '3': 3, '4': 1, '5': 13, '10': 'rows'},
  ],
};

/// Descriptor for `TerminalAttached`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalAttachedDescriptor = $convert.base64Decode(
    'ChBUZXJtaW5hbEF0dGFjaGVkEjoKFHRlcm1pbmFsX3Nlc3Npb25fdWlkGAEgASgJQgi6SAVyA7'
    'ABAVISdGVybWluYWxTZXNzaW9uVWlkEhgKB2NvbHVtbnMYAiABKA1SB2NvbHVtbnMSEgoEcm93'
    'cxgDIAEoDVIEcm93cw==');

@$core.Deprecated('Use terminalOutputDescriptor instead')
const TerminalOutput$json = {
  '1': 'TerminalOutput',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 12, '8': {}, '10': 'data'},
  ],
};

/// Descriptor for `TerminalOutput`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalOutputDescriptor = $convert.base64Decode(
    'Cg5UZXJtaW5hbE91dHB1dBIfCgRkYXRhGAEgASgMQgu6SAh6BhABGICAQFIEZGF0YQ==');

@$core.Deprecated('Use terminalExitedDescriptor instead')
const TerminalExited$json = {
  '1': 'TerminalExited',
  '2': [
    {
      '1': 'exit_code',
      '3': 1,
      '4': 1,
      '5': 5,
      '9': 0,
      '10': 'exitCode',
      '17': true
    },
    {'1': 'signal', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'signal'},
  ],
  '8': [
    {'1': '_exit_code'},
  ],
};

/// Descriptor for `TerminalExited`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalExitedDescriptor = $convert.base64Decode(
    'Cg5UZXJtaW5hbEV4aXRlZBIgCglleGl0X2NvZGUYASABKAVIAFIIZXhpdENvZGWIAQESHwoGc2'
    'lnbmFsGAIgASgJQge6SARyAhhAUgZzaWduYWxCDAoKX2V4aXRfY29kZQ==');

@$core.Deprecated('Use terminalErrorDescriptor instead')
const TerminalError$json = {
  '1': 'TerminalError',
  '2': [
    {'1': 'code', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'code'},
    {'1': 'message', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'message'},
  ],
};

/// Descriptor for `TerminalError`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalErrorDescriptor = $convert.base64Decode(
    'Cg1UZXJtaW5hbEVycm9yEh0KBGNvZGUYASABKAlCCbpIBnIEEAEYQFIEY29kZRIkCgdtZXNzYW'
    'dlGAIgASgJQgq6SAdyBRABGIAEUgdtZXNzYWdl');

@$core.Deprecated('Use terminalPongDescriptor instead')
const TerminalPong$json = {
  '1': 'TerminalPong',
  '2': [
    {'1': 'nonce', '3': 1, '4': 1, '5': 4, '10': 'nonce'},
  ],
};

/// Descriptor for `TerminalPong`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalPongDescriptor =
    $convert.base64Decode('CgxUZXJtaW5hbFBvbmcSFAoFbm9uY2UYASABKARSBW5vbmNl');

@$core.Deprecated('Use terminalServerFrameDescriptor instead')
const TerminalServerFrame$json = {
  '1': 'TerminalServerFrame',
  '2': [
    {
      '1': 'attached',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalAttached',
      '9': 0,
      '10': 'attached'
    },
    {
      '1': 'output',
      '3': 2,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalOutput',
      '9': 0,
      '10': 'output'
    },
    {
      '1': 'exited',
      '3': 3,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalExited',
      '9': 0,
      '10': 'exited'
    },
    {
      '1': 'error',
      '3': 4,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalError',
      '9': 0,
      '10': 'error'
    },
    {
      '1': 'pong',
      '3': 5,
      '4': 1,
      '5': 11,
      '6': '.super_manager.terminal.v1.TerminalPong',
      '9': 0,
      '10': 'pong'
    },
  ],
  '8': [
    {'1': 'payload'},
  ],
};

/// Descriptor for `TerminalServerFrame`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalServerFrameDescriptor = $convert.base64Decode(
    'ChNUZXJtaW5hbFNlcnZlckZyYW1lEkkKCGF0dGFjaGVkGAEgASgLMisuc3VwZXJfbWFuYWdlci'
    '50ZXJtaW5hbC52MS5UZXJtaW5hbEF0dGFjaGVkSABSCGF0dGFjaGVkEkMKBm91dHB1dBgCIAEo'
    'CzIpLnN1cGVyX21hbmFnZXIudGVybWluYWwudjEuVGVybWluYWxPdXRwdXRIAFIGb3V0cHV0Ek'
    'MKBmV4aXRlZBgDIAEoCzIpLnN1cGVyX21hbmFnZXIudGVybWluYWwudjEuVGVybWluYWxFeGl0'
    'ZWRIAFIGZXhpdGVkEkAKBWVycm9yGAQgASgLMiguc3VwZXJfbWFuYWdlci50ZXJtaW5hbC52MS'
    '5UZXJtaW5hbEVycm9ySABSBWVycm9yEj0KBHBvbmcYBSABKAsyJy5zdXBlcl9tYW5hZ2VyLnRl'
    'cm1pbmFsLnYxLlRlcm1pbmFsUG9uZ0gAUgRwb25nQgkKB3BheWxvYWQ=');
