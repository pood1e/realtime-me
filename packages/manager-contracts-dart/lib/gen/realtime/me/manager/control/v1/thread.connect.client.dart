//
//  Generated code. Do not modify.
//  source: realtime/me/manager/control/v1/thread.proto
//

import "package:connectrpc/connect.dart" as connect;
import "thread.pb.dart" as realtimememanagercontrolv1thread;
import "thread.connect.spec.dart" as specs;

/// ThreadService manages durable structured conversations.
extension type ThreadServiceClient (connect.Transport _transport) {
  /// CreateThread creates a conversation for one workspace and runtime.
  Future<realtimememanagercontrolv1thread.CreateThreadResponse> createThread(
    realtimememanagercontrolv1thread.CreateThreadRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ThreadService.createThread,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetThread returns one conversation.
  Future<realtimememanagercontrolv1thread.GetThreadResponse> getThread(
    realtimememanagercontrolv1thread.GetThreadRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ThreadService.getThread,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListThreads returns conversations in one workspace.
  Future<realtimememanagercontrolv1thread.ListThreadsResponse> listThreads(
    realtimememanagercontrolv1thread.ListThreadsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ThreadService.listThreads,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteThread deletes a conversation and its semantic history.
  Future<realtimememanagercontrolv1thread.DeleteThreadResponse> deleteThread(
    realtimememanagercontrolv1thread.DeleteThreadRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.ThreadService.deleteThread,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
