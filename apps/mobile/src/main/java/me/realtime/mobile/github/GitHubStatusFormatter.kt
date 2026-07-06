package me.realtime.mobile.github

import me.realtime.protocol.v1.ChargeState
import me.realtime.protocol.v1.WatchSnapshot
import me.realtime.protocol.v1.WristState
import java.time.Duration
import java.time.Instant
import java.util.Locale

class GitHubStatusFormatter {
    fun format(snapshot: WatchSnapshot, now: Instant): GitHubStatus {
        val message = buildMessage(snapshot)
        return GitHubStatus(
            message = message,
            emoji = emoji(snapshot),
            expiresAt = now.plus(STATUS_TTL),
        )
    }

    private fun buildMessage(snapshot: WatchSnapshot): String {
        val segments = buildList {
            if (snapshot.watchState.wristState != WristState.WRIST_STATE_OFF_WRIST &&
                snapshot.heartRate.beatsPerMinute > 0
            ) {
                add("❤️${snapshot.heartRate.beatsPerMinute}")
            }
            if (snapshot.activityTotals.steps > 0) {
                add("👣${compactCount(snapshot.activityTotals.steps)}")
            }
            if (snapshot.watchState.batteryPercent > 0) {
                add("${batteryEmoji(snapshot)}${snapshot.watchState.batteryPercent}%")
            }
            if (snapshot.watchState.wristState == WristState.WRIST_STATE_OFF_WRIST) {
                add("💤")
            }
        }
        val message = segments.ifEmpty { listOf("⌚synced") }.joinToString(separator = " · ")
        return message.take(MAX_MESSAGE_LENGTH)
    }

    private fun batteryEmoji(snapshot: WatchSnapshot): String = when {
        snapshot.watchState.chargeState == ChargeState.CHARGE_STATE_CHARGING -> "🔌"
        snapshot.watchState.batteryPercent in 1..14 -> "🪫"
        else -> "🔋"
    }

    private fun emoji(snapshot: WatchSnapshot): String = when {
        snapshot.watchState.wristState == WristState.WRIST_STATE_OFF_WRIST -> "💤"
        snapshot.watchState.chargeState == ChargeState.CHARGE_STATE_CHARGING -> "🔌"
        snapshot.watchState.batteryPercent in 1..14 -> "🪫"
        else -> "⌚"
    }

    private fun compactCount(value: Int): String {
        if (value < 1000) return value.toString()
        return String.format(Locale.US, "%.1fk", value / 1000.0)
            .replace(".0k", "k")
    }

    private companion object {
        const val MAX_MESSAGE_LENGTH = 80
        val STATUS_TTL: Duration = Duration.ofMinutes(20)
    }
}
