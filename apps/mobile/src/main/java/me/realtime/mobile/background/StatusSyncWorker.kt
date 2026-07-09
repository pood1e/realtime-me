package me.realtime.mobile.background

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import me.realtime.mobile.state.StatusGatewayTokenStore
import me.realtime.mobile.status.StatusGatewayPushResult

class StatusSyncWorker(
    appContext: Context,
    workerParameters: WorkerParameters,
) : CoroutineWorker(appContext, workerParameters) {
    override suspend fun doWork(): Result {
        if (!StatusGatewayTokenStore(applicationContext).hasToken()) return Result.success()

        // doWork runs on Dispatchers.Default, one thread per core. The push blocks
        // on HttpURLConnection, so it belongs on the pool that expects to wait.
        val result = withContext(Dispatchers.IO) { StatusSyncRunner(applicationContext).syncLatest() }
        return when (result) {
            StatusGatewayPushResult.Failure -> Result.retry()
            StatusGatewayPushResult.Disabled,
            StatusGatewayPushResult.Success,
            -> Result.success()
        }
    }
}
