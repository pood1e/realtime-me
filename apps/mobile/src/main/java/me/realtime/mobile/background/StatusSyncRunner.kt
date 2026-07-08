package me.realtime.mobile.background

import android.content.Context
import android.util.Log
import kotlinx.coroutines.CancellationException
import me.realtime.mobile.state.WatchSnapshotProcessor
import me.realtime.mobile.status.StatusGatewayPushResult
import me.realtime.mobile.status.StatusGatewayPusher
import me.realtime.mobile.wear.WatchSnapshotReader

/**
 * The single funnel for every status push (foreground service, watch-data
 * listener, and the periodic worker). A sync attempt is total: it always
 * resolves to a [StatusGatewayPushResult] and never throws, so a transient
 * failure deep in the Android or network stack cannot crash the process from
 * the background coroutines that drive it.
 */
class StatusSyncRunner(context: Context) {
    private val appContext = context.applicationContext
    private val processor = WatchSnapshotProcessor(appContext)
    private val pusher = StatusGatewayPusher(appContext)
    private val reader = WatchSnapshotReader(appContext)

    suspend fun syncLatest(): StatusGatewayPushResult = guarded {
        runCatching { reader.latestPayload() }
            .getOrNull()
            ?.let(processor::process)
        pusher.pushLatest()
    }

    fun syncPayload(payload: ByteArray): StatusGatewayPushResult = guarded {
        processor.process(payload)
        pusher.pushLatest()
    }

    private inline fun guarded(block: () -> StatusGatewayPushResult): StatusGatewayPushResult =
        try {
            block()
        } catch (cancellation: CancellationException) {
            throw cancellation
        } catch (throwable: Throwable) {
            Log.w(TAG, "Status sync failed: ${throwable.javaClass.simpleName}")
            StatusGatewayPushResult.Failure
        }

    private companion object {
        const val TAG = "RealtimeStatus"
    }
}
