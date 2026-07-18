import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:realtime_me_manager_contracts/gen/realtime/me/manager/control/v1/terminal.pb.dart';
import 'package:realtime_me_manager_contracts/gen/realtime/me/manager/control/v1/workspace.pb.dart';

import '../../core/network/remote_session.dart';
import '../../core/state/app_session.dart';
import '../../ui/common.dart';

typedef _TerminalEntry = ({Workspace workspace, TerminalSession terminal});

class TerminalSessionsPane extends ConsumerStatefulWidget {
  final VoidCallback onOpenAgents;

  const TerminalSessionsPane({required this.onOpenAgents, super.key});

  @override
  ConsumerState<TerminalSessionsPane> createState() =>
      _TerminalSessionsPaneState();
}

class _TerminalSessionsPaneState extends ConsumerState<TerminalSessionsPane> {
  late Future<List<_TerminalEntry>> _terminals;

  RemoteSession get _session => ref.read(appSessionProvider).requireValue!;

  @override
  void initState() {
    super.initState();
    _terminals = _load();
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<_TerminalEntry>>(
      future: _terminals,
      builder: (context, snapshot) {
        if (snapshot.connectionState != ConnectionState.done) {
          return const Center(child: CircularProgressIndicator());
        }
        if (snapshot.hasError) {
          return ErrorState(
            message: readableError(snapshot.error!),
            onRetry: _refresh,
          );
        }
        final terminals = snapshot.data!;
        if (terminals.isEmpty) {
          return EmptyState(
            icon: Icons.terminal_rounded,
            title: '还没有终端',
            message: '从 Agent 页进入工作区，即可创建 tmux 终端。',
            action: FilledButton.icon(
              onPressed: widget.onOpenAgents,
              icon: const Icon(Icons.folder_open_rounded),
              label: const Text('打开 Agent'),
            ),
          );
        }
        return RefreshIndicator(
          onRefresh: _refresh,
          child: ListView.separated(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 96),
            itemCount: terminals.length + 1,
            separatorBuilder: (_, _) => const SizedBox(height: 8),
            itemBuilder: (context, index) {
              if (index == 0) {
                return Padding(
                  padding: const EdgeInsets.fromLTRB(8, 4, 8, 12),
                  child: Row(
                    children: [
                      Expanded(
                        child: Text(
                          '终端',
                          style: Theme.of(context).textTheme.headlineSmall,
                        ),
                      ),
                      Text(
                        '${terminals.length}',
                        style: Theme.of(context).textTheme.labelLarge,
                      ),
                    ],
                  ),
                );
              }
              final entry = terminals[index - 1];
              final running =
                  entry.terminal.state ==
                  TerminalSessionState.TERMINAL_SESSION_STATE_RUNNING;
              return Card(
                child: ListTile(
                  minTileHeight: 84,
                  leading: const CircleAvatar(
                    child: Icon(Icons.terminal_rounded),
                  ),
                  title: Text(entry.terminal.displayName),
                  subtitle: Text(
                    '${entry.workspace.displayName} · ${entry.terminal.cwd}',
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  trailing: StatusPill(
                    label: running ? '运行中' : '已关闭',
                    color: running
                        ? Theme.of(context).colorScheme.primary
                        : Theme.of(context).colorScheme.outline,
                    icon: running
                        ? Icons.play_arrow_rounded
                        : Icons.stop_rounded,
                  ),
                  onTap: running
                      ? () => context.push('/terminals/${entry.terminal.uid}')
                      : null,
                ),
              );
            },
          ),
        );
      },
    );
  }

  Future<List<_TerminalEntry>> _load() async {
    final workspaces = await _session.control.listWorkspaces();
    final lists = await Future.wait(workspaces.map(_loadWorkspaceTerminals));
    return lists.expand((entries) => entries).toList(growable: false);
  }

  Future<List<_TerminalEntry>> _loadWorkspaceTerminals(
    Workspace workspace,
  ) async {
    final terminals = await _session.control.listTerminalSessions(
      workspace.uid,
    );
    return terminals
        .map((terminal) => (workspace: workspace, terminal: terminal))
        .toList(growable: false);
  }

  Future<void> _refresh() async {
    final next = _load();
    setState(() => _terminals = next);
    await next;
  }
}
