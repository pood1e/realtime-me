package me.realtime.mobile.background

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import me.realtime.mobile.state.GitHubTokenStore
import me.realtime.mobile.state.SnapshotProcessor
import me.realtime.mobile.state.WatchSnapshotStore
import me.realtime.mobile.wear.WatchSnapshotReader
import me.realtime.protocol.toJavaInstant
import me.realtime.protocol.v1.ReportWatchSnapshotRequest
import me.realtime.protocol.v1.WatchSnapshot
import java.time.Duration
import java.time.Instant

class GitHubStatusSyncWorker(
    appContext: Context,
    workerParameters: WorkerParameters,
) : CoroutineWorker(appContext, workerParameters) {
    override suspend fun doWork(): Result {
        if (!GitHubTokenStore(applicationContext).hasToken()) return Result.success()

        val processor = SnapshotProcessor(applicationContext)
        val now = Instant.now()
        latestDataLayerSnapshot(now)?.let { snapshot ->
            return processor.processSnapshot(snapshot).toWorkResult()
        }

        val storedSnapshot = WatchSnapshotStore(applicationContext).latest() ?: return Result.success()
        if (Duration.between(storedSnapshot.receivedAt, now) > MAX_SNAPSHOT_AGE) return Result.success()

        return processor.processSnapshot(
            snapshot = storedSnapshot.snapshot,
            receivedAt = storedSnapshot.receivedAt,
        ).toWorkResult()
    }

    private suspend fun latestDataLayerSnapshot(now: Instant): WatchSnapshot? {
        val payload = runCatching { WatchSnapshotReader(applicationContext).latestPayload() }.getOrNull()
            ?: return null
        val request = runCatching { ReportWatchSnapshotRequest.parseFrom(payload) }.getOrNull()
            ?: return null
        if (!request.hasWatchSnapshot()) return null
        return request.watchSnapshot.takeIf { it.isFresh(now) }
    }

    private fun WatchSnapshot.isFresh(now: Instant): Boolean {
        if (!hasRecordTime()) return false
        val age = Duration.between(recordTime.toJavaInstant(), now)
        return age >= Duration.ZERO && age <= MAX_SNAPSHOT_AGE
    }

    private fun me.realtime.mobile.state.SnapshotProcessResult.toWorkResult(): Result {
        return if (retryable) Result.retry() else Result.success()
    }

    private companion object {
        val MAX_SNAPSHOT_AGE: Duration = Duration.ofMinutes(10)
    }
}
