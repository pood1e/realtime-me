import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/network/remote_session.dart';
import '../../core/state/app_session.dart';
import '../../gen/super_manager/control/v1/runtime.pb.dart';
import '../../gen/super_manager/control/v1/terminal.pb.dart';
import '../../gen/super_manager/control/v1/thread.pb.dart';
import '../../gen/super_manager/control/v1/workspace.pb.dart';
import '../../ui/common.dart';

typedef _WorkspaceData = ({
  Workspace workspace,
  List<Runtime> runtimes,
  List<Thread> threads,
  List<TerminalSession> terminals,
});

class WorkspaceScreen extends ConsumerStatefulWidget {
  final String workspaceUid;

  const WorkspaceScreen({required this.workspaceUid, super.key});

  @override
  ConsumerState<WorkspaceScreen> createState() => _WorkspaceScreenState();
}

class _WorkspaceScreenState extends ConsumerState<WorkspaceScreen> {
  late Future<_WorkspaceData> _data;

  RemoteSession get _session => ref.read(appSessionProvider).requireValue!;

  @override
  void initState() {
    super.initState();
    _data = _load();
  }

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: 2,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('工作区'),
          bottom: const TabBar(
            tabs: [
              Tab(icon: Icon(Icons.forum_outlined), text: '对话'),
              Tab(icon: Icon(Icons.terminal_rounded), text: '终端'),
            ],
          ),
        ),
        body: FutureBuilder<_WorkspaceData>(
          future: _data,
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
            final data = snapshot.data!;
            return Column(
              children: [
                _WorkspaceHeader(workspace: data.workspace),
                Expanded(
                  child: TabBarView(
                    children: [
                      _ThreadsView(
                        data: data,
                        onCreate: () => _createThread(data),
                        onDelete: _deleteThread,
                      ),
                      _TerminalsView(
                        terminals: data.terminals,
                        onCreate: () => _createTerminal(data.workspace),
                        onDelete: _deleteTerminal,
                      ),
                    ],
                  ),
                ),
              ],
            );
          },
        ),
      ),
    );
  }

  Future<_WorkspaceData> _load() async {
    final results = await Future.wait<Object>([
      _session.control.getWorkspace(widget.workspaceUid),
      _session.control.listRuntimes(),
      _session.control.listThreads(widget.workspaceUid),
      _session.control.listTerminalSessions(widget.workspaceUid),
    ]);
    return (
      workspace: results[0] as Workspace,
      runtimes: results[1] as List<Runtime>,
      threads: results[2] as List<Thread>,
      terminals: results[3] as List<TerminalSession>,
    );
  }

  Future<void> _refresh() async {
    final next = _load();
    setState(() => _data = next);
    await next;
  }

  Future<void> _createThread(_WorkspaceData data) async {
    final available = data.runtimes
        .where(
          (runtime) =>
              runtime.availability ==
              RuntimeAvailability.RUNTIME_AVAILABILITY_AVAILABLE,
        )
        .toList();
    if (available.isEmpty) {
      _showError(StateError('没有可用的 Codex 或 Claude Code runtime'));
      return;
    }
    final name = TextEditingController(text: '新对话');
    var selectedRuntime = available.first.uid;
    final value = await showDialog<(String, String)>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          title: const Text('创建 Agent 对话'),
          content: SizedBox(
            width: 420,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: name,
                  autofocus: true,
                  maxLength: 128,
                  decoration: const InputDecoration(labelText: '对话名称'),
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<String>(
                  initialValue: selectedRuntime,
                  decoration: const InputDecoration(labelText: '运行时'),
                  items: [
                    for (final runtime in available)
                      DropdownMenuItem(
                        value: runtime.uid,
                        child: Text(
                          '${runtime.displayName} · ${runtime.version}',
                        ),
                      ),
                  ],
                  onChanged: (value) {
                    if (value != null) {
                      setDialogState(() => selectedRuntime = value);
                    }
                  },
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('取消'),
            ),
            FilledButton(
              onPressed: () {
                if (name.text.trim().isNotEmpty) {
                  Navigator.pop(context, (name.text, selectedRuntime));
                }
              },
              child: const Text('创建'),
            ),
          ],
        ),
      ),
    );
    name.dispose();
    if (value == null || !mounted) {
      return;
    }
    try {
      final thread = await _session.control.createThread(
        workspaceUid: data.workspace.uid,
        runtimeUid: value.$2,
        displayName: value.$1,
      );
      if (mounted) {
        await context.push('/threads/${thread.uid}');
        await _refresh();
      }
    } on Object catch (error) {
      _showError(error);
    }
  }

  Future<void> _deleteThread(Thread thread) async {
    final confirmed = await _confirm(
      title: '删除对话？',
      message: '“${thread.displayName}”及其语义历史会被永久删除。',
    );
    if (!confirmed) {
      return;
    }
    try {
      await _session.control.deleteThread(thread.uid);
      await _refresh();
    } on Object catch (error) {
      _showError(error);
    }
  }

  Future<void> _createTerminal(Workspace workspace) async {
    final name = TextEditingController(text: 'Shell');
    final value = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('创建终端'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (workspace.activeExecutionUid.isNotEmpty) ...[
              Text(
                'Structured Agent 正在写入此工作区。终端可绕过应用的单写者限制，请避免同时修改相同文件。',
                style: TextStyle(color: Theme.of(context).colorScheme.error),
              ),
              const SizedBox(height: 16),
            ],
            TextField(
              controller: name,
              autofocus: true,
              maxLength: 128,
              decoration: const InputDecoration(labelText: '终端名称'),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () {
              if (name.text.trim().isNotEmpty) {
                Navigator.pop(context, name.text);
              }
            },
            child: const Text('创建'),
          ),
        ],
      ),
    );
    name.dispose();
    if (value == null || !mounted) {
      return;
    }
    try {
      final terminal = await _session.control.createTerminalSession(
        workspaceUid: workspace.uid,
        displayName: value,
      );
      if (mounted) {
        await context.push('/terminals/${terminal.uid}');
        await _refresh();
      }
    } on Object catch (error) {
      _showError(error);
    }
  }

  Future<void> _deleteTerminal(TerminalSession terminal) async {
    final confirmed = await _confirm(
      title: '关闭终端？',
      message: '“${terminal.displayName}”中的 shell 和正在运行的进程会被终止。',
    );
    if (!confirmed) {
      return;
    }
    try {
      await _session.control.deleteTerminalSession(terminal.uid);
      await _refresh();
    } on Object catch (error) {
      _showError(error);
    }
  }

  Future<bool> _confirm({
    required String title,
    required String message,
  }) async {
    return await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: Text(title),
            content: Text(message),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context, false),
                child: const Text('取消'),
              ),
              FilledButton(
                onPressed: () => Navigator.pop(context, true),
                child: const Text('确认'),
              ),
            ],
          ),
        ) ??
        false;
  }

  void _showError(Object error) {
    if (mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(readableError(error))));
    }
  }
}

class _WorkspaceHeader extends StatelessWidget {
  final Workspace workspace;

  const _WorkspaceHeader({required this.workspace});

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(20, 16, 20, 14),
      color: Theme.of(context).colorScheme.surfaceContainerLow,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            workspace.displayName,
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 4),
          SelectableText(
            workspace.path,
            style: const TextStyle(fontFamily: 'monospace'),
          ),
          if (workspace.activeExecutionUid.isNotEmpty) ...[
            const SizedBox(height: 10),
            StatusPill(
              label: 'Structured Agent 正在写入',
              color: scheme.tertiary,
              icon: Icons.warning_amber_rounded,
            ),
          ],
        ],
      ),
    );
  }
}

class _ThreadsView extends StatelessWidget {
  final _WorkspaceData data;
  final VoidCallback onCreate;
  final ValueChanged<Thread> onDelete;

  const _ThreadsView({
    required this.data,
    required this.onCreate,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    if (data.threads.isEmpty) {
      return EmptyState(
        icon: Icons.forum_outlined,
        title: '还没有 Agent 对话',
        message: '选择已安装的 Codex 或 Claude Code，创建结构化对话。',
        action: FilledButton.icon(
          onPressed: onCreate,
          icon: const Icon(Icons.add_comment_outlined),
          label: const Text('新建对话'),
        ),
      );
    }
    final runtimeNames = {
      for (final runtime in data.runtimes) runtime.uid: runtime.displayName,
    };
    return ListView.separated(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 88),
      itemCount: data.threads.length + 1,
      separatorBuilder: (_, _) => const SizedBox(height: 8),
      itemBuilder: (context, index) {
        if (index == 0) {
          return Align(
            alignment: Alignment.centerRight,
            child: FilledButton.tonalIcon(
              onPressed: onCreate,
              icon: const Icon(Icons.add_rounded),
              label: const Text('新建对话'),
            ),
          );
        }
        final thread = data.threads[index - 1];
        return Card(
          child: ListTile(
            minTileHeight: 78,
            leading: CircleAvatar(child: Icon(_threadIcon(thread.state))),
            title: Text(thread.displayName),
            subtitle: Text(runtimeNames[thread.runtimeUid] ?? '未知运行时'),
            trailing: PopupMenuButton<String>(
              tooltip: '对话操作',
              onSelected: (value) {
                if (value == 'delete') {
                  onDelete(thread);
                }
              },
              itemBuilder: (_) => const [
                PopupMenuItem(value: 'delete', child: Text('删除对话')),
              ],
            ),
            onTap: () => context.push('/threads/${thread.uid}'),
          ),
        );
      },
    );
  }
}

class _TerminalsView extends StatelessWidget {
  final List<TerminalSession> terminals;
  final VoidCallback onCreate;
  final ValueChanged<TerminalSession> onDelete;

  const _TerminalsView({
    required this.terminals,
    required this.onCreate,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    if (terminals.isEmpty) {
      return EmptyState(
        icon: Icons.terminal_rounded,
        title: '还没有终端',
        message: '终端由 tmux 持有，应用断线或 API 重启后仍可重新附着。',
        action: FilledButton.icon(
          onPressed: onCreate,
          icon: const Icon(Icons.add_rounded),
          label: const Text('创建终端'),
        ),
      );
    }
    return ListView.separated(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 88),
      itemCount: terminals.length + 1,
      separatorBuilder: (_, _) => const SizedBox(height: 8),
      itemBuilder: (context, index) {
        if (index == 0) {
          return Align(
            alignment: Alignment.centerRight,
            child: FilledButton.tonalIcon(
              onPressed: onCreate,
              icon: const Icon(Icons.add_rounded),
              label: const Text('创建终端'),
            ),
          );
        }
        final terminal = terminals[index - 1];
        final running =
            terminal.state ==
            TerminalSessionState.TERMINAL_SESSION_STATE_RUNNING;
        return Card(
          child: ListTile(
            minTileHeight: 78,
            leading: const CircleAvatar(child: Icon(Icons.terminal_rounded)),
            title: Text(terminal.displayName),
            subtitle: Text(terminal.cwd, overflow: TextOverflow.ellipsis),
            trailing: PopupMenuButton<String>(
              tooltip: '终端操作',
              onSelected: (value) {
                if (value == 'delete') {
                  onDelete(terminal);
                }
              },
              itemBuilder: (_) => const [
                PopupMenuItem(value: 'delete', child: Text('关闭终端')),
              ],
            ),
            enabled: running,
            onTap: running
                ? () => context.push('/terminals/${terminal.uid}')
                : null,
          ),
        );
      },
    );
  }
}

IconData _threadIcon(ThreadState state) {
  return switch (state) {
    ThreadState.THREAD_STATE_RUNNING => Icons.sync_rounded,
    ThreadState.THREAD_STATE_INPUT_REQUIRED => Icons.help_outline_rounded,
    ThreadState.THREAD_STATE_LOST => Icons.error_outline_rounded,
    _ => Icons.chat_bubble_outline_rounded,
  };
}
