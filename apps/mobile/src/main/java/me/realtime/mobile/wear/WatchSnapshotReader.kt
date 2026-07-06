package me.realtime.mobile.wear

import android.content.Context
import com.google.android.gms.tasks.Task
import com.google.android.gms.wearable.DataMapItem
import com.google.android.gms.wearable.Wearable
import kotlinx.coroutines.suspendCancellableCoroutine
import me.realtime.protocol.DataLayerContract
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

class WatchSnapshotReader(context: Context) {
    private val dataClient = Wearable.getDataClient(context)

    suspend fun latestPayload(): ByteArray? {
        val dataItems = dataClient.dataItems.await()
        dataItems.use { buffer ->
            for (dataItem in buffer) {
                if (dataItem.uri.path != DataLayerContract.WATCH_SNAPSHOT_PATH) continue
                return DataMapItem.fromDataItem(dataItem)
                    .dataMap
                    .getByteArray(DataLayerContract.SNAPSHOT_BYTES_KEY)
            }
        }
        return null
    }

    private suspend fun <T> Task<T>.await(): T = suspendCancellableCoroutine { continuation ->
        addOnSuccessListener { result -> continuation.resume(result) }
        addOnFailureListener { error -> continuation.resumeWithException(error) }
        addOnCanceledListener { continuation.cancel() }
    }
}
