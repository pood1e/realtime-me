package me.realtime.mobile.status

import android.util.Log
import me.realtime.mobile.BuildConfig
import me.realtime.protocol.v1.EnrollDeviceRequest
import me.realtime.protocol.v1.EnrollDeviceResponse
import me.realtime.protocol.v1.ReportMobileStatusRequest
import java.io.IOException
import java.net.HttpURLConnection
import java.net.URL

/**
 * Talks to the gateway's ConnectRPC services. Unary Connect calls are a plain
 * POST whose body is the binary-serialized request message (Content-Type
 * application/proto) and whose 2xx response body is the serialized response
 * message, so the javalite runtime needs no JSON support.
 */
class StatusGatewayClient(
    private val endpoints: List<String> = configuredEndpoints(),
) {
    /** Enrolls the phone and returns the gateway-assigned device uid, or null. */
    fun enroll(token: String, request: EnrollDeviceRequest): String? {
        for (endpoint in endpoints) {
            val body = post(endpoint, ENROLL_PROCEDURE, token, request.toByteArray()) ?: continue
            val uid = runCatching { EnrollDeviceResponse.parseFrom(body).deviceUid }.getOrNull()
            if (!uid.isNullOrEmpty()) return uid
        }
        return null
    }

    fun reportMobile(token: String, request: ReportMobileStatusRequest): StatusGatewayPushResult {
        if (endpoints.isEmpty()) {
            Log.w(TAG, "Status gateway endpoint is not configured")
            return StatusGatewayPushResult.Disabled
        }
        for (endpoint in endpoints) {
            if (post(endpoint, REPORT_PROCEDURE, token, request.toByteArray()) != null) {
                return StatusGatewayPushResult.Success
            }
        }
        Log.w(TAG, "Status gateway push failed for all configured endpoints")
        return StatusGatewayPushResult.Failure
    }

    private fun post(endpoint: String, procedure: String, token: String, body: ByteArray): ByteArray? {
        var connection: HttpURLConnection? = null
        return try {
            val url = URL("${endpoint.trimEnd('/')}/$procedure")
            connection = (url.openConnection() as HttpURLConnection).apply {
                requestMethod = "POST"
                connectTimeout = TIMEOUT_MS
                readTimeout = TIMEOUT_MS
                doOutput = true
                setRequestProperty("Accept", CONTENT_TYPE)
                setRequestProperty("Content-Type", CONTENT_TYPE)
                setRequestProperty("Authorization", "Bearer $token")
            }
            connection.outputStream.use { it.write(body) }
            val responseCode = connection.responseCode
            if (responseCode !in 200..299) {
                Log.w(TAG, "Status gateway rejected $procedure with HTTP $responseCode")
                return null
            }
            connection.inputStream.use { it.readBytes() }
        } catch (exception: IOException) {
            Log.w(TAG, "Status gateway network error: ${exception.javaClass.simpleName}")
            null
        } catch (exception: RuntimeException) {
            Log.w(TAG, "Status gateway request error: ${exception.javaClass.simpleName}")
            null
        } finally {
            connection?.disconnect()
        }
    }

    private companion object {
        const val TAG = "RealtimeStatus"
        const val TIMEOUT_MS = 5_000
        const val CONTENT_TYPE = "application/proto"
        const val ENROLL_PROCEDURE = "realtime.me.v1.EnrollmentService/EnrollDevice"
        const val REPORT_PROCEDURE = "realtime.me.v1.IngestService/ReportMobileStatus"

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
