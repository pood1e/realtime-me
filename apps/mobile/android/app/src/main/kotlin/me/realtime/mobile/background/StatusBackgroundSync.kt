package me.realtime.mobile.background

import android.content.Context
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.NetworkType
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import me.realtime.mobile.state.StatusGatewayTokenStore
import java.util.concurrent.TimeUnit

/**
 * Arms the phone's background sync. While the app can run a foreground service,
 * [StatusForegroundService] is the push engine (event-driven + a low-frequency
 * heartbeat, which also covers the first push on start). The periodic worker is
 * the safety net that keeps reporting when the OS has torn that service down.
 */
object StatusBackgroundSync {
    fun ensureActive(context: Context) {
        val appContext = context.applicationContext
        if (!StatusGatewayTokenStore(appContext).hasToken()) {
            cancel(appContext)
            StatusForegroundService.stop(appContext)
            return
        }

        ensureScheduled(appContext)
        StatusForegroundService.start(appContext)
    }

    fun ensureScheduled(context: Context) {
        val appContext = context.applicationContext
        if (!StatusGatewayTokenStore(appContext).hasToken()) {
            cancel(appContext)
            return
        }

        WorkManager.getInstance(appContext)
            .enqueueUniquePeriodicWork(
                PERIODIC_WORK_NAME,
                ExistingPeriodicWorkPolicy.UPDATE,
                PeriodicWorkRequestBuilder<StatusSyncWorker>(
                    PERIODIC_INTERVAL_MINUTES,
                    TimeUnit.MINUTES,
                )
                    .setConstraints(networkConstraints())
                    .build(),
            )
    }

    fun cancel(context: Context) {
        WorkManager.getInstance(context.applicationContext).cancelUniqueWork(PERIODIC_WORK_NAME)
    }

    private fun networkConstraints(): Constraints = Constraints.Builder()
        .setRequiredNetworkType(NetworkType.CONNECTED)
        .build()

    private const val PERIODIC_WORK_NAME = "status_gateway_sync_periodic"
    private const val PERIODIC_INTERVAL_MINUTES = 15L
}
