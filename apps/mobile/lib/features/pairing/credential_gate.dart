import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/state/app_session.dart';
import 'pairing_screen.dart';

class CredentialGate extends ConsumerWidget {
  final Widget child;

  const CredentialGate({required this.child, super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return ref
        .watch(appSessionProvider)
        .when(
          data: (session) => session == null ? const PairingScreen() : child,
          loading: () => const _LoadingScreen(),
          error: (_, _) => const _CredentialErrorScreen(),
        );
  }
}

class _LoadingScreen extends StatelessWidget {
  const _LoadingScreen();

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: SafeArea(
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.terminal_rounded, size: 48),
              SizedBox(height: 24),
              CircularProgressIndicator(),
            ],
          ),
        ),
      ),
    );
  }
}

class _CredentialErrorScreen extends ConsumerWidget {
  const _CredentialErrorScreen();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 420),
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    Icons.key_off_rounded,
                    size: 56,
                    color: Theme.of(context).colorScheme.error,
                  ),
                  const SizedBox(height: 20),
                  Text(
                    '无法读取设备凭据',
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 12),
                  const Text(
                    '本地安全存储或设备证书不可用。清除本机凭据后可重新配对。',
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 24),
                  FilledButton.icon(
                    onPressed: () => ref
                        .read(appSessionProvider.notifier)
                        .forgetLocalCredentials(),
                    icon: const Icon(Icons.restart_alt_rounded),
                    label: const Text('清除并重新配对'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
