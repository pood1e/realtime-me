package me.realtime.watch.health

import android.os.SystemClock
import android.util.Log
import androidx.health.services.client.PassiveListenerService
import androidx.health.services.client.data.DataPointContainer
import androidx.health.services.client.data.DataType
import me.realtime.protocol.v1.ReportWatchSnapshotRequest
import me.realtime.protocol.v1.WatchSnapshot
import me.realtime.protocol.v1.WristState
import me.realtime.watch.state.WatchSnapshotRepository
import me.realtime.watch.wear.SnapshotPublisher
import java.time.Instant
import kotlin.math.roundToInt

class PassiveHealthService : PassiveListenerService() {
    override fun onNewDataPointsReceived(dataPoints: DataPointContainer) {
        val repository = WatchSnapshotRepository(this)
        val snapshot = PassiveHealthSnapshotUpdater(repository, bootInstant()).update(dataPoints) ?: return
        publish(snapshot)
    }

    override fun onPermissionLost() {
        Log.w(TAG, "Passive Health Services permission lost")
    }

    private fun publish(snapshot: WatchSnapshot) {
        val payload = ReportWatchSnapshotRequest.newBuilder()
            .setWatchSnapshot(snapshot)
            .build()
        SnapshotPublisher(this).publishIfAllowed(payload)
    }

    private fun bootInstant(): Instant {
        return Instant.ofEpochMilli(System.currentTimeMillis() - SystemClock.elapsedRealtime())
    }

    private companion object {
        const val TAG = "PassiveHealthService"
    }
}

private class PassiveHealthSnapshotUpdater(
    private val repository: WatchSnapshotRepository,
    private val bootInstant: Instant,
) {
    fun update(dataPoints: DataPointContainer): WatchSnapshot? {
        var snapshot: WatchSnapshot? = null
        latestHeartRate(dataPoints)?.let { sample ->
            repository.updateWristState(WristState.WRIST_STATE_ON_WRIST)
            snapshot = repository.updateHeartRate(sample.beatsPerMinute, sample.sampleTime)
        }
        latestDailySteps(dataPoints)?.let { sample ->
            snapshot = repository.updateSteps(sample.steps, sample.sampleTime)
        }
        return snapshot
    }

    private fun latestHeartRate(dataPoints: DataPointContainer): HeartRateSample? {
        return dataPoints.getData(DataType.HEART_RATE_BPM)
            .maxByOrNull { it.timeDurationFromBoot }
            ?.let { point ->
                val beatsPerMinute = point.value.roundToInt()
                if (beatsPerMinute <= 0) return@let null
                HeartRateSample(beatsPerMinute, point.getTimeInstant(bootInstant))
            }
    }

    private fun latestDailySteps(dataPoints: DataPointContainer): DailyStepsSample? {
        return dataPoints.getData(DataType.STEPS_DAILY)
            .maxByOrNull { it.endDurationFromBoot }
            ?.let { point ->
                val steps = point.value.toInt().coerceAtLeast(0)
                DailyStepsSample(steps, point.getEndInstant(bootInstant))
            }
    }
}

private data class HeartRateSample(
    val beatsPerMinute: Int,
    val sampleTime: Instant,
)

private data class DailyStepsSample(
    val steps: Int,
    val sampleTime: Instant,
)
