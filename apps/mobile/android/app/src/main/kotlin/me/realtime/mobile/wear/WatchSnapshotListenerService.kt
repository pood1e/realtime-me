package me.realtime.mobile.wear

import com.google.android.gms.wearable.DataEvent
import com.google.android.gms.wearable.DataEventBuffer
import com.google.android.gms.wearable.DataMapItem
import com.google.android.gms.wearable.WearableListenerService
import me.realtime.mobile.background.StatusSyncRunner
import me.realtime.protocol.DataLayerContract

class WatchSnapshotListenerService : WearableListenerService() {
    // Wearable delivers this callback on a background thread and may tear the
    // service down as soon as it returns, so the push runs to completion here.
    // Launching it into a service-scoped coroutine let onDestroy cancel an
    // in-flight enroll-plus-report, leaving the snapshot for the 15-minute
    // worker to rediscover.
    override fun onDataChanged(dataEvents: DataEventBuffer) {
        val payloads = try {
            dataEvents.mapNotNull { event -> event.snapshotPayload() }
        } finally {
            dataEvents.release()
        }
        if (payloads.isEmpty()) return
        StatusSyncRunner(applicationContext).syncPayloads(payloads)
    }

    private fun DataEvent.snapshotPayload(): ByteArray? {
        if (type != DataEvent.TYPE_CHANGED) return null
        if (dataItem.uri.path != DataLayerContract.WATCH_SNAPSHOT_PATH) return null
        return DataMapItem.fromDataItem(dataItem)
            .dataMap
            .getByteArray(DataLayerContract.SNAPSHOT_BYTES_KEY)
    }
}
