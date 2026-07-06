package me.realtime.mobile.background

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import me.realtime.mobile.state.StatusGatewayTokenStore
import me.realtime.mobile.status.StatusGatewayPushResult

class StatusSyncWorker(
    appContext: Context,
    workerParameters: WorkerParameters,
) : CoroutineWorker(appContext, workerParameters) {
    override suspend fun doWork(): Result {
        if (!StatusGatewayTokenStore(applicationContext).hasToken()) return Result.success()

        return when (StatusSyncRunner(applicationContext).syncLatest()) {
            StatusGatewayPushResult.Failure -> Result.retry()
            StatusGatewayPushResult.Disabled,
            StatusGatewayPushResult.Success,
            -> Result.success()
        }
    }
}
