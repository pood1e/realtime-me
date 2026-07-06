package me.realtime.mobile.status

import android.content.Context
import android.util.Log
import me.realtime.mobile.state.StatusGatewayTokenStore
import me.realtime.mobile.state.StoredWatchSnapshot
import me.realtime.mobile.state.WatchSnapshotStore
import java.time.Duration
import java.time.Instant

class StatusGatewayPusher(context: Context) {
    private val appContext = context.applicationContext
    private val tokenStore = StatusGatewayTokenStore(appContext)
    private val snapshotStore = WatchSnapshotStore(appContext)
    private val payloadBuilder = StatusGatewayPayloadBuilder(appContext)
    private val client = StatusGatewayClient()

    fun pushLatest(): StatusGatewayPushResult {
        val token = tokenStore.token() ?: run {
            Log.w(TAG, "Status gateway token is not configured")
            return StatusGatewayPushResult.Disabled
        }
        val payload = payloadBuilder.build(snapshotStore.latest()?.takeIfFresh())
        return client.push(token, payload)
    }

    private fun StoredWatchSnapshot.takeIfFresh(now: Instant = Instant.now()): StoredWatchSnapshot? {
        val age = Duration.between(receivedAt, now)
        return takeIf { age >= Duration.ZERO && age <= MAX_WATCH_SNAPSHOT_AGE }
    }

    private companion object {
        const val TAG = "RealtimeStatus"
        val MAX_WATCH_SNAPSHOT_AGE: Duration = Duration.ofMinutes(2)
    }
}
