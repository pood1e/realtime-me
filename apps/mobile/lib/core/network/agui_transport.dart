import 'dart:convert';
import 'dart:io';

import 'package:ag_ui/ag_ui.dart';

final class AguiEnvelope {
  final int sequence;
  final BaseEvent event;

  const AguiEnvelope({required this.sequence, required this.event});
}

final class AguiTransport {
  static const _connectTimeout = Duration(seconds: 20);
  static const _maxErrorBytes = 64 * 1024;

  final Uri _baseUri;
  final String _token;
  final HttpClient _client;

  factory AguiTransport({
    required Uri baseUri,
    required String token,
    required HttpClient client,
  }) => AguiTransport._(baseUri, token, client);

  const AguiTransport._(this._baseUri, this._token, this._client);

  Stream<AguiEnvelope> run(RunAgentInput input, {int after = 0}) {
    return _open(
      method: 'POST',
      uri: _resolve('/v1/ag-ui/runs', {'after': '$after'}),
      body: input.toJson(),
    );
  }

  Stream<AguiEnvelope> replay(String threadUid, {int after = 0}) {
    return _open(
      method: 'GET',
      uri: _resolve('/v1/ag-ui/threads/$threadUid/events', {'after': '$after'}),
    );
  }

  Future<AgentCapabilities> getCapabilities(String threadUid) async {
    final request = await _client
        .getUrl(_resolve('/v1/ag-ui/threads/$threadUid/capabilities'))
        .timeout(_connectTimeout);
    _setHeaders(request);
    final response = await request.close().timeout(_connectTimeout);
    final body = await _readBounded(response, _maxErrorBytes);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw HttpException(_errorMessage(response.statusCode, body));
    }
    final decoded = jsonDecode(utf8.decode(body));
    if (decoded is! Map<String, dynamic>) {
      throw const FormatException('能力响应不是 JSON 对象');
    }
    return AgentCapabilities.fromJson(decoded);
  }

  Stream<AguiEnvelope> _open({
    required String method,
    required Uri uri,
    Map<String, dynamic>? body,
  }) async* {
    final request = await _client.openUrl(method, uri).timeout(_connectTimeout);
    _setHeaders(request);
    request.headers.set(HttpHeaders.acceptHeader, 'text/event-stream');
    if (body != null) {
      final bytes = utf8.encode(jsonEncode(body));
      request.headers.contentType = ContentType.json;
      request.contentLength = bytes.length;
      request.add(bytes);
    }
    final response = await request.close().timeout(_connectTimeout);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      final errorBody = await _readBounded(response, _maxErrorBytes);
      throw HttpException(
        _errorMessage(response.statusCode, errorBody),
        uri: uri,
      );
    }
    final messages = SseParser(
      maxDataCodeUnits: 4 * 1024 * 1024,
    ).parseBytes(response);
    await for (final message in messages) {
      final data = message.data;
      if (data == null || data.isEmpty || data.trim() == ':') {
        continue;
      }
      final sequence = int.tryParse(message.id ?? '');
      if (sequence == null || sequence <= 0) {
        throw const FormatException('SSE 事件缺少有效 sequence');
      }
      final decoded = jsonDecode(data);
      if (decoded is! Map<String, dynamic>) {
        throw const FormatException('AG-UI 事件不是 JSON 对象');
      }
      yield AguiEnvelope(sequence: sequence, event: _decodeEvent(decoded));
    }
  }

  Uri _resolve(String path, [Map<String, String>? query]) =>
      _baseUri.replace(path: path, queryParameters: query);

  void _setHeaders(HttpClientRequest request) {
    request.headers.set(HttpHeaders.authorizationHeader, 'Bearer $_token');
    request.headers.set(HttpHeaders.cacheControlHeader, 'no-cache');
  }
}

BaseEvent _decodeEvent(Map<String, dynamic> json) {
  final event = BaseEvent.fromJson(json);
  switch (event) {
    case TextMessageStartEvent(messageId: final id) ||
        TextMessageContentEvent(messageId: final id) ||
        TextMessageEndEvent(messageId: final id) ||
        ReasoningStartEvent(messageId: final id) ||
        ReasoningMessageStartEvent(messageId: final id) ||
        ReasoningMessageContentEvent(messageId: final id) ||
        ReasoningMessageEndEvent(messageId: final id) ||
        ReasoningEndEvent(messageId: final id):
      _requireIdentifier(id, 'messageId');
    case ToolCallStartEvent(toolCallId: final id, toolCallName: final name):
      _requireIdentifier(id, 'toolCallId');
      _requireIdentifier(name, 'toolCallName');
    case ToolCallArgsEvent(toolCallId: final id) ||
        ToolCallEndEvent(toolCallId: final id):
      _requireIdentifier(id, 'toolCallId');
    case ToolCallResultEvent(
      messageId: final messageId,
      toolCallId: final toolCallId,
    ):
      _requireIdentifier(messageId, 'messageId');
      _requireIdentifier(toolCallId, 'toolCallId');
    case ActivitySnapshotEvent(
          messageId: final messageId,
          activityType: final activityType,
        ) ||
        ActivityDeltaEvent(
          messageId: final messageId,
          activityType: final activityType,
        ):
      _requireIdentifier(messageId, 'messageId');
      _requireIdentifier(activityType, 'activityType');
    case RunStartedEvent(threadId: final threadId, runId: final runId) ||
        RunFinishedEvent(threadId: final threadId, runId: final runId):
      _requireIdentifier(threadId, 'threadId');
      _requireIdentifier(runId, 'runId');
    case StepStartedEvent(stepName: final name) ||
        StepFinishedEvent(stepName: final name) ||
        CustomEvent(name: final name):
      _requireIdentifier(name, event is CustomEvent ? 'name' : 'stepName');
    default:
      break;
  }
  return event;
}

void _requireIdentifier(String value, String field) {
  if (value.trim().isEmpty) {
    throw FormatException('AG-UI 事件的 $field 不能为空');
  }
}

Future<List<int>> _readBounded(HttpClientResponse response, int limit) async {
  final bytes = <int>[];
  await for (final chunk in response) {
    final remaining = limit - bytes.length;
    if (remaining <= 0) {
      break;
    }
    bytes.addAll(chunk.length <= remaining ? chunk : chunk.take(remaining));
  }
  return bytes;
}

String _errorMessage(int statusCode, List<int> body) {
  final text = utf8.decode(body, allowMalformed: true);
  try {
    final value = jsonDecode(text);
    if (value is Map<String, dynamic> && value['error'] is String) {
      return 'HTTP $statusCode: ${value['error']}';
    }
  } on FormatException {
    // Use the bounded plain-text response below.
  }
  return 'HTTP $statusCode: ${text.trim().isEmpty ? '请求失败' : text.trim()}';
}
