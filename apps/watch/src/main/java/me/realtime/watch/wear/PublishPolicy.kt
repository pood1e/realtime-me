package me.realtime.watch.wear

import android.content.Context
import me.realtime.protocol.v1.ReportWatchSnapshotRequest
import me.realtime.protocol.v1.WristState
import java.time.Duration

class PublishPolicy(context: Context) {
    private val preferences = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun shouldPublish(payload: ReportWatchSnapshotRequest, nowMillis: Long = System.currentTimeMillis()): Boolean {
        val snapshot = payload.watchSnapshot
        val lastPublishMillis = preferences.getLong(LAST_PUBLISH_TIME_KEY, 0L)
        val currentSignature = signature(payload)
        val lastSignature = preferences.getString(LAST_SIGNATURE_KEY, null)
        val elapsed = Duration.ofMillis(nowMillis - lastPublishMillis)
        val lowPowerMode = snapshot.watchState.batteryPercent in 1..14 ||
            snapshot.watchState.wristState == WristState.WRIST_STATE_OFF_WRIST
        val minInterval = if (lowPowerMode) LOW_POWER_MIN_INTERVAL else NORMAL_MIN_INTERVAL
        val maxInterval = if (lowPowerMode) LOW_POWER_MAX_INTERVAL else NORMAL_MAX_INTERVAL

        if (lastPublishMillis == 0L) return true
        if (addsFirstHealthReading(lastSignature, currentSignature)) return true
        if (elapsed < minInterval) return false
        if (elapsed >= maxInterval) return true
        return currentSignature != lastSignature
    }

    fun markPublished(payload: ReportWatchSnapshotRequest, nowMillis: Long = System.currentTimeMillis()) {
        preferences.edit()
            .putLong(LAST_PUBLISH_TIME_KEY, nowMillis)
            .putString(LAST_SIGNATURE_KEY, signature(payload))
            .apply()
    }

    private fun addsFirstHealthReading(lastSignature: String?, currentSignature: String): Boolean {
        if (lastSignature == null) return false
        val lastParts = lastSignature.split(SIGNATURE_SEPARATOR)
        val currentParts = currentSignature.split(SIGNATURE_SEPARATOR)
        if (lastParts.size != SIGNATURE_PART_COUNT || currentParts.size != SIGNATURE_PART_COUNT) return true

        val heartRateBecameAvailable = lastParts[HEART_RATE_PRESENT_INDEX] == UNAVAILABLE &&
            currentParts[HEART_RATE_PRESENT_INDEX] == AVAILABLE
        val stepsBecameAvailable = lastParts[STEPS_PRESENT_INDEX] == UNAVAILABLE &&
            currentParts[STEPS_PRESENT_INDEX] == AVAILABLE
        return heartRateBecameAvailable || stepsBecameAvailable
    }

    private fun signature(snapshot: ReportWatchSnapshotRequest): String {
        val watchSnapshot = snapshot.watchSnapshot
        val heartRateBucket = watchSnapshot.heartRate.beatsPerMinute / HEART_RATE_BUCKET_SIZE
        val stepBucket = watchSnapshot.activityTotals.steps / STEPS_BUCKET_SIZE
        val batteryBucket = watchSnapshot.watchState.batteryPercent / BATTERY_BUCKET_SIZE
        return listOf(
            watchSnapshot.hasHeartRate().availabilitySignature(),
            heartRateBucket,
            watchSnapshot.hasActivityTotals().availabilitySignature(),
            stepBucket,
            batteryBucket,
            watchSnapshot.watchState.chargeState.number,
            watchSnapshot.watchState.wristState.number,
        ).joinToString(separator = SIGNATURE_SEPARATOR)
    }

    private fun Boolean.availabilitySignature(): String = if (this) AVAILABLE else UNAVAILABLE

    private companion object {
        const val PREFS_NAME = "publish_policy"
        const val LAST_PUBLISH_TIME_KEY = "last_publish_time_ms"
        const val LAST_SIGNATURE_KEY = "last_signature"
        const val HEART_RATE_BUCKET_SIZE = 1
        const val STEPS_BUCKET_SIZE = 1
        const val BATTERY_BUCKET_SIZE = 5
        const val HEART_RATE_PRESENT_INDEX = 0
        const val STEPS_PRESENT_INDEX = 2
        const val SIGNATURE_PART_COUNT = 7
        const val AVAILABLE = "1"
        const val UNAVAILABLE = "0"
        const val SIGNATURE_SEPARATOR = ":"
        val NORMAL_MIN_INTERVAL: Duration = Duration.ofSeconds(2)
        val NORMAL_MAX_INTERVAL: Duration = Duration.ofMinutes(1)
        val LOW_POWER_MIN_INTERVAL: Duration = Duration.ofSeconds(30)
        val LOW_POWER_MAX_INTERVAL: Duration = Duration.ofMinutes(5)
    }
}
