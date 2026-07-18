import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../../core/security/device_credentials.dart';
import '../../core/state/app_session.dart';

enum _PairingMode { scan, paste }

class PairingScreen extends ConsumerStatefulWidget {
  final bool closeOnSuccess;

  const PairingScreen({this.closeOnSuccess = false, super.key});

  @override
  ConsumerState<PairingScreen> createState() => _PairingScreenState();
}

class _PairingScreenState extends ConsumerState<PairingScreen> {
  final _scanner = MobileScannerController(
    formats: const [BarcodeFormat.qrCode],
    detectionSpeed: DetectionSpeed.noDuplicates,
  );
  final _payload = TextEditingController();
  final _deviceName = TextEditingController(text: 'Android');
  _PairingMode _mode = _PairingMode.scan;
  PairingOffer? _offer;
  bool _busy = false;
  bool _scanLocked = false;

  @override
  void dispose() {
    _scanner.dispose();
    _payload.dispose();
    _deviceName.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 560),
            child: ListView(
              padding: const EdgeInsets.fromLTRB(24, 32, 24, 40),
              children: [
                Align(
                  alignment: Alignment.centerLeft,
                  child: DecoratedBox(
                    decoration: BoxDecoration(
                      color: theme.colorScheme.primaryContainer,
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: const Padding(
                      padding: EdgeInsets.all(14),
                      child: Icon(Icons.terminal_rounded, size: 32),
                    ),
                  ),
                ),
                const SizedBox(height: 24),
                Text('连接工作机', style: theme.textTheme.headlineMedium),
                const SizedBox(height: 8),
                Text(
                  '在 Linux 服务端运行 smctl pair create，然后扫描一次性二维码。',
                  style: theme.textTheme.bodyLarge?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                ),
                const SizedBox(height: 28),
                SegmentedButton<_PairingMode>(
                  segments: const [
                    ButtonSegment(
                      value: _PairingMode.scan,
                      icon: Icon(Icons.qr_code_scanner_rounded),
                      label: Text('扫描'),
                    ),
                    ButtonSegment(
                      value: _PairingMode.paste,
                      icon: Icon(Icons.content_paste_rounded),
                      label: Text('粘贴'),
                    ),
                  ],
                  selected: {_mode},
                  onSelectionChanged: _busy
                      ? null
                      : (selection) {
                          final mode = selection.first;
                          setState(() {
                            _mode = mode;
                            _offer = null;
                            _scanLocked = false;
                          });
                          unawaited(
                            mode == _PairingMode.scan
                                ? _scanner.start()
                                : _scanner.stop(),
                          );
                        },
                ),
                const SizedBox(height: 20),
                if (_mode == _PairingMode.scan)
                  _buildScanner()
                else
                  _buildPasteField(),
                if (_offer case final offer?) ...[
                  const SizedBox(height: 20),
                  _OfferSummary(offer: offer),
                  const SizedBox(height: 20),
                  TextField(
                    controller: _deviceName,
                    enabled: !_busy,
                    textInputAction: TextInputAction.done,
                    maxLength: 128,
                    decoration: const InputDecoration(
                      labelText: '设备名称',
                      prefixIcon: Icon(Icons.phone_android_rounded),
                    ),
                  ),
                  const SizedBox(height: 8),
                  FilledButton.icon(
                    onPressed: _busy ? null : _pair,
                    icon: _busy
                        ? const SizedBox.square(
                            dimension: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.link_rounded),
                    label: Text(_busy ? '正在配对' : '确认连接'),
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildScanner() {
    return AspectRatio(
      aspectRatio: 4 / 3,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: Stack(
          fit: StackFit.expand,
          children: [
            MobileScanner(
              controller: _scanner,
              onDetect: (capture) {
                if (_scanLocked || capture.barcodes.isEmpty) {
                  return;
                }
                final value = capture.barcodes.first.rawValue;
                if (value == null) {
                  return;
                }
                _scanLocked = true;
                _acceptPayload(value);
              },
              errorBuilder: (context, error) => ColoredBox(
                color: Theme.of(context).colorScheme.surfaceContainerHighest,
                child: const Center(child: Text('相机不可用，请改用粘贴配对内容')),
              ),
            ),
            IgnorePointer(
              child: Center(
                child: Container(
                  width: 220,
                  height: 220,
                  decoration: BoxDecoration(
                    border: Border.all(color: Colors.white, width: 3),
                    borderRadius: BorderRadius.circular(20),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPasteField() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          controller: _payload,
          enabled: !_busy,
          minLines: 4,
          maxLines: 8,
          autocorrect: false,
          enableSuggestions: false,
          decoration: const InputDecoration(
            labelText: '配对内容',
            alignLabelWithHint: true,
          ),
        ),
        const SizedBox(height: 12),
        OutlinedButton.icon(
          onPressed: _busy ? null : () => _acceptPayload(_payload.text),
          icon: const Icon(Icons.verified_user_outlined),
          label: const Text('校验配对内容'),
        ),
      ],
    );
  }

  void _acceptPayload(String payload) {
    try {
      final offer = PairingOffer.parse(payload);
      setState(() => _offer = offer);
      unawaited(_scanner.stop());
    } on Object catch (error) {
      _scanLocked = false;
      _showError(error);
    }
  }

  Future<void> _pair() async {
    final offer = _offer;
    if (offer == null) {
      return;
    }
    setState(() => _busy = true);
    try {
      await ref.read(appSessionProvider.notifier).pair(offer, _deviceName.text);
      if (mounted && widget.closeOnSuccess) {
        if (context.canPop()) {
          context.pop();
        } else {
          context.go('/');
        }
      }
    } on Object catch (error) {
      if (mounted) {
        _showError(error);
        setState(() => _busy = false);
      }
    }
  }

  void _showError(Object error) {
    if (!mounted) {
      return;
    }
    final message = switch (error) {
      FormatException(:final message) => message,
      _ => '连接失败，请检查 DDNS、端口和服务状态',
    };
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }
}

class _OfferSummary extends StatelessWidget {
  final PairingOffer offer;

  const _OfferSummary({required this.offer});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Row(
              children: [
                Icon(Icons.shield_outlined),
                SizedBox(width: 10),
                Text('已验证一次性配对信息'),
              ],
            ),
            const SizedBox(height: 12),
            SelectableText(
              offer.serviceUri.toString(),
              style: const TextStyle(fontFamily: 'monospace'),
            ),
            const SizedBox(height: 6),
            Text('有效至 ${offer.expireTime.toLocal().toIso8601String()}'),
          ],
        ),
      ),
    );
  }
}
