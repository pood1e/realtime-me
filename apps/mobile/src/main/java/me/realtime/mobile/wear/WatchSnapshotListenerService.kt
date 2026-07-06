package me.realtime.mobile.wear

import com.google.android.gms.wearable.DataEvent
import com.google.android.gms.wearable.DataEventBuffer
import com.google.android.gms.wearable.DataMapItem
import com.google.android.gms.wearable.MessageEvent
import com.google.android.gms.wearable.WearableListenerService
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch
import me.realtime.mobile.state.WatchSnapshotProcessor
import me.realtime.protocol.DataLayerContract

class WatchSnapshotListenerService : WearableListenerService() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    override fun onDataChanged(dataEvents: DataEventBuffer) {
        try {
            for (event in dataEvents) {
                val payload = event.snapshotPayload() ?: continue
                scope.launch {
                    WatchSnapshotProcessor(applicationContext).process(payload)
                }
            }
        } finally {
            dataEvents.release()
        }
    }

    override fun onMessageReceived(messageEvent: MessageEvent) {
        if (messageEvent.path != DataLayerContract.WATCH_SNAPSHOT_PATH) return
        scope.launch {
            WatchSnapshotProcessor(applicationContext).process(messageEvent.data)
        }
    }

    override fun onDestroy() {
        scope.cancel()
        super.onDestroy()
    }

    private fun DataEvent.snapshotPayload(): ByteArray? {
        if (type != DataEvent.TYPE_CHANGED) return null
        if (dataItem.uri.path != DataLayerContract.WATCH_SNAPSHOT_PATH) return null
        return DataMapItem.fromDataItem(dataItem)
            .dataMap
            .getByteArray(DataLayerContract.SNAPSHOT_BYTES_KEY)
    }
}
