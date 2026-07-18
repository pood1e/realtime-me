// This is a generated file - do not edit.
//
// Generated from realtime/me/manager/control/v1/terminal.proto.

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

@$core.Deprecated('Use terminalSessionStateDescriptor instead')
const TerminalSessionState$json = {
  '1': 'TerminalSessionState',
  '2': [
    {'1': 'TERMINAL_SESSION_STATE_UNSPECIFIED', '2': 0},
    {'1': 'TERMINAL_SESSION_STATE_RUNNING', '2': 1},
    {'1': 'TERMINAL_SESSION_STATE_CLOSED', '2': 2},
  ],
};

/// Descriptor for `TerminalSessionState`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List terminalSessionStateDescriptor = $convert.base64Decode(
    'ChRUZXJtaW5hbFNlc3Npb25TdGF0ZRImCiJURVJNSU5BTF9TRVNTSU9OX1NUQVRFX1VOU1BFQ0'
    'lGSUVEEAASIgoeVEVSTUlOQUxfU0VTU0lPTl9TVEFURV9SVU5OSU5HEAESIQodVEVSTUlOQUxf'
    'U0VTU0lPTl9TVEFURV9DTE9TRUQQAg==');

@$core.Deprecated('Use terminalSessionDescriptor instead')
const TerminalSession$json = {
  '1': 'TerminalSession',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
    {
      '1': 'workspace_uid',
      '3': 2,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'workspaceUid'
    },
    {'1': 'display_name', '3': 3, '4': 1, '5': 9, '8': {}, '10': 'displayName'},
    {'1': 'cwd', '3': 4, '4': 1, '5': 9, '8': {}, '10': 'cwd'},
    {
      '1': 'state',
      '3': 5,
      '4': 1,
      '5': 14,
      '6': '.realtime.me.manager.control.v1.TerminalSessionState',
      '8': {},
      '10': 'state'
    },
    {
      '1': 'create_time',
      '3': 6,
      '4': 1,
      '5': 11,
      '6': '.google.protobuf.Timestamp',
      '10': 'createTime'
    },
  ],
};

/// Descriptor for `TerminalSession`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List terminalSessionDescriptor = $convert.base64Decode(
    'Cg9UZXJtaW5hbFNlc3Npb24SGgoDdWlkGAEgASgJQgi6SAVyA7ABAVIDdWlkEi0KDXdvcmtzcG'
    'FjZV91aWQYAiABKAlCCLpIBXIDsAEBUgx3b3Jrc3BhY2VVaWQSLQoMZGlzcGxheV9uYW1lGAMg'
    'ASgJQgq6SAdyBRABGIABUgtkaXNwbGF5TmFtZRIgCgNjd2QYBCABKAlCDrpIC3IJEAEYgCAyAl'
    '4vUgNjd2QSVAoFc3RhdGUYBSABKA4yNC5yZWFsdGltZS5tZS5tYW5hZ2VyLmNvbnRyb2wudjEu'
    'VGVybWluYWxTZXNzaW9uU3RhdGVCCLpIBYIBAhABUgVzdGF0ZRI7CgtjcmVhdGVfdGltZRgGIA'
    'EoCzIaLmdvb2dsZS5wcm90b2J1Zi5UaW1lc3RhbXBSCmNyZWF0ZVRpbWU=');

@$core.Deprecated('Use createTerminalSessionRequestDescriptor instead')
const CreateTerminalSessionRequest$json = {
  '1': 'CreateTerminalSessionRequest',
  '2': [
    {
      '1': 'workspace_uid',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'workspaceUid'
    },
    {'1': 'display_name', '3': 2, '4': 1, '5': 9, '8': {}, '10': 'displayName'},
    {'1': 'columns', '3': 3, '4': 1, '5': 13, '8': {}, '10': 'columns'},
    {'1': 'rows', '3': 4, '4': 1, '5': 13, '8': {}, '10': 'rows'},
  ],
};

/// Descriptor for `CreateTerminalSessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTerminalSessionRequestDescriptor = $convert.base64Decode(
    'ChxDcmVhdGVUZXJtaW5hbFNlc3Npb25SZXF1ZXN0Ei0KDXdvcmtzcGFjZV91aWQYASABKAlCCL'
    'pIBXIDsAEBUgx3b3Jrc3BhY2VVaWQSLQoMZGlzcGxheV9uYW1lGAIgASgJQgq6SAdyBRABGIAB'
    'UgtkaXNwbGF5TmFtZRIkCgdjb2x1bW5zGAMgASgNQgq6SAcqBRjoBygUUgdjb2x1bW5zEh4KBH'
    'Jvd3MYBCABKA1CCrpIByoFGOgHKAVSBHJvd3M=');

@$core.Deprecated('Use createTerminalSessionResponseDescriptor instead')
const CreateTerminalSessionResponse$json = {
  '1': 'CreateTerminalSessionResponse',
  '2': [
    {
      '1': 'terminal_session',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.manager.control.v1.TerminalSession',
      '8': {},
      '10': 'terminalSession'
    },
  ],
};

/// Descriptor for `CreateTerminalSessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createTerminalSessionResponseDescriptor =
    $convert.base64Decode(
        'Ch1DcmVhdGVUZXJtaW5hbFNlc3Npb25SZXNwb25zZRJiChB0ZXJtaW5hbF9zZXNzaW9uGAEgAS'
        'gLMi8ucmVhbHRpbWUubWUubWFuYWdlci5jb250cm9sLnYxLlRlcm1pbmFsU2Vzc2lvbkIGukgD'
        'yAEBUg90ZXJtaW5hbFNlc3Npb24=');

@$core.Deprecated('Use getTerminalSessionRequestDescriptor instead')
const GetTerminalSessionRequest$json = {
  '1': 'GetTerminalSessionRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
  ],
};

/// Descriptor for `GetTerminalSessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTerminalSessionRequestDescriptor =
    $convert.base64Decode(
        'ChlHZXRUZXJtaW5hbFNlc3Npb25SZXF1ZXN0EhoKA3VpZBgBIAEoCUIIukgFcgOwAQFSA3VpZA'
        '==');

@$core.Deprecated('Use getTerminalSessionResponseDescriptor instead')
const GetTerminalSessionResponse$json = {
  '1': 'GetTerminalSessionResponse',
  '2': [
    {
      '1': 'terminal_session',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.manager.control.v1.TerminalSession',
      '8': {},
      '10': 'terminalSession'
    },
  ],
};

/// Descriptor for `GetTerminalSessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getTerminalSessionResponseDescriptor =
    $convert.base64Decode(
        'ChpHZXRUZXJtaW5hbFNlc3Npb25SZXNwb25zZRJiChB0ZXJtaW5hbF9zZXNzaW9uGAEgASgLMi'
        '8ucmVhbHRpbWUubWUubWFuYWdlci5jb250cm9sLnYxLlRlcm1pbmFsU2Vzc2lvbkIGukgDyAEB'
        'Ug90ZXJtaW5hbFNlc3Npb24=');

@$core.Deprecated('Use listTerminalSessionsRequestDescriptor instead')
const ListTerminalSessionsRequest$json = {
  '1': 'ListTerminalSessionsRequest',
  '2': [
    {
      '1': 'workspace_uid',
      '3': 1,
      '4': 1,
      '5': 9,
      '8': {},
      '10': 'workspaceUid'
    },
  ],
};

/// Descriptor for `ListTerminalSessionsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTerminalSessionsRequestDescriptor =
    $convert.base64Decode(
        'ChtMaXN0VGVybWluYWxTZXNzaW9uc1JlcXVlc3QSLQoNd29ya3NwYWNlX3VpZBgBIAEoCUIIuk'
        'gFcgOwAQFSDHdvcmtzcGFjZVVpZA==');

@$core.Deprecated('Use listTerminalSessionsResponseDescriptor instead')
const ListTerminalSessionsResponse$json = {
  '1': 'ListTerminalSessionsResponse',
  '2': [
    {
      '1': 'terminal_sessions',
      '3': 1,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.manager.control.v1.TerminalSession',
      '10': 'terminalSessions'
    },
  ],
};

/// Descriptor for `ListTerminalSessionsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List listTerminalSessionsResponseDescriptor =
    $convert.base64Decode(
        'ChxMaXN0VGVybWluYWxTZXNzaW9uc1Jlc3BvbnNlElwKEXRlcm1pbmFsX3Nlc3Npb25zGAEgAy'
        'gLMi8ucmVhbHRpbWUubWUubWFuYWdlci5jb250cm9sLnYxLlRlcm1pbmFsU2Vzc2lvblIQdGVy'
        'bWluYWxTZXNzaW9ucw==');

@$core.Deprecated('Use deleteTerminalSessionRequestDescriptor instead')
const DeleteTerminalSessionRequest$json = {
  '1': 'DeleteTerminalSessionRequest',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '8': {}, '10': 'uid'},
  ],
};

/// Descriptor for `DeleteTerminalSessionRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteTerminalSessionRequestDescriptor =
    $convert.base64Decode(
        'ChxEZWxldGVUZXJtaW5hbFNlc3Npb25SZXF1ZXN0EhoKA3VpZBgBIAEoCUIIukgFcgOwAQFSA3'
        'VpZA==');

@$core.Deprecated('Use deleteTerminalSessionResponseDescriptor instead')
const DeleteTerminalSessionResponse$json = {
  '1': 'DeleteTerminalSessionResponse',
};

/// Descriptor for `DeleteTerminalSessionResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List deleteTerminalSessionResponseDescriptor =
    $convert.base64Decode('Ch1EZWxldGVUZXJtaW5hbFNlc3Npb25SZXNwb25zZQ==');

const $core.Map<$core.String, $core.dynamic> TerminalServiceBase$json = {
  '1': 'TerminalService',
  '2': [
    {
      '1': 'CreateTerminalSession',
      '2': '.realtime.me.manager.control.v1.CreateTerminalSessionRequest',
      '3': '.realtime.me.manager.control.v1.CreateTerminalSessionResponse'
    },
    {
      '1': 'GetTerminalSession',
      '2': '.realtime.me.manager.control.v1.GetTerminalSessionRequest',
      '3': '.realtime.me.manager.control.v1.GetTerminalSessionResponse'
    },
    {
      '1': 'ListTerminalSessions',
      '2': '.realtime.me.manager.control.v1.ListTerminalSessionsRequest',
      '3': '.realtime.me.manager.control.v1.ListTerminalSessionsResponse'
    },
    {
      '1': 'DeleteTerminalSession',
      '2': '.realtime.me.manager.control.v1.DeleteTerminalSessionRequest',
      '3': '.realtime.me.manager.control.v1.DeleteTerminalSessionResponse'
    },
  ],
};

@$core.Deprecated('Use terminalServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    TerminalServiceBase$messageJson = {
  '.realtime.me.manager.control.v1.CreateTerminalSessionRequest':
      CreateTerminalSessionRequest$json,
  '.realtime.me.manager.control.v1.CreateTerminalSessionResponse':
      CreateTerminalSessionResponse$json,
  '.realtime.me.manager.control.v1.TerminalSession': TerminalSession$json,
  '.google.protobuf.Timestamp': $0.Timestamp$json,
  '.realtime.me.manager.control.v1.GetTerminalSessionRequest':
      GetTerminalSessionRequest$json,
  '.realtime.me.manager.control.v1.GetTerminalSessionResponse':
      GetTerminalSessionResponse$json,
  '.realtime.me.manager.control.v1.ListTerminalSessionsRequest':
      ListTerminalSessionsRequest$json,
  '.realtime.me.manager.control.v1.ListTerminalSessionsResponse':
      ListTerminalSessionsResponse$json,
  '.realtime.me.manager.control.v1.DeleteTerminalSessionRequest':
      DeleteTerminalSessionRequest$json,
  '.realtime.me.manager.control.v1.DeleteTerminalSessionResponse':
      DeleteTerminalSessionResponse$json,
};

/// Descriptor for `TerminalService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List terminalServiceDescriptor = $convert.base64Decode(
    'Cg9UZXJtaW5hbFNlcnZpY2USlAEKFUNyZWF0ZVRlcm1pbmFsU2Vzc2lvbhI8LnJlYWx0aW1lLm'
    '1lLm1hbmFnZXIuY29udHJvbC52MS5DcmVhdGVUZXJtaW5hbFNlc3Npb25SZXF1ZXN0Gj0ucmVh'
    'bHRpbWUubWUubWFuYWdlci5jb250cm9sLnYxLkNyZWF0ZVRlcm1pbmFsU2Vzc2lvblJlc3Bvbn'
    'NlEosBChJHZXRUZXJtaW5hbFNlc3Npb24SOS5yZWFsdGltZS5tZS5tYW5hZ2VyLmNvbnRyb2wu'
    'djEuR2V0VGVybWluYWxTZXNzaW9uUmVxdWVzdBo6LnJlYWx0aW1lLm1lLm1hbmFnZXIuY29udH'
    'JvbC52MS5HZXRUZXJtaW5hbFNlc3Npb25SZXNwb25zZRKRAQoUTGlzdFRlcm1pbmFsU2Vzc2lv'
    'bnMSOy5yZWFsdGltZS5tZS5tYW5hZ2VyLmNvbnRyb2wudjEuTGlzdFRlcm1pbmFsU2Vzc2lvbn'
    'NSZXF1ZXN0GjwucmVhbHRpbWUubWUubWFuYWdlci5jb250cm9sLnYxLkxpc3RUZXJtaW5hbFNl'
    'c3Npb25zUmVzcG9uc2USlAEKFURlbGV0ZVRlcm1pbmFsU2Vzc2lvbhI8LnJlYWx0aW1lLm1lLm'
    '1hbmFnZXIuY29udHJvbC52MS5EZWxldGVUZXJtaW5hbFNlc3Npb25SZXF1ZXN0Gj0ucmVhbHRp'
    'bWUubWUubWFuYWdlci5jb250cm9sLnYxLkRlbGV0ZVRlcm1pbmFsU2Vzc2lvblJlc3BvbnNl');
