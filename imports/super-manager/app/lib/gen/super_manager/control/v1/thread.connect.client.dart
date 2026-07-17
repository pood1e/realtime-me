//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/thread.proto
//

import "package:connectrpc/connect.dart" as connect;
import "thread.pb.dart" as super_managercontrolv1thread;
import "thread.connect.spec.dart" as specs;

/// ThreadService manages durable structured conversations.
extension type ThreadServiceClient (connect.Transport _transport) {
  /// CreateThread creates a conversation for one workspace and runtime.
  Future<super_managercontrolv1thread.CreateThreadResponse> createThread(
    super_managercontrolv1thread.CreateThreadRequest input, {
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
  Future<super_managercontrolv1thread.GetThreadResponse> getThread(
    super_managercontrolv1thread.GetThreadRequest input, {
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
  Future<super_managercontrolv1thread.ListThreadsResponse> listThreads(
    super_managercontrolv1thread.ListThreadsRequest input, {
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
  Future<super_managercontrolv1thread.DeleteThreadResponse> deleteThread(
    super_managercontrolv1thread.DeleteThreadRequest input, {
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
