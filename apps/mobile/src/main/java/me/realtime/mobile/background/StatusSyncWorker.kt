package me.realtime.mobile.background

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import me.realtime.mobile.state.StatusGatewayTokenStore
import me.realtime.mobile.state.WatchSnapshotProcessor
import me.realtime.mobile.status.StatusGatewayPushResult
import me.realtime.mobile.status.StatusGatewayPusher
import me.realtime.mobile.wear.WatchSnapshotReader

class StatusSyncWorker(
    appContext: Context,
    workerParameters: WorkerParameters,
) : CoroutineWorker(appContext, workerParameters) {
    override suspend fun doWork(): Result {
        if (!StatusGatewayTokenStore(applicationContext).hasToken()) return Result.success()

        val payload = runCatching { WatchSnapshotReader(applicationContext).latestPayload() }.getOrNull()
        val processed = payload?.let { WatchSnapshotProcessor(applicationContext).process(it) } == true
        if (processed) return Result.success()

        return when (StatusGatewayPusher(applicationContext).pushLatest()) {
            StatusGatewayPushResult.Failure -> Result.retry()
            StatusGatewayPushResult.Disabled,
            StatusGatewayPushResult.Success,
            -> Result.success()
        }
    }
}
