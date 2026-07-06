package me.realtime.mobile.background

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

class SyncStartupReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action !in SUPPORTED_ACTIONS) return
        StatusBackgroundSync.ensureActive(context)
    }

    private companion object {
        val SUPPORTED_ACTIONS = setOf(Intent.ACTION_BOOT_COMPLETED, Intent.ACTION_MY_PACKAGE_REPLACED)
    }
}
