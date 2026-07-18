package me.realtime.mobile.nintendo

import android.content.Context
import android.os.SystemClock
import android.util.Log
import com.google.protobuf.Timestamp
import me.realtime.status.v1.OnlineState
import me.realtime.status.v1.SwitchPresence
import org.json.JSONObject
import java.io.File
import java.net.HttpURLConnection
import java.net.URL
import java.util.Base64
import java.util.concurrent.TimeUnit

class NintendoSwitchPresenceReader(context: Context) {
    private val cipher = CoralCipher()
    private val nintendoDataDirectory = runCatching {
        context.createPackageContext(NSO_PACKAGE, Context.CONTEXT_IGNORE_SECURITY)
            .applicationInfo
            .dataDir
    }.getOrNull()

    fun read(): SwitchPresence? = runCatching {
        var token = usableCoralToken() ?: run {
            Log.w(TAG, "Coral token unavailable")
            return null
        }
        val userId = userId() ?: run {
            Log.w(TAG, "Coral user id unavailable")
            return null
        }
        val body = "{\"parameter\":{\"id\":$userId}}"
        var response = coralCall(token, SHOW_SELF_PATH, body)
        if (response.optInt("status") == INVALID_TOKEN_STATUS) {
            token = refreshCoralToken(token) ?: return null
            response = coralCall(token, SHOW_SELF_PATH, body)
        }
        parsePresence(response).also { presence ->
            if (presence != null) clearRefreshBackoff()
            Log.i(TAG, "Switch presence parsed=${presence != null} state=${presence?.state} game=${presence?.gameName.orEmpty()}")
        }
    }.onFailure { throwable ->
        Log.w(TAG, "Switch presence read failed: ${throwable.javaClass.simpleName}")
    }.getOrNull()

    private fun coralCall(token: String, path: String, bodyText: String): JSONObject {
        val url = "$API_BASE$path"
        val encrypted = cipher.encryptRequest(APP_VERSION, url, bodyText)
        val connection = (URL(url).openConnection() as HttpURLConnection).apply {
            requestMethod = "POST"
            connectTimeout = TIMEOUT_MS
            readTimeout = TIMEOUT_MS
            doOutput = true
            setRequestProperty("User-Agent", USER_AGENT)
            setRequestProperty("Authorization", "Bearer $token")
            setRequestProperty("Content-Type", "application/octet-stream")
        }
        return try {
            connection.outputStream.use { it.write(encrypted) }
            val responseCode = connection.responseCode
            val stream = if (responseCode in 200..299) {
                connection.inputStream
            } else {
                connection.errorStream ?: return JSONObject()
            }
            val plaintext = cipher.decryptBody(stream.use { it.readBytes() })
            JSONObject(plaintext)
        } finally {
            connection.disconnect()
        }
    }

    private fun parsePresence(response: JSONObject): SwitchPresence? {
        val presence = response.optJSONObject("result")?.optJSONObject("presence") ?: return null
        val builder = SwitchPresence.newBuilder()
            .setState(presence.optString("state").toOnlineState())
            .setFetchTime(timestamp(System.currentTimeMillis() / 1000))
        presence.optEpochSeconds("updatedAt")?.let { builder.setPresenceUpdateTime(timestamp(it)) }
        presence.optEpochSeconds("logoutAt")?.let { builder.setLogoutTime(timestamp(it)) }
        presence.optJSONObject("game")?.let { game ->
            game.optString("name").takeIf(String::isNotBlank)?.let(builder::setGameName)
            firstNonBlank(game.optString("titleId"), game.optString("id"))?.let(builder::setTitleId)
            game.optString("imageUri").takeIf(String::isNotBlank)?.let(builder::setImageUri)
        }
        return builder.build()
    }

    private fun usableCoralToken(): String? {
        val token = coralToken()
        token?.takeIf { it.ttlSeconds > 0L }?.let { return it.value }
        return refreshCoralToken(token?.value)
    }

    private fun refreshCoralToken(previousToken: String?): String? = synchronized(REFRESH_LOCK) {
        coralToken()
            ?.takeIf { it.ttlSeconds > 0L && it.value != previousToken }
            ?.let { return@synchronized it.value }

        refreshRetryDelaySeconds().takeIf { it > 0L }?.let { delaySeconds ->
            Log.w(TAG, "Nintendo credential refresh deferred for ${delaySeconds}s")
            return@synchronized null
        }

        val shouldRestorePreviousApp = requestNsoRefresh() ?: run {
            recordRefreshFailure()
            return@synchronized null
        }
        try {
            repeat(REFRESH_POLL_ATTEMPTS) {
                Thread.sleep(REFRESH_POLL_INTERVAL_MS)
                coralToken()
                    ?.takeIf { it.ttlSeconds > 0L && it.value != previousToken }
                    ?.let {
                        clearRefreshBackoff()
                        return@synchronized it.value
                    }
            }
        } finally {
            restorePreviousApp(shouldRestorePreviousApp)
        }
        recordRefreshFailure()
        null
    }

    private fun requestNsoRefresh(): Boolean? {
        val previousAppWasNso = isNsoResumed()
        val result = rootCommand(START_NSO_COMMAND)?.toString(Charsets.UTF_8) ?: return null
        if (result.contains("Error:")) {
            Log.w(TAG, "Nintendo credential refresh activity failed")
            return null
        }
        Log.i(TAG, "Started Nintendo Switch Online credential refresh")
        return !previousAppWasNso
    }

    private fun restorePreviousApp(shouldRestore: Boolean) {
        if (!shouldRestore) return
        repeat(RESTORE_BACK_ATTEMPTS) {
            if (!isNsoResumed()) return
            rootCommand(BACK_COMMAND)
            Thread.sleep(RESTORE_BACK_DELAY_MS)
        }
    }

    private fun isNsoResumed(): Boolean = rootCommand(RESUMED_ACTIVITY_COMMAND)
        ?.toString(Charsets.UTF_8)
        ?.contains(" $NSO_PACKAGE/") == true

    private fun refreshRetryDelaySeconds(): Long {
        val remainingMs = nextRefreshAttemptAtMs - SystemClock.elapsedRealtime()
        return if (remainingMs > 0L) (remainingMs + 999L) / 1_000L else 0L
    }

    private fun recordRefreshFailure() {
        val multiplier = 1L shl refreshFailureCount.coerceAtMost(MAX_BACKOFF_EXPONENT)
        val delayMs = (REFRESH_INITIAL_BACKOFF_MS * multiplier).coerceAtMost(REFRESH_MAX_BACKOFF_MS)
        refreshFailureCount++
        nextRefreshAttemptAtMs = SystemClock.elapsedRealtime() + delayMs
        Log.w(TAG, "Nintendo credential refresh failed; retrying in ${delayMs / 60_000L}m")
    }

    private fun clearRefreshBackoff() {
        synchronized(REFRESH_LOCK) {
            refreshFailureCount = 0
            nextRefreshAttemptAtMs = 0L
        }
    }

    private fun coralToken(): CoralToken? {
        val path = nintendoFile(TOKEN_DATASTORE_RELATIVE_PATH) ?: return null
        val data = suRead(path)?.toString(Charsets.ISO_8859_1) ?: return null
        val candidates = TOKEN_PATTERN.findAll(data).map { it.value }.toList()
        return candidates
            .mapNotNull { token ->
                val payload = jwtPayload(token)
                val ttl = payload.ttlSeconds() ?: return@mapNotNull null
                if (payload.optString("iss") != CORAL_TOKEN_ISS) return@mapNotNull null
                CoralToken(token, ttl)
            }
            .maxByOrNull(CoralToken::ttlSeconds)
    }

    private fun userId(): Long? {
        val path = nintendoFile(LOGIN_USER_RELATIVE_PATH) ?: return null
        val data = suRead(path)?.toString(Charsets.UTF_8) ?: return null
        return JSONObject(data).optLong("userId").takeIf { it > 0L }
    }

    private fun nintendoFile(relativePath: String): String? {
        val dataDirectory = nintendoDataDirectory ?: return null
        return File(dataDirectory, relativePath).path
    }

    private fun suRead(path: String): ByteArray? {
        return rootCommand("cat '$path'", path.substringAfterLast('/'))?.takeIf { it.isNotEmpty() }
    }

    private fun rootCommand(command: String, operation: String = "root command"): ByteArray? {
        val process = ProcessBuilder("su", "-mm", "-c", command).start()
        if (!process.waitFor(SU_TIMEOUT_SECONDS, TimeUnit.SECONDS)) {
            Log.w(TAG, "$operation timed out")
            process.destroyForcibly()
            return null
        }
        val output = process.inputStream.readBytes()
        if (process.exitValue() != 0) {
            val error = process.errorStream.readBytes().toString(Charsets.UTF_8).trim()
            Log.w(TAG, "$operation failed rc=${process.exitValue()} err=${error.take(240)}")
            return null
        }
        return output
    }

    private fun jwtPayload(token: String): JSONObject {
        val parts = token.split('.')
        if (parts.size < 2) return JSONObject()
        val payload = parts[1].padEnd(parts[1].length + ((4 - parts[1].length % 4) % 4), '=')
        return JSONObject(String(Base64.getUrlDecoder().decode(payload), Charsets.UTF_8))
    }

    private fun JSONObject.ttlSeconds(): Long? {
        val exp = optLong("exp", 0L)
        if (exp <= 0L) return null
        return exp - System.currentTimeMillis() / 1000
    }

    private fun String.toOnlineState(): OnlineState = when (uppercase()) {
        "ONLINE" -> OnlineState.ONLINE_STATE_ONLINE
        "OFFLINE" -> OnlineState.ONLINE_STATE_OFFLINE
        else -> OnlineState.ONLINE_STATE_UNSPECIFIED
    }

    private fun JSONObject.optEpochSeconds(name: String): Long? = when {
        !has(name) || isNull(name) -> null
        else -> optLong(name).takeIf { it > 0L }
    }

    private fun timestamp(seconds: Long): Timestamp = Timestamp.newBuilder().setSeconds(seconds).build()

    private fun firstNonBlank(vararg values: String): String? = values.firstOrNull { it.isNotBlank() }

    private data class CoralToken(val value: String, val ttlSeconds: Long)

    private companion object {
        const val TAG = "SwitchPresence"
        const val API_BASE = "https://api-lp1.znc.srv.nintendo.net"
        const val CORAL_TOKEN_ISS = "api-lp1.znc.srv.nintendo.net"
        const val APP_VERSION = "3.4.0"
        const val USER_AGENT = "com.nintendo.znca/$APP_VERSION(Android/12)"
        const val SHOW_SELF_PATH = "/v4/User/ShowSelf"
        const val NSO_PACKAGE = "com.nintendo.znca"
        const val NSO_BOOT_ACTIVITY = "com.nintendo.coral.ui.boot.BootActivity"
        const val INVALID_TOKEN_STATUS = 9403
        const val TOKEN_DATASTORE_RELATIVE_PATH = "files/datastore/preferences_token_datastore.preferences_pb"
        const val LOGIN_USER_RELATIVE_PATH = "files/datastore/login_user.pb"
        const val TIMEOUT_MS = 10_000
        const val SU_TIMEOUT_SECONDS = 5L
        const val REFRESH_POLL_ATTEMPTS = 30
        const val REFRESH_POLL_INTERVAL_MS = 1_000L
        const val RESTORE_BACK_ATTEMPTS = 2
        const val RESTORE_BACK_DELAY_MS = 1_000L
        const val REFRESH_INITIAL_BACKOFF_MS = 5 * 60_000L
        const val REFRESH_MAX_BACKOFF_MS = 60 * 60_000L
        const val MAX_BACKOFF_EXPONENT = 4
        const val RESUMED_ACTIVITY_COMMAND =
            "dumpsys activity activities | grep -m1 'mResumedActivity:'"
        const val START_NSO_COMMAND =
            "am start --activity-no-animation -n $NSO_PACKAGE/$NSO_BOOT_ACTIVITY 2>&1"
        const val BACK_COMMAND = "input keyevent KEYCODE_BACK"
        val REFRESH_LOCK = Any()
        val TOKEN_PATTERN = Regex("[A-Za-z0-9_\\-.]{80,}")
        var refreshFailureCount = 0
        var nextRefreshAttemptAtMs = 0L
    }
}
