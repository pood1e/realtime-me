import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:realtime_me_status_contracts/gen/realtime/me/status/v1/watch.pb.dart';

import '../../ui/common.dart';
import 'status_settings.dart';
import 'status_state.dart';

class StatusPane extends ConsumerWidget {
  const StatusPane({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return ref
        .watch(statusStateProvider)
        .when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (error, _) => ErrorState(
            message: readableError(error),
            onRetry: () => ref.read(statusStateProvider.notifier).retry(),
          ),
          data: (state) => _StatusContent(state: state),
        );
  }
}

class _StatusContent extends ConsumerWidget {
  final StatusState state;

  const _StatusContent({required this.state});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final horizontal = math.max(16.0, (constraints.maxWidth - 760) / 2);
        return RefreshIndicator(
          onRefresh: ref.read(statusStateProvider.notifier).refresh,
          child: ListView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: EdgeInsets.fromLTRB(horizontal, 16, horizontal, 96),
            children: [
              _SyncBanner(state: state),
              const SizedBox(height: 24),
              Row(
                children: [
                  Expanded(
                    child: Text(
                      'Wear OS',
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                  ),
                  if (state.receivedAt case final receivedAt?)
                    Text(
                      _relativeTime(receivedAt),
                      style: Theme.of(context).textTheme.labelLarge?.copyWith(
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      ),
                    ),
                ],
              ),
              const SizedBox(height: 10),
              if (state.watch case final watch?)
                _WatchSnapshotView(watch: watch, receivedAt: state.receivedAt)
              else
                const _NoWatchSnapshot(),
            ],
          ),
        );
      },
    );
  }
}

class _SyncBanner extends ConsumerWidget {
  final StatusState state;

  const _SyncBanner({required this.state});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final scheme = Theme.of(context).colorScheme;
    final configured = state.hasToken;
    final ready = configured && state.hasRequiredPermissions;
    return Card(
      color: ready ? scheme.primaryContainer : scheme.tertiaryContainer,
      child: Padding(
        padding: const EdgeInsets.all(18),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(
              ready ? Icons.cloud_done_rounded : Icons.cloud_off_rounded,
              color: ready
                  ? scheme.onPrimaryContainer
                  : scheme.onTertiaryContainer,
              size: 28,
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    configured ? '原生后台同步已启用' : '尚未连接 Status Gateway',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 4),
                  Text(
                    configured
                        ? '关闭 Flutter 后，Wear 数据仍由 Android 服务持续上报。'
                        : '配置 Ingest Token 后，手机和手表状态会在后台持续上报。',
                  ),
                  if (!configured) ...[
                    const SizedBox(height: 14),
                    FilledButton.icon(
                      onPressed: () => showStatusTokenDialog(context, ref),
                      icon: const Icon(Icons.key_rounded),
                      label: const Text('配置 Token'),
                    ),
                  ] else if (!state.hasRequiredPermissions) ...[
                    const SizedBox(height: 14),
                    FilledButton.tonalIcon(
                      onPressed: () => requestStatusPermissions(context, ref),
                      icon: const Icon(Icons.notifications_active_outlined),
                      label: const Text('授权系统权限'),
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _WatchSnapshotView extends StatelessWidget {
  final WatchSnapshot watch;
  final DateTime? receivedAt;

  const _WatchSnapshotView({required this.watch, required this.receivedAt});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        LayoutBuilder(
          builder: (context, constraints) {
            final width = constraints.maxWidth >= 640
                ? (constraints.maxWidth - 16) / 3
                : (constraints.maxWidth - 8) / 2;
            return Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                SizedBox(
                  width: width,
                  child: _MetricCard(
                    icon: Icons.favorite_rounded,
                    label: '心率',
                    value: watch.hasHeartRate()
                        ? '${watch.heartRate.beatsPerMinute}'
                        : '—',
                    unit: watch.hasHeartRate() ? 'BPM' : null,
                  ),
                ),
                SizedBox(
                  width: width,
                  child: _MetricCard(
                    icon: Icons.directions_walk_rounded,
                    label: '今日步数',
                    value: watch.hasActivityTotals()
                        ? '${watch.activityTotals.steps}'
                        : '—',
                    unit: watch.hasActivityTotals() ? '步' : null,
                  ),
                ),
                SizedBox(
                  width: width,
                  child: _MetricCard(
                    icon:
                        watch.hasWatchState() &&
                            watch.watchState.chargeState ==
                                ChargeState.CHARGE_STATE_CHARGING
                        ? Icons.battery_charging_full_rounded
                        : Icons.battery_5_bar_rounded,
                    label: '手表电量',
                    value: watch.hasWatchState()
                        ? '${watch.watchState.batteryPercent}'
                        : '—',
                    unit: watch.hasWatchState() ? '%' : null,
                  ),
                ),
              ],
            );
          },
        ),
        const SizedBox(height: 8),
        Card(
          child: ListTile(
            minTileHeight: 84,
            leading: CircleAvatar(
              backgroundColor: Theme.of(context).colorScheme.secondaryContainer,
              child: const Icon(Icons.watch_rounded),
            ),
            title: Text(
              watch.hasDeviceInfo() && watch.deviceInfo.displayName.isNotEmpty
                  ? watch.deviceInfo.displayName
                  : 'Wear OS 手表',
            ),
            subtitle: Text(_watchDetails(watch, receivedAt)),
            trailing: const Icon(Icons.bluetooth_connected_rounded),
          ),
        ),
      ],
    );
  }
}

class _MetricCard extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  final String? unit;

  const _MetricCard({
    required this.icon,
    required this.label,
    required this.value,
    required this.unit,
  });

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, color: scheme.primary),
            const SizedBox(height: 16),
            Text(label, style: Theme.of(context).textTheme.labelLarge),
            const SizedBox(height: 4),
            Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Flexible(
                  child: Text(
                    value,
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      fontFeatures: const [FontFeature.tabularFigures()],
                    ),
                  ),
                ),
                if (unit != null) ...[
                  const SizedBox(width: 4),
                  Padding(
                    padding: const EdgeInsets.only(bottom: 4),
                    child: Text(unit!),
                  ),
                ],
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _NoWatchSnapshot extends StatelessWidget {
  const _NoWatchSnapshot();

  @override
  Widget build(BuildContext context) {
    return const Card(
      child: Padding(
        padding: EdgeInsets.symmetric(vertical: 32),
        child: EmptyState(
          icon: Icons.watch_off_outlined,
          title: '等待手表数据',
          message: '确认手机与 Wear OS 手表已配对，然后下拉刷新。',
        ),
      ),
    );
  }
}

String _watchDetails(WatchSnapshot watch, DateTime? receivedAt) {
  final details = <String>[];
  if (watch.hasDeviceInfo() && watch.deviceInfo.model.isNotEmpty) {
    details.add(watch.deviceInfo.model);
  }
  if (receivedAt != null) {
    details.add('${_relativeTime(receivedAt)}收到');
  }
  return details.isEmpty ? '最近的 Data Layer 快照' : details.join(' · ');
}

String _relativeTime(DateTime time) {
  final elapsed = DateTime.now().difference(time);
  if (elapsed.isNegative || elapsed.inSeconds < 30) {
    return '刚刚';
  }
  if (elapsed.inMinutes < 1) {
    return '${elapsed.inSeconds} 秒前';
  }
  if (elapsed.inHours < 1) {
    return '${elapsed.inMinutes} 分钟前';
  }
  if (elapsed.inDays < 1) {
    return '${elapsed.inHours} 小时前';
  }
  return '${elapsed.inDays} 天前';
}
