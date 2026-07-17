import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../network/pairing_client.dart';
import '../network/remote_session.dart';
import '../security/credential_store.dart';
import '../security/device_credentials.dart';

final credentialStoreProvider = Provider<CredentialStore>(
  (ref) => CredentialStore(),
);
final pairingClientProvider = Provider<PairingClient>((ref) => PairingClient());

final appSessionProvider =
    AsyncNotifierProvider<AppSessionController, RemoteSession?>(
      AppSessionController.new,
    );

final class AppSessionController extends AsyncNotifier<RemoteSession?> {
  RemoteSession? _session;

  @override
  Future<RemoteSession?> build() async {
    ref.onDispose(() => _session?.close());
    final credentials = await ref.read(credentialStoreProvider).read();
    if (credentials == null) {
      return null;
    }
    return _session = RemoteSession(credentials);
  }

  Future<void> pair(PairingOffer offer, String displayName) async {
    final credentials = await ref
        .read(pairingClientProvider)
        .pair(offer, displayName);
    final next = RemoteSession(credentials);
    try {
      await ref.read(credentialStoreProvider).write(credentials);
    } catch (_) {
      next.close();
      rethrow;
    }
    final previous = _session;
    _session = next;
    state = AsyncData(next);
    previous?.close();
  }

  Future<void> disconnect() async {
    final current = _session;
    if (current != null) {
      await current.control.revokeDevice(current.credentials.deviceUid);
    }
    await forgetLocalCredentials();
  }

  Future<void> forgetLocalCredentials() async {
    await ref.read(credentialStoreProvider).clear();
    _session?.close();
    _session = null;
    state = const AsyncData(null);
  }
}
