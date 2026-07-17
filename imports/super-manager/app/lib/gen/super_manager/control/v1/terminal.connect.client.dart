//
//  Generated code. Do not modify.
//  source: super_manager/control/v1/terminal.proto
//

import "package:connectrpc/connect.dart" as connect;
import "terminal.pb.dart" as super_managercontrolv1terminal;
import "terminal.connect.spec.dart" as specs;

/// TerminalService manages tmux-backed raw shell sessions.
extension type TerminalServiceClient (connect.Transport _transport) {
  /// CreateTerminalSession creates a raw shell owned by Super Manager.
  Future<super_managercontrolv1terminal.CreateTerminalSessionResponse> createTerminalSession(
    super_managercontrolv1terminal.CreateTerminalSessionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TerminalService.createTerminalSession,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// GetTerminalSession returns one raw shell session.
  Future<super_managercontrolv1terminal.GetTerminalSessionResponse> getTerminalSession(
    super_managercontrolv1terminal.GetTerminalSessionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TerminalService.getTerminalSession,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// ListTerminalSessions returns raw shell sessions in one workspace.
  Future<super_managercontrolv1terminal.ListTerminalSessionsResponse> listTerminalSessions(
    super_managercontrolv1terminal.ListTerminalSessionsRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TerminalService.listTerminalSessions,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }

  /// DeleteTerminalSession closes the shell and removes its record.
  Future<super_managercontrolv1terminal.DeleteTerminalSessionResponse> deleteTerminalSession(
    super_managercontrolv1terminal.DeleteTerminalSessionRequest input, {
    connect.Headers? headers,
    connect.AbortSignal? signal,
    Function(connect.Headers)? onHeader,
    Function(connect.Headers)? onTrailer,
  }) {
    return connect.Client(_transport).unary(
      specs.TerminalService.deleteTerminalSession,
      input,
      signal: signal,
      headers: headers,
      onHeader: onHeader,
      onTrailer: onTrailer,
    );
  }
}
