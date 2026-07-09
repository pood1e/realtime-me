package me.realtime.mobile.state

import android.content.Context
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.withContext
import me.realtime.mobile.background.StatusBackgroundSync
import me.realtime.mobile.status.StatusDeviceIdentity
import me.realtime.mobile.wear.WatchSnapshotReader

/**
 * Owns the phone's local status stores and exposes them reactively, so the UI
 * observes changes instead of polling. It is also the single place that (re)arms
 * background sync when the gateway token changes.
 */
class StatusRepository(context: Context) {
    private val appContext = context.applicationContext
    private val tokenStore = StatusGatewayTokenStore(appContext)
    private val identityStore = StatusDeviceIdentity(appContext)
    private val snapshotStore = WatchSnapshotStore(appContext)
    private val snapshotReader = WatchSnapshotReader(appContext)
    private val snapshotProcessor = WatchSnapshotProcessor(appContext)

    fun hasToken(): Boolean = tokenStore.hasToken()

    fun watchSnapshots(): Flow<StoredWatchSnapshot?> = snapshotStore.changes()

    /** Returns false when the device's Keystore refused to store the token. */
    suspend fun saveToken(token: String): Boolean = withContext(Dispatchers.IO) {
        if (!tokenStore.save(token)) return@withContext false
        StatusBackgroundSync.ensureActive(appContext)
        true
    }

    suspend fun clearToken(): Unit = withContext(Dispatchers.IO) {
        tokenStore.clear()
        // The uid belongs to the gateway that minted it. Disconnecting means the
        // next connection may be to a different gateway, which would reject it.
        identityStore.clear()
        StatusBackgroundSync.ensureActive(appContext)
    }

    suspend fun ensureSyncActive(): Unit = withContext(Dispatchers.IO) {
        StatusBackgroundSync.ensureActive(appContext)
    }

    // Pulls whatever snapshot the watch has already published into the store so
    // the reactive flow reflects it even when no change event fired (e.g. the
    // screen was opened after the data item was last delivered).
    suspend fun refreshFromWatch(): Unit = withContext(Dispatchers.IO) {
        val payload = runCatching { snapshotReader.latestPayload() }.getOrNull() ?: return@withContext
        snapshotProcessor.process(payload)
    }
}
