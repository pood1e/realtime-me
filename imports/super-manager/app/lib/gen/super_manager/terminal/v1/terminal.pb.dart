// This is a generated file - do not edit.
//
// Generated from super_manager/terminal/v1/terminal.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

/// TerminalAttach attaches the WebSocket to one registered terminal session.
class TerminalAttach extends $pb.GeneratedMessage {
  factory TerminalAttach({
    $core.String? terminalSessionUid,
  }) {
    final result = create();
    if (terminalSessionUid != null)
      result.terminalSessionUid = terminalSessionUid;
    return result;
  }

  TerminalAttach._();

  factory TerminalAttach.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalAttach.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalAttach',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'terminalSessionUid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalAttach clone() => TerminalAttach()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalAttach copyWith(void Function(TerminalAttach) updates) =>
      super.copyWith((message) => updates(message as TerminalAttach))
          as TerminalAttach;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalAttach create() => TerminalAttach._();
  @$core.override
  TerminalAttach createEmptyInstance() => create();
  static $pb.PbList<TerminalAttach> createRepeated() =>
      $pb.PbList<TerminalAttach>();
  @$core.pragma('dart2js:noInline')
  static TerminalAttach getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalAttach>(create);
  static TerminalAttach? _defaultInstance;

  /// terminal_session_uid is the terminal session UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get terminalSessionUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set terminalSessionUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTerminalSessionUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearTerminalSessionUid() => $_clearField(1);
}

/// TerminalInput carries raw terminal input bytes.
class TerminalInput extends $pb.GeneratedMessage {
  factory TerminalInput({
    $core.List<$core.int>? data,
  }) {
    final result = create();
    if (data != null) result.data = data;
    return result;
  }

  TerminalInput._();

  factory TerminalInput.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalInput.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalInput',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..a<$core.List<$core.int>>(
        1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalInput clone() => TerminalInput()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalInput copyWith(void Function(TerminalInput) updates) =>
      super.copyWith((message) => updates(message as TerminalInput))
          as TerminalInput;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalInput create() => TerminalInput._();
  @$core.override
  TerminalInput createEmptyInstance() => create();
  static $pb.PbList<TerminalInput> createRepeated() =>
      $pb.PbList<TerminalInput>();
  @$core.pragma('dart2js:noInline')
  static TerminalInput getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalInput>(create);
  static TerminalInput? _defaultInstance;

  /// data is raw user input and is never written to the application audit log.
  @$pb.TagNumber(1)
  $core.List<$core.int> get data => $_getN(0);
  @$pb.TagNumber(1)
  set data($core.List<$core.int> value) => $_setBytes(0, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
}

/// TerminalResize changes the attached PTY dimensions.
class TerminalResize extends $pb.GeneratedMessage {
  factory TerminalResize({
    $core.int? columns,
    $core.int? rows,
  }) {
    final result = create();
    if (columns != null) result.columns = columns;
    if (rows != null) result.rows = rows;
    return result;
  }

  TerminalResize._();

  factory TerminalResize.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalResize.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalResize',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'columns', $pb.PbFieldType.OU3)
    ..a<$core.int>(2, _omitFieldNames ? '' : 'rows', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalResize clone() => TerminalResize()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalResize copyWith(void Function(TerminalResize) updates) =>
      super.copyWith((message) => updates(message as TerminalResize))
          as TerminalResize;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalResize create() => TerminalResize._();
  @$core.override
  TerminalResize createEmptyInstance() => create();
  static $pb.PbList<TerminalResize> createRepeated() =>
      $pb.PbList<TerminalResize>();
  @$core.pragma('dart2js:noInline')
  static TerminalResize getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalResize>(create);
  static TerminalResize? _defaultInstance;

  /// columns is the terminal width.
  @$pb.TagNumber(1)
  $core.int get columns => $_getIZ(0);
  @$pb.TagNumber(1)
  set columns($core.int value) => $_setUnsignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasColumns() => $_has(0);
  @$pb.TagNumber(1)
  void clearColumns() => $_clearField(1);

  /// rows is the terminal height.
  @$pb.TagNumber(2)
  $core.int get rows => $_getIZ(1);
  @$pb.TagNumber(2)
  set rows($core.int value) => $_setUnsignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRows() => $_has(1);
  @$pb.TagNumber(2)
  void clearRows() => $_clearField(2);
}

/// TerminalDetach detaches this WebSocket without stopping the shell.
class TerminalDetach extends $pb.GeneratedMessage {
  factory TerminalDetach() => create();

  TerminalDetach._();

  factory TerminalDetach.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalDetach.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalDetach',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalDetach clone() => TerminalDetach()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalDetach copyWith(void Function(TerminalDetach) updates) =>
      super.copyWith((message) => updates(message as TerminalDetach))
          as TerminalDetach;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalDetach create() => TerminalDetach._();
  @$core.override
  TerminalDetach createEmptyInstance() => create();
  static $pb.PbList<TerminalDetach> createRepeated() =>
      $pb.PbList<TerminalDetach>();
  @$core.pragma('dart2js:noInline')
  static TerminalDetach getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalDetach>(create);
  static TerminalDetach? _defaultInstance;
}

/// TerminalClose stops the tmux-backed shell.
class TerminalClose extends $pb.GeneratedMessage {
  factory TerminalClose() => create();

  TerminalClose._();

  factory TerminalClose.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalClose.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalClose',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalClose clone() => TerminalClose()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalClose copyWith(void Function(TerminalClose) updates) =>
      super.copyWith((message) => updates(message as TerminalClose))
          as TerminalClose;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalClose create() => TerminalClose._();
  @$core.override
  TerminalClose createEmptyInstance() => create();
  static $pb.PbList<TerminalClose> createRepeated() =>
      $pb.PbList<TerminalClose>();
  @$core.pragma('dart2js:noInline')
  static TerminalClose getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalClose>(create);
  static TerminalClose? _defaultInstance;
}

/// TerminalPing checks liveness without changing shell state.
class TerminalPing extends $pb.GeneratedMessage {
  factory TerminalPing({
    $fixnum.Int64? nonce,
  }) {
    final result = create();
    if (nonce != null) result.nonce = nonce;
    return result;
  }

  TerminalPing._();

  factory TerminalPing.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalPing.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalPing',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, _omitFieldNames ? '' : 'nonce', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalPing clone() => TerminalPing()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalPing copyWith(void Function(TerminalPing) updates) =>
      super.copyWith((message) => updates(message as TerminalPing))
          as TerminalPing;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalPing create() => TerminalPing._();
  @$core.override
  TerminalPing createEmptyInstance() => create();
  static $pb.PbList<TerminalPing> createRepeated() =>
      $pb.PbList<TerminalPing>();
  @$core.pragma('dart2js:noInline')
  static TerminalPing getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalPing>(create);
  static TerminalPing? _defaultInstance;

  /// nonce is echoed by TerminalPong.
  @$pb.TagNumber(1)
  $fixnum.Int64 get nonce => $_getI64(0);
  @$pb.TagNumber(1)
  set nonce($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasNonce() => $_has(0);
  @$pb.TagNumber(1)
  void clearNonce() => $_clearField(1);
}

enum TerminalClientFrame_Payload {
  attach,
  input,
  resize,
  detach,
  close,
  ping,
  notSet
}

/// TerminalClientFrame is one binary Protobuf frame sent by the Flutter client.
class TerminalClientFrame extends $pb.GeneratedMessage {
  factory TerminalClientFrame({
    TerminalAttach? attach,
    TerminalInput? input,
    TerminalResize? resize,
    TerminalDetach? detach,
    TerminalClose? close,
    TerminalPing? ping,
  }) {
    final result = create();
    if (attach != null) result.attach = attach;
    if (input != null) result.input = input;
    if (resize != null) result.resize = resize;
    if (detach != null) result.detach = detach;
    if (close != null) result.close = close;
    if (ping != null) result.ping = ping;
    return result;
  }

  TerminalClientFrame._();

  factory TerminalClientFrame.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalClientFrame.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, TerminalClientFrame_Payload>
      _TerminalClientFrame_PayloadByTag = {
    1: TerminalClientFrame_Payload.attach,
    2: TerminalClientFrame_Payload.input,
    3: TerminalClientFrame_Payload.resize,
    4: TerminalClientFrame_Payload.detach,
    5: TerminalClientFrame_Payload.close,
    6: TerminalClientFrame_Payload.ping,
    0: TerminalClientFrame_Payload.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalClientFrame',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..oo(0, [1, 2, 3, 4, 5, 6])
    ..aOM<TerminalAttach>(1, _omitFieldNames ? '' : 'attach',
        subBuilder: TerminalAttach.create)
    ..aOM<TerminalInput>(2, _omitFieldNames ? '' : 'input',
        subBuilder: TerminalInput.create)
    ..aOM<TerminalResize>(3, _omitFieldNames ? '' : 'resize',
        subBuilder: TerminalResize.create)
    ..aOM<TerminalDetach>(4, _omitFieldNames ? '' : 'detach',
        subBuilder: TerminalDetach.create)
    ..aOM<TerminalClose>(5, _omitFieldNames ? '' : 'close',
        subBuilder: TerminalClose.create)
    ..aOM<TerminalPing>(6, _omitFieldNames ? '' : 'ping',
        subBuilder: TerminalPing.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalClientFrame clone() => TerminalClientFrame()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalClientFrame copyWith(void Function(TerminalClientFrame) updates) =>
      super.copyWith((message) => updates(message as TerminalClientFrame))
          as TerminalClientFrame;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalClientFrame create() => TerminalClientFrame._();
  @$core.override
  TerminalClientFrame createEmptyInstance() => create();
  static $pb.PbList<TerminalClientFrame> createRepeated() =>
      $pb.PbList<TerminalClientFrame>();
  @$core.pragma('dart2js:noInline')
  static TerminalClientFrame getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalClientFrame>(create);
  static TerminalClientFrame? _defaultInstance;

  TerminalClientFrame_Payload whichPayload() =>
      _TerminalClientFrame_PayloadByTag[$_whichOneof(0)]!;
  void clearPayload() => $_clearField($_whichOneof(0));

  /// attach must be the first frame on a new WebSocket.
  @$pb.TagNumber(1)
  TerminalAttach get attach => $_getN(0);
  @$pb.TagNumber(1)
  set attach(TerminalAttach value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAttach() => $_has(0);
  @$pb.TagNumber(1)
  void clearAttach() => $_clearField(1);
  @$pb.TagNumber(1)
  TerminalAttach ensureAttach() => $_ensure(0);

  /// input forwards raw bytes to the PTY.
  @$pb.TagNumber(2)
  TerminalInput get input => $_getN(1);
  @$pb.TagNumber(2)
  set input(TerminalInput value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasInput() => $_has(1);
  @$pb.TagNumber(2)
  void clearInput() => $_clearField(2);
  @$pb.TagNumber(2)
  TerminalInput ensureInput() => $_ensure(1);

  /// resize updates the PTY dimensions.
  @$pb.TagNumber(3)
  TerminalResize get resize => $_getN(2);
  @$pb.TagNumber(3)
  set resize(TerminalResize value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasResize() => $_has(2);
  @$pb.TagNumber(3)
  void clearResize() => $_clearField(3);
  @$pb.TagNumber(3)
  TerminalResize ensureResize() => $_ensure(2);

  /// detach closes only the current attachment.
  @$pb.TagNumber(4)
  TerminalDetach get detach => $_getN(3);
  @$pb.TagNumber(4)
  set detach(TerminalDetach value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasDetach() => $_has(3);
  @$pb.TagNumber(4)
  void clearDetach() => $_clearField(4);
  @$pb.TagNumber(4)
  TerminalDetach ensureDetach() => $_ensure(3);

  /// close stops the backing shell.
  @$pb.TagNumber(5)
  TerminalClose get close => $_getN(4);
  @$pb.TagNumber(5)
  set close(TerminalClose value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasClose() => $_has(4);
  @$pb.TagNumber(5)
  void clearClose() => $_clearField(5);
  @$pb.TagNumber(5)
  TerminalClose ensureClose() => $_ensure(4);

  /// ping checks the attachment connection.
  @$pb.TagNumber(6)
  TerminalPing get ping => $_getN(5);
  @$pb.TagNumber(6)
  set ping(TerminalPing value) => $_setField(6, value);
  @$pb.TagNumber(6)
  $core.bool hasPing() => $_has(5);
  @$pb.TagNumber(6)
  void clearPing() => $_clearField(6);
  @$pb.TagNumber(6)
  TerminalPing ensurePing() => $_ensure(5);
}

/// TerminalAttached confirms a successful PTY attachment.
class TerminalAttached extends $pb.GeneratedMessage {
  factory TerminalAttached({
    $core.String? terminalSessionUid,
    $core.int? columns,
    $core.int? rows,
  }) {
    final result = create();
    if (terminalSessionUid != null)
      result.terminalSessionUid = terminalSessionUid;
    if (columns != null) result.columns = columns;
    if (rows != null) result.rows = rows;
    return result;
  }

  TerminalAttached._();

  factory TerminalAttached.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalAttached.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalAttached',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'terminalSessionUid')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'columns', $pb.PbFieldType.OU3)
    ..a<$core.int>(3, _omitFieldNames ? '' : 'rows', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalAttached clone() => TerminalAttached()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalAttached copyWith(void Function(TerminalAttached) updates) =>
      super.copyWith((message) => updates(message as TerminalAttached))
          as TerminalAttached;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalAttached create() => TerminalAttached._();
  @$core.override
  TerminalAttached createEmptyInstance() => create();
  static $pb.PbList<TerminalAttached> createRepeated() =>
      $pb.PbList<TerminalAttached>();
  @$core.pragma('dart2js:noInline')
  static TerminalAttached getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalAttached>(create);
  static TerminalAttached? _defaultInstance;

  /// terminal_session_uid is the attached terminal session UUIDv4 identifier.
  @$pb.TagNumber(1)
  $core.String get terminalSessionUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set terminalSessionUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTerminalSessionUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearTerminalSessionUid() => $_clearField(1);

  /// columns is the current terminal width.
  @$pb.TagNumber(2)
  $core.int get columns => $_getIZ(1);
  @$pb.TagNumber(2)
  set columns($core.int value) => $_setUnsignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasColumns() => $_has(1);
  @$pb.TagNumber(2)
  void clearColumns() => $_clearField(2);

  /// rows is the current terminal height.
  @$pb.TagNumber(3)
  $core.int get rows => $_getIZ(2);
  @$pb.TagNumber(3)
  set rows($core.int value) => $_setUnsignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRows() => $_has(2);
  @$pb.TagNumber(3)
  void clearRows() => $_clearField(3);
}

/// TerminalOutput carries raw PTY output bytes.
class TerminalOutput extends $pb.GeneratedMessage {
  factory TerminalOutput({
    $core.List<$core.int>? data,
  }) {
    final result = create();
    if (data != null) result.data = data;
    return result;
  }

  TerminalOutput._();

  factory TerminalOutput.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalOutput.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalOutput',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..a<$core.List<$core.int>>(
        1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalOutput clone() => TerminalOutput()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalOutput copyWith(void Function(TerminalOutput) updates) =>
      super.copyWith((message) => updates(message as TerminalOutput))
          as TerminalOutput;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalOutput create() => TerminalOutput._();
  @$core.override
  TerminalOutput createEmptyInstance() => create();
  static $pb.PbList<TerminalOutput> createRepeated() =>
      $pb.PbList<TerminalOutput>();
  @$core.pragma('dart2js:noInline')
  static TerminalOutput getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalOutput>(create);
  static TerminalOutput? _defaultInstance;

  /// data is the raw PTY byte sequence.
  @$pb.TagNumber(1)
  $core.List<$core.int> get data => $_getN(0);
  @$pb.TagNumber(1)
  set data($core.List<$core.int> value) => $_setBytes(0, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
}

/// TerminalExited reports that the backing shell has ended.
class TerminalExited extends $pb.GeneratedMessage {
  factory TerminalExited({
    $core.int? exitCode,
    $core.String? signal,
  }) {
    final result = create();
    if (exitCode != null) result.exitCode = exitCode;
    if (signal != null) result.signal = signal;
    return result;
  }

  TerminalExited._();

  factory TerminalExited.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalExited.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalExited',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..a<$core.int>(1, _omitFieldNames ? '' : 'exitCode', $pb.PbFieldType.O3)
    ..aOS(2, _omitFieldNames ? '' : 'signal')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalExited clone() => TerminalExited()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalExited copyWith(void Function(TerminalExited) updates) =>
      super.copyWith((message) => updates(message as TerminalExited))
          as TerminalExited;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalExited create() => TerminalExited._();
  @$core.override
  TerminalExited createEmptyInstance() => create();
  static $pb.PbList<TerminalExited> createRepeated() =>
      $pb.PbList<TerminalExited>();
  @$core.pragma('dart2js:noInline')
  static TerminalExited getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalExited>(create);
  static TerminalExited? _defaultInstance;

  /// exit_code is the shell exit code when available.
  @$pb.TagNumber(1)
  $core.int get exitCode => $_getIZ(0);
  @$pb.TagNumber(1)
  set exitCode($core.int value) => $_setSignedInt32(0, value);
  @$pb.TagNumber(1)
  $core.bool hasExitCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearExitCode() => $_clearField(1);

  /// signal is the terminating signal name when available.
  @$pb.TagNumber(2)
  $core.String get signal => $_getSZ(1);
  @$pb.TagNumber(2)
  set signal($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasSignal() => $_has(1);
  @$pb.TagNumber(2)
  void clearSignal() => $_clearField(2);
}

/// TerminalError reports a bounded attachment protocol error.
class TerminalError extends $pb.GeneratedMessage {
  factory TerminalError({
    $core.String? code,
    $core.String? message,
  }) {
    final result = create();
    if (code != null) result.code = code;
    if (message != null) result.message = message;
    return result;
  }

  TerminalError._();

  factory TerminalError.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalError.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalError',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'code')
    ..aOS(2, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalError clone() => TerminalError()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalError copyWith(void Function(TerminalError) updates) =>
      super.copyWith((message) => updates(message as TerminalError))
          as TerminalError;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalError create() => TerminalError._();
  @$core.override
  TerminalError createEmptyInstance() => create();
  static $pb.PbList<TerminalError> createRepeated() =>
      $pb.PbList<TerminalError>();
  @$core.pragma('dart2js:noInline')
  static TerminalError getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalError>(create);
  static TerminalError? _defaultInstance;

  /// code is a stable machine-readable error code.
  @$pb.TagNumber(1)
  $core.String get code => $_getSZ(0);
  @$pb.TagNumber(1)
  set code($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasCode() => $_has(0);
  @$pb.TagNumber(1)
  void clearCode() => $_clearField(1);

  /// message is a non-secret human-readable diagnostic.
  @$pb.TagNumber(2)
  $core.String get message => $_getSZ(1);
  @$pb.TagNumber(2)
  set message($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMessage() => $_has(1);
  @$pb.TagNumber(2)
  void clearMessage() => $_clearField(2);
}

/// TerminalPong echoes a client liveness nonce.
class TerminalPong extends $pb.GeneratedMessage {
  factory TerminalPong({
    $fixnum.Int64? nonce,
  }) {
    final result = create();
    if (nonce != null) result.nonce = nonce;
    return result;
  }

  TerminalPong._();

  factory TerminalPong.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalPong.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalPong',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..a<$fixnum.Int64>(1, _omitFieldNames ? '' : 'nonce', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalPong clone() => TerminalPong()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalPong copyWith(void Function(TerminalPong) updates) =>
      super.copyWith((message) => updates(message as TerminalPong))
          as TerminalPong;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalPong create() => TerminalPong._();
  @$core.override
  TerminalPong createEmptyInstance() => create();
  static $pb.PbList<TerminalPong> createRepeated() =>
      $pb.PbList<TerminalPong>();
  @$core.pragma('dart2js:noInline')
  static TerminalPong getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalPong>(create);
  static TerminalPong? _defaultInstance;

  /// nonce is copied from TerminalPing.
  @$pb.TagNumber(1)
  $fixnum.Int64 get nonce => $_getI64(0);
  @$pb.TagNumber(1)
  set nonce($fixnum.Int64 value) => $_setInt64(0, value);
  @$pb.TagNumber(1)
  $core.bool hasNonce() => $_has(0);
  @$pb.TagNumber(1)
  void clearNonce() => $_clearField(1);
}

enum TerminalServerFrame_Payload {
  attached,
  output,
  exited,
  error,
  pong,
  notSet
}

/// TerminalServerFrame is one binary Protobuf frame sent by the server.
class TerminalServerFrame extends $pb.GeneratedMessage {
  factory TerminalServerFrame({
    TerminalAttached? attached,
    TerminalOutput? output,
    TerminalExited? exited,
    TerminalError? error,
    TerminalPong? pong,
  }) {
    final result = create();
    if (attached != null) result.attached = attached;
    if (output != null) result.output = output;
    if (exited != null) result.exited = exited;
    if (error != null) result.error = error;
    if (pong != null) result.pong = pong;
    return result;
  }

  TerminalServerFrame._();

  factory TerminalServerFrame.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory TerminalServerFrame.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, TerminalServerFrame_Payload>
      _TerminalServerFrame_PayloadByTag = {
    1: TerminalServerFrame_Payload.attached,
    2: TerminalServerFrame_Payload.output,
    3: TerminalServerFrame_Payload.exited,
    4: TerminalServerFrame_Payload.error,
    5: TerminalServerFrame_Payload.pong,
    0: TerminalServerFrame_Payload.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'TerminalServerFrame',
      package: const $pb.PackageName(
          _omitMessageNames ? '' : 'super_manager.terminal.v1'),
      createEmptyInstance: create)
    ..oo(0, [1, 2, 3, 4, 5])
    ..aOM<TerminalAttached>(1, _omitFieldNames ? '' : 'attached',
        subBuilder: TerminalAttached.create)
    ..aOM<TerminalOutput>(2, _omitFieldNames ? '' : 'output',
        subBuilder: TerminalOutput.create)
    ..aOM<TerminalExited>(3, _omitFieldNames ? '' : 'exited',
        subBuilder: TerminalExited.create)
    ..aOM<TerminalError>(4, _omitFieldNames ? '' : 'error',
        subBuilder: TerminalError.create)
    ..aOM<TerminalPong>(5, _omitFieldNames ? '' : 'pong',
        subBuilder: TerminalPong.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalServerFrame clone() => TerminalServerFrame()..mergeFromMessage(this);
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  TerminalServerFrame copyWith(void Function(TerminalServerFrame) updates) =>
      super.copyWith((message) => updates(message as TerminalServerFrame))
          as TerminalServerFrame;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static TerminalServerFrame create() => TerminalServerFrame._();
  @$core.override
  TerminalServerFrame createEmptyInstance() => create();
  static $pb.PbList<TerminalServerFrame> createRepeated() =>
      $pb.PbList<TerminalServerFrame>();
  @$core.pragma('dart2js:noInline')
  static TerminalServerFrame getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<TerminalServerFrame>(create);
  static TerminalServerFrame? _defaultInstance;

  TerminalServerFrame_Payload whichPayload() =>
      _TerminalServerFrame_PayloadByTag[$_whichOneof(0)]!;
  void clearPayload() => $_clearField($_whichOneof(0));

  /// attached confirms a successful attachment.
  @$pb.TagNumber(1)
  TerminalAttached get attached => $_getN(0);
  @$pb.TagNumber(1)
  set attached(TerminalAttached value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasAttached() => $_has(0);
  @$pb.TagNumber(1)
  void clearAttached() => $_clearField(1);
  @$pb.TagNumber(1)
  TerminalAttached ensureAttached() => $_ensure(0);

  /// output carries raw PTY bytes.
  @$pb.TagNumber(2)
  TerminalOutput get output => $_getN(1);
  @$pb.TagNumber(2)
  set output(TerminalOutput value) => $_setField(2, value);
  @$pb.TagNumber(2)
  $core.bool hasOutput() => $_has(1);
  @$pb.TagNumber(2)
  void clearOutput() => $_clearField(2);
  @$pb.TagNumber(2)
  TerminalOutput ensureOutput() => $_ensure(1);

  /// exited reports that the shell ended.
  @$pb.TagNumber(3)
  TerminalExited get exited => $_getN(2);
  @$pb.TagNumber(3)
  set exited(TerminalExited value) => $_setField(3, value);
  @$pb.TagNumber(3)
  $core.bool hasExited() => $_has(2);
  @$pb.TagNumber(3)
  void clearExited() => $_clearField(3);
  @$pb.TagNumber(3)
  TerminalExited ensureExited() => $_ensure(2);

  /// error reports an attachment protocol failure.
  @$pb.TagNumber(4)
  TerminalError get error => $_getN(3);
  @$pb.TagNumber(4)
  set error(TerminalError value) => $_setField(4, value);
  @$pb.TagNumber(4)
  $core.bool hasError() => $_has(3);
  @$pb.TagNumber(4)
  void clearError() => $_clearField(4);
  @$pb.TagNumber(4)
  TerminalError ensureError() => $_ensure(3);

  /// pong answers a client ping.
  @$pb.TagNumber(5)
  TerminalPong get pong => $_getN(4);
  @$pb.TagNumber(5)
  set pong(TerminalPong value) => $_setField(5, value);
  @$pb.TagNumber(5)
  $core.bool hasPong() => $_has(4);
  @$pb.TagNumber(5)
  void clearPong() => $_clearField(5);
  @$pb.TagNumber(5)
  TerminalPong ensurePong() => $_ensure(4);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
