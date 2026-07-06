package me.realtime.mobile.status

import android.util.Log
import me.realtime.mobile.BuildConfig
import org.json.JSONObject
import java.io.IOException
import java.net.HttpURLConnection
import java.net.URL
import java.nio.charset.StandardCharsets

class StatusGatewayClient(
    private val endpoints: List<String> = configuredEndpoints(),
) {
    fun push(token: String, payload: JSONObject): StatusGatewayPushResult {
        if (endpoints.isEmpty()) {
            Log.w(TAG, "Status gateway endpoint is not configured")
            return StatusGatewayPushResult.Disabled
        }
        for (endpoint in endpoints) {
            if (post(endpoint, token, payload)) return StatusGatewayPushResult.Success
        }
        Log.w(TAG, "Status gateway push failed for all configured endpoints")
        return StatusGatewayPushResult.Failure
    }

    private fun post(endpoint: String, token: String, payload: JSONObject): Boolean {
        var connection: HttpURLConnection? = null

        return try {
            connection = (URL("${endpoint.trimEnd('/')}/api/ingest/mobile").openConnection() as HttpURLConnection).apply {
                requestMethod = "POST"
                connectTimeout = TIMEOUT_MS
                readTimeout = TIMEOUT_MS
                doOutput = true
                setRequestProperty("Accept", "application/json")
                setRequestProperty("Content-Type", "application/json; charset=utf-8")
                setRequestProperty("Authorization", "Bearer $token")
            }
            connection.outputStream.use { output -> output.write(payload.toString().toByteArray(StandardCharsets.UTF_8)) }
            val responseCode = connection.responseCode
            if (responseCode !in 200..299) {
                Log.w(TAG, "Status gateway rejected payload with HTTP $responseCode")
            }
            responseCode in 200..299
        } catch (exception: IOException) {
            Log.w(TAG, "Status gateway network error: ${exception.javaClass.simpleName}")
            false
        } catch (exception: RuntimeException) {
            Log.w(TAG, "Status gateway request error: ${exception.javaClass.simpleName}")
            false
        } finally {
            connection?.disconnect()
        }
    }

    private companion object {
        const val TAG = "RealtimeStatus"
        const val TIMEOUT_MS = 5_000

        fun configuredEndpoints(): List<String> = buildList {
            BuildConfig.STATUS_GATEWAY_LAN_URL.trim().takeIf { it.isNotEmpty() }?.let(::add)
            BuildConfig.STATUS_GATEWAY_PUBLIC_URL.trim().takeIf { it.isNotEmpty() }?.let(::add)
        }.distinct()
    }
}

sealed class StatusGatewayPushResult {
    data object Success : StatusGatewayPushResult()
    data object Failure : StatusGatewayPushResult()
    data object Disabled : StatusGatewayPushResult()
}
