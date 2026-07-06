package me.realtime.mobile.state

import android.content.Context
import me.realtime.mobile.github.GitHubStatusClient
import me.realtime.mobile.github.GitHubStatusFormatter
import me.realtime.mobile.github.GitHubUpdateResult
import me.realtime.protocol.v1.ReportWatchSnapshotRequest
import me.realtime.protocol.v1.WatchSnapshot
import java.time.Clock
import java.time.Instant

class SnapshotProcessor(
    private val tokenStore: GitHubTokenStore,
    private val syncStore: StatusSyncStore,
    private val snapshotStore: WatchSnapshotStore,
    private val statusFormatter: GitHubStatusFormatter,
    private val statusClient: GitHubStatusClient,
    private val clock: Clock = Clock.systemUTC(),
) {
    constructor(context: Context) : this(
        tokenStore = GitHubTokenStore(context),
        syncStore = StatusSyncStore(context),
        snapshotStore = WatchSnapshotStore(context),
        statusFormatter = GitHubStatusFormatter(),
        statusClient = GitHubStatusClient(),
    )

    suspend fun process(payloadBytes: ByteArray): SnapshotProcessResult {
        val payload = runCatching { ReportWatchSnapshotRequest.parseFrom(payloadBytes) }.getOrElse {
            syncStore.markError("Invalid watch snapshot")
            return SnapshotProcessResult("Invalid watch snapshot")
        }
        if (!payload.hasWatchSnapshot()) {
            syncStore.markError("Watch snapshot is missing")
            return SnapshotProcessResult("Watch snapshot is missing")
        }

        return processSnapshot(payload.watchSnapshot)
    }

    suspend fun processSnapshot(snapshot: WatchSnapshot, receivedAt: Instant = clock.instant()): SnapshotProcessResult {
        val now = clock.instant()
        snapshotStore.save(snapshot, receivedAt)

        val token = tokenStore.token()
        if (token == null) {
            syncStore.markError("GitHub token is not configured")
            return SnapshotProcessResult("GitHub token is not configured")
        }

        val status = statusFormatter.format(snapshot, now)
        if (!syncStore.shouldUpdate(status, now)) {
            syncStore.markSkipped(status.message)
            return SnapshotProcessResult("Skipped; status changed too recently")
        }

        syncStore.markAttempt(now)
        return when (val result = statusClient.changeStatus(token, status)) {
            is GitHubUpdateResult.Success -> {
                syncStore.markUpdated(status, now)
                SnapshotProcessResult("GitHub status updated")
            }
            is GitHubUpdateResult.Failure -> {
                syncStore.markError(result.message)
                SnapshotProcessResult(result.message, retryable = result.retryable)
            }
        }
    }
}

data class SnapshotProcessResult(
    val message: String,
    val retryable: Boolean = false,
)
