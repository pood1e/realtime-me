package me.realtime.watch.wear

import android.content.Context
import android.util.Log
import com.google.android.gms.wearable.PutDataMapRequest
import com.google.android.gms.wearable.Wearable
import me.realtime.protocol.DataLayerContract
import me.realtime.protocol.v1.ReportWatchSnapshotRequest

class SnapshotPublisher(private val context: Context) {
    private val policy = PublishPolicy(context)

    fun publishIfAllowed(payload: ReportWatchSnapshotRequest) {
        if (!policy.shouldPublish(payload)) return
        publish(payload)
        policy.markPublished(payload)
    }

    fun publish(payload: ReportWatchSnapshotRequest) {
        val payloadBytes = payload.toByteArray()
        val request = PutDataMapRequest.create(DataLayerContract.WATCH_SNAPSHOT_PATH).apply {
            dataMap.putByteArray(DataLayerContract.SNAPSHOT_BYTES_KEY, payloadBytes)
            dataMap.putLong(DataLayerContract.PUBLISHED_TIME_KEY, System.currentTimeMillis())
        }.asPutDataRequest().setUrgent()

        Wearable.getDataClient(context)
            .putDataItem(request)
            .addOnFailureListener { error ->
                Log.w(TAG, "Unable to publish watch snapshot as a data item", error)
            }

        Wearable.getNodeClient(context)
            .connectedNodes
            .addOnSuccessListener { nodes ->
                nodes.forEach { node ->
                    Wearable.getMessageClient(context)
                        .sendMessage(node.id, DataLayerContract.WATCH_SNAPSHOT_PATH, payloadBytes)
                        .addOnFailureListener { error ->
                            Log.w(TAG, "Unable to publish watch snapshot as a message", error)
                        }
                }
            }
            .addOnFailureListener { error ->
                Log.w(TAG, "Unable to resolve connected phone nodes", error)
            }
    }

    private companion object {
        const val TAG = "SnapshotPublisher"
    }
}
