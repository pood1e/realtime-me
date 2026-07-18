// This is a generated file - do not edit.
//
// Generated from realtime/me/site/v1/profile.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use getProfileRequestDescriptor instead')
const GetProfileRequest$json = {
  '1': 'GetProfileRequest',
};

/// Descriptor for `GetProfileRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getProfileRequestDescriptor =
    $convert.base64Decode('ChFHZXRQcm9maWxlUmVxdWVzdA==');

@$core.Deprecated('Use getProfileResponseDescriptor instead')
const GetProfileResponse$json = {
  '1': 'GetProfileResponse',
  '2': [
    {
      '1': 'profile',
      '3': 1,
      '4': 1,
      '5': 11,
      '6': '.realtime.me.site.v1.Profile',
      '10': 'profile'
    },
  ],
};

/// Descriptor for `GetProfileResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getProfileResponseDescriptor = $convert.base64Decode(
    'ChJHZXRQcm9maWxlUmVzcG9uc2USNgoHcHJvZmlsZRgBIAEoCzIcLnJlYWx0aW1lLm1lLnNpdG'
    'UudjEuUHJvZmlsZVIHcHJvZmlsZQ==');

@$core.Deprecated('Use profileDescriptor instead')
const Profile$json = {
  '1': 'Profile',
  '2': [
    {'1': 'display_name', '3': 1, '4': 1, '5': 9, '10': 'displayName'},
    {'1': 'avatar_url', '3': 4, '4': 1, '5': 9, '10': 'avatarUrl'},
    {'1': 'github_login', '3': 6, '4': 1, '5': 9, '10': 'githubLogin'},
    {
      '1': 'links',
      '3': 7,
      '4': 3,
      '5': 11,
      '6': '.realtime.me.site.v1.ProfileLink',
      '10': 'links'
    },
  ],
  '9': [
    {'1': 2, '2': 3},
    {'1': 3, '2': 4},
    {'1': 5, '2': 6},
  ],
  '10': ['headline', 'bio', 'location'],
};

/// Descriptor for `Profile`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileDescriptor = $convert.base64Decode(
    'CgdQcm9maWxlEiEKDGRpc3BsYXlfbmFtZRgBIAEoCVILZGlzcGxheU5hbWUSHQoKYXZhdGFyX3'
    'VybBgEIAEoCVIJYXZhdGFyVXJsEiEKDGdpdGh1Yl9sb2dpbhgGIAEoCVILZ2l0aHViTG9naW4S'
    'NgoFbGlua3MYByADKAsyIC5yZWFsdGltZS5tZS5zaXRlLnYxLlByb2ZpbGVMaW5rUgVsaW5rc0'
    'oECAIQA0oECAMQBEoECAUQBlIIaGVhZGxpbmVSA2Jpb1IIbG9jYXRpb24=');

@$core.Deprecated('Use profileLinkDescriptor instead')
const ProfileLink$json = {
  '1': 'ProfileLink',
  '2': [
    {'1': 'label', '3': 1, '4': 1, '5': 9, '10': 'label'},
    {'1': 'uri', '3': 2, '4': 1, '5': 9, '10': 'uri'},
    {'1': 'platform', '3': 3, '4': 1, '5': 9, '10': 'platform'},
  ],
};

/// Descriptor for `ProfileLink`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List profileLinkDescriptor = $convert.base64Decode(
    'CgtQcm9maWxlTGluaxIUCgVsYWJlbBgBIAEoCVIFbGFiZWwSEAoDdXJpGAIgASgJUgN1cmkSGg'
    'oIcGxhdGZvcm0YAyABKAlSCHBsYXRmb3Jt');

const $core.Map<$core.String, $core.dynamic> ProfileServiceBase$json = {
  '1': 'ProfileService',
  '2': [
    {
      '1': 'GetProfile',
      '2': '.realtime.me.site.v1.GetProfileRequest',
      '3': '.realtime.me.site.v1.GetProfileResponse'
    },
  ],
};

@$core.Deprecated('Use profileServiceDescriptor instead')
const $core.Map<$core.String, $core.Map<$core.String, $core.dynamic>>
    ProfileServiceBase$messageJson = {
  '.realtime.me.site.v1.GetProfileRequest': GetProfileRequest$json,
  '.realtime.me.site.v1.GetProfileResponse': GetProfileResponse$json,
  '.realtime.me.site.v1.Profile': Profile$json,
  '.realtime.me.site.v1.ProfileLink': ProfileLink$json,
};

/// Descriptor for `ProfileService`. Decode as a `google.protobuf.ServiceDescriptorProto`.
final $typed_data.Uint8List profileServiceDescriptor = $convert.base64Decode(
    'Cg5Qcm9maWxlU2VydmljZRJdCgpHZXRQcm9maWxlEiYucmVhbHRpbWUubWUuc2l0ZS52MS5HZX'
    'RQcm9maWxlUmVxdWVzdBonLnJlYWx0aW1lLm1lLnNpdGUudjEuR2V0UHJvZmlsZVJlc3BvbnNl');
