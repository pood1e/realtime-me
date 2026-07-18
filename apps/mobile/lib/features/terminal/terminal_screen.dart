import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:xterm/xterm.dart';

import '../../core/network/remote_session.dart';
import '../../core/network/terminal_socket.dart';
import '../../core/state/app_session.dart';
import 'package:realtime_me_manager_contracts/gen/realtime/me/manager/control/v1/terminal.pb.dart';
import '../../ui/common.dart';

const _maxReconnectAttempts = 8;

enum _ConnectionState { connecting, connected, disconnected, closed }

class TerminalScreen extends ConsumerStatefulWidget {
  final String terminalUid;

  const TerminalScreen({required this.terminalUid, super.key});

  @override
  ConsumerState<TerminalScreen> createState() => _TerminalScreenState();
}

class _TerminalScreenState extends ConsumerState<TerminalScreen> {
  late final Terminal _terminal;
  late Future<TerminalSession> _terminalSession;
  TerminalSocket? _socket;
  StreamSubscription<String>? _outputSubscription;
  Timer? _reconnectTimer;
  Timer? _stableConnectionTimer;
  _ConnectionState _connectionState = _ConnectionState.connecting;
  String? _connectionError;
  int _reconnectAttempts = 0;
  bool _disposed = false;
  bool _closing = false;

  RemoteSession get _session => ref.read(appSessionProvider).requireValue!;

  @override
  void initState() {
    super.initState();
    _terminal = Terminal(
      maxLines: 5000,
      onOutput: _write,
      onResize: (width, height, pixelWidth, pixelHeight) =>
          _resize(width, height),
    );
    _terminalSession = _load();
    unawaited(_initialize());
  }

  @override
  void dispose() {
    _disposed = true;
    _reconnectTimer?.cancel();
    _stableConnectionTimer?.cancel();
    unawaited(_outputSubscription?.cancel());
    final socket = _socket;
    _socket = null;
    if (socket != null) {
      unawaited(socket.detach());
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: FutureBuilder<TerminalSession>(
          future: _terminalSession,
          builder: (context, snapshot) =>
              Text(snapshot.data?.displayName ?? '终端'),
        ),
        actions: [
          Tooltip(
            message: _connectionLabel,
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8),
              child: Icon(_connectionIcon, color: _connectionColor(context)),
            ),
          ),
          IconButton(
            tooltip: '终端说明',
            onPressed: _showTerminalInfo,
            icon: const Icon(Icons.info_outline_rounded),
          ),
          IconButton(
            tooltip: '关闭终端',
            onPressed: _closing ? null : _confirmClose,
            icon: _closing
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.close_rounded),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: FutureBuilder<TerminalSession>(
        future: _terminalSession,
        builder: (context, snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snapshot.hasError) {
            return ErrorState(
              message: readableError(snapshot.error!),
              onRetry: _reload,
            );
          }
          if (snapshot.data!.state !=
                  TerminalSessionState.TERMINAL_SESSION_STATE_RUNNING ||
              _connectionState == _ConnectionState.closed) {
            return const EmptyState(
              icon: Icons.terminal_rounded,
              title: '终端已关闭',
              message: '返回工作区创建新的终端会话。',
            );
          }
          return Column(
            children: [
              if (_connectionState == _ConnectionState.disconnected)
                MaterialBanner(
                  leading: const Icon(Icons.link_off_rounded),
                  content: Text(_connectionError ?? '终端连接已断开'),
                  actions: [
                    TextButton(onPressed: _retry, child: const Text('重连')),
                  ],
                ),
              Expanded(
                child: ColoredBox(
                  color: Colors.black,
                  child: TerminalView(
                    _terminal,
                    autofocus: true,
                    readOnly: _connectionState != _ConnectionState.connected,
                    padding: const EdgeInsets.all(8),
                    keyboardType: TextInputType.text,
                  ),
                ),
              ),
              _TerminalKeys(
                enabled: _connectionState == _ConnectionState.connected,
                send: _write,
              ),
            ],
          );
        },
      ),
    );
  }

  Future<TerminalSession> _load() =>
      _session.control.getTerminalSession(widget.terminalUid);

  Future<void> _initialize() async {
    try {
      final terminalSession = await _terminalSession;
      if (terminalSession.state ==
          TerminalSessionState.TERMINAL_SESSION_STATE_RUNNING) {
        await _connect();
      } else if (mounted) {
        setState(() => _connectionState = _ConnectionState.closed);
      }
    } on Object {
      // FutureBuilder renders the resource load error.
    }
  }

  Future<void> _connect() async {
    if (_disposed || _closing || _socket != null) {
      return;
    }
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    if (mounted) {
      setState(() {
        _connectionState = _ConnectionState.connecting;
        _connectionError = null;
      });
    }
    try {
      final socket = await _session.connectTerminal(widget.terminalUid);
      if (_disposed || _closing) {
        await socket.detach();
        return;
      }
      _socket = socket;
      _stableConnectionTimer?.cancel();
      _stableConnectionTimer = Timer(const Duration(seconds: 30), () {
        if (identical(_socket, socket)) {
          _reconnectAttempts = 0;
        }
      });
      _outputSubscription = utf8.decoder
          .bind(socket.output)
          .listen(
            _terminal.write,
            onError: (Object error, StackTrace stackTrace) {
              _connectionError = readableError(error);
            },
            onDone: () => _handleDisconnected(socket),
            cancelOnError: false,
          );
      unawaited(socket.done.then((_) => _handleDisconnected(socket)));
      if (mounted) {
        setState(() {
          _connectionState = _ConnectionState.connected;
          _connectionError = null;
        });
      }
    } on Object catch (error) {
      _scheduleReconnect(error);
    }
  }

  void _handleDisconnected(TerminalSocket socket) {
    if (!identical(_socket, socket)) {
      return;
    }
    _socket = null;
    _stableConnectionTimer?.cancel();
    _stableConnectionTimer = null;
    unawaited(_outputSubscription?.cancel());
    _outputSubscription = null;
    if (_disposed || _closing) {
      return;
    }
    if (socket.exited) {
      _connectionState = _ConnectionState.closed;
      if (mounted) {
        setState(() {});
      }
      return;
    }
    _scheduleReconnect(StateError(_connectionError ?? '终端连接已关闭'));
  }

  void _scheduleReconnect(Object error) {
    if (_disposed || _closing || _reconnectTimer != null) {
      return;
    }
    _connectionState = _ConnectionState.disconnected;
    _connectionError = readableError(error);
    _reconnectAttempts += 1;
    if (mounted) {
      setState(() {});
    }
    if (_reconnectAttempts > _maxReconnectAttempts) {
      return;
    }
    final seconds = switch (_reconnectAttempts) {
      1 => 1,
      2 => 2,
      3 => 4,
      4 => 8,
      _ => 10,
    };
    _reconnectTimer = Timer(Duration(seconds: seconds), _connect);
  }

  void _retry() {
    _reconnectAttempts = 0;
    unawaited(_connect());
  }

  void _write(String data) {
    if (_connectionState == _ConnectionState.connected) {
      _socket?.write(data);
    }
  }

  void _resize(int width, int height) {
    if (_connectionState == _ConnectionState.connected) {
      _socket?.resize(width, height);
    }
  }

  Future<void> _confirmClose() async {
    final confirmed =
        await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: const Text('关闭终端？'),
            content: const Text('Shell 和其中正在运行的进程会被终止。'),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context, false),
                child: const Text('取消'),
              ),
              FilledButton(
                onPressed: () => Navigator.pop(context, true),
                child: const Text('关闭'),
              ),
            ],
          ),
        ) ??
        false;
    if (!confirmed || !mounted) {
      return;
    }
    setState(() => _closing = true);
    _reconnectTimer?.cancel();
    _stableConnectionTimer?.cancel();
    try {
      final socket = _socket;
      _socket = null;
      if (socket != null) {
        await socket.closeTerminal();
      } else {
        await _session.control.deleteTerminalSession(widget.terminalUid);
      }
      if (mounted) {
        setState(() => _connectionState = _ConnectionState.closed);
        context.pop();
      }
    } on Object catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text(readableError(error))));
        setState(() => _closing = false);
      }
    }
  }

  Future<void> _showTerminalInfo() {
    return showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('原始终端'),
        content: const Text(
          '终端直接连接 tmux-backed PTY。其字符流不会进入 Agent 对话历史，也不会写入应用审计日志。',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('知道了'),
          ),
        ],
      ),
    );
  }

  void _reload() {
    setState(() => _terminalSession = _load());
    unawaited(_initialize());
  }

  String get _connectionLabel => switch (_connectionState) {
    _ConnectionState.connecting => '正在连接',
    _ConnectionState.connected => '终端已连接',
    _ConnectionState.disconnected => '终端已断开',
    _ConnectionState.closed => '终端已关闭',
  };

  IconData get _connectionIcon => switch (_connectionState) {
    _ConnectionState.connecting => Icons.sync_rounded,
    _ConnectionState.connected => Icons.link_rounded,
    _ConnectionState.disconnected => Icons.link_off_rounded,
    _ConnectionState.closed => Icons.cancel_outlined,
  };

  Color _connectionColor(BuildContext context) => switch (_connectionState) {
    _ConnectionState.connected => Theme.of(context).colorScheme.primary,
    _ConnectionState.connecting => Theme.of(context).colorScheme.tertiary,
    _ConnectionState.disconnected ||
    _ConnectionState.closed => Theme.of(context).colorScheme.error,
  };
}

class _TerminalKeys extends StatelessWidget {
  final bool enabled;
  final ValueChanged<String> send;

  const _TerminalKeys({required this.enabled, required this.send});

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      top: false,
      child: Material(
        color: Theme.of(context).colorScheme.surfaceContainer,
        child: SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          child: Row(
            children: [
              _key('Ctrl+C', '\x03'),
              _key('Ctrl+D', '\x04'),
              _key('Tab', '\t'),
              _key('Esc', '\x1b'),
              _key('↑', '\x1b[A'),
              _key('↓', '\x1b[B'),
            ],
          ),
        ),
      ),
    );
  }

  Widget _key(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: OutlinedButton(
        onPressed: enabled ? () => send(value) : null,
        child: Text(label),
      ),
    );
  }
}
