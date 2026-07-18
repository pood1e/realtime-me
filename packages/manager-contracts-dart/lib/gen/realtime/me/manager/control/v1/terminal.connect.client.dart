//
//  Generated code. Do not modify.
//  source: realtime/me/manager/control/v1/terminal.proto
//

import "package:connectrpc/connect.dart" as connect;
import "terminal.pb.dart" as realtimememanagercontrolv1terminal;
import "terminal.connect.spec.dart" as specs;

/// TerminalService manages tmux-backed raw shell sessions.
extension type TerminalServiceClient (connect.Transport _transport) {
  /// CreateTerminalSession creates a raw shell owned by Super Manager.
  Future<realtimememanagercontrolv1terminal.CreateTerminalSessionResponse> createTerminalSession(
    realtimememanagercontrolv1terminal.CreateTerminalSessionRequest input, {
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
  Future<realtimememanagercontrolv1terminal.GetTerminalSessionResponse> getTerminalSession(
    realtimememanagercontrolv1terminal.GetTerminalSessionRequest input, {
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
  Future<realtimememanagercontrolv1terminal.ListTerminalSessionsResponse> listTerminalSessions(
    realtimememanagercontrolv1terminal.ListTerminalSessionsRequest input, {
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
  Future<realtimememanagercontrolv1terminal.DeleteTerminalSessionResponse> deleteTerminalSession(
    realtimememanagercontrolv1terminal.DeleteTerminalSessionRequest input, {
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
