package me.realtime.mobile.github

import kotlinx.coroutines.delay
import org.json.JSONObject
import java.io.IOException
import java.net.HttpURLConnection
import java.net.URL
import java.nio.charset.StandardCharsets

class GitHubStatusClient(
    private val endpoint: URL = URL("https://api.github.com/graphql"),
) {
    fun viewerProfile(token: String): GitHubProfileResult {
        return when (val response = sendRequest(requestBody(VIEWER_QUERY), retryableMutation = false, token = token)) {
            is GraphQLResponse.Success -> parseViewerProfile(response.body)
            is GraphQLResponse.Failure -> GitHubProfileResult.Failure(response.retryable, response.message)
        }
    }

    suspend fun changeStatus(token: String, status: GitHubStatus): GitHubUpdateResult {
        var lastFailure: GitHubUpdateResult.Failure? = null
        repeat(MAX_ATTEMPTS) { attempt ->
            val result = sendStatus(token, status)
            if (result is GitHubUpdateResult.Success) return result
            result as GitHubUpdateResult.Failure
            lastFailure = result
            if (!result.retryable || attempt == MAX_ATTEMPTS - 1) return result
            delay(RETRY_BACKOFF_MS * (attempt + 1))
        }
        return lastFailure ?: GitHubUpdateResult.Failure(retryable = false, message = "GitHub update failed")
    }

    private fun sendStatus(token: String, status: GitHubStatus): GitHubUpdateResult {
        return when (val response = sendRequest(requestBody(CHANGE_STATUS_MUTATION, changeStatusInput(status)), retryableMutation = true, token = token)) {
            is GraphQLResponse.Success -> GitHubUpdateResult.Success
            is GraphQLResponse.Failure -> GitHubUpdateResult.Failure(response.retryable, response.message)
        }
    }

    private fun sendRequest(body: String, retryableMutation: Boolean, token: String): GraphQLResponse {
        val connection = (endpoint.openConnection() as HttpURLConnection).apply {
            requestMethod = "POST"
            connectTimeout = TIMEOUT_MS
            readTimeout = TIMEOUT_MS
            doOutput = true
            setRequestProperty("Accept", "application/vnd.github+json")
            setRequestProperty("Content-Type", "application/json; charset=utf-8")
            setRequestProperty("User-Agent", "realtime-me-android")
            setRequestProperty("Authorization", "Bearer $token")
        }

        return try {
            connection.outputStream.use { output -> output.write(body.toByteArray(StandardCharsets.UTF_8)) }
            val responseCode = connection.responseCode
            val responseBody = readResponse(connection)
            if (responseCode in 200..299 && !JSONObject(responseBody).has("errors")) {
                GraphQLResponse.Success(responseBody)
            } else {
                GraphQLResponse.Failure(
                    retryable = responseCode == HTTP_TOO_MANY_REQUESTS || responseCode >= HTTP_SERVER_ERROR,
                    message = rejectionMessage(responseBody, responseCode, retryableMutation),
                )
            }
        } catch (_: IOException) {
            GraphQLResponse.Failure(retryable = true, message = "Network error while connecting to GitHub")
        } catch (_: RuntimeException) {
            GraphQLResponse.Failure(retryable = false, message = "Invalid GitHub response")
        } finally {
            connection.disconnect()
        }
    }

    private fun parseViewerProfile(body: String): GitHubProfileResult {
        val viewer = JSONObject(body)
            .getJSONObject("data")
            .getJSONObject("viewer")
        val status = viewer.optJSONObject("status")
        return GitHubProfileResult.Success(
            GitHubProfile(
                login = viewer.getString("login"),
                status = status?.let {
                    GitHubProfileStatus(
                        message = it.nullableString("message"),
                        emoji = it.nullableString("emoji"),
                        expiresAt = it.nullableString("expiresAt"),
                    )
                },
            ),
        )
    }

    private fun rejectionMessage(body: String, responseCode: Int, retryableMutation: Boolean): String {
        val operation = if (retryableMutation) "status update" else "profile check"
        val errorMessage = runCatching {
            JSONObject(body)
                .optJSONArray("errors")
                ?.optJSONObject(0)
                ?.nullableString("message")
        }.getOrNull()
        return errorMessage?.let { "GitHub rejected the $operation: $it" }
            ?: "GitHub rejected the $operation (HTTP $responseCode)"
    }

    private fun requestBody(query: String, variables: JSONObject = JSONObject()): String {
        return JSONObject()
            .put("query", query)
            .put("variables", variables)
            .toString()
    }

    private fun changeStatusInput(status: GitHubStatus): JSONObject {
        val input = JSONObject()
            .put("message", status.message)
            .put("emoji", status.emoji)
            .put("expiresAt", status.expiresAt.toString())
            .put("limitedAvailability", false)
        return JSONObject().put("input", input)
    }

    private fun readResponse(connection: HttpURLConnection): String {
        val stream = if (connection.responseCode in 200..299) {
            connection.inputStream
        } else {
            connection.errorStream ?: connection.inputStream
        }
        return stream.bufferedReader(StandardCharsets.UTF_8).use { it.readText() }
    }

    private fun JSONObject.nullableString(name: String): String? {
        if (isNull(name)) return null
        return optString(name).takeIf { it.isNotBlank() }
    }

    private companion object {
        const val TIMEOUT_MS = 10_000
        const val MAX_ATTEMPTS = 2
        const val RETRY_BACKOFF_MS = 500L
        const val HTTP_TOO_MANY_REQUESTS = 429
        const val HTTP_SERVER_ERROR = 500
        const val VIEWER_QUERY = """
            query ViewerStatus {
              viewer {
                login
                status {
                  message
                  emoji
                  expiresAt
                }
              }
            }
        """
        const val CHANGE_STATUS_MUTATION = """
            mutation ChangeUserStatus(${'$'}input: ChangeUserStatusInput!) {
              changeUserStatus(input: ${'$'}input) {
                status {
                  message
                  emoji
                  expiresAt
                }
              }
            }
        """
    }
}

data class GitHubProfile(
    val login: String,
    val status: GitHubProfileStatus?,
)

data class GitHubProfileStatus(
    val message: String?,
    val emoji: String?,
    val expiresAt: String?,
)

sealed class GitHubProfileResult {
    data class Success(val profile: GitHubProfile) : GitHubProfileResult()

    data class Failure(
        val retryable: Boolean,
        val message: String,
    ) : GitHubProfileResult()
}

private sealed class GraphQLResponse {
    data class Success(val body: String) : GraphQLResponse()

    data class Failure(
        val retryable: Boolean,
        val message: String,
    ) : GraphQLResponse()
}

sealed class GitHubUpdateResult {
    data object Success : GitHubUpdateResult()

    data class Failure(
        val retryable: Boolean,
        val message: String,
    ) : GitHubUpdateResult()
}
