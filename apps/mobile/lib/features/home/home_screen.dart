import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/state/app_session.dart';
import 'runtime_status_pane.dart';
import 'workspaces_pane.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  int _index = 0;

  @override
  Widget build(BuildContext context) {
    final session = ref.watch(appSessionProvider).value!;
    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            Icon(
              Icons.terminal_rounded,
              color: Theme.of(context).colorScheme.primary,
            ),
            const SizedBox(width: 10),
            const Text('Super Manager'),
          ],
        ),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 16),
            child: Tooltip(
              message: session.credentials.serviceUri.toString(),
              child: const Icon(Icons.lock_rounded, size: 20),
            ),
          ),
          PopupMenuButton<_HomeAction>(
            tooltip: '连接设置',
            onSelected: (action) {
              if (action == _HomeAction.forgetLocal) {
                _confirmForgetLocal();
              }
            },
            itemBuilder: (context) => const [
              PopupMenuItem(
                value: _HomeAction.forgetLocal,
                child: ListTile(
                  leading: Icon(Icons.restart_alt_rounded),
                  title: Text('仅清除本机凭据'),
                  contentPadding: EdgeInsets.zero,
                ),
              ),
            ],
          ),
        ],
      ),
      body: IndexedStack(
        index: _index,
        children: const [WorkspacesPane(), RuntimeStatusPane()],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (value) => setState(() => _index = value),
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.folder_outlined),
            selectedIcon: Icon(Icons.folder_rounded),
            label: '工作区',
          ),
          NavigationDestination(
            icon: Icon(Icons.monitor_heart_outlined),
            selectedIcon: Icon(Icons.monitor_heart_rounded),
            label: '状态',
          ),
        ],
      ),
    );
  }

  Future<void> _confirmForgetLocal() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('仅清除本机凭据？'),
        content: const Text(
          '仅在证书过期、设备已被吊销或服务器无法连接时使用。服务器上的设备令牌不会被吊销，之后应在 Linux 主机执行 smctl device revoke。',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('清除'),
          ),
        ],
      ),
    );
    if (confirmed == true && mounted) {
      await ref.read(appSessionProvider.notifier).forgetLocalCredentials();
    }
  }
}

enum _HomeAction { forgetLocal }
