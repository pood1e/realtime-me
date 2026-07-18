import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/network/remote_session.dart';
import '../../core/state/app_session.dart';
import 'package:realtime_me_manager_contracts/gen/realtime/me/manager/control/v1/workspace.pb.dart';
import '../../ui/common.dart';

class WorkspacesPane extends ConsumerStatefulWidget {
  const WorkspacesPane({super.key});

  @override
  ConsumerState<WorkspacesPane> createState() => _WorkspacesPaneState();
}

class _WorkspacesPaneState extends ConsumerState<WorkspacesPane> {
  late Future<List<Workspace>> _workspaces;

  RemoteSession get _session => ref.read(appSessionProvider).requireValue!;

  @override
  void initState() {
    super.initState();
    _workspaces = _session.control.listWorkspaces();
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<List<Workspace>>(
      future: _workspaces,
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
        final workspaces = snapshot.data!;
        if (workspaces.isEmpty) {
          return EmptyState(
            icon: Icons.create_new_folder_outlined,
            title: '还没有工作区',
            message: '登记 Linux 主机上的项目绝对路径，之后可创建 Agent 对话和终端。',
            action: FilledButton.icon(
              onPressed: _createWorkspace,
              icon: const Icon(Icons.add_rounded),
              label: const Text('添加工作区'),
            ),
          );
        }
        return RefreshIndicator(
          onRefresh: _refresh,
          child: ListView.separated(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 96),
            itemCount: workspaces.length + 1,
            separatorBuilder: (_, _) => const SizedBox(height: 8),
            itemBuilder: (context, index) {
              if (index == 0) {
                return Padding(
                  padding: const EdgeInsets.fromLTRB(8, 4, 8, 12),
                  child: Row(
                    children: [
                      Expanded(
                        child: Text(
                          '项目',
                          style: Theme.of(context).textTheme.headlineSmall,
                        ),
                      ),
                      IconButton.filledTonal(
                        onPressed: _createWorkspace,
                        tooltip: '添加工作区',
                        icon: const Icon(Icons.add_rounded),
                      ),
                    ],
                  ),
                );
              }
              final workspace = workspaces[index - 1];
              return Card(
                child: ListTile(
                  minTileHeight: 84,
                  leading: CircleAvatar(
                    backgroundColor: Theme.of(
                      context,
                    ).colorScheme.secondaryContainer,
                    child: const Icon(Icons.folder_outlined),
                  ),
                  title: Text(workspace.displayName),
                  subtitle: Text(
                    workspace.path,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(fontFamily: 'monospace'),
                  ),
                  trailing: workspace.activeExecutionUid.isEmpty
                      ? const Icon(Icons.chevron_right_rounded)
                      : StatusPill(
                          label: '运行中',
                          color: Theme.of(context).colorScheme.tertiary,
                          icon: Icons.sync_rounded,
                        ),
                  onTap: () => context.push('/workspaces/${workspace.uid}'),
                  onLongPress: () => _deleteWorkspace(workspace),
                ),
              );
            },
          ),
        );
      },
    );
  }

  Future<void> _refresh() async {
    final next = _session.control.listWorkspaces();
    setState(() => _workspaces = next);
    await next;
  }

  Future<void> _createWorkspace() async {
    final name = TextEditingController();
    final path = TextEditingController();
    final value = await showDialog<(String, String)>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('添加工作区'),
        content: SizedBox(
          width: 420,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: name,
                autofocus: true,
                maxLength: 128,
                decoration: const InputDecoration(labelText: '名称'),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: path,
                autocorrect: false,
                decoration: const InputDecoration(
                  labelText: 'Linux 绝对路径',
                  hintText: '/srv/workspaces/project',
                ),
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
              if (name.text.trim().isNotEmpty &&
                  path.text.trim().startsWith('/')) {
                Navigator.pop(context, (name.text, path.text));
              }
            },
            child: const Text('添加'),
          ),
        ],
      ),
    );
    name.dispose();
    path.dispose();
    if (value == null || !mounted) {
      return;
    }
    try {
      await _session.control.createWorkspace(value.$1, value.$2);
      await _refresh();
    } on Object catch (error) {
      _showError(error);
    }
  }

  Future<void> _deleteWorkspace(Workspace workspace) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('移除工作区？'),
        content: Text('只移除“${workspace.displayName}”的登记，不会删除服务器文件。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('移除'),
          ),
        ],
      ),
    );
    if (confirmed != true) {
      return;
    }
    try {
      await _session.control.deleteWorkspace(workspace.uid);
      await _refresh();
    } on Object catch (error) {
      _showError(error);
    }
  }

  void _showError(Object error) {
    if (mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(readableError(error))));
    }
  }
}
