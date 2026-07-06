package me.realtime.mobile.state

import android.content.Context
import me.realtime.protocol.toJavaInstant
import me.realtime.protocol.v1.ReportWatchSnapshotRequest
import me.realtime.protocol.v1.WatchSnapshot
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

    private fun WatchSnapshot.isFresh(now: Instant): Boolean {
        if (!hasRecordTime()) return true
        val age = Duration.between(recordTime.toJavaInstant(), now)
        return age >= Duration.ZERO && age <= MAX_SNAPSHOT_AGE
    }

    private companion object {
        val MAX_SNAPSHOT_AGE: Duration = Duration.ofMinutes(10)
    }
}
