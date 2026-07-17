//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/terminal.proto
//

import "package:connectrpc/connect.dart" as connect;
import "terminal.pb.dart" as super_managercontrolv1terminal;

/// TerminalService manages tmux-backed raw shell sessions.
abstract final class TerminalService {
  /// Fully-qualified name of the TerminalService service.
  static const name = 'super_manager.control.v1.TerminalService';

  /// CreateTerminalSession creates a raw shell owned by Super Manager.
  static const createTerminalSession = connect.Spec(
    '/$name/CreateTerminalSession',
    connect.StreamType.unary,
    super_managercontrolv1terminal.CreateTerminalSessionRequest.new,
    super_managercontrolv1terminal.CreateTerminalSessionResponse.new,
  );

  /// GetTerminalSession returns one raw shell session.
  static const getTerminalSession = connect.Spec(
    '/$name/GetTerminalSession',
    connect.StreamType.unary,
    super_managercontrolv1terminal.GetTerminalSessionRequest.new,
    super_managercontrolv1terminal.GetTerminalSessionResponse.new,
  );

  /// ListTerminalSessions returns raw shell sessions in one workspace.
  static const listTerminalSessions = connect.Spec(
    '/$name/ListTerminalSessions',
    connect.StreamType.unary,
    super_managercontrolv1terminal.ListTerminalSessionsRequest.new,
    super_managercontrolv1terminal.ListTerminalSessionsResponse.new,
  );

  /// DeleteTerminalSession closes the shell and removes its record.
  static const deleteTerminalSession = connect.Spec(
    '/$name/DeleteTerminalSession',
    connect.StreamType.unary,
    super_managercontrolv1terminal.DeleteTerminalSessionRequest.new,
    super_managercontrolv1terminal.DeleteTerminalSessionResponse.new,
  );
}
