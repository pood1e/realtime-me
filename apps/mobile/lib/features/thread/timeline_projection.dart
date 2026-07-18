import 'dart:convert';

import 'package:ag_ui/ag_ui.dart';

const _maxMessageCodeUnits = 128 * 1024;
const _maxToolCodeUnits = 64 * 1024;
const _maxReasoningCodeUnits = 32 * 1024;
const _maxActivityCodeUnits = 32 * 1024;
const _truncatedSuffix = '\n\n…内容已在本机截断';

enum TimelineMessageRole { user, assistant }

sealed class TimelineItem {
  final String id;

  const TimelineItem({required this.id});
}

final class TimelineMessageItem extends TimelineItem {
  final TimelineMessageRole role;
  String content;
  bool complete;

  TimelineMessageItem({
    required super.id,
    required this.role,
    required this.content,
    required this.complete,
  });
}

final class TimelineToolItem extends TimelineItem {
  final String name;
  String arguments;
  String result;
  bool complete;

  TimelineToolItem({
    required super.id,
    required this.name,
    this.arguments = '',
    this.result = '',
    this.complete = false,
  });
}

final class TimelineReasoningItem extends TimelineItem {
  String content;
  bool complete;

  TimelineReasoningItem({
    required super.id,
    this.content = '',
    this.complete = false,
  });
}

final class TimelineActivityItem extends TimelineItem {
  final String activityType;
  Object? content;

  TimelineActivityItem({
    required super.id,
    required this.activityType,
    required this.content,
  });

  String get displayContent =>
      _bounded(_prettyJson(content), _maxActivityCodeUnits);
}

enum TimelineNoticeKind { steering, error }

final class TimelineNoticeItem extends TimelineItem {
  final TimelineNoticeKind kind;
  final String title;
  final String message;

  const TimelineNoticeItem({
    required super.id,
    required this.kind,
    required this.title,
    required this.message,
  });
}

final class PendingInterruptBatch {
  final String runId;
  final List<Interrupt> interrupts;

  const PendingInterruptBatch({required this.runId, required this.interrupts});
}

final class TimelineProjection {
  final List<TimelineItem> items = [];
  final Map<String, TimelineMessageItem> _messages = {};
  final Map<String, TimelineToolItem> _tools = {};
  final Map<String, TimelineReasoningItem> _reasoning = {};
  final Map<String, TimelineActivityItem> _activities = {};

  int sequence = 0;
  bool running = false;
  String? activeRunId;
  PendingInterruptBatch? pendingInterrupts;

  void apply(int nextSequence, BaseEvent event) {
    if (nextSequence <= sequence) {
      return;
    }
    sequence = nextSequence;
    switch (event) {
      case RunStartedEvent():
        _applyRunStarted(event);
      case TextMessageStartEvent():
        _startTextMessage(event);
      case TextMessageContentEvent():
        _appendText(event.messageId, event.delta);
      case TextMessageChunkEvent():
        _applyTextChunk(event);
      case TextMessageEndEvent():
        _messages[event.messageId]?.complete = true;
      case ToolCallStartEvent():
        _startTool(event.toolCallId, event.toolCallName);
      case ToolCallArgsEvent():
        _appendToolArguments(event.toolCallId, event.delta);
      case ToolCallChunkEvent():
        _applyToolChunk(event);
      case ToolCallEndEvent():
        _tools[event.toolCallId]?.complete = true;
      case ToolCallResultEvent():
        final tool = _ensureTool(event.toolCallId, '工具');
        tool.result = _appendBounded(
          tool.result,
          event.content,
          _maxToolCodeUnits,
        );
        tool.complete = true;
      case ReasoningStartEvent():
        _ensureReasoning(event.messageId);
      case ReasoningMessageStartEvent():
        _ensureReasoning(event.messageId);
      case ReasoningMessageContentEvent():
        _appendReasoning(event.messageId, event.delta);
      case ReasoningMessageChunkEvent():
        final id = event.messageId;
        final delta = event.delta;
        if (id != null && delta != null) {
          _appendReasoning(id, delta);
        }
      case ReasoningMessageEndEvent():
        _reasoning[event.messageId]?.complete = true;
      case ReasoningEndEvent():
        _reasoning[event.messageId]?.complete = true;
      case ActivitySnapshotEvent():
        _applyActivity(event);
      case CustomEvent():
        _applyCustom(event);
      case RunFinishedEvent():
        _applyRunFinished(event);
      case RunErrorEvent():
        running = false;
        activeRunId = null;
        pendingInterrupts = null;
        items.add(
          TimelineNoticeItem(
            id: 'error-$nextSequence',
            kind: TimelineNoticeKind.error,
            title: event.code == 'CANCELED' ? '任务已停止' : '执行失败',
            message: event.message,
          ),
        );
      default:
        break;
    }
  }

  void _applyRunStarted(RunStartedEvent event) {
    running = true;
    activeRunId = event.runId;
    if (event.parentRunId != null) {
      pendingInterrupts = null;
    }
    final messages = event.input?.messages ?? const <Message>[];
    for (var index = 0; index < messages.length; index++) {
      final message = messages[index];
      if (message is! UserMessage || message.content == null) {
        continue;
      }
      final id = message.id ?? '${event.runId}-user-$index';
      if (_messages.containsKey(id)) {
        continue;
      }
      final item = TimelineMessageItem(
        id: id,
        role: TimelineMessageRole.user,
        content: _bounded(message.content!, _maxMessageCodeUnits),
        complete: true,
      );
      _messages[id] = item;
      items.add(item);
    }
  }

  void _startTextMessage(TextMessageStartEvent event) {
    if (_messages.containsKey(event.messageId)) {
      return;
    }
    final item = TimelineMessageItem(
      id: event.messageId,
      role: event.role == TextMessageRole.user
          ? TimelineMessageRole.user
          : TimelineMessageRole.assistant,
      content: '',
      complete: false,
    );
    _messages[event.messageId] = item;
    items.add(item);
  }

  void _appendText(String id, String delta) {
    final message =
        _messages[id] ??
        TimelineMessageItem(
          id: id,
          role: TimelineMessageRole.assistant,
          content: '',
          complete: false,
        );
    if (!_messages.containsKey(id)) {
      _messages[id] = message;
      items.add(message);
    }
    message.content = _appendBounded(
      message.content,
      delta,
      _maxMessageCodeUnits,
    );
  }

  void _applyTextChunk(TextMessageChunkEvent event) {
    final id = event.messageId;
    if (id == null) {
      return;
    }
    if (!_messages.containsKey(id)) {
      _startTextMessage(
        TextMessageStartEvent(
          messageId: id,
          role: event.role ?? TextMessageRole.assistant,
          name: event.name,
        ),
      );
    }
    if (event.delta != null) {
      _appendText(id, event.delta!);
    }
  }

  void _startTool(String id, String name) {
    if (_tools.containsKey(id)) {
      return;
    }
    final item = TimelineToolItem(id: id, name: name);
    _tools[id] = item;
    items.add(item);
  }

  TimelineToolItem _ensureTool(String id, String name) {
    final existing = _tools[id];
    if (existing != null) {
      return existing;
    }
    final item = TimelineToolItem(id: id, name: name);
    _tools[id] = item;
    items.add(item);
    return item;
  }

  void _appendToolArguments(String id, String delta) {
    final tool = _ensureTool(id, '工具');
    tool.arguments = _appendBounded(tool.arguments, delta, _maxToolCodeUnits);
  }

  void _applyToolChunk(ToolCallChunkEvent event) {
    final id = event.toolCallId;
    if (id == null) {
      return;
    }
    final tool = _ensureTool(id, event.toolCallName ?? '工具');
    if (event.delta != null) {
      tool.arguments = _appendBounded(
        tool.arguments,
        event.delta!,
        _maxToolCodeUnits,
      );
    }
  }

  TimelineReasoningItem _ensureReasoning(String id) {
    final existing = _reasoning[id];
    if (existing != null) {
      return existing;
    }
    final item = TimelineReasoningItem(id: id);
    _reasoning[id] = item;
    items.add(item);
    return item;
  }

  void _appendReasoning(String id, String delta) {
    final item = _ensureReasoning(id);
    item.content = _appendBounded(item.content, delta, _maxReasoningCodeUnits);
  }

  void _applyActivity(ActivitySnapshotEvent event) {
    final existing = _activities[event.messageId];
    if (existing == null) {
      final item = TimelineActivityItem(
        id: event.messageId,
        activityType: event.activityType,
        content: event.content,
      );
      _activities[event.messageId] = item;
      items.add(item);
      return;
    }
    existing.content = event.replace
        ? event.content
        : _mergeActivity(existing.content, event.content);
  }

  void _applyCustom(CustomEvent event) {
    if (event.name != 'realtime.me.manager.steer') {
      return;
    }
    final value = event.value;
    final instruction =
        value is Map<String, dynamic> && value['instruction'] is String
        ? value['instruction']! as String
        : '';
    items.add(
      TimelineNoticeItem(
        id: 'steer-$sequence',
        kind: TimelineNoticeKind.steering,
        title: '追加指令',
        message: _bounded(instruction, _maxMessageCodeUnits),
      ),
    );
  }

  void _applyRunFinished(RunFinishedEvent event) {
    running = false;
    activeRunId = null;
    final outcome = event.outcome;
    if (outcome is RunFinishedInterruptOutcome) {
      pendingInterrupts = PendingInterruptBatch(
        runId: event.runId,
        interrupts: List.unmodifiable(outcome.interrupts),
      );
      return;
    }
    pendingInterrupts = null;
  }
}

Object? _mergeActivity(Object? current, Object? next) {
  if (current is! Map<String, dynamic> || next is! Map<String, dynamic>) {
    return next;
  }
  final merged = <String, dynamic>{...current};
  for (final entry in next.entries) {
    final previous = merged[entry.key];
    if (entry.key == 'delta' && previous is String && entry.value is String) {
      merged[entry.key] = '$previous${entry.value}';
    } else {
      merged[entry.key] = entry.value;
    }
  }
  return merged;
}

String _prettyJson(Object? value) {
  if (value is String) {
    return value;
  }
  try {
    return const JsonEncoder.withIndent('  ').convert(value);
  } on JsonUnsupportedObjectError {
    return value.toString();
  }
}

String _appendBounded(String current, String delta, int limit) {
  if (current.endsWith(_truncatedSuffix)) {
    return current;
  }
  return _bounded('$current$delta', limit);
}

String _bounded(String value, int limit) {
  if (value.length <= limit) {
    return value;
  }
  return '${value.substring(0, limit)}$_truncatedSuffix';
}
