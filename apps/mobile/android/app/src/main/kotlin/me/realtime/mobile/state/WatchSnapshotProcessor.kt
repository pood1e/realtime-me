package me.realtime.mobile.state

import android.content.Context
import me.realtime.protocol.toJavaInstant
import me.realtime.status.v1.ReportWatchSnapshotRequest
import me.realtime.status.v1.WatchSnapshot
import java.time.Duration
import java.time.Instant

class WatchSnapshotProcessor(context: Context) {
    private val snapshotStore = WatchSnapshotStore(context.applicationContext)

    fun process(payloadBytes: ByteArray): Boolean {
        val payload = runCatching { ReportWatchSnapshotRequest.parseFrom(payloadBytes) }.getOrNull() ?: return false
        if (!payload.hasWatchSnapshot()) return false
        return processSnapshot(payload.watchSnapshot)
    }

    fun processSnapshot(snapshot: WatchSnapshot, receivedAt: Instant = Instant.now()): Boolean {
        if (!snapshot.isFresh(receivedAt)) return false
        snapshotStore.save(snapshot, receivedAt)
        return true
    }

    /**
     * The watch stamps [WatchSnapshot.getRecordTime] from its own clock, which the
     * phone only ever approximately agrees with: a watch running seconds ahead
     * hands the phone a snapshot from its future. Refusing those would silently
     * discard every reading for as long as the drift pointed that way, so the
     * window opens a little either side of now and still shuts on a stale one.
     */
    private fun WatchSnapshot.isFresh(now: Instant): Boolean {
        if (!hasRecordTime()) return true
        val age = Duration.between(recordTime.toJavaInstant(), now)
        return age >= MAX_CLOCK_SKEW.negated() && age <= MAX_SNAPSHOT_AGE
    }

    private companion object {
        val MAX_SNAPSHOT_AGE: Duration = Duration.ofMinutes(10)
        val MAX_CLOCK_SKEW: Duration = Duration.ofMinutes(1)
    }
}
