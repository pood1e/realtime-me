package me.realtime.watch.boot

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import me.realtime.watch.health.PassiveHealthRegistration

class PassiveHealthStartupReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action !in SUPPORTED_ACTIONS) return
        if (!PassiveHealthRegistration.hasRequiredPermissions(context)) return
        PassiveHealthRegistration.enqueue(context.applicationContext)
    }

    private companion object {
        val SUPPORTED_ACTIONS = setOf(Intent.ACTION_BOOT_COMPLETED, Intent.ACTION_MY_PACKAGE_REPLACED)
    }
}
