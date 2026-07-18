// This is a generated file - do not edit.
//
// Generated from realtime/me/status/v1/status_types.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:protobuf/protobuf.dart' as $pb;

/// DeviceKind is the broad category of a reporting device.
class DeviceKind extends $pb.ProtobufEnum {
  /// Kind is not known.
  static const DeviceKind DEVICE_KIND_UNSPECIFIED =
      DeviceKind._(0, _omitEnumNames ? '' : 'DEVICE_KIND_UNSPECIFIED');

  /// A physical host machine.
  static const DeviceKind DEVICE_KIND_HOST =
      DeviceKind._(1, _omitEnumNames ? '' : 'DEVICE_KIND_HOST');

  /// A virtual machine running on a host.
  static const DeviceKind DEVICE_KIND_VIRTUAL_MACHINE =
      DeviceKind._(2, _omitEnumNames ? '' : 'DEVICE_KIND_VIRTUAL_MACHINE');

  /// A phone companion.
  static const DeviceKind DEVICE_KIND_PHONE =
      DeviceKind._(3, _omitEnumNames ? '' : 'DEVICE_KIND_PHONE');

  /// A watch paired to a phone.
  static const DeviceKind DEVICE_KIND_WATCH =
      DeviceKind._(4, _omitEnumNames ? '' : 'DEVICE_KIND_WATCH');

  static const $core.List<DeviceKind> values = <DeviceKind>[
    DEVICE_KIND_UNSPECIFIED,
    DEVICE_KIND_HOST,
    DEVICE_KIND_VIRTUAL_MACHINE,
    DEVICE_KIND_PHONE,
    DEVICE_KIND_WATCH,
  ];

  static final $core.List<DeviceKind?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 4);
  static DeviceKind? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DeviceKind._(super.value, super.name);
}

/// DeviceRole is the operational role a device plays on the status surface.
class DeviceRole extends $pb.ProtobufEnum {
  /// Role is not known.
  static const DeviceRole DEVICE_ROLE_UNSPECIFIED =
      DeviceRole._(0, _omitEnumNames ? '' : 'DEVICE_ROLE_UNSPECIFIED');

  /// The always-on server that also hosts the gateway.
  static const DeviceRole DEVICE_ROLE_SERVER =
      DeviceRole._(1, _omitEnumNames ? '' : 'DEVICE_ROLE_SERVER');

  /// A personal desktop or laptop.
  static const DeviceRole DEVICE_ROLE_DESKTOP =
      DeviceRole._(2, _omitEnumNames ? '' : 'DEVICE_ROLE_DESKTOP');

  /// A virtual machine guest.
  static const DeviceRole DEVICE_ROLE_VM =
      DeviceRole._(3, _omitEnumNames ? '' : 'DEVICE_ROLE_VM');

  static const $core.List<DeviceRole> values = <DeviceRole>[
    DEVICE_ROLE_UNSPECIFIED,
    DEVICE_ROLE_SERVER,
    DEVICE_ROLE_DESKTOP,
    DEVICE_ROLE_VM,
  ];

  static final $core.List<DeviceRole?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static DeviceRole? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const DeviceRole._(super.value, super.name);
}

/// OnlineState is whether a device is currently reachable.
class OnlineState extends $pb.ProtobufEnum {
  /// Reachability is not known.
  static const OnlineState ONLINE_STATE_UNSPECIFIED =
      OnlineState._(0, _omitEnumNames ? '' : 'ONLINE_STATE_UNSPECIFIED');

  /// The device is reachable.
  static const OnlineState ONLINE_STATE_ONLINE =
      OnlineState._(1, _omitEnumNames ? '' : 'ONLINE_STATE_ONLINE');

  /// The device is not reachable.
  static const OnlineState ONLINE_STATE_OFFLINE =
      OnlineState._(2, _omitEnumNames ? '' : 'ONLINE_STATE_OFFLINE');

  static const $core.List<OnlineState> values = <OnlineState>[
    ONLINE_STATE_UNSPECIFIED,
    ONLINE_STATE_ONLINE,
    ONLINE_STATE_OFFLINE,
  ];

  static final $core.List<OnlineState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 2);
  static OnlineState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const OnlineState._(super.value, super.name);
}

/// NetworkState is a phone's current network transport.
class NetworkState extends $pb.ProtobufEnum {
  /// Transport is not known.
  static const NetworkState NETWORK_STATE_UNSPECIFIED =
      NetworkState._(0, _omitEnumNames ? '' : 'NETWORK_STATE_UNSPECIFIED');

  /// No network is available.
  static const NetworkState NETWORK_STATE_OFFLINE =
      NetworkState._(1, _omitEnumNames ? '' : 'NETWORK_STATE_OFFLINE');

  /// Connected over Wi-Fi.
  static const NetworkState NETWORK_STATE_WIFI =
      NetworkState._(2, _omitEnumNames ? '' : 'NETWORK_STATE_WIFI');

  /// Connected over a cellular network.
  static const NetworkState NETWORK_STATE_CELLULAR =
      NetworkState._(3, _omitEnumNames ? '' : 'NETWORK_STATE_CELLULAR');

  /// Connected through a VPN.
  static const NetworkState NETWORK_STATE_VPN =
      NetworkState._(4, _omitEnumNames ? '' : 'NETWORK_STATE_VPN');

  /// Online over an unspecified transport.
  static const NetworkState NETWORK_STATE_ONLINE =
      NetworkState._(5, _omitEnumNames ? '' : 'NETWORK_STATE_ONLINE');

  static const $core.List<NetworkState> values = <NetworkState>[
    NETWORK_STATE_UNSPECIFIED,
    NETWORK_STATE_OFFLINE,
    NETWORK_STATE_WIFI,
    NETWORK_STATE_CELLULAR,
    NETWORK_STATE_VPN,
    NETWORK_STATE_ONLINE,
  ];

  static final $core.List<NetworkState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 5);
  static NetworkState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const NetworkState._(super.value, super.name);
}

/// AgentState is the run state of a coding agent.
class AgentState extends $pb.ProtobufEnum {
  /// State is not known.
  static const AgentState AGENT_STATE_UNSPECIFIED =
      AgentState._(0, _omitEnumNames ? '' : 'AGENT_STATE_UNSPECIFIED');

  /// The agent is idle.
  static const AgentState AGENT_STATE_IDLE =
      AgentState._(1, _omitEnumNames ? '' : 'AGENT_STATE_IDLE');

  /// The agent is actively running a task.
  static const AgentState AGENT_STATE_RUNNING =
      AgentState._(2, _omitEnumNames ? '' : 'AGENT_STATE_RUNNING');

  /// The agent's last task failed.
  static const AgentState AGENT_STATE_FAILED =
      AgentState._(3, _omitEnumNames ? '' : 'AGENT_STATE_FAILED');

  static const $core.List<AgentState> values = <AgentState>[
    AGENT_STATE_UNSPECIFIED,
    AGENT_STATE_IDLE,
    AGENT_STATE_RUNNING,
    AGENT_STATE_FAILED,
  ];

  static final $core.List<AgentState?> _byValue =
      $pb.ProtobufEnum.$_initByValueList(values, 3);
  static AgentState? valueOf($core.int value) =>
      value < 0 || value >= _byValue.length ? null : _byValue[value];

  const AgentState._(super.value, super.name);
}

const $core.bool _omitEnumNames =
    $core.bool.fromEnvironment('protobuf.omit_enum_names');
