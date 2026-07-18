import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../ui/common.dart';
import 'status_state.dart';

class StatusSettingsSection extends ConsumerWidget {
  const StatusSettingsSection({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncState = ref.watch(statusStateProvider);
    final state = asyncState.value;
    final scheme = Theme.of(context).colorScheme;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Status 上报', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 10),
        Card(
          child: Column(
            children: [
              ListTile(
                minTileHeight: 72,
                leading: const CircleAvatar(
                  child: Icon(Icons.cloud_upload_outlined),
                ),
                title: const Text('Gateway Token'),
                subtitle: Text(
                  state?.hasToken == true
                      ? '已加密保存在 Android Keystore'
                      : '配置后由原生服务持续上报',
                ),
                trailing: StatusPill(
                  label: state?.hasToken == true ? '已配置' : '未配置',
                  color: state?.hasToken == true
                      ? scheme.primary
                      : scheme.error,
                  icon: state?.hasToken == true
                      ? Icons.check_rounded
                      : Icons.key_off_rounded,
                ),
                onTap: asyncState.isLoading
                    ? null
                    : () => showStatusTokenDialog(context, ref),
              ),
              const Divider(height: 1),
              ListTile(
                minTileHeight: 72,
                leading: const CircleAvatar(
                  child: Icon(Icons.notifications_active_outlined),
                ),
                title: const Text('系统权限'),
                subtitle: const Text('通知用于前台同步，蓝牙用于附件状态'),
                trailing: StatusPill(
                  label: state?.hasRequiredPermissions == true ? '已授权' : '待授权',
                  color: state?.hasRequiredPermissions == true
                      ? scheme.primary
                      : scheme.tertiary,
                  icon: state?.hasRequiredPermissions == true
                      ? Icons.check_rounded
                      : Icons.warning_amber_rounded,
                ),
                onTap: state?.hasRequiredPermissions == false
                    ? () => requestStatusPermissions(context, ref)
                    : null,
              ),
            ],
          ),
        ),
        if (state?.hasToken == true) ...[
          const SizedBox(height: 8),
          Align(
            alignment: Alignment.centerRight,
            child: TextButton.icon(
              onPressed: () => clearStatusToken(context, ref),
              icon: const Icon(Icons.link_off_rounded),
              label: const Text('停止 Status 上报'),
            ),
          ),
        ],
      ],
    );
  }
}

Future<void> showStatusTokenDialog(BuildContext context, WidgetRef ref) async {
  final controller = TextEditingController();
  final formKey = GlobalKey<FormState>();
  var obscure = true;
  final token = await showDialog<String>(
    context: context,
    builder: (context) => StatefulBuilder(
      builder: (context, setState) => AlertDialog(
        title: const Text('配置 Status Gateway'),
        content: SizedBox(
          width: 420,
          child: Form(
            key: formKey,
            child: TextFormField(
              controller: controller,
              autofocus: true,
              obscureText: obscure,
              autocorrect: false,
              enableSuggestions: false,
              textInputAction: TextInputAction.done,
              decoration: InputDecoration(
                labelText: 'Ingest Token',
                helperText: 'Token 仅保存在本机 Android Keystore',
                suffixIcon: IconButton(
                  onPressed: () => setState(() => obscure = !obscure),
                  tooltip: obscure ? '显示 Token' : '隐藏 Token',
                  icon: Icon(
                    obscure
                        ? Icons.visibility_outlined
                        : Icons.visibility_off_outlined,
                  ),
                ),
              ),
              validator: (value) => value == null || value.trim().isEmpty
                  ? '请输入 Ingest Token'
                  : null,
              onFieldSubmitted: (_) {
                if (formKey.currentState!.validate()) {
                  Navigator.pop(context, controller.text.trim());
                }
              },
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: () {
              if (formKey.currentState!.validate()) {
                Navigator.pop(context, controller.text.trim());
              }
            },
            child: const Text('保存'),
          ),
        ],
      ),
    ),
  );
  controller.dispose();
  if (token == null || !context.mounted) {
    return;
  }
  try {
    await ref.read(statusStateProvider.notifier).saveToken(token);
    if (context.mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Status 上报已启用')));
    }
  } on Object catch (error) {
    if (context.mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(readableError(error))));
    }
  }
}

Future<void> requestStatusPermissions(
  BuildContext context,
  WidgetRef ref,
) async {
  try {
    final granted = await ref
        .read(statusStateProvider.notifier)
        .requestPermissions();
    if (!granted && context.mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('未授权时仍可上报，但通知和蓝牙状态可能不完整')));
    }
  } on Object catch (error) {
    if (context.mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(readableError(error))));
    }
  }
}

Future<void> clearStatusToken(BuildContext context, WidgetRef ref) async {
  final confirmed = await showDialog<bool>(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('停止 Status 上报？'),
      content: const Text('本机将删除 Status Token、停止前台同步，并在下次连接时重新登记设备。'),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context, false),
          child: const Text('取消'),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, true),
          child: const Text('停止并删除'),
        ),
      ],
    ),
  );
  if (confirmed != true || !context.mounted) {
    return;
  }
  try {
    await ref.read(statusStateProvider.notifier).clearToken();
  } on Object catch (error) {
    if (context.mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(readableError(error))));
    }
  }
}
