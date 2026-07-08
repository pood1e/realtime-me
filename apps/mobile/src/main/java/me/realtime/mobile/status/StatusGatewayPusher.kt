package me.realtime.mobile.status

import android.content.Context
import android.util.Log
import me.realtime.mobile.state.StatusGatewayTokenStore
import me.realtime.mobile.state.WatchSnapshotStore

class StatusGatewayPusher(context: Context) {
    private val appContext = context.applicationContext
    private val tokenStore = StatusGatewayTokenStore(appContext)
    private val identityStore = StatusDeviceIdentity(appContext)
    private val snapshotStore = WatchSnapshotStore(appContext)
    private val payloadBuilder = StatusGatewayPayloadBuilder(appContext)
    private val client = StatusGatewayClient()

    fun pushLatest(): StatusGatewayPushResult {
        val token = tokenStore.token() ?: run {
            Log.w(TAG, "Status gateway token is not configured")
            return StatusGatewayPushResult.Disabled
        }
        val deviceUid = ensureDeviceUid(token) ?: run {
            Log.w(TAG, "Status gateway enrollment did not return a device uid")
            return StatusGatewayPushResult.Failure
        }
        // Always report the latest known watch snapshot; its record_time conveys
        // freshness, so the page shows last-known state with a timestamp instead
        // of dropping the watch whenever it goes idle.
        val request = payloadBuilder.build(deviceUid, snapshotStore.latest())
        return client.reportMobile(token, request)
    }

    // Enroll once to obtain the gateway-owned device uid, then reuse the cached
    // value on every subsequent report.
    private fun ensureDeviceUid(token: String): String? {
        identityStore.uid()?.let { return it }
        val uid = client.enroll(token, payloadBuilder.enrollRequest()) ?: return null
        identityStore.save(uid)
        return uid
    }

    private companion object {
        const val TAG = "RealtimeStatus"
    }
}
