package me.realtime.mobile

import android.Manifest
import android.content.ActivityNotFoundException
import android.content.ClipData
import android.content.ClipboardManager
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.enableEdgeToEdge
import androidx.activity.compose.setContent
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.DirectionsWalk
import androidx.compose.material.icons.outlined.AccessTime
import androidx.compose.material.icons.outlined.BatteryChargingFull
import androidx.compose.material.icons.outlined.BatteryFull
import androidx.compose.material.icons.outlined.CheckCircle
import androidx.compose.material.icons.outlined.ContentPaste
import androidx.compose.material.icons.outlined.Delete
import androidx.compose.material.icons.outlined.ErrorOutline
import androidx.compose.material.icons.outlined.Favorite
import androidx.compose.material.icons.outlined.HourglassEmpty
import androidx.compose.material.icons.outlined.LinkOff
import androidx.compose.material.icons.outlined.OpenInBrowser
import androidx.compose.material.icons.outlined.WatchOff
import androidx.compose.material3.AssistChip
import androidx.compose.material3.ElevatedCard
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilledIconButton
import androidx.compose.material3.FilledTonalIconButton
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.core.net.toUri
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.MainScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import me.realtime.mobile.background.GitHubBackgroundSync
import me.realtime.mobile.background.GitHubStatusForegroundService
import me.realtime.mobile.github.GitHubProfile
import me.realtime.mobile.github.GitHubProfileResult
import me.realtime.mobile.github.GitHubStatusClient
import me.realtime.mobile.state.GitHubTokenStore
import me.realtime.mobile.state.SnapshotProcessor
import me.realtime.mobile.state.StatusSyncStore
import me.realtime.mobile.state.StoredWatchSnapshot
import me.realtime.mobile.state.WatchSnapshotStore
import me.realtime.mobile.wear.WatchSnapshotReader
import me.realtime.protocol.v1.ChargeState
import me.realtime.protocol.v1.WatchSnapshot
import me.realtime.protocol.v1.WristState
import java.text.NumberFormat
import java.time.Instant
import java.time.LocalDate
import java.time.ZoneId
import java.time.format.DateTimeFormatter

class MainActivity : ComponentActivity() {
    private val scope = MainScope()
    private val numberFormat = NumberFormat.getIntegerInstance()
    private val zoneId = ZoneId.systemDefault()
    private val timeFormatter = DateTimeFormatter.ofPattern("HH:mm:ss").withZone(zoneId)
    private val dateTimeFormatter = DateTimeFormatter.ofPattern("MMM d, HH:mm:ss").withZone(zoneId)
    private var refreshJob: Job? = null

    private var githubAccount by mutableStateOf<GitHubAccountUiState>(GitHubAccountUiState.Missing)
    private var githubSyncFailed by mutableStateOf(false)
    private var watchData by mutableStateOf<WatchDataUiState>(WatchDataUiState.Empty)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        requestNotificationPermissionIfNeeded()
        enableEdgeToEdge()
        setContent {
            RealtimeMobileScreen(
                githubAccount = githubAccount,
                githubSyncFailed = githubSyncFailed,
                watchData = watchData,
                onRequestGitHubToken = ::requestGitHubToken,
                onImportCopiedToken = ::importCopiedToken,
                onDisconnectGitHub = ::disconnectGitHub,
            )
        }
    }

    override fun onStart() {
        super.onStart()
        startBackgroundSyncIfConfigured()
        startRefreshLoop()
    }

    override fun onResume() {
        super.onResume()
        refreshStatus()
        refreshGitHubAccount()
    }

    override fun onStop() {
        refreshJob?.cancel()
        refreshJob = null
        super.onStop()
    }

    override fun onDestroy() {
        scope.cancel()
        super.onDestroy()
    }

    private fun requestGitHubToken() {
        val intent = Intent(Intent.ACTION_VIEW, GITHUB_TOKEN_REQUEST_URI.toUri())
        try {
            startActivity(intent)
        } catch (_: ActivityNotFoundException) {
            Toast.makeText(this, R.string.github_token_browser_missing, Toast.LENGTH_LONG).show()
        }
    }

    private fun importCopiedToken() {
        val clipboard = getSystemService(ClipboardManager::class.java)
        val token = clipboard?.primaryClip
            ?.takeIf { it.itemCount > 0 }
            ?.getItemAt(0)
            ?.coerceToText(this)
            ?.toString()
            ?.trim()
            .orEmpty()
        if (!token.looksLikeGitHubToken()) {
            Toast.makeText(this, R.string.github_token_clipboard_missing, Toast.LENGTH_SHORT).show()
            return
        }

        scope.launch {
            withContext(Dispatchers.IO) {
                GitHubTokenStore(applicationContext).save(token)
                StatusSyncStore(applicationContext).clearError()
            }
            clipboard?.clearToken()
            withContext(Dispatchers.IO) { GitHubBackgroundSync.ensureActive(applicationContext) }
            Toast.makeText(this@MainActivity, R.string.github_token_saved, Toast.LENGTH_SHORT).show()
            refreshStatus()
            refreshGitHubAccount()
        }
    }

    private fun disconnectGitHub() {
        scope.launch {
            withContext(Dispatchers.IO) {
                GitHubTokenStore(applicationContext).clear()
                StatusSyncStore(applicationContext).clearError()
                GitHubBackgroundSync.cancel(applicationContext)
                GitHubStatusForegroundService.stop(applicationContext)
            }
            githubAccount = GitHubAccountUiState.Missing
            githubSyncFailed = false
            Toast.makeText(this@MainActivity, getString(R.string.github_disconnected), Toast.LENGTH_SHORT).show()
            refreshStatus()
        }
    }

    private fun requestNotificationPermissionIfNeeded() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.TIRAMISU) return
        if (checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) == PackageManager.PERMISSION_GRANTED) return
        requestPermissions(arrayOf(Manifest.permission.POST_NOTIFICATIONS), NOTIFICATION_PERMISSION_REQUEST_CODE)
    }

    private fun startBackgroundSyncIfConfigured() {
        scope.launch(Dispatchers.IO) {
            GitHubBackgroundSync.ensureActive(applicationContext)
        }
    }

    private fun refreshGitHubAccount() {
        if (githubAccount is GitHubAccountUiState.Checking) return

        scope.launch {
            val token = withContext(Dispatchers.IO) { GitHubTokenStore(applicationContext).token() }
            if (token == null) {
                githubAccount = GitHubAccountUiState.Missing
                return@launch
            }

            githubAccount = GitHubAccountUiState.Checking
            githubAccount = when (val result = withContext(Dispatchers.IO) { GitHubStatusClient().viewerProfile(token) }) {
                is GitHubProfileResult.Success -> result.profile.toUiState()
                is GitHubProfileResult.Failure -> GitHubAccountUiState.Error
            }
        }
    }

    private fun startRefreshLoop() {
        if (refreshJob?.isActive == true) return
        refreshJob = scope.launch {
            while (isActive) {
                refreshStatusOnce()
                delay(REFRESH_INTERVAL_MS)
            }
        }
    }

    private fun refreshStatus() {
        scope.launch { refreshStatusOnce() }
    }

    private suspend fun refreshStatusOnce() {
        val state = withContext(Dispatchers.IO) {
            refreshLatestWatchSnapshot()
            StatusState(
                hasToken = GitHubTokenStore(applicationContext).hasToken(),
                githubSyncFailed = StatusSyncStore(applicationContext).hasError(),
                watchData = WatchSnapshotStore(applicationContext).latest()?.toUiState() ?: WatchDataUiState.Empty,
            )
        }
        applyTokenPresence(state.hasToken)
        githubSyncFailed = state.hasToken && state.githubSyncFailed
        watchData = state.watchData
    }

    private fun applyTokenPresence(hasToken: Boolean) {
        githubAccount = when {
            !hasToken -> GitHubAccountUiState.Missing
            githubAccount is GitHubAccountUiState.Missing -> GitHubAccountUiState.Stored
            else -> githubAccount
        }
    }

    private suspend fun refreshLatestWatchSnapshot() {
        val payload = runCatching { WatchSnapshotReader(applicationContext).latestPayload() }.getOrNull() ?: return
        runCatching { SnapshotProcessor(applicationContext).process(payload) }
    }

    private fun StoredWatchSnapshot.toUiState(): WatchDataUiState.Loaded {
        val snapshot = snapshot
        return WatchDataUiState.Loaded(
            heartRate = formatHeartRate(snapshot),
            steps = formatSteps(snapshot),
            battery = formatBattery(snapshot),
            isCharging = isCharging(snapshot),
            isOffWrist = isOffWrist(snapshot),
            updatedAt = formatReceivedAt(receivedAt),
        )
    }

    private fun formatHeartRate(snapshot: WatchSnapshot): String {
        if (!snapshot.hasHeartRate() || snapshot.heartRate.beatsPerMinute <= 0) return MISSING_VALUE
        return getString(R.string.watch_metric_heart_rate_value, snapshot.heartRate.beatsPerMinute)
    }

    private fun formatSteps(snapshot: WatchSnapshot): String {
        if (!snapshot.hasActivityTotals() || snapshot.activityTotals.steps < 0) return MISSING_VALUE
        return numberFormat.format(snapshot.activityTotals.steps)
    }

    private fun formatBattery(snapshot: WatchSnapshot): String {
        if (!snapshot.hasWatchState()) return MISSING_VALUE
        return getString(R.string.watch_metric_battery_value, snapshot.watchState.batteryPercent)
    }

    private fun isCharging(snapshot: WatchSnapshot): Boolean {
        return snapshot.hasWatchState() && snapshot.watchState.chargeState == ChargeState.CHARGE_STATE_CHARGING
    }

    private fun isOffWrist(snapshot: WatchSnapshot): Boolean {
        return snapshot.hasWatchState() && snapshot.watchState.wristState == WristState.WRIST_STATE_OFF_WRIST
    }

    private fun formatReceivedAt(receivedAt: Instant): String {
        val receivedDate = receivedAt.atZone(zoneId).toLocalDate()
        return if (receivedDate == LocalDate.now(zoneId)) {
            timeFormatter.format(receivedAt)
        } else {
            dateTimeFormatter.format(receivedAt)
        }
    }

    private fun GitHubProfile.toUiState(): GitHubAccountUiState.Connected {
        val statusText = status?.let { githubStatus ->
            listOfNotNull(githubStatus.emoji, githubStatus.message)
                .filter { it.isNotBlank() }
                .joinToString(separator = " ")
                .takeIf { it.isNotBlank() }
        } ?: getString(R.string.github_status_empty)
        return GitHubAccountUiState.Connected(
            login = "@$login",
            status = statusText,
        )
    }

    private fun String.looksLikeGitHubToken(): Boolean = GITHUB_TOKEN_PREFIXES.any(::startsWith)

    private fun ClipboardManager.clearToken() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            clearPrimaryClip()
        } else {
            setPrimaryClip(ClipData.newPlainText("", ""))
        }
    }

    private data class StatusState(
        val hasToken: Boolean,
        val githubSyncFailed: Boolean,
        val watchData: WatchDataUiState,
    )

    private companion object {
        const val REFRESH_INTERVAL_MS = 2_000L
        const val NOTIFICATION_PERMISSION_REQUEST_CODE = 1002
        const val MISSING_VALUE = "—"
        const val GITHUB_TOKEN_REQUEST_URI = "https://github.com/settings/tokens/new?description=Realtime%20Me%20Pixel%20Watch%20status%20publisher&scopes=user"
        val GITHUB_TOKEN_PREFIXES = listOf("github_pat_", "ghp_")
    }
}

private sealed interface WatchDataUiState {
    data object Empty : WatchDataUiState

    data class Loaded(
        val heartRate: String,
        val steps: String,
        val battery: String,
        val isCharging: Boolean,
        val isOffWrist: Boolean,
        val updatedAt: String,
    ) : WatchDataUiState
}

private sealed interface GitHubAccountUiState {
    val hasToken: Boolean

    data object Missing : GitHubAccountUiState {
        override val hasToken: Boolean = false
    }

    data object Stored : GitHubAccountUiState {
        override val hasToken: Boolean = true
    }

    data object Checking : GitHubAccountUiState {
        override val hasToken: Boolean = true
    }

    data class Connected(
        val login: String,
        val status: String,
    ) : GitHubAccountUiState {
        override val hasToken: Boolean = true
    }

    data object Error : GitHubAccountUiState {
        override val hasToken: Boolean = true
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun RealtimeMobileScreen(
    githubAccount: GitHubAccountUiState,
    githubSyncFailed: Boolean,
    watchData: WatchDataUiState,
    onRequestGitHubToken: () -> Unit,
    onImportCopiedToken: () -> Unit,
    onDisconnectGitHub: () -> Unit,
) {
    MaterialTheme {
        Scaffold(
            topBar = {
                TopAppBar(title = { Text(text = stringResource(R.string.app_name)) })
            },
        ) { contentPadding ->
            Surface(modifier = Modifier.fillMaxSize()) {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(contentPadding)
                        .verticalScroll(rememberScrollState())
                        .padding(24.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp),
                ) {
                    GitHubConnectionCard(
                        githubAccount = githubAccount,
                        githubSyncFailed = githubSyncFailed,
                        onRequestGitHubToken = onRequestGitHubToken,
                        onImportCopiedToken = onImportCopiedToken,
                        onDisconnectGitHub = onDisconnectGitHub,
                    )
                    CurrentWatchCard(watchData = watchData)
                }
            }
        }
    }
}

@OptIn(ExperimentalFoundationApi::class, ExperimentalLayoutApi::class)
@Composable
private fun GitHubConnectionCard(
    githubAccount: GitHubAccountUiState,
    githubSyncFailed: Boolean,
    onRequestGitHubToken: () -> Unit,
    onImportCopiedToken: () -> Unit,
    onDisconnectGitHub: () -> Unit,
) {
    var showDisconnectAction by remember(githubAccount.hasToken) { mutableStateOf(false) }
    ElevatedCard(
        modifier = Modifier
            .fillMaxWidth()
            .combinedClickable(
                onClick = {},
                onLongClickLabel = stringResource(R.string.disconnect_github),
                onLongClick = {
                    if (githubAccount.hasToken) showDisconnectAction = !showDisconnectAction
                },
            ),
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            IconSectionHeader(
                title = stringResource(R.string.github_card_title),
                icon = githubAccount.statusIcon(githubSyncFailed),
                contentDescription = githubAccount.statusLabel(githubSyncFailed),
                tint = githubAccount.statusIconTint(githubSyncFailed),
            )
            Text(
                text = githubAccount.description(),
                style = MaterialTheme.typography.bodyMedium,
            )
            FlowRow(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                if (!githubAccount.hasToken || githubAccount == GitHubAccountUiState.Error || githubSyncFailed) {
                    ActionIconButton(
                        icon = Icons.Outlined.OpenInBrowser,
                        contentDescription = stringResource(R.string.create_github_token),
                        onClick = onRequestGitHubToken,
                        style = ActionButtonStyle.Primary,
                    )
                    ActionIconButton(
                        icon = Icons.Outlined.ContentPaste,
                        contentDescription = stringResource(R.string.save_copied_token),
                        onClick = onImportCopiedToken,
                    )
                }
                if (githubAccount.hasToken && showDisconnectAction) {
                    ActionIconButton(
                        icon = Icons.Outlined.Delete,
                        contentDescription = stringResource(R.string.disconnect_github),
                        onClick = onDisconnectGitHub,
                    )
                }
            }
        }
    }
}

private fun GitHubAccountUiState.statusIcon(githubSyncFailed: Boolean): ImageVector = when {
    githubSyncFailed && hasToken -> Icons.Outlined.ErrorOutline
    this == GitHubAccountUiState.Missing -> Icons.Outlined.LinkOff
    this == GitHubAccountUiState.Stored -> Icons.Outlined.HourglassEmpty
    this == GitHubAccountUiState.Checking -> Icons.Outlined.HourglassEmpty
    this is GitHubAccountUiState.Connected -> Icons.Outlined.CheckCircle
    else -> Icons.Outlined.ErrorOutline
}

@Composable
private fun GitHubAccountUiState.statusIconTint(githubSyncFailed: Boolean): Color {
    return if ((githubSyncFailed && hasToken) || this == GitHubAccountUiState.Error) {
        MaterialTheme.colorScheme.error
    } else {
        MaterialTheme.colorScheme.primary
    }
}

@Composable
private fun GitHubAccountUiState.statusLabel(githubSyncFailed: Boolean): String = when {
    githubSyncFailed && hasToken -> stringResource(R.string.github_sync_failed_chip)
    this == GitHubAccountUiState.Missing -> stringResource(R.string.github_not_connected_chip)
    this == GitHubAccountUiState.Stored -> stringResource(R.string.github_stored_chip)
    this == GitHubAccountUiState.Checking -> stringResource(R.string.github_checking_chip)
    this is GitHubAccountUiState.Connected -> stringResource(R.string.github_connected_chip)
    else -> stringResource(R.string.github_error_chip)
}

@Composable
private fun GitHubAccountUiState.description(): String = when (this) {
    GitHubAccountUiState.Missing -> stringResource(R.string.github_token_missing_description)
    GitHubAccountUiState.Stored -> stringResource(R.string.github_token_stored_description)
    GitHubAccountUiState.Checking -> stringResource(R.string.github_token_checking_description)
    is GitHubAccountUiState.Connected -> stringResource(R.string.github_connected_description, login)
    GitHubAccountUiState.Error -> stringResource(R.string.github_token_error_description)
}

@Composable
private fun CurrentWatchCard(watchData: WatchDataUiState) {
    ElevatedCard(modifier = Modifier.fillMaxWidth()) {
        Column(
            modifier = Modifier.padding(20.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            SectionHeader(
                title = stringResource(R.string.current_watch_title),
                status = watchData.statusLabel(),
                statusIcon = watchData.statusIcon(),
            )
            when (watchData) {
                WatchDataUiState.Empty -> EmptyWatchData()
                is WatchDataUiState.Loaded -> WatchMetrics(watchData)
            }
        }
    }
}


@Composable
private fun WatchDataUiState.statusLabel(): String = when (this) {
    WatchDataUiState.Empty -> stringResource(R.string.watch_waiting_chip)
    is WatchDataUiState.Loaded -> updatedAt
}

private fun WatchDataUiState.statusIcon(): ImageVector = when (this) {
    WatchDataUiState.Empty -> Icons.Outlined.HourglassEmpty
    is WatchDataUiState.Loaded -> Icons.Outlined.AccessTime
}

@Composable
private fun EmptyWatchData() {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(
            text = stringResource(R.string.watch_data_empty_title),
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.SemiBold,
        )
        Text(
            text = stringResource(R.string.watch_data_empty_description),
            style = MaterialTheme.typography.bodyMedium,
        )
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun WatchMetrics(watchData: WatchDataUiState.Loaded) {
    FlowRow(
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        if (!watchData.isOffWrist) {
            MetricBadge(
                icon = Icons.Outlined.Favorite,
                value = watchData.heartRate,
                contentDescription = stringResource(R.string.watch_metric_heart_rate),
            )
        }
        MetricBadge(
            icon = Icons.AutoMirrored.Outlined.DirectionsWalk,
            value = watchData.steps,
            contentDescription = stringResource(R.string.watch_metric_steps),
        )
        MetricBadge(
            icon = Icons.Outlined.BatteryFull,
            value = watchData.battery,
            contentDescription = stringResource(R.string.watch_metric_battery),
        )
        if (watchData.isCharging) {
            IconBadge(
                icon = Icons.Outlined.BatteryChargingFull,
                contentDescription = stringResource(R.string.watch_metric_charge_state),
            )
        }
        if (watchData.isOffWrist) {
            IconBadge(
                icon = Icons.Outlined.WatchOff,
                contentDescription = stringResource(R.string.watch_metric_wrist_state),
            )
        }
    }
}

@Composable
private fun MetricBadge(
    icon: ImageVector,
    value: String,
    contentDescription: String,
) {
    AssistChip(
        onClick = {},
        leadingIcon = {
            Icon(
                imageVector = icon,
                contentDescription = contentDescription,
                modifier = Modifier.size(18.dp),
            )
        },
        label = {
            Text(
                text = value,
                style = MaterialTheme.typography.labelLarge,
                fontWeight = FontWeight.SemiBold,
            )
        },
    )
}

@Composable
private fun IconBadge(
    icon: ImageVector,
    contentDescription: String,
) {
    Surface(
        shape = MaterialTheme.shapes.small,
        color = MaterialTheme.colorScheme.secondaryContainer,
        contentColor = MaterialTheme.colorScheme.onSecondaryContainer,
    ) {
        Icon(
            imageVector = icon,
            contentDescription = contentDescription,
            modifier = Modifier
                .padding(8.dp)
                .size(18.dp),
        )
    }
}

@Composable
private fun ActionIconButton(
    icon: ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    style: ActionButtonStyle = ActionButtonStyle.Tonal,
) {
    val modifier = Modifier.size(56.dp)
    val content: @Composable () -> Unit = {
        Icon(
            imageVector = icon,
            contentDescription = contentDescription,
            modifier = Modifier.size(26.dp),
        )
    }
    when (style) {
        ActionButtonStyle.Primary -> FilledIconButton(
            onClick = onClick,
            modifier = modifier,
            content = content,
        )
        ActionButtonStyle.Tonal -> FilledTonalIconButton(
            onClick = onClick,
            modifier = modifier,
            content = content,
        )
    }
}

private enum class ActionButtonStyle {
    Primary,
    Tonal,
}

@Composable
private fun IconSectionHeader(
    title: String,
    icon: ImageVector,
    contentDescription: String,
    tint: Color = MaterialTheme.colorScheme.primary,
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
        )
        Icon(
            imageVector = icon,
            contentDescription = contentDescription,
            modifier = Modifier.size(24.dp),
            tint = tint,
        )
    }
}

@Composable
private fun SectionHeader(
    title: String,
    status: String,
    statusIcon: ImageVector,
) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
        )
        AssistChip(
            onClick = {},
            leadingIcon = {
                Icon(
                    imageVector = statusIcon,
                    contentDescription = null,
                    modifier = Modifier.size(18.dp),
                )
            },
            label = { Text(text = status) },
        )
    }
}
