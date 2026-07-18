import 'package:pigeon/pigeon.dart';

@ConfigurePigeon(
  PigeonOptions(
    dartOut: 'lib/core/platform/status_bridge.g.dart',
    dartOptions: DartOptions(),
    dartPackageName: 'realtime_me',
    kotlinOut:
        'android/app/src/main/kotlin/me/realtime/mobile/platform/StatusBridge.g.kt',
    kotlinOptions: KotlinOptions(package: 'me.realtime.mobile.platform'),
    ignoreLints: false,
  ),
)
class StatusSnapshotData {
  StatusSnapshotData({
    required this.revision,
    required this.receivedAtEpochMillis,
    this.protobufBytes,
  });

  int revision;
  int receivedAtEpochMillis;
  Uint8List? protobufBytes;
}

@HostApi()
abstract class StatusHostApi {
  StatusSnapshotData getSnapshot();

  bool hasToken();

  bool hasRequiredPermissions();

  @async
  bool saveToken(String token);

  @async
  void clearToken();

  @async
  void refresh();

  @async
  bool requestPermissions();
}

@EventChannelApi()
abstract class StatusEventChannels {
  StatusSnapshotData snapshots();
}
