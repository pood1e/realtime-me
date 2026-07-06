package me.realtime.mobile.debug

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.widget.Toast
import me.realtime.mobile.R
import me.realtime.mobile.background.StatusBackgroundSync
import me.realtime.mobile.state.StatusGatewayTokenStore

class DebugTokenReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action != ACTION_SET_STATUS_GATEWAY_TOKEN) return
        val token = intent.getStringExtra(EXTRA_TOKEN)?.trim().orEmpty()
        if (token.isEmpty()) {
            Toast.makeText(context, context.getString(R.string.debug_token_missing), Toast.LENGTH_SHORT).show()
            return
        }
        val appContext = context.applicationContext
        StatusGatewayTokenStore(appContext).save(token)
        StatusBackgroundSync.ensureActive(appContext)
        Toast.makeText(context, context.getString(R.string.debug_token_saved), Toast.LENGTH_SHORT).show()
    }

    private companion object {
        const val ACTION_SET_STATUS_GATEWAY_TOKEN = "me.realtime.mobile.debug.SET_STATUS_GATEWAY_TOKEN"
        const val EXTRA_TOKEN = "token"
    }
}
