import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/state/app_session.dart';
import '../../ui/common.dart';
import '../status/status_pane.dart';
import '../status/status_settings.dart';
import 'runtime_status_pane.dart';
import 'terminal_sessions_pane.dart';
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
    final session = ref.watch(appSessionProvider).value;
    final destination = _destinations[_index];
    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            Icon(
              destination.selectedIcon,
              color: Theme.of(context).colorScheme.primary,
            ),
            const SizedBox(width: 10),
            Text(destination.title),
          ],
        ),
        actions: [
          if (session != null) ...[
            Padding(
              padding: const EdgeInsets.only(right: 8),
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
                    title: Text('仅清除 Manager 凭据'),
                    contentPadding: EdgeInsets.zero,
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
      body: IndexedStack(
        index: _index,
        children: [
          const StatusPane(),
          const _ManagerGate(child: WorkspacesPane()),
          _ManagerGate(
            child: TerminalSessionsPane(
              onOpenAgents: () => setState(() => _index = 1),
            ),
          ),
          const _ManagerGate(
            disconnected: _DisconnectedSettingsPane(),
            child: RuntimeStatusPane(header: StatusSettingsSection()),
          ),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (value) => setState(() => _index = value),
        destinations: _destinations
            .map(
              (destination) => NavigationDestination(
                icon: Icon(destination.icon),
                selectedIcon: Icon(destination.selectedIcon),
                label: destination.label,
              ),
            )
            .toList(growable: false),
      ),
    );
  }

  Future<void> _confirmForgetLocal() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('仅清除 Manager 凭据？'),
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

const _destinations = [
  (
    title: 'Realtime Status',
    label: '状态',
    icon: Icons.monitor_heart_outlined,
    selectedIcon: Icons.monitor_heart_rounded,
  ),
  (
    title: 'Agent 工作区',
    label: 'Agent',
    icon: Icons.auto_awesome_outlined,
    selectedIcon: Icons.auto_awesome_rounded,
  ),
  (
    title: '远程终端',
    label: '终端',
    icon: Icons.terminal_outlined,
    selectedIcon: Icons.terminal_rounded,
  ),
  (
    title: '设置',
    label: '设置',
    icon: Icons.settings_outlined,
    selectedIcon: Icons.settings_rounded,
  ),
];

class _ManagerGate extends ConsumerWidget {
  final Widget child;
  final Widget disconnected;

  const _ManagerGate({
    required this.child,
    this.disconnected = const _ManagerConnectPane(),
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return ref
        .watch(appSessionProvider)
        .when(
          data: (session) => session == null ? disconnected : child,
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (error, _) => ErrorState(
            message: readableError(error),
            onRetry: () => ref.invalidate(appSessionProvider),
          ),
        );
  }
}

class _ManagerConnectPane extends StatelessWidget {
  const _ManagerConnectPane();

  @override
  Widget build(BuildContext context) {
    return EmptyState(
      icon: Icons.phonelink_lock_outlined,
      title: '尚未连接 Manager',
      message: '扫描 Linux 主机生成的一次性配对码，启用 Agent 与远程终端。',
      action: FilledButton.icon(
        onPressed: () => context.push('/pairing'),
        icon: const Icon(Icons.qr_code_scanner_rounded),
        label: const Text('连接工作机'),
      ),
    );
  }
}

class _DisconnectedSettingsPane extends StatelessWidget {
  const _DisconnectedSettingsPane();

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 96),
      children: [
        const StatusSettingsSection(),
        const SizedBox(height: 24),
        Text('Manager', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 10),
        Card(
          child: ListTile(
            minTileHeight: 80,
            leading: const CircleAvatar(
              child: Icon(Icons.phonelink_lock_outlined),
            ),
            title: const Text('连接工作机'),
            subtitle: const Text('启用 Agent、终端和运行时管理'),
            trailing: const Icon(Icons.chevron_right_rounded),
            onTap: () => context.push('/pairing'),
          ),
        ),
      ],
    );
  }
}
