package me.realtime.mobile.status

import android.util.Log
import me.realtime.mobile.BuildConfig
import me.realtime.status.v1.EnrollDeviceRequest
import me.realtime.status.v1.EnrollDeviceResponse
import me.realtime.status.v1.ReportMobileStatusRequest
import org.json.JSONObject
import java.io.IOException
import java.net.HttpURLConnection
import java.net.URL
import java.nio.charset.StandardCharsets

/**
 * Talks to the gateway's ConnectRPC services. Unary Connect calls are a plain
 * POST whose body is the binary-serialized request message (Content-Type
 * application/proto) and whose 2xx response body is the serialized response
 * message, so the javalite runtime needs no JSON support.
 */
class StatusGatewayClient(
    private val endpoints: List<String> = configuredEndpoints(),
) {
    val isConfigured: Boolean = endpoints.isNotEmpty()

    /** Enrolls the phone and returns the gateway-assigned device uid, or null. */
    fun enroll(token: String, request: EnrollDeviceRequest): String? {
        for (endpoint in endpoints) {
            val body = (post(endpoint, ENROLL_PROCEDURE, token, request.toByteArray()) as? Response.Body)?.bytes ?: continue
            val uid = runCatching { EnrollDeviceResponse.parseFrom(body).deviceUid }.getOrNull()
            if (!uid.isNullOrEmpty()) return uid
        }
        return null
    }

    internal fun reportMobile(token: String, request: ReportMobileStatusRequest): ReportOutcome {
        for (endpoint in endpoints) {
            when (post(endpoint, REPORT_PROCEDURE, token, request.toByteArray())) {
                is Response.Body -> return ReportOutcome.Success
                Response.DeviceUnenrolled -> return ReportOutcome.DeviceUnenrolled
                Response.Unreachable -> continue
            }
        }
        Log.w(TAG, "Status gateway push failed for all configured endpoints")
        return ReportOutcome.Failure
    }

    private fun post(endpoint: String, procedure: String, token: String, body: ByteArray): Response {
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
            if (responseCode in 200..299) {
                return Response.Body(connection.inputStream.use { it.readBytes() })
            }
            val errorCode = connection.connectErrorCode()
            Log.w(TAG, "Status gateway rejected $procedure with HTTP $responseCode ($errorCode)")
            if (errorCode == NOT_FOUND_CODE) Response.DeviceUnenrolled else Response.Unreachable
        } catch (exception: IOException) {
            Log.w(TAG, "Status gateway network error: ${exception.javaClass.simpleName}")
            Response.Unreachable
        } catch (exception: RuntimeException) {
            Log.w(TAG, "Status gateway request error: ${exception.javaClass.simpleName}")
            Response.Unreachable
        } finally {
            connection?.disconnect()
        }
    }

    // A Connect unary error carries its code in a JSON body whatever the request
    // codec, and the HTTP status alone cannot distinguish the gateway's answer
    // from an intermediary's. Read the code the protocol actually defines.
    private fun HttpURLConnection.connectErrorCode(): String? {
        val body = runCatching { errorStream?.use { it.readBytes() } }.getOrNull() ?: return null
        return runCatching {
            JSONObject(String(body, StandardCharsets.UTF_8)).optString("code")
        }.getOrNull()?.takeIf { it.isNotEmpty() }
    }

    private sealed interface Response {
        class Body(val bytes: ByteArray) : Response
        data object DeviceUnenrolled : Response
        data object Unreachable : Response
    }

    private companion object {
        const val TAG = "RealtimeStatus"
        const val TIMEOUT_MS = 5_000
        const val NOT_FOUND_CODE = "not_found"
        const val CONTENT_TYPE = "application/proto"
        const val ENROLL_PROCEDURE = "realtime.me.v1.EnrollmentService/EnrollDevice"
        const val REPORT_PROCEDURE = "realtime.me.v1.IngestService/ReportMobileStatus"

        fun configuredEndpoints(): List<String> = buildList {
            BuildConfig.STATUS_GATEWAY_LAN_URL.trim().takeIf { it.isNotEmpty() }?.let(::add)
            BuildConfig.STATUS_GATEWAY_PUBLIC_URL.trim().takeIf { it.isNotEmpty() }?.let(::add)
        }.distinct()
    }
}

/** How the gateway answered one report attempt. Internal to the status layer. */
internal enum class ReportOutcome {
    Success,
    Failure,

    /** The gateway does not know the device uid the phone presented. */
    DeviceUnenrolled,
}

sealed class StatusGatewayPushResult {
    data object Success : StatusGatewayPushResult()
    data object Failure : StatusGatewayPushResult()
    data object Disabled : StatusGatewayPushResult()
}
