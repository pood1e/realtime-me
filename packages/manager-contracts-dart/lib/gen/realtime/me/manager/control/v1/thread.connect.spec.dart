//
//  Generated code. Do not modify.
//  source: realtime/me/manager/control/v1/thread.proto
//

import "package:connectrpc/connect.dart" as connect;
import "thread.pb.dart" as realtimememanagercontrolv1thread;

/// ThreadService manages durable structured conversations.
abstract final class ThreadService {
  /// Fully-qualified name of the ThreadService service.
  static const name = 'realtime.me.manager.control.v1.ThreadService';

  /// CreateThread creates a conversation for one workspace and runtime.
  static const createThread = connect.Spec(
    '/$name/CreateThread',
    connect.StreamType.unary,
    realtimememanagercontrolv1thread.CreateThreadRequest.new,
    realtimememanagercontrolv1thread.CreateThreadResponse.new,
  );

  /// GetThread returns one conversation.
  static const getThread = connect.Spec(
    '/$name/GetThread',
    connect.StreamType.unary,
    realtimememanagercontrolv1thread.GetThreadRequest.new,
    realtimememanagercontrolv1thread.GetThreadResponse.new,
  );

  /// ListThreads returns conversations in one workspace.
  static const listThreads = connect.Spec(
    '/$name/ListThreads',
    connect.StreamType.unary,
    realtimememanagercontrolv1thread.ListThreadsRequest.new,
    realtimememanagercontrolv1thread.ListThreadsResponse.new,
  );

  /// DeleteThread deletes a conversation and its semantic history.
  static const deleteThread = connect.Spec(
    '/$name/DeleteThread',
    connect.StreamType.unary,
    realtimememanagercontrolv1thread.DeleteThreadRequest.new,
    realtimememanagercontrolv1thread.DeleteThreadResponse.new,
  );
}
