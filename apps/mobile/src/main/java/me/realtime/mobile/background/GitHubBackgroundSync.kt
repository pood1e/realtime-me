package me.realtime.mobile.background

import android.content.Context
import androidx.work.BackoffPolicy
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import me.realtime.mobile.state.GitHubTokenStore
import java.util.concurrent.TimeUnit

object GitHubBackgroundSync {
    fun ensureActive(context: Context) {
        val appContext = context.applicationContext
        if (!GitHubTokenStore(appContext).hasToken()) {
            cancel(appContext)
            GitHubStatusForegroundService.stop(appContext)
            return
        }

        ensureScheduled(appContext)
        enqueueNow(appContext)
        GitHubStatusForegroundService.start(appContext)
    }

    fun ensureScheduled(context: Context) {
        val appContext = context.applicationContext
        if (!GitHubTokenStore(appContext).hasToken()) {
            cancel(appContext)
            return
        }

        WorkManager.getInstance(appContext)
            .enqueueUniquePeriodicWork(
                PERIODIC_WORK_NAME,
                ExistingPeriodicWorkPolicy.UPDATE,
                PeriodicWorkRequestBuilder<GitHubStatusSyncWorker>(
                    PERIODIC_INTERVAL_MINUTES,
                    TimeUnit.MINUTES,
                )
                    .setConstraints(networkConstraints())
                    .build(),
            )
    }

    fun enqueueNow(context: Context) {
        val appContext = context.applicationContext
        if (!GitHubTokenStore(appContext).hasToken()) return

        WorkManager.getInstance(appContext)
            .enqueueUniqueWork(
                IMMEDIATE_WORK_NAME,
                ExistingWorkPolicy.REPLACE,
                OneTimeWorkRequestBuilder<GitHubStatusSyncWorker>()
                    .setConstraints(networkConstraints())
                    .setBackoffCriteria(
                        BackoffPolicy.EXPONENTIAL,
                        RETRY_BACKOFF_SECONDS,
                        TimeUnit.SECONDS,
                    )
                    .build(),
            )
    }

    fun cancel(context: Context) {
        val workManager = WorkManager.getInstance(context.applicationContext)
        workManager.cancelUniqueWork(IMMEDIATE_WORK_NAME)
        workManager.cancelUniqueWork(PERIODIC_WORK_NAME)
    }

    private fun networkConstraints(): Constraints = Constraints.Builder()
        .setRequiredNetworkType(NetworkType.CONNECTED)
        .build()

    private const val IMMEDIATE_WORK_NAME = "github_status_sync_now"
    private const val PERIODIC_WORK_NAME = "github_status_sync_periodic"
    private const val PERIODIC_INTERVAL_MINUTES = 15L
    private const val RETRY_BACKOFF_SECONDS = 30L
}
