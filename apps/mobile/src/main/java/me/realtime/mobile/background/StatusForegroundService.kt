package me.realtime.mobile.background

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.content.pm.ServiceInfo
import android.net.ConnectivityManager
import android.net.Network
import android.os.Build
import android.os.IBinder
import androidx.core.content.ContextCompat
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import me.realtime.mobile.R
import me.realtime.mobile.state.StatusGatewayTokenStore
import java.time.Duration

/**
 * Keeps the phone's status fresh on the gateway. Pushes are event-driven — watch
 * data arrives via [me.realtime.mobile.wear.WatchSnapshotListenerService], and
 * this service pushes on connectivity and charging changes — with a single
 * low-frequency heartbeat as the fallback. Every trigger funnels through one
 * conflated channel so bursts coalesce into a single serialized push. The
 * gateway expires a phone report after 15 minutes, so the heartbeat has to stay
 * comfortably inside that window.
 */
class StatusForegroundService : Service() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)
    private val pushRequests = Channel<Unit>(Channel.CONFLATED)
    private var started = false
    private var foregroundActive = false
    private var networkCallback: ConnectivityManager.NetworkCallback? = null
    private var chargingReceiver: BroadcastReceiver? = null

    override fun onCreate() {
        super.onCreate()
        foregroundActive = startForegroundNotification()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        if (!foregroundActive || !hasSyncToken()) {
            stopSelf(startId)
            return START_NOT_STICKY
        }

        if (!started) {
            started = true
            registerEventTriggers()
            scope.launch { pushConsumer() }
            scope.launch { heartbeat() }
        }
        requestPush()
        return START_STICKY
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onDestroy() {
        unregisterEventTriggers()
        scope.cancel()
        super.onDestroy()
    }

    // The single push path: coalesced requests are handled one at a time so an
    // event burst never fans out into overlapping network calls.
    private suspend fun pushConsumer() {
        for (request in pushRequests) {
            if (!hasSyncToken()) {
                stopSelf()
                return
            }
            StatusSyncRunner(applicationContext).syncLatest()
        }
    }

    private suspend fun heartbeat() {
        while (scope.isActive) {
            delay(HEARTBEAT_INTERVAL.toMillis())
            requestPush()
        }
    }

    private fun requestPush() {
        pushRequests.trySend(Unit)
    }

    private fun registerEventTriggers() {
        val connectivityManager = getSystemService(ConnectivityManager::class.java)
        val callback = object : ConnectivityManager.NetworkCallback() {
            override fun onAvailable(network: Network) = requestPush()
            override fun onLost(network: Network) = requestPush()
        }
        runCatching { connectivityManager?.registerDefaultNetworkCallback(callback) }
            .onSuccess { networkCallback = callback }

        val receiver = object : BroadcastReceiver() {
            override fun onReceive(context: Context?, intent: Intent?) = requestPush()
        }
        val filter = IntentFilter().apply {
            addAction(Intent.ACTION_POWER_CONNECTED)
            addAction(Intent.ACTION_POWER_DISCONNECTED)
        }
        ContextCompat.registerReceiver(this, receiver, filter, ContextCompat.RECEIVER_NOT_EXPORTED)
        chargingReceiver = receiver
    }

    private fun unregisterEventTriggers() {
        networkCallback?.let { callback ->
            runCatching { getSystemService(ConnectivityManager::class.java)?.unregisterNetworkCallback(callback) }
        }
        networkCallback = null
        chargingReceiver?.let { receiver -> runCatching { unregisterReceiver(receiver) } }
        chargingReceiver = null
    }

    private fun hasSyncToken(): Boolean = StatusGatewayTokenStore(this).hasToken()

    // The OS can recreate this sticky service in the background, where starting a
    // foreground service is disallowed and startForeground() throws. Treat that as
    // a soft failure and let the service go — the periodic worker keeps reporting
    // until the app is foregrounded and can arm the service again.
    private fun startForegroundNotification(): Boolean = runCatching {
        createNotificationChannel()
        val notification = Notification.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_launcher)
            .setContentTitle(getString(R.string.status_sync_notification_title))
            .setOngoing(true)
            .setCategory(Notification.CATEGORY_SERVICE)
            .build()

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            startForeground(NOTIFICATION_ID, notification, ServiceInfo.FOREGROUND_SERVICE_TYPE_SPECIAL_USE)
        } else {
            startForeground(NOTIFICATION_ID, notification)
        }
    }.isSuccess

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
        // The event triggers below carry every real transition; this is only the
        // fallback for a phone whose state has not changed. At 30s it was ~2,880
        // pushes a day on an idle phone, each one a full report with Bluetooth
        // profile scans.
        private val HEARTBEAT_INTERVAL: Duration = Duration.ofMinutes(5)

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
