package me.realtime.watch.state

import android.content.Context
import me.realtime.protocol.toJavaInstant
import me.realtime.protocol.toProtoTimestamp
import me.realtime.protocol.v1.ActivityTotals
import me.realtime.protocol.v1.HeartRateSample
import me.realtime.protocol.v1.WatchSnapshot
import java.time.Duration
import java.time.Instant
import java.time.ZoneId
import java.util.Base64
import java.util.UUID

class WatchSnapshotRepository(private val context: Context) {
    private val preferences = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun updateHeartRate(beatsPerMinute: Int, sampleTime: Instant): WatchSnapshot = updateSnapshot { builder, _ ->
        builder.heartRate = HeartRateSample.newBuilder()
            .setBeatsPerMinute(beatsPerMinute)
            .setSampleTime(sampleTime.toProtoTimestamp())
            .build()
    }

    // Called when the sensor reports it has left the wrist, or that it cannot
    // trust what it read. An absent heart rate says nobody is wearing this.
    fun clearHeartRate(): WatchSnapshot = updateSnapshot { builder, _ ->
        builder.clearHeartRate()
    }

    fun updateSteps(steps: Int, sampleTime: Instant): WatchSnapshot = updateSnapshot { builder, _ ->
        builder.activityTotals = builder.activityTotals.toBuilder()
            .setSteps(steps)
            .setSampleTime(sampleTime.toProtoTimestamp())
            .build()
    }

    // Re-reads battery and charge state without touching the health samples.
    fun refreshDeviceState(): WatchSnapshot = updateSnapshot { _, _ -> }

    fun currentSnapshot(): WatchSnapshot? {
        val encoded = preferences.getString(SNAPSHOT_KEY, null) ?: return null
        return runCatching {
            WatchSnapshot.parseFrom(Base64.getDecoder().decode(encoded))
        }.getOrNull()
    }

    // The Health Services binder thread and the registration worker both mutate
    // the single persisted snapshot through separate repository instances that
    // share one SharedPreferences file, so the read-modify-write is serialized
    // on a process-wide lock to avoid lost updates.
    private fun updateSnapshot(applyUpdate: (WatchSnapshot.Builder, Instant) -> Unit): WatchSnapshot =
        synchronized(SNAPSHOT_LOCK) {
            val builder = currentSnapshot()?.toBuilder() ?: WatchSnapshot.newBuilder()
            val now = Instant.now()
            applyUpdate(builder, now)
            resetStaleActivityTotals(builder, now)
            retireStaleHeartRate(builder, now)
            builder.snapshotId = UUID.randomUUID().toString()
            builder.recordTime = now.toProtoTimestamp()
            builder.watchState = WatchStateReader.read(context)
            builder.deviceInfo = DeviceInfoReader.read()
            builder.build().also(::save)
        }

    private fun save(snapshot: WatchSnapshot) {
        preferences.edit()
            .putString(SNAPSHOT_KEY, Base64.getEncoder().encodeToString(snapshot.toByteArray()))
            .apply()
    }

    // A heart rate is a reading, not a property of the watch. Once the sensor
    // stops measuring, the last number would ride every later snapshot -- the
    // battery refresh republishes it on a timer -- and a resting heart rate from
    // this morning is indistinguishable on the page from one taken just now. A
    // reading nothing has refreshed within the window is no longer a heart rate.
    private fun retireStaleHeartRate(builder: WatchSnapshot.Builder, now: Instant) {
        if (!builder.hasHeartRate()) return
        val sampledAt = builder.heartRate.sampleTime.toJavaInstant()
        if (Duration.between(sampledAt, now) > HEART_RATE_MAX_AGE) builder.clearHeartRate()
    }

    private fun resetStaleActivityTotals(builder: WatchSnapshot.Builder, now: Instant) {
        if (!builder.hasActivityTotals()) return
        val sampleDate = builder.activityTotals.sampleTime.toJavaInstant()
            .atZone(ZoneId.systemDefault())
            .toLocalDate()
        val today = now.atZone(ZoneId.systemDefault()).toLocalDate()
        if (sampleDate == today) return
        builder.activityTotals = ActivityTotals.newBuilder()
            .setSteps(0)
            .setSampleTime(now.toProtoTimestamp())
            .build()
    }

    private companion object {
        const val PREFS_NAME = "watch_snapshot_repository"
        const val SNAPSHOT_KEY = "latest_snapshot"

        // Passive monitoring batches its samples, so a worn watch can stay quiet
        // for several minutes. This has to outlast that silence and still retire
        // a reading long before it becomes a lie.
        private val HEART_RATE_MAX_AGE: Duration = Duration.ofMinutes(15)

        // Guards the persisted-snapshot read-modify-write across all instances.
        private val SNAPSHOT_LOCK = Any()
    }
}
