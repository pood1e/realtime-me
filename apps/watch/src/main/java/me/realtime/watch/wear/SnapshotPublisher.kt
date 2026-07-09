package me.realtime.watch.wear

import android.content.Context
import android.util.Log
import com.google.android.gms.wearable.PutDataMapRequest
import com.google.android.gms.wearable.Wearable
import me.realtime.protocol.DataLayerContract
import me.realtime.protocol.v1.ReportWatchSnapshotRequest

class SnapshotPublisher(private val context: Context) {
    private val policy = PublishPolicy(context)

    // The Health Services binder thread and the registration worker publish
    // through separate instances that share one SharedPreferences file, so the
    // gate's read-modify-write runs under a process-wide lock. Without it both
    // threads can clear the same gate and publish the same snapshot twice.
    fun publishIfAllowed(payload: ReportWatchSnapshotRequest) {
        synchronized(PUBLISH_LOCK) {
            if (!policy.shouldPublish(payload)) return
            publish(payload)
            policy.markPublished(payload)
        }
    }

    private fun publish(payload: ReportWatchSnapshotRequest) {
        // Publish over a single transport (the durable Data Layer item). The
        // per-publish timestamp guarantees the item content changes so delivery
        // is not suppressed; a parallel MessageClient send would only cause the
        // phone to process — and forward to the gateway — every snapshot twice.
        val request = PutDataMapRequest.create(DataLayerContract.WATCH_SNAPSHOT_PATH).apply {
            dataMap.putByteArray(DataLayerContract.SNAPSHOT_BYTES_KEY, payload.toByteArray())
            dataMap.putLong(DataLayerContract.PUBLISHED_TIME_KEY, System.currentTimeMillis())
        }.asPutDataRequest().setUrgent()

        Wearable.getDataClient(context)
            .putDataItem(request)
            .addOnFailureListener { error ->
                Log.w(TAG, "Unable to publish watch snapshot as a data item", error)
            }
    }

    private companion object {
        const val TAG = "SnapshotPublisher"

        // Guards the publish gate's read-modify-write across all instances.
        private val PUBLISH_LOCK = Any()
    }
}
