import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:realtime_me_status_contracts/gen/realtime/me/status/v1/watch.pb.dart';

import '../../core/platform/status_bridge.g.dart' as bridge;

final statusPlatformProvider = Provider<StatusPlatformClient>(
  (ref) => StatusPlatformClient(),
);

final statusStateProvider =
    AsyncNotifierProvider<StatusStateController, StatusState>(
      StatusStateController.new,
    );

class StatusState {
  final WatchSnapshot? watch;
  final DateTime? receivedAt;
  final int revision;
  final bool hasToken;
  final bool hasRequiredPermissions;

  const StatusState({
    required this.watch,
    required this.receivedAt,
    required this.revision,
    required this.hasToken,
    required this.hasRequiredPermissions,
  });

  StatusState withSnapshot(StatusSnapshot snapshot) => StatusState(
    watch: snapshot.watch,
    receivedAt: snapshot.receivedAt,
    revision: snapshot.revision,
    hasToken: hasToken,
    hasRequiredPermissions: hasRequiredPermissions,
  );

  StatusState copyWith({bool? hasToken, bool? hasRequiredPermissions}) {
    return StatusState(
      watch: watch,
      receivedAt: receivedAt,
      revision: revision,
      hasToken: hasToken ?? this.hasToken,
      hasRequiredPermissions:
          hasRequiredPermissions ?? this.hasRequiredPermissions,
    );
  }
}

class StatusSnapshot {
  final WatchSnapshot? watch;
  final DateTime? receivedAt;
  final int revision;

  const StatusSnapshot({
    required this.watch,
    required this.receivedAt,
    required this.revision,
  });

  factory StatusSnapshot.fromBridge(bridge.StatusSnapshotData data) {
    final bytes = data.protobufBytes;
    return StatusSnapshot(
      watch: bytes == null ? null : WatchSnapshot.fromBuffer(bytes),
      receivedAt: data.receivedAtEpochMillis <= 0
          ? null
          : DateTime.fromMillisecondsSinceEpoch(
              data.receivedAtEpochMillis,
              isUtc: true,
            ).toLocal(),
      revision: data.revision,
    );
  }
}

class StatusPlatformClient {
  final bridge.StatusHostApi _host;

  StatusPlatformClient({bridge.StatusHostApi? host})
    : _host = host ?? bridge.StatusHostApi();

  Stream<StatusSnapshot> get snapshots =>
      bridge.snapshots().map(StatusSnapshot.fromBridge);

  Future<StatusState> load() async {
    final values = await Future.wait<Object>([
      _host.getSnapshot(),
      _host.hasToken(),
      _host.hasRequiredPermissions(),
    ]);
    final snapshot = StatusSnapshot.fromBridge(
      values[0] as bridge.StatusSnapshotData,
    );
    return StatusState(
      watch: snapshot.watch,
      receivedAt: snapshot.receivedAt,
      revision: snapshot.revision,
      hasToken: values[1] as bool,
      hasRequiredPermissions: values[2] as bool,
    );
  }

  Future<bool> saveToken(String token) => _host.saveToken(token);

  Future<void> clearToken() => _host.clearToken();

  Future<void> refresh() => _host.refresh();

  Future<bool> requestPermissions() => _host.requestPermissions();
}

final class StatusStateController extends AsyncNotifier<StatusState> {
  late StatusPlatformClient _platform;
  StreamSubscription<StatusSnapshot>? _subscription;

  @override
  Future<StatusState> build() async {
    _platform = ref.watch(statusPlatformProvider);
    ref.onDispose(() => _subscription?.cancel());
    return _connect();
  }

  Future<StatusState> _connect() async {
    await _platform.refresh();
    final initial = await _platform.load();
    await _subscription?.cancel();
    _subscription = _platform.snapshots.listen(
      _applySnapshot,
      onError: (Object error, StackTrace stackTrace) {
        state = AsyncError(error, stackTrace);
      },
    );
    return initial;
  }

  Future<void> retry() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(_connect);
  }

  Future<void> refresh() async {
    await _platform.refresh();
    state = await AsyncValue.guard(_platform.load);
  }

  Future<void> saveToken(String token) async {
    final saved = await _platform.saveToken(token);
    if (!saved) {
      throw StateError('Android Keystore 无法保存 Status Token');
    }
    _update((current) => current.copyWith(hasToken: true));
  }

  Future<void> clearToken() async {
    await _platform.clearToken();
    _update((current) => current.copyWith(hasToken: false));
  }

  Future<bool> requestPermissions() async {
    final granted = await _platform.requestPermissions();
    _update((current) => current.copyWith(hasRequiredPermissions: granted));
    return granted;
  }

  void _applySnapshot(StatusSnapshot snapshot) {
    _update((current) => current.withSnapshot(snapshot));
  }

  void _update(StatusState Function(StatusState current) update) {
    final current = state.value;
    if (current != null) {
      state = AsyncData(update(current));
    }
  }
}
