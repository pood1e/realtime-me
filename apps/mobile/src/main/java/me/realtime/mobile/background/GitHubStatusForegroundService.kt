package me.realtime.mobile.background

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.ServiceInfo
import android.os.Build
import android.os.IBinder
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import me.realtime.mobile.R
import me.realtime.mobile.state.GitHubTokenStore
import me.realtime.mobile.state.SnapshotProcessor
import me.realtime.mobile.state.WatchSnapshotStore
import me.realtime.mobile.wear.WatchSnapshotReader
import me.realtime.protocol.toJavaInstant
import me.realtime.protocol.v1.ReportWatchSnapshotRequest
import me.realtime.protocol.v1.WatchSnapshot
import java.time.Duration
import java.time.Instant

class GitHubStatusForegroundService : Service() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var started = false

    override fun onCreate() {
        super.onCreate()
        startForegroundNotification()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        if (!GitHubTokenStore(this).hasToken()) {
            stopSelf(startId)
            return START_NOT_STICKY
        }

        if (!started) {
            started = true
            scope.launch { syncLoop() }
        }
        return START_STICKY
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onDestroy() {
        scope.cancel()
        super.onDestroy()
    }

    private suspend fun syncLoop() {
        while (scope.isActive) {
            if (!GitHubTokenStore(applicationContext).hasToken()) {
                stopSelf()
                return
            }
            syncOnce()
            delay(SYNC_INTERVAL.toMillis())
        }
    }

    private suspend fun syncOnce() {
        val processor = SnapshotProcessor(applicationContext)
        latestDataLayerSnapshot()?.let { snapshot ->
            processor.processSnapshot(snapshot)
            return
        }

        val storedSnapshot = WatchSnapshotStore(applicationContext).latest() ?: return
        if (Duration.between(storedSnapshot.receivedAt, Instant.now()) <= MAX_SNAPSHOT_AGE) {
            processor.processSnapshot(
                snapshot = storedSnapshot.snapshot,
                receivedAt = storedSnapshot.receivedAt,
            )
        }
    }

    private suspend fun latestDataLayerSnapshot(): WatchSnapshot? {
        val payload = runCatching { WatchSnapshotReader(applicationContext).latestPayload() }.getOrNull()
            ?: return null
        val request = runCatching { ReportWatchSnapshotRequest.parseFrom(payload) }.getOrNull()
            ?: return null
        if (!request.hasWatchSnapshot()) return null
        return request.watchSnapshot.takeIf { snapshot -> snapshot.isFresh() }
    }

    private fun WatchSnapshot.isFresh(): Boolean {
        if (!hasRecordTime()) return false
        val age = Duration.between(recordTime.toJavaInstant(), Instant.now())
        return age >= Duration.ZERO && age <= MAX_SNAPSHOT_AGE
    }

    private fun startForegroundNotification() {
        createNotificationChannel()
        val notification = Notification.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_launcher)
            .setContentTitle(getString(R.string.github_sync_notification_title))
            .setOngoing(true)
            .setCategory(Notification.CATEGORY_SERVICE)
            .build()

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            startForeground(NOTIFICATION_ID, notification, ServiceInfo.FOREGROUND_SERVICE_TYPE_DATA_SYNC)
        } else {
            startForeground(NOTIFICATION_ID, notification)
        }
    }

    private fun createNotificationChannel() {
        val manager = getSystemService(NotificationManager::class.java)
        val channel = NotificationChannel(
            CHANNEL_ID,
            getString(R.string.github_sync_notification_channel),
            NotificationManager.IMPORTANCE_LOW,
        )
        manager.createNotificationChannel(channel)
    }

    companion object {
        private const val CHANNEL_ID = "github_status_sync"
        private const val NOTIFICATION_ID = 200
        private val SYNC_INTERVAL: Duration = Duration.ofSeconds(10)
        private val MAX_SNAPSHOT_AGE: Duration = Duration.ofMinutes(10)

        fun start(context: Context) {
            runCatching {
                context.startForegroundService(Intent(context, GitHubStatusForegroundService::class.java))
            }
        }

        fun stop(context: Context) {
            context.stopService(Intent(context, GitHubStatusForegroundService::class.java))
        }
    }
}
