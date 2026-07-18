//
//  Generated code. Do not modify.
//  source: realtime/me/manager/control/v1/terminal.proto
//

import "package:connectrpc/connect.dart" as connect;
import "terminal.pb.dart" as realtimememanagercontrolv1terminal;

/// TerminalService manages tmux-backed raw shell sessions.
abstract final class TerminalService {
  /// Fully-qualified name of the TerminalService service.
  static const name = 'realtime.me.manager.control.v1.TerminalService';

  /// CreateTerminalSession creates a raw shell owned by Super Manager.
  static const createTerminalSession = connect.Spec(
    '/$name/CreateTerminalSession',
    connect.StreamType.unary,
    realtimememanagercontrolv1terminal.CreateTerminalSessionRequest.new,
    realtimememanagercontrolv1terminal.CreateTerminalSessionResponse.new,
  );

  /// GetTerminalSession returns one raw shell session.
  static const getTerminalSession = connect.Spec(
    '/$name/GetTerminalSession',
    connect.StreamType.unary,
    realtimememanagercontrolv1terminal.GetTerminalSessionRequest.new,
    realtimememanagercontrolv1terminal.GetTerminalSessionResponse.new,
  );

  /// ListTerminalSessions returns raw shell sessions in one workspace.
  static const listTerminalSessions = connect.Spec(
    '/$name/ListTerminalSessions',
    connect.StreamType.unary,
    realtimememanagercontrolv1terminal.ListTerminalSessionsRequest.new,
    realtimememanagercontrolv1terminal.ListTerminalSessionsResponse.new,
  );

  /// DeleteTerminalSession closes the shell and removes its record.
  static const deleteTerminalSession = connect.Spec(
    '/$name/DeleteTerminalSession',
    connect.StreamType.unary,
    realtimememanagercontrolv1terminal.DeleteTerminalSessionRequest.new,
    realtimememanagercontrolv1terminal.DeleteTerminalSessionResponse.new,
  );
}
