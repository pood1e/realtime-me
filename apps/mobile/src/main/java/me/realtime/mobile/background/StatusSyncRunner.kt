package me.realtime.mobile.background

import android.content.Context
import me.realtime.mobile.state.WatchSnapshotProcessor
import me.realtime.mobile.status.StatusGatewayPushResult
import me.realtime.mobile.status.StatusGatewayPusher
import me.realtime.mobile.wear.WatchSnapshotReader

class StatusSyncRunner(context: Context) {
    private val appContext = context.applicationContext
    private val processor = WatchSnapshotProcessor(appContext)
    private val pusher = StatusGatewayPusher(appContext)
    private val reader = WatchSnapshotReader(appContext)

    suspend fun syncLatest(): StatusGatewayPushResult {
        runCatching { reader.latestPayload() }
            .getOrNull()
            ?.let(processor::process)
        return pusher.pushLatest()
    }

    fun syncPayload(payload: ByteArray): StatusGatewayPushResult {
        processor.process(payload)
        return pusher.pushLatest()
    }
}
