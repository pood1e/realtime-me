package me.realtime.mobile.status

import android.content.Context
import androidx.core.content.edit
import java.util.UUID

class StatusDeviceIdentity(context: Context) {
    private val preferences = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    fun id(): String {
        preferences.getString(DEVICE_ID_KEY, null)?.let { return it }
        return "phone-${UUID.randomUUID()}".also { value ->
            preferences.edit { putString(DEVICE_ID_KEY, value) }
        }
    }

    private companion object {
        const val PREFS_NAME = "status_device_identity"
        const val DEVICE_ID_KEY = "device_id"
    }
}
