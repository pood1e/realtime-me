package me.realtime.mobile.status

import android.content.Context
import androidx.core.content.edit

/**
 * Stores the gateway-assigned device uid. The identifier is minted by the
 * gateway during enrollment; the phone only persists and echoes it back and
 * never constructs one itself.
 */
class StatusDeviceIdentity(context: Context) {
    private val preferences = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun uid(): String? = preferences.getString(DEVICE_UID_KEY, null)?.takeIf { it.isNotEmpty() }

    fun save(uid: String) {
        preferences.edit { putString(DEVICE_UID_KEY, uid) }
    }

    fun clear() {
        preferences.edit { remove(DEVICE_UID_KEY) }
    }

    private companion object {
        const val PREFS_NAME = "status_device_identity"
        const val DEVICE_UID_KEY = "device_uid"
    }
}
