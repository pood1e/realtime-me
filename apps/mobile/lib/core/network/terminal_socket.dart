import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:fixnum/fixnum.dart';
import 'package:web_socket_channel/io.dart';

import 'package:realtime_me_manager_contracts/gen/realtime/me/manager/terminal/v1/terminal.pb.dart';

final class TerminalSocket {
  final IOWebSocketChannel _channel;
  final HttpClient _httpClient;
  final String _terminalUid;
  final StreamController<List<int>> _output = StreamController();
  final Completer<void> _ready = Completer();
  final Completer<void> _done = Completer();
  late final StreamSubscription<dynamic> _subscription;
  bool _closed = false;
  bool _finished = false;
  bool _exited = false;
  int? _exitCode;

  TerminalSocket._(this._channel, this._httpClient, this._terminalUid) {
    _subscription = _channel.stream.listen(
      _handleFrame,
      onError: (Object error, StackTrace stackTrace) {
        if (!_ready.isCompleted) {
          _ready.completeError(error, stackTrace);
        }
        if (!_output.isClosed) {
          _output.addError(error, stackTrace);
        }
      },
      onDone: () {
        unawaited(_finish());
      },
      cancelOnError: false,
    );
  }

  static Future<TerminalSocket> connect({
    required Uri serviceUri,
    required String token,
    required String terminalUid,
    required HttpClient httpClient,
  }) async {
    final uri = serviceUri.replace(
      scheme: 'wss',
      path: '/v1/terminal',
      query: null,
      fragment: null,
    );
    IOWebSocketChannel? channel;
    TerminalSocket? socket;
    try {
      channel = IOWebSocketChannel.connect(
        uri,
        headers: {HttpHeaders.authorizationHeader: 'Bearer $token'},
        pingInterval: const Duration(seconds: 20),
        connectTimeout: const Duration(seconds: 20),
        customClient: httpClient,
      );
      await channel.ready;
      socket = TerminalSocket._(channel, httpClient, terminalUid);
      socket._send(
        TerminalClientFrame(
          attach: TerminalAttach(terminalSessionUid: terminalUid),
        ),
      );
      await socket._ready.future.timeout(const Duration(seconds: 20));
      return socket;
    } on Object {
      if (socket != null) {
        try {
          await socket._abort();
        } on Object {
          httpClient.close(force: true);
        }
      } else {
        try {
          await channel?.sink.close();
        } on Object {
          // The WebSocket setup failure remains the actionable error.
        } finally {
          httpClient.close(force: true);
        }
      }
      rethrow;
    }
  }

  Stream<List<int>> get output => _output.stream;
  Future<void> get done => _done.future;
  bool get exited => _exited;
  int? get exitCode => _exitCode;

  void write(String value) {
    final bytes = utf8.encode(value);
    var offset = 0;
    while (offset < bytes.length) {
      var end = (offset + 65536).clamp(0, bytes.length).toInt();
      while (end < bytes.length &&
          end > offset &&
          (bytes[end] & 0xC0) == 0x80) {
        end -= 1;
      }
      if (end == offset) {
        end = (offset + 65536).clamp(0, bytes.length).toInt();
      }
      _send(
        TerminalClientFrame(
          input: TerminalInput(
            data: Uint8List.fromList(bytes.sublist(offset, end)),
          ),
        ),
      );
      offset = end;
    }
  }

  void resize(int columns, int rows) {
    _send(
      TerminalClientFrame(
        resize: TerminalResize(
          columns: columns.clamp(20, 1000).toInt(),
          rows: rows.clamp(5, 1000).toInt(),
        ),
      ),
    );
  }

  void ping(int nonce) {
    _send(TerminalClientFrame(ping: TerminalPing(nonce: Int64(nonce))));
  }

  Future<void> detach() =>
      _close(TerminalClientFrame(detach: TerminalDetach()));

  Future<void> closeTerminal() =>
      _close(TerminalClientFrame(close: TerminalClose()));

  void _handleFrame(dynamic data) {
    try {
      _decodeFrame(data);
    } on Object catch (error, stackTrace) {
      if (!_output.isClosed) {
        _output.addError(error, stackTrace);
      }
      unawaited(_closeProtocolError());
    }
  }

  void _decodeFrame(dynamic data) {
    if (data is! List<int>) {
      throw const FormatException('终端服务返回了非二进制帧');
    }
    final frame = TerminalServerFrame.fromBuffer(data);
    switch (frame.whichPayload()) {
      case TerminalServerFrame_Payload.output:
        _output.add(Uint8List.fromList(frame.output.data));
      case TerminalServerFrame_Payload.error:
        final error = StateError('${frame.error.code}: ${frame.error.message}');
        if (!_ready.isCompleted) {
          _ready.completeError(error);
        }
        _output.addError(error);
      case TerminalServerFrame_Payload.exited:
        _exited = true;
        _exitCode = frame.exited.hasExitCode() ? frame.exited.exitCode : null;
        if (!_ready.isCompleted) {
          _ready.completeError(StateError('终端在连接完成前退出'));
        }
        if (!_done.isCompleted) {
          _done.complete();
        }
      case TerminalServerFrame_Payload.attached:
        if (frame.attached.terminalSessionUid != _terminalUid) {
          throw const FormatException('终端服务确认了错误的会话');
        }
        if (!_ready.isCompleted) {
          _ready.complete();
        }
      case TerminalServerFrame_Payload.pong:
        break;
      case TerminalServerFrame_Payload.notSet:
        throw const FormatException('终端服务返回了空帧');
    }
  }

  void _send(TerminalClientFrame frame) {
    if (!_closed) {
      _channel.sink.add(frame.writeToBuffer());
    }
  }

  Future<void> _close(TerminalClientFrame finalFrame) async {
    if (_closed) {
      await _finish();
      return;
    }
    _send(finalFrame);
    _closed = true;
    try {
      await _channel.sink.close();
    } finally {
      try {
        await _subscription.cancel();
      } finally {
        await _finish();
      }
    }
  }

  Future<void> _closeProtocolError() async {
    if (_closed) {
      await _finish();
      return;
    }
    _closed = true;
    try {
      await _channel.sink.close(1002, 'invalid terminal frame');
    } finally {
      try {
        await _subscription.cancel();
      } finally {
        await _finish();
      }
    }
  }

  Future<void> _abort() async {
    if (_closed) {
      await _finish();
      return;
    }
    _closed = true;
    try {
      await _channel.sink.close(1001, 'connection setup failed');
    } finally {
      try {
        await _subscription.cancel();
      } finally {
        await _finish();
      }
    }
  }

  Future<void> _finish() async {
    if (_finished) {
      return;
    }
    _finished = true;
    _closed = true;
    _httpClient.close(force: true);
    if (!_ready.isCompleted) {
      _ready.completeError(StateError('终端连接在确认前关闭'));
    }
    if (!_done.isCompleted) {
      _done.complete();
    }
    if (!_output.isClosed) {
      await _output.close();
    }
  }
}
