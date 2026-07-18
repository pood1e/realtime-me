package me.realtime.mobile.platform

import android.content.Context
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch
import me.realtime.mobile.state.StatusRepository
import me.realtime.mobile.state.StoredWatchSnapshot

class StatusBridgeHost(
    context: Context,
    private val hasRequiredPermissions: () -> Boolean,
    private val requestPermissions: ((Result<Boolean>) -> Unit) -> Unit,
) : StatusHostApi {
    private val repository = StatusRepository(context.applicationContext)
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)

    override fun getSnapshot(): StatusSnapshotData = repository.latestSnapshot().toBridgeData()

    override fun hasToken(): Boolean = repository.hasToken()

    override fun hasRequiredPermissions(): Boolean = hasRequiredPermissions.invoke()

    override fun saveToken(token: String, callback: (Result<Boolean>) -> Unit) = launch(callback) {
        repository.saveToken(token)
    }

    override fun clearToken(callback: (Result<Unit>) -> Unit) = launch(callback) {
        repository.clearToken()
    }

    override fun refresh(callback: (Result<Unit>) -> Unit) = launch(callback) {
        repository.refresh()
    }

    override fun requestPermissions(callback: (Result<Boolean>) -> Unit) {
        requestPermissions.invoke(callback)
    }

    fun close() {
        scope.cancel()
    }

    private fun <T> launch(callback: (Result<T>) -> Unit, block: suspend () -> T) {
        scope.launch {
            callback(
                try {
                    Result.success(block())
                } catch (error: Throwable) {
                    Result.failure(error)
                },
            )
        }
    }
}

class StatusSnapshotEvents(context: Context) : SnapshotsStreamHandler() {
    private val repository = StatusRepository(context.applicationContext)
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)
    private var collection: Job? = null

    override fun onListen(p0: Any?, sink: PigeonEventSink<StatusSnapshotData>) {
        collection?.cancel()
        collection = scope.launch {
            repository.watchSnapshots().collectLatest { snapshot ->
                sink.success(snapshot.toBridgeData())
            }
        }
    }

    override fun onCancel(p0: Any?) {
        collection?.cancel()
        collection = null
    }

    fun close() {
        collection?.cancel()
        scope.cancel()
    }
}

private fun StoredWatchSnapshot?.toBridgeData(): StatusSnapshotData = StatusSnapshotData(
    revision = this?.revision ?: 0L,
    receivedAtEpochMillis = this?.receivedAt?.toEpochMilli() ?: 0L,
    protobufBytes = this?.snapshot?.toByteArray(),
)
