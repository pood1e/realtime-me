package me.realtime.mobile.state

import android.content.Context

class StatusGatewayTokenStore(context: Context) {
    private val store = EncryptedSecretStore(
        context = context,
        prefsName = PREFS_NAME,
        keyAlias = KEY_ALIAS,
    )

    fun save(token: String) = store.save(token)

    fun token(): String? = store.value()

    fun hasToken(): Boolean = store.hasValue()

    fun clear() = store.clear()

    private companion object {
        const val PREFS_NAME = "status_gateway_secrets"
        const val KEY_ALIAS = "realtime_me_status_gateway_token"
    }
}
