// This is a generated file - do not edit.
//
// Generated from super_manager/control/v1/device.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_relative_imports

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;
import 'package:protobuf/well_known_types/google/protobuf/timestamp.pb.dart'
    as $0;

import 'device.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'device.pbenum.dart';

/// Device is one paired Flutter client.
class Device extends $pb.GeneratedMessage {
  factory Device({
    $core.String? uid,
    $core.String? displayName,
    DeviceStatus? status,
    $core.String? certificateSerial,
    $0.Timestamp? createTime,
    $0.Timestamp? expireTime,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (displayName != null) result.displayName = displayName;
    if (status != null) result.status = status;
    if (certificateSerial != null) result.certificateSerial = certificateSerial;
    if (createTime != null) result.createTime = createTime;
    if (expireTime != null) result.expireTime = expireTime;
    return result;
  }

  Device._();

  factory Device.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Device.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Device',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..aE<DeviceStatus>(3, _omitFieldNames ? '' : 'status',
        enumValues: DeviceStatus.values)
    ..aOS(4, _omitFieldNames ? '' : 'certificateSerial')
    ..aOM<$0.Timestamp>(5, _omitFieldNames ? '' : 'createTime',
        subBuilder: $0.Timestamp.create)
    ..aOM<$0.Timestamp>(6, _omitFieldNames ? '' : 'expireTime',
        subBuilder: $0.Timestamp.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Device clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Device copyWith(void Function(Device) updates) =>
      super.copyWith((message) => updates(message as Device)) as Device;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Device create() => Device._();
  @$core.override
  Device createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Device getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Device>(create);
  static Device? _defaultInstance;

  /// uid is the server-assigned UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  /// display_name is the human-readable client label.
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);

  /// status controls whether the inner bearer token remains valid.
  @$pb.TagNumber(3)
  DeviceStatus get status => $_getN(2);
  @$pb.TagNumber(3)
  set status(DeviceStatus value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasStatus() => $_has(2);
  @$pb.TagNumber(3)
  void clearStatus() => $_clearField(3);

  /// certificate_serial is the non-secret X.509 serial used for audit correlation.
  @$pb.TagNumber(4)
  $core.String get certificateSerial => $_getSZ(3);
  @$pb.TagNumber(4)
  set certificateSerial($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCertificateSerial() => $_has(3);
  @$pb.TagNumber(4)
  void clearCertificateSerial() => $_clearField(4);

  /// create_time is the pairing time.
  @$pb.TagNumber(5)
  $0.Timestamp get createTime => $_getN(4);
  @$pb.TagNumber(5)
  set createTime($0.Timestamp value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasCreateTime() => $_has(4);
  @$pb.TagNumber(5)
  void clearCreateTime() => $_clearField(5);
  @$pb.TagNumber(5)
  $0.Timestamp ensureCreateTime() => $_ensure(4);

  /// expire_time is the credential expiry time.
  @$pb.TagNumber(6)
  $0.Timestamp get expireTime => $_getN(5);
  @$pb.TagNumber(6)
  set expireTime($0.Timestamp value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasExpireTime() => $_has(5);
  @$pb.TagNumber(6)
  void clearExpireTime() => $_clearField(6);
  @$pb.TagNumber(6)
  $0.Timestamp ensureExpireTime() => $_ensure(5);
}

/// PairDeviceRequest redeems one locally created, single-use pairing secret.
class PairDeviceRequest extends $pb.GeneratedMessage {
  factory PairDeviceRequest({
    $core.List<$core.int>? pairingSecret,
    $core.String? displayName,
  }) {
    final result = create();
    if (pairingSecret != null) result.pairingSecret = pairingSecret;
    if (displayName != null) result.displayName = displayName;
    return result;
  }

  PairDeviceRequest._();

  factory PairDeviceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PairDeviceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PairDeviceRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..a<$core.List<$core.int>>(
        1, _omitFieldNames ? '' : 'pairingSecret', $pb.PbFieldType.OY)
    ..aOS(2, _omitFieldNames ? '' : 'displayName')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PairDeviceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PairDeviceRequest copyWith(void Function(PairDeviceRequest) updates) =>
      super.copyWith((message) => updates(message as PairDeviceRequest))
          as PairDeviceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PairDeviceRequest create() => PairDeviceRequest._();
  @$core.override
  PairDeviceRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static PairDeviceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PairDeviceRequest>(create);
  static PairDeviceRequest? _defaultInstance;

  /// pairing_secret is the 32-byte secret encoded in the local QR payload.
  @$pb.TagNumber(1)
  $core.List<$core.int> get pairingSecret => $_getN(0);
  @$pb.TagNumber(1)
  set pairingSecret($core.List<$core.int> value) => $_setBytes(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPairingSecret() => $_has(0);
  @$pb.TagNumber(1)
  void clearPairingSecret() => $_clearField(1);

  /// display_name is the human-readable Android device label.
  @$pb.TagNumber(2)
  $core.String get displayName => $_getSZ(1);
  @$pb.TagNumber(2)
  set displayName($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDisplayName() => $_has(1);
  @$pb.TagNumber(2)
  void clearDisplayName() => $_clearField(2);
}

/// PairDeviceResponse returns credentials exactly once to the paired device.
class PairDeviceResponse extends $pb.GeneratedMessage {
  factory PairDeviceResponse({
    Device? device,
    $core.List<$core.int>? devicePkcs12,
    $core.String? pkcs12Password,
    $core.String? deviceToken,
    $core.List<$core.int>? caCertificatePem,
  }) {
    final result = create();
    if (device != null) result.device = device;
    if (devicePkcs12 != null) result.devicePkcs12 = devicePkcs12;
    if (pkcs12Password != null) result.pkcs12Password = pkcs12Password;
    if (deviceToken != null) result.deviceToken = deviceToken;
    if (caCertificatePem != null) result.caCertificatePem = caCertificatePem;
    return result;
  }

  PairDeviceResponse._();

  factory PairDeviceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PairDeviceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PairDeviceResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOM<Device>(1, _omitFieldNames ? '' : 'device', subBuilder: Device.create)
    ..a<$core.List<$core.int>>(
        2, _omitFieldNames ? '' : 'devicePkcs12', $pb.PbFieldType.OY)
    ..aOS(3, _omitFieldNames ? '' : 'pkcs12Password')
    ..aOS(4, _omitFieldNames ? '' : 'deviceToken')
    ..a<$core.List<$core.int>>(
        5, _omitFieldNames ? '' : 'caCertificatePem', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PairDeviceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PairDeviceResponse copyWith(void Function(PairDeviceResponse) updates) =>
      super.copyWith((message) => updates(message as PairDeviceResponse))
          as PairDeviceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PairDeviceResponse create() => PairDeviceResponse._();
  @$core.override
  PairDeviceResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static PairDeviceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PairDeviceResponse>(create);
  static PairDeviceResponse? _defaultInstance;

  /// device is the newly paired client record.
  @$pb.TagNumber(1)
  Device get device => $_getN(0);
  @$pb.TagNumber(1)
  set device(Device value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasDevice() => $_has(0);
  @$pb.TagNumber(1)
  void clearDevice() => $_clearField(1);
  @$pb.TagNumber(1)
  Device ensureDevice() => $_ensure(0);

  /// device_pkcs12 contains the device certificate and private key.
  @$pb.TagNumber(2)
  $core.List<$core.int> get devicePkcs12 => $_getN(1);
  @$pb.TagNumber(2)
  set devicePkcs12($core.List<$core.int> value) => $_setBytes(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDevicePkcs12() => $_has(1);
  @$pb.TagNumber(2)
  void clearDevicePkcs12() => $_clearField(2);

  /// pkcs12_password decrypts device_pkcs12 and must be stored in secure storage.
  @$pb.TagNumber(3)
  $core.String get pkcs12Password => $_getSZ(2);
  @$pb.TagNumber(3)
  set pkcs12Password($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPkcs12Password() => $_has(2);
  @$pb.TagNumber(3)
  void clearPkcs12Password() => $_clearField(3);

  /// device_token is the revocable inner bearer credential.
  @$pb.TagNumber(4)
  $core.String get deviceToken => $_getSZ(3);
  @$pb.TagNumber(4)
  set deviceToken($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasDeviceToken() => $_has(3);
  @$pb.TagNumber(4)
  void clearDeviceToken() => $_clearField(4);

  /// ca_certificate_pem is the private CA certificate used to validate the relay IP endpoint.
  @$pb.TagNumber(5)
  $core.List<$core.int> get caCertificatePem => $_getN(4);
  @$pb.TagNumber(5)
  set caCertificatePem($core.List<$core.int> value) => $_setBytes(4, value);
  @$pb.TagNumber(5)
  $core.bool hasCaCertificatePem() => $_has(4);
  @$pb.TagNumber(5)
  void clearCaCertificatePem() => $_clearField(5);
}

/// ListDevicesRequest requests paired devices.
class ListDevicesRequest extends $pb.GeneratedMessage {
  factory ListDevicesRequest() => create();

  ListDevicesRequest._();

  factory ListDevicesRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListDevicesRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListDevicesRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDevicesRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDevicesRequest copyWith(void Function(ListDevicesRequest) updates) =>
      super.copyWith((message) => updates(message as ListDevicesRequest))
          as ListDevicesRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListDevicesRequest create() => ListDevicesRequest._();
  @$core.override
  ListDevicesRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListDevicesRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListDevicesRequest>(create);
  static ListDevicesRequest? _defaultInstance;
}

/// ListDevicesResponse returns paired devices without credentials.
class ListDevicesResponse extends $pb.GeneratedMessage {
  factory ListDevicesResponse({
    $core.Iterable<Device>? devices,
  }) {
    final result = create();
    if (devices != null) result.devices.addAll(devices);
    return result;
  }

  ListDevicesResponse._();

  factory ListDevicesResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ListDevicesResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ListDevicesResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..pPM<Device>(1, _omitFieldNames ? '' : 'devices',
        subBuilder: Device.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDevicesResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ListDevicesResponse copyWith(void Function(ListDevicesResponse) updates) =>
      super.copyWith((message) => updates(message as ListDevicesResponse))
          as ListDevicesResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ListDevicesResponse create() => ListDevicesResponse._();
  @$core.override
  ListDevicesResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ListDevicesResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<ListDevicesResponse>(create);
  static ListDevicesResponse? _defaultInstance;

  /// devices contains all paired client records.
  @$pb.TagNumber(1)
  $pb.PbList<Device> get devices => $_getList(0);
}

/// DeleteDeviceRequest revokes one paired device.
class DeleteDeviceRequest extends $pb.GeneratedMessage {
  factory DeleteDeviceRequest({
    $core.String? uid,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    return result;
  }

  DeleteDeviceRequest._();

  factory DeleteDeviceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteDeviceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteDeviceRequest',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDeviceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDeviceRequest copyWith(void Function(DeleteDeviceRequest) updates) =>
      super.copyWith((message) => updates(message as DeleteDeviceRequest))
          as DeleteDeviceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteDeviceRequest create() => DeleteDeviceRequest._();
  @$core.override
  DeleteDeviceRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static DeleteDeviceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteDeviceRequest>(create);
  static DeleteDeviceRequest? _defaultInstance;

  /// uid is the device UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);
}

/// DeleteDeviceResponse confirms that the device was revoked.
class DeleteDeviceResponse extends $pb.GeneratedMessage {
  factory DeleteDeviceResponse() => create();

  DeleteDeviceResponse._();

  factory DeleteDeviceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory DeleteDeviceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'DeleteDeviceResponse',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.control.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDeviceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  DeleteDeviceResponse copyWith(void Function(DeleteDeviceResponse) updates) =>
      super.copyWith((message) => updates(message as DeleteDeviceResponse))
          as DeleteDeviceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static DeleteDeviceResponse create() => DeleteDeviceResponse._();
  @$core.override
  DeleteDeviceResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static DeleteDeviceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<DeleteDeviceResponse>(create);
  static DeleteDeviceResponse? _defaultInstance;
}

/// PairingService exposes the only API available without an existing device certificate.
class PairingServiceApi {
  final $pb.RpcClient _client;

  PairingServiceApi(this._client);

  /// PairDevice atomically redeems a locally generated one-time secret.
  $async.Future<PairDeviceResponse> pairDevice(
          $pb.ClientContext? ctx, PairDeviceRequest request) =>
      _client.invoke<PairDeviceResponse>(
          ctx, 'PairingService', 'PairDevice', request, PairDeviceResponse());
}

/// DeviceService manages paired device revocation.
class DeviceServiceApi {
  final $pb.RpcClient _client;

  DeviceServiceApi(this._client);

  /// ListDevices returns paired devices without credentials.
  $async.Future<ListDevicesResponse> listDevices(
          $pb.ClientContext? ctx, ListDevicesRequest request) =>
      _client.invoke<ListDevicesResponse>(
          ctx, 'DeviceService', 'ListDevices', request, ListDevicesResponse());

  /// DeleteDevice revokes a device immediately through the inner token layer.
  $async.Future<DeleteDeviceResponse> deleteDevice(
          $pb.ClientContext? ctx, DeleteDeviceRequest request) =>
      _client.invoke<DeleteDeviceResponse>(ctx, 'DeviceService', 'DeleteDevice',
          request, DeleteDeviceResponse());
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
