import 'dart:async';
import 'dart:collection';
import 'dart:convert';

import 'package:ag_ui/ag_ui.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../core/network/agui_transport.dart';
import '../../core/network/remote_session.dart';
import '../../core/state/app_session.dart';
import '../../gen/super_manager/control/v1/thread.pb.dart';
import '../../ui/common.dart';
import 'interrupt_panel.dart';
import 'timeline_projection.dart';
import 'timeline_view.dart';

const _uuid = Uuid();
const _maxPromptBytes = 128 * 1024;
const _maxReplayRetries = 8;

class ThreadScreen extends ConsumerStatefulWidget {
  final String threadUid;

  const ThreadScreen({required this.threadUid, super.key});

  @override
  ConsumerState<ThreadScreen> createState() => _ThreadScreenState();
}

class _ThreadScreenState extends ConsumerState<ThreadScreen> {
  final _projection = TimelineProjection();
  final _pendingEvents = SplayTreeMap<int, BaseEvent>();
  final _composer = TextEditingController();
  final _scrollController = ScrollController();
  final _composerFocus = FocusNode();

  late Future<Thread> _thread;
  StreamSubscription<AguiEnvelope>? _replaySubscription;
  StreamSubscription<AguiEnvelope>? _runSubscription;
  Timer? _reconnectTimer;
  AgentCapabilities? _capabilities;
  String? _streamError;
  int _replayAttempts = 0;
  bool _replayConnected = false;
  bool _submitting = false;
  bool _stopping = false;
  bool _disposed = false;

  RemoteSession get _session => ref.read(appSessionProvider).requireValue!;
  bool get _canSteer => _capabilities?.humanInTheLoop?.interventions == true;

  @override
  void initState() {
    super.initState();
    _thread = _session.control.getThread(widget.threadUid);
    unawaited(_loadCapabilities());
    _connectReplay();
  }

  @override
  void dispose() {
    _disposed = true;
    _reconnectTimer?.cancel();
    unawaited(_replaySubscription?.cancel());
    unawaited(_runSubscription?.cancel());
    _composer.dispose();
    _scrollController.dispose();
    _composerFocus.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final pending = _projection.pendingInterrupts;
    return Scaffold(
      appBar: AppBar(
        title: FutureBuilder<Thread>(
          future: _thread,
          builder: (context, snapshot) =>
              Text(snapshot.data?.displayName ?? 'Agent 对话'),
        ),
        actions: [
          Tooltip(
            message: _connectionDescription,
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8),
              child: Icon(
                _replayConnected
                    ? Icons.cloud_done_outlined
                    : Icons.cloud_off_outlined,
                color: _replayConnected
                    ? Theme.of(context).colorScheme.primary
                    : Theme.of(context).colorScheme.error,
              ),
            ),
          ),
          if (!_replayConnected && _reconnectTimer == null)
            IconButton(
              tooltip: '重新连接',
              onPressed: _retryReplay,
              icon: const Icon(Icons.refresh_rounded),
            ),
          if (_projection.running || pending != null)
            IconButton(
              tooltip: '停止任务',
              onPressed: _stopping ? null : _confirmStop,
              icon: _stopping
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.stop_circle_outlined),
            ),
          const SizedBox(width: 8),
        ],
      ),
      body: FutureBuilder<Thread>(
        future: _thread,
        builder: (context, snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snapshot.hasError) {
            return ErrorState(
              message: readableError(snapshot.error!),
              onRetry: _reloadThread,
            );
          }
          return Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 1000),
              child: Column(
                children: [
                  if (_streamError != null && !_replayConnected)
                    MaterialBanner(
                      content: Text(_streamError!),
                      leading: const Icon(Icons.sync_problem_rounded),
                      actions: [
                        TextButton(
                          onPressed: _retryReplay,
                          child: const Text('重连'),
                        ),
                      ],
                    ),
                  Expanded(
                    child: TimelineView(
                      items: _projection.items,
                      controller: _scrollController,
                    ),
                  ),
                  if (pending != null)
                    InterruptPanel(
                      key: ValueKey(pending.runId),
                      batch: pending,
                      onSubmit: _resume,
                    )
                  else
                    _Composer(
                      controller: _composer,
                      focusNode: _composerFocus,
                      running: _projection.running,
                      canSteer: _canSteer,
                      submitting: _submitting,
                      onSend: _sendComposer,
                    ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  String get _connectionDescription {
    if (_replayConnected) {
      return '事件流已连接';
    }
    if (_reconnectTimer != null) {
      return '事件流正在重连';
    }
    return '事件流已断开';
  }

  Future<void> _loadCapabilities() async {
    try {
      final capabilities = await _session.agui.getCapabilities(
        widget.threadUid,
      );
      if (mounted) {
        setState(() => _capabilities = capabilities);
      }
    } on Object catch (_) {
      // The event stream remains usable; unsupported steering stays disabled.
    }
  }

  void _connectReplay() {
    if (_disposed) {
      return;
    }
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    unawaited(_replaySubscription?.cancel());
    final stream = _session.agui.replay(
      widget.threadUid,
      after: _projection.sequence,
    );
    _replaySubscription = stream.listen(
      (envelope) {
        _replayAttempts = 0;
        if (!_replayConnected && mounted) {
          setState(() {
            _replayConnected = true;
            _streamError = null;
          });
        }
        _ingest(envelope);
      },
      onError: (Object error, StackTrace stackTrace) => _handleReplayEnd(error),
      onDone: () => _handleReplayEnd(StateError('事件流连接已关闭')),
      cancelOnError: true,
    );
    if (mounted) {
      setState(() {
        _replayConnected = true;
        _streamError = null;
      });
    }
  }

  void _handleReplayEnd(Object error) {
    if (_disposed || _reconnectTimer != null) {
      return;
    }
    _replayConnected = false;
    _streamError = readableError(error);
    _replayAttempts += 1;
    if (mounted) {
      setState(() {});
    }
    if (_replayAttempts > _maxReplayRetries) {
      return;
    }
    final seconds = switch (_replayAttempts) {
      1 => 1,
      2 => 2,
      3 => 4,
      4 => 8,
      _ => 10,
    };
    _reconnectTimer = Timer(Duration(seconds: seconds), _connectReplay);
  }

  void _retryReplay() {
    _replayAttempts = 0;
    _connectReplay();
  }

  void _ingest(AguiEnvelope envelope) {
    if (envelope.sequence <= _projection.sequence) {
      return;
    }
    _pendingEvents.putIfAbsent(envelope.sequence, () => envelope.event);
    var changed = false;
    while (true) {
      final nextSequence = _projection.sequence + 1;
      final next = _pendingEvents.remove(nextSequence);
      if (next == null) {
        break;
      }
      _projection.apply(nextSequence, next);
      changed = true;
    }
    if (changed && mounted) {
      setState(() {});
      _scrollAfterBuild();
    }
  }

  void _scrollAfterBuild() {
    final shouldFollow =
        !_scrollController.hasClients ||
        _scrollController.position.maxScrollExtent -
                _scrollController.position.pixels <
            180;
    if (!shouldFollow) {
      return;
    }
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || !_scrollController.hasClients) {
        return;
      }
      final target = _scrollController.position.maxScrollExtent;
      if (MediaQuery.disableAnimationsOf(context)) {
        _scrollController.jumpTo(target);
      } else {
        unawaited(
          _scrollController.animateTo(
            target,
            duration: const Duration(milliseconds: 180),
            curve: Curves.easeOut,
          ),
        );
      }
    });
  }

  Future<void> _sendComposer() async {
    final text = _composer.text.trim();
    if (text.isEmpty || _submitting) {
      return;
    }
    if (utf8.encode(text).length > _maxPromptBytes) {
      _showError(StateError('输入不能超过 128 KiB'));
      return;
    }
    if (_projection.running) {
      await _steer(text);
      return;
    }
    final input = RunAgentInput(
      threadId: widget.threadUid,
      runId: _uuid.v4(),
      state: const <String, dynamic>{},
      messages: [UserMessage(id: _uuid.v4(), content: text)],
      tools: const [],
      context: const [],
      forwardedProps: const <String, dynamic>{},
    );
    try {
      await _startRun(input);
      _composer.clear();
    } on Object catch (error) {
      _showError(error);
    }
  }

  Future<void> _steer(String instruction) async {
    if (!_canSteer) {
      _showError(StateError('当前 runtime 不支持追加指令'));
      return;
    }
    setState(() => _submitting = true);
    try {
      final execution = await _session.control.steerActiveExecution(
        widget.threadUid,
        instruction,
      );
      if (execution == null) {
        throw StateError('没有可追加指令的运行中任务');
      }
      _composer.clear();
    } on Object catch (error) {
      _showError(error);
    } finally {
      if (mounted) {
        setState(() => _submitting = false);
      }
    }
  }

  Future<void> _resume(List<ResumeEntry> entries) {
    final pending = _projection.pendingInterrupts;
    if (pending == null) {
      throw StateError('提问已不再等待回答');
    }
    return _startRun(
      RunAgentInput(
        threadId: widget.threadUid,
        runId: _uuid.v4(),
        parentRunId: pending.runId,
        state: const <String, dynamic>{},
        messages: const [],
        tools: const [],
        context: const [],
        forwardedProps: const <String, dynamic>{},
        resume: entries,
      ),
    );
  }

  Future<void> _startRun(RunAgentInput input) async {
    if (_submitting || _runSubscription != null) {
      throw StateError('已有请求正在提交');
    }
    setState(() => _submitting = true);
    final accepted = Completer<void>();
    late final StreamSubscription<AguiEnvelope> subscription;
    subscription = _session.agui
        .run(input, after: _projection.sequence)
        .listen(
          (envelope) {
            if (!accepted.isCompleted) {
              accepted.complete();
              if (mounted) {
                setState(() => _submitting = false);
              }
            }
            _ingest(envelope);
          },
          onError: (Object error, StackTrace stackTrace) {
            if (!accepted.isCompleted) {
              accepted.completeError(error, stackTrace);
            }
            if (mounted) {
              setState(() => _submitting = false);
            }
            if (identical(_runSubscription, subscription)) {
              _runSubscription = null;
            }
          },
          onDone: () {
            if (!accepted.isCompleted) {
              accepted.completeError(StateError('服务端未确认本次请求'));
            }
            if (mounted) {
              setState(() => _submitting = false);
            }
            if (identical(_runSubscription, subscription)) {
              _runSubscription = null;
            }
          },
          cancelOnError: true,
        );
    _runSubscription = subscription;
    try {
      await accepted.future;
    } on Object {
      await subscription.cancel();
      if (identical(_runSubscription, subscription)) {
        _runSubscription = null;
      }
      if (mounted) {
        setState(() => _submitting = false);
      }
      rethrow;
    }
  }

  Future<void> _confirmStop() async {
    final confirmed =
        await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: const Text('停止当前任务？'),
            content: const Text('正在执行的 Codex 或 Claude Code 任务会被中断。'),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context, false),
                child: const Text('继续执行'),
              ),
              FilledButton(
                onPressed: () => Navigator.pop(context, true),
                child: const Text('停止'),
              ),
            ],
          ),
        ) ??
        false;
    if (!confirmed || !mounted) {
      return;
    }
    setState(() => _stopping = true);
    try {
      final execution = await _session.control.cancelActiveExecution(
        widget.threadUid,
      );
      if (execution == null) {
        throw StateError('任务已经结束');
      }
    } on Object catch (error) {
      _showError(error);
    } finally {
      if (mounted) {
        setState(() => _stopping = false);
      }
    }
  }

  void _reloadThread() {
    setState(() => _thread = _session.control.getThread(widget.threadUid));
  }

  void _showError(Object error) {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(readableError(error))));
  }
}

class _Composer extends StatelessWidget {
  final TextEditingController controller;
  final FocusNode focusNode;
  final bool running;
  final bool canSteer;
  final bool submitting;
  final Future<void> Function() onSend;

  const _Composer({
    required this.controller,
    required this.focusNode,
    required this.running,
    required this.canSteer,
    required this.submitting,
    required this.onSend,
  });

  @override
  Widget build(BuildContext context) {
    final enabled = !submitting && (!running || canSteer);
    return SafeArea(
      top: false,
      child: Material(
        color: Theme.of(context).colorScheme.surfaceContainer,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Expanded(
                child: TextField(
                  controller: controller,
                  focusNode: focusNode,
                  enabled: enabled,
                  minLines: 1,
                  maxLines: 6,
                  keyboardType: TextInputType.multiline,
                  textCapitalization: TextCapitalization.sentences,
                  decoration: InputDecoration(
                    labelText: running ? '追加指令' : '任务目标',
                    hintText: running && !canSteer
                        ? '当前 runtime 不支持执行中追加指令'
                        : '描述要完成的工作…',
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Tooltip(
                message: running ? '追加指令' : '发送任务',
                child: IconButton.filled(
                  onPressed: enabled ? () => unawaited(onSend()) : null,
                  icon: submitting
                      ? const SizedBox(
                          width: 20,
                          height: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : Icon(
                          running
                              ? Icons.subdirectory_arrow_right_rounded
                              : Icons.send_rounded,
                        ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
