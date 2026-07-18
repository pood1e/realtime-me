import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/network/remote_session.dart';
import '../../core/state/app_session.dart';
import '../../gen/super_manager/control/v1/device.pb.dart';
import '../../gen/super_manager/control/v1/runtime.pb.dart';
import '../../ui/common.dart';

typedef _StatusData = ({
  List<Runtime> runtimes,
  Map<String, QuotaSnapshot> quotas,
  List<Device> devices,
});

class RuntimeStatusPane extends ConsumerStatefulWidget {
  const RuntimeStatusPane({super.key});

  @override
  ConsumerState<RuntimeStatusPane> createState() => _RuntimeStatusPaneState();
}

class _RuntimeStatusPaneState extends ConsumerState<RuntimeStatusPane> {
  late Future<_StatusData> _status;

  RemoteSession get _session => ref.read(appSessionProvider).requireValue!;

  @override
  void initState() {
    super.initState();
    _status = _load();
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<_StatusData>(
      future: _status,
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
        return RefreshIndicator(
          onRefresh: _refresh,
          child: ListView(
            padding: const EdgeInsets.fromLTRB(16, 16, 16, 96),
            children: [
              _EndpointHeader(session: _session),
              const SizedBox(height: 24),
              _SectionTitle(title: '运行时', count: data.runtimes.length),
              const SizedBox(height: 8),
              for (final runtime in data.runtimes) ...[
                _RuntimeCard(runtime: runtime, quota: data.quotas[runtime.uid]),
                const SizedBox(height: 8),
              ],
              const SizedBox(height: 20),
              _SectionTitle(title: '已配对设备', count: data.devices.length),
              const SizedBox(height: 8),
              Card(
                child: Column(
                  children: [
                    for (
                      var index = 0;
                      index < data.devices.length;
                      index++
                    ) ...[
                      _DeviceTile(
                        device: data.devices[index],
                        currentUid: _session.credentials.deviceUid,
                        onRevoke: () => _revokeDevice(data.devices[index]),
                      ),
                      if (index != data.devices.length - 1)
                        const Divider(height: 1),
                    ],
                  ],
                ),
              ),
              const SizedBox(height: 24),
              OutlinedButton.icon(
                onPressed: _disconnect,
                icon: const Icon(Icons.logout_rounded),
                label: const Text('吊销本机并退出'),
              ),
            ],
          ),
        );
      },
    );
  }

  Future<_StatusData> _load() async {
    final results = await Future.wait<Object>([
      _session.control.listRuntimes(),
      _session.control.listDevices(),
    ]);
    final runtimes = results[0] as List<Runtime>;
    final quotas = await Future.wait(
      runtimes.map((runtime) => _session.control.getRuntimeQuota(runtime.uid)),
    );
    return (
      runtimes: runtimes,
      quotas: {for (final quota in quotas) quota.runtimeUid: quota},
      devices: results[1] as List<Device>,
    );
  }

  Future<void> _refresh() async {
    final next = _load();
    setState(() => _status = next);
    await next;
  }

  Future<void> _revokeDevice(Device device) async {
    if (device.uid == _session.credentials.deviceUid) {
      await _disconnect();
      return;
    }
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('吊销设备？'),
        content: Text('“${device.displayName}”将立即失去访问权限。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('吊销'),
          ),
        ],
      ),
    );
    if (confirmed != true) {
      return;
    }
    try {
      await _session.control.revokeDevice(device.uid);
      await _refresh();
    } on Object catch (error) {
      _showError(error);
    }
  }

  Future<void> _disconnect() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('退出并重新配对？'),
        content: const Text('服务器会吊销本机证书和令牌。再次使用需要在 Linux 主机上生成新的配对码。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('吊销并退出'),
          ),
        ],
      ),
    );
    if (confirmed != true) {
      return;
    }
    try {
      await ref.read(appSessionProvider.notifier).disconnect();
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

class _EndpointHeader extends StatelessWidget {
  final RemoteSession session;

  const _EndpointHeader({required this.session});

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('连接', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 10),
        Card(
          child: ListTile(
            leading: const CircleAvatar(child: Icon(Icons.home_work_outlined)),
            title: SelectableText(session.credentials.serviceUri.toString()),
            subtitle: const Text('DDNS 直连 · mTLS + 设备令牌'),
            trailing: StatusPill(
              label: '已保护',
              color: scheme.primary,
              icon: Icons.lock_rounded,
            ),
          ),
        ),
      ],
    );
  }
}

class _SectionTitle extends StatelessWidget {
  final String title;
  final int count;

  const _SectionTitle({required this.title, required this.count});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: Text(title, style: Theme.of(context).textTheme.titleLarge),
        ),
        Text('$count', style: Theme.of(context).textTheme.labelLarge),
      ],
    );
  }
}

class _RuntimeCard extends StatelessWidget {
  final Runtime runtime;
  final QuotaSnapshot? quota;

  const _RuntimeCard({required this.runtime, required this.quota});

  @override
  Widget build(BuildContext context) {
    final available =
        runtime.availability ==
        RuntimeAvailability.RUNTIME_AVAILABILITY_AVAILABLE;
    final statusColor = available
        ? Theme.of(context).colorScheme.primary
        : Theme.of(context).colorScheme.error;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                CircleAvatar(child: Icon(_runtimeIcon(runtime))),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        runtime.displayName,
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                      Text(
                        'CLI ${runtime.version.isEmpty ? '未知' : runtime.version}',
                      ),
                    ],
                  ),
                ),
                StatusPill(
                  label: available ? '可用' : '不可用',
                  color: statusColor,
                  icon: available
                      ? Icons.check_rounded
                      : Icons.error_outline_rounded,
                ),
              ],
            ),
            if (runtime.diagnostic.isNotEmpty) ...[
              const SizedBox(height: 12),
              Text(runtime.diagnostic),
            ],
            if (quota != null) ...[
              const SizedBox(height: 16),
              _QuotaView(quota: quota!),
            ],
          ],
        ),
      ),
    );
  }
}

class _QuotaView extends StatelessWidget {
  final QuotaSnapshot quota;

  const _QuotaView({required this.quota});

  @override
  Widget build(BuildContext context) {
    if (!quota.hasUsedRatio()) {
      return Text(
        '额度：${_freshnessLabel(quota.freshness)}',
        style: TextStyle(color: Theme.of(context).colorScheme.onSurfaceVariant),
      );
    }
    final percentage = (quota.usedRatio * 100).round();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Expanded(child: Text('额度已使用')),
            Text('$percentage% · ${_freshnessLabel(quota.freshness)}'),
          ],
        ),
        const SizedBox(height: 8),
        LinearProgressIndicator(value: quota.usedRatio.clamp(0, 1).toDouble()),
      ],
    );
  }
}

class _DeviceTile extends StatelessWidget {
  final Device device;
  final String currentUid;
  final VoidCallback onRevoke;

  const _DeviceTile({
    required this.device,
    required this.currentUid,
    required this.onRevoke,
  });

  @override
  Widget build(BuildContext context) {
    final current = device.uid == currentUid;
    final expiry = device.expireTime.toDateTime(toLocal: true);
    final expiryLabel =
        '${expiry.year.toString().padLeft(4, '0')}-${expiry.month.toString().padLeft(2, '0')}-${expiry.day.toString().padLeft(2, '0')}';
    return ListTile(
      minTileHeight: 72,
      leading: const CircleAvatar(child: Icon(Icons.phone_android_rounded)),
      title: Text(device.displayName),
      subtitle: Text(
        current ? '当前设备 · 证书有效至 $expiryLabel' : '证书有效至 $expiryLabel',
      ),
      trailing: IconButton(
        onPressed: onRevoke,
        tooltip: current ? '退出本机' : '吊销设备',
        icon: Icon(
          current ? Icons.logout_rounded : Icons.phonelink_erase_rounded,
        ),
      ),
    );
  }
}

IconData _runtimeIcon(Runtime runtime) {
  return switch (runtime.kind) {
    RuntimeKind.RUNTIME_KIND_CODEX => Icons.code_rounded,
    RuntimeKind.RUNTIME_KIND_CLAUDE_CODE => Icons.auto_awesome_outlined,
    _ => Icons.memory_rounded,
  };
}

String _freshnessLabel(QuotaFreshness freshness) {
  return switch (freshness) {
    QuotaFreshness.QUOTA_FRESHNESS_FRESH => '较新',
    QuotaFreshness.QUOTA_FRESHNESS_STALE => '较旧',
    _ => '不可用',
  };
}
