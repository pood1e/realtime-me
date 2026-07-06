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
import me.realtime.mobile.state.StatusGatewayTokenStore
import java.time.Duration

class StatusForegroundService : Service() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private var started = false

    override fun onCreate() {
        super.onCreate()
        startForegroundNotification()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        if (!hasSyncToken()) {
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
            if (!hasSyncToken()) {
                stopSelf()
                return
            }
            syncOnce()
            delay(SYNC_INTERVAL.toMillis())
        }
    }

    private suspend fun syncOnce() {
        StatusSyncRunner(applicationContext).syncLatest()
    }

    private fun hasSyncToken(): Boolean = StatusGatewayTokenStore(this).hasToken()

    private fun startForegroundNotification() {
        createNotificationChannel()
        val notification = Notification.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_launcher)
            .setContentTitle(getString(R.string.status_sync_notification_title))
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
            getString(R.string.status_sync_notification_channel),
            NotificationManager.IMPORTANCE_LOW,
        )
        manager.createNotificationChannel(channel)
    }

    companion object {
        private const val CHANNEL_ID = "status_gateway_sync"
        private const val NOTIFICATION_ID = 200
        private val SYNC_INTERVAL: Duration = Duration.ofSeconds(10)

        fun start(context: Context) {
            runCatching {
                context.startForegroundService(Intent(context, StatusForegroundService::class.java))
            }
        }

        fun stop(context: Context) {
            context.stopService(Intent(context, StatusForegroundService::class.java))
        }
    }
}
