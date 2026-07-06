package me.realtime.mobile.github

import java.time.Instant

data class GitHubStatus(
    val message: String,
    val emoji: String,
    val expiresAt: Instant,
) {
    val signature: String = listOf(message, emoji).joinToString(separator = "|")
}
