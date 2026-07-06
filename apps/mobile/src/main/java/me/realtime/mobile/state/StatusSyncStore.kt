package me.realtime.mobile.state

import android.content.Context
import androidx.core.content.edit
import me.realtime.mobile.github.GitHubStatus
import java.time.Duration
import java.time.Instant

class StatusSyncStore(context: Context) {
    private val preferences = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun shouldUpdate(status: GitHubStatus, now: Instant): Boolean {
        val lastAttempt = preferences.getLong(LAST_ATTEMPT_TIME_KEY, 0L)
        if (lastAttempt > 0 && Duration.between(Instant.ofEpochMilli(lastAttempt), now) < MIN_UPDATE_INTERVAL) {
            return false
        }

        val lastUpdate = preferences.getLong(LAST_UPDATE_TIME_KEY, 0L)
        if (lastUpdate == 0L) return true

        val elapsed = Duration.between(Instant.ofEpochMilli(lastUpdate), now)
        if (elapsed < MIN_UPDATE_INTERVAL) return false

        val lastSignature = preferences.getString(LAST_SIGNATURE_KEY, null)
        if (lastSignature == status.signature && elapsed < SAME_STATUS_REFRESH_INTERVAL) return false

        return true
    }

    fun markAttempt(now: Instant) {
        preferences.edit {
            putLong(LAST_ATTEMPT_TIME_KEY, now.toEpochMilli())
        }
    }

    fun markUpdated(status: GitHubStatus, now: Instant) {
        preferences.edit {
            putString(LAST_SIGNATURE_KEY, status.signature)
            putLong(LAST_ATTEMPT_TIME_KEY, now.toEpochMilli())
            putLong(LAST_UPDATE_TIME_KEY, now.toEpochMilli())
            putString(LAST_MESSAGE_KEY, status.message)
            remove(LAST_ERROR_KEY)
        }
    }

    fun markSkipped(message: String) {
        preferences.edit {
            putString(LAST_MESSAGE_KEY, message)
        }
    }

    fun markError(message: String) {
        preferences.edit {
            putString(LAST_ERROR_KEY, message)
        }
    }

    fun hasError(): Boolean {
        return preferences.getString(LAST_ERROR_KEY, null) != null
    }

    fun clearError() {
        preferences.edit {
            remove(LAST_ERROR_KEY)
        }
    }

    fun summary(): String {
        val lastMessage = preferences.getString(LAST_MESSAGE_KEY, "No watch snapshot processed yet")
        val lastError = preferences.getString(LAST_ERROR_KEY, null)
        val lastUpdate = preferences.getLong(LAST_UPDATE_TIME_KEY, 0L)
        val updateText = if (lastUpdate == 0L) "never" else Instant.ofEpochMilli(lastUpdate).toString()
        return buildString {
            append("Last GitHub update: ").append(updateText).append('\n')
            append("Last status: ").append(lastMessage)
            if (lastError != null) append('\n').append("Last error: ").append(lastError)
        }
    }

    private companion object {
        const val PREFS_NAME = "status_sync"
        const val LAST_SIGNATURE_KEY = "last_signature"
        const val LAST_ATTEMPT_TIME_KEY = "last_attempt_time_ms"
        const val LAST_UPDATE_TIME_KEY = "last_update_time_ms"
        const val LAST_MESSAGE_KEY = "last_message"
        const val LAST_ERROR_KEY = "last_error"
        val MIN_UPDATE_INTERVAL: Duration = Duration.ofSeconds(10)
        val SAME_STATUS_REFRESH_INTERVAL: Duration = Duration.ofSeconds(10)
    }
}
