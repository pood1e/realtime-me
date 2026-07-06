package me.realtime.mobile.state

import android.content.Context
import androidx.core.content.edit
import me.realtime.protocol.v1.WatchSnapshot
import java.time.Instant
import java.util.Base64

class WatchSnapshotStore(context: Context) {
    private val preferences = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun save(snapshot: WatchSnapshot, receivedAt: Instant) {
        preferences.edit {
            putString(SNAPSHOT_BYTES_KEY, Base64.getEncoder().encodeToString(snapshot.toByteArray()))
            putLong(RECEIVED_AT_KEY, receivedAt.toEpochMilli())
        }
    }

    fun latest(): StoredWatchSnapshot? {
        val encodedSnapshot = preferences.getString(SNAPSHOT_BYTES_KEY, null) ?: return null
        val receivedAt = preferences.getLong(RECEIVED_AT_KEY, 0L).takeIf { it > 0 } ?: return null
        return runCatching {
            StoredWatchSnapshot(
                snapshot = WatchSnapshot.parseFrom(Base64.getDecoder().decode(encodedSnapshot)),
                receivedAt = Instant.ofEpochMilli(receivedAt),
            )
        }.getOrNull()
    }

    private companion object {
        const val PREFS_NAME = "watch_snapshot"
        const val SNAPSHOT_BYTES_KEY = "snapshot_bytes"
        const val RECEIVED_AT_KEY = "received_at_ms"
    }
}

data class StoredWatchSnapshot(
    val snapshot: WatchSnapshot,
    val receivedAt: Instant,
)
