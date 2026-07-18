//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/thread.proto
//

import "package:connectrpc/connect.dart" as connect;
import "thread.pb.dart" as super_managercontrolv1thread;

/// ThreadService manages durable structured conversations.
abstract final class ThreadService {
  /// Fully-qualified name of the ThreadService service.
  static const name = 'super_manager.control.v1.ThreadService';

  /// CreateThread creates a conversation for one workspace and runtime.
  static const createThread = connect.Spec(
    '/$name/CreateThread',
    connect.StreamType.unary,
    super_managercontrolv1thread.CreateThreadRequest.new,
    super_managercontrolv1thread.CreateThreadResponse.new,
  );

  /// GetThread returns one conversation.
  static const getThread = connect.Spec(
    '/$name/GetThread',
    connect.StreamType.unary,
    super_managercontrolv1thread.GetThreadRequest.new,
    super_managercontrolv1thread.GetThreadResponse.new,
  );

  /// ListThreads returns conversations in one workspace.
  static const listThreads = connect.Spec(
    '/$name/ListThreads',
    connect.StreamType.unary,
    super_managercontrolv1thread.ListThreadsRequest.new,
    super_managercontrolv1thread.ListThreadsResponse.new,
  );

  /// DeleteThread deletes a conversation and its semantic history.
  static const deleteThread = connect.Spec(
    '/$name/DeleteThread',
    connect.StreamType.unary,
    super_managercontrolv1thread.DeleteThreadRequest.new,
    super_managercontrolv1thread.DeleteThreadResponse.new,
  );
}
