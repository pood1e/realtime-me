package me.realtime.watch.state

import android.content.Context
import me.realtime.protocol.toProtoTimestamp
import me.realtime.protocol.v1.ActivityTotals
import me.realtime.protocol.v1.HeartRateSample
import me.realtime.protocol.v1.WatchSnapshot
import me.realtime.protocol.v1.WristState
import java.time.Instant
import java.util.Base64
import java.util.UUID

class WatchSnapshotRepository(private val context: Context) {
    private val preferences = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun updateHeartRate(beatsPerMinute: Int, sampleTime: Instant): WatchSnapshot = updateSnapshot { builder ->
        builder.heartRate = HeartRateSample.newBuilder()
            .setBeatsPerMinute(beatsPerMinute)
            .setSampleTime(sampleTime.toProtoTimestamp())
            .build()
    }

    fun updateSteps(steps: Int, sampleTime: Instant): WatchSnapshot = updateSnapshot { builder ->
        builder.activityTotals = builder.activityTotals.toBuilder()
            .setSteps(steps)
            .setSampleTime(sampleTime.toProtoTimestamp())
            .build()
    }

    fun updateWristState(wristState: WristState): WatchSnapshot = updateSnapshot { builder ->
        builder.watchState = WatchStateReader.read(context, wristState)
    }

    fun refreshDeviceState(includeStepTotal: Boolean = false): WatchSnapshot = updateSnapshot { builder ->
        if (includeStepTotal && !builder.hasActivityTotals()) {
            builder.activityTotals = ActivityTotals.newBuilder()
                .setSteps(0)
                .setSampleTime(Instant.now().toProtoTimestamp())
                .build()
        }
        builder.watchState = WatchStateReader.read(context, builder.watchState.wristState)
    }

    fun currentSnapshot(): WatchSnapshot? {
        val encoded = preferences.getString(SNAPSHOT_KEY, null) ?: return null
        return runCatching {
            WatchSnapshot.parseFrom(Base64.getDecoder().decode(encoded))
        }.getOrNull()
    }

    private fun updateSnapshot(applyUpdate: (WatchSnapshot.Builder) -> Unit): WatchSnapshot {
        val builder = currentSnapshot()?.toBuilder() ?: WatchSnapshot.newBuilder()
        applyUpdate(builder)
        builder.snapshotId = UUID.randomUUID().toString()
        builder.recordTime = Instant.now().toProtoTimestamp()
        if (!builder.hasWatchState()) {
            builder.watchState = WatchStateReader.read(context)
        } else {
            builder.watchState = WatchStateReader.read(context, builder.watchState.wristState)
        }
        builder.deviceInfo = DeviceInfoReader.read()
        return builder.build().also(::save)
    }

    private fun save(snapshot: WatchSnapshot) {
        preferences.edit()
            .putString(SNAPSHOT_KEY, Base64.getEncoder().encodeToString(snapshot.toByteArray()))
            .apply()
    }

    private companion object {
        const val PREFS_NAME = "watch_snapshot_repository"
        const val SNAPSHOT_KEY = "latest_snapshot"
    }
}
