package me.realtime.mobile

import android.Manifest
import android.content.ClipData
import android.content.ClipboardManager
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.viewModels
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.Image
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
import androidx.compose.material.icons.outlined.Favorite
import androidx.compose.material.icons.outlined.HourglassEmpty
import androidx.compose.material.icons.outlined.LinkOff
import androidx.compose.material.icons.outlined.WatchOff
import androidx.compose.material3.AssistChip
import androidx.compose.material3.ElevatedCard
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilledTonalIconButton
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp

class MainActivity : ComponentActivity() {
    private val viewModel: MainViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        requestRuntimePermissionsIfNeeded()
        enableEdgeToEdge()
        setContent {
            val uiState by viewModel.uiState.collectAsState()
            LaunchedEffect(Unit) {
                viewModel.uiEvents.collect { event -> handleEvent(event) }
            }
            RealtimeMobileScreen(
                statusGatewayConnected = uiState.statusGatewayConnected,
                watchData = uiState.watchData,
                onImportStatusGatewayToken = ::importStatusGatewayToken,
                onDisconnectStatusGateway = viewModel::disconnect,
            )
        }
    }

    override fun onStart() {
        super.onStart()
        viewModel.onStart()
    }

    override fun onResume() {
        super.onResume()
        viewModel.refresh()
    }

    private fun importStatusGatewayToken() {
        val clipboardToken = getSystemService(ClipboardManager::class.java)
            ?.primaryClip
            ?.takeIf { it.itemCount > 0 }
            ?.getItemAt(0)
            ?.coerceToText(this)
            ?.toString()
        viewModel.importToken(clipboardToken)
    }

    private fun handleEvent(event: MainEvent) {
        when (event) {
            MainEvent.TokenSaved -> {
                getSystemService(ClipboardManager::class.java)?.clearToken()
                toast(R.string.status_gateway_token_saved)
            }
            MainEvent.TokenMissing -> toast(R.string.status_gateway_token_clipboard_missing)
            MainEvent.Disconnected -> toast(R.string.status_gateway_disconnected)
        }
    }

    private fun toast(messageId: Int) {
        Toast.makeText(this, messageId, Toast.LENGTH_SHORT).show()
    }

    private fun requestRuntimePermissionsIfNeeded() {
        val permissions = buildList {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU &&
                checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED
            ) {
                add(Manifest.permission.POST_NOTIFICATIONS)
            }
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S &&
                checkSelfPermission(Manifest.permission.BLUETOOTH_CONNECT) != PackageManager.PERMISSION_GRANTED
            ) {
                add(Manifest.permission.BLUETOOTH_CONNECT)
            }
        }
        if (permissions.isNotEmpty()) {
            requestPermissions(permissions.toTypedArray(), RUNTIME_PERMISSION_REQUEST_CODE)
        }
    }

    private fun ClipboardManager.clearToken() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            clearPrimaryClip()
        } else {
            setPrimaryClip(ClipData.newPlainText("", ""))
        }
    }

    private companion object {
        const val RUNTIME_PERMISSION_REQUEST_CODE = 1002
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun RealtimeMobileScreen(
    statusGatewayConnected: Boolean,
    watchData: WatchDataUiState,
    onImportStatusGatewayToken: () -> Unit,
    onDisconnectStatusGateway: () -> Unit,
) {
    MaterialTheme {
        Scaffold(
            topBar = {
                TopAppBar(
                    title = {
                        Row(
                            horizontalArrangement = Arrangement.spacedBy(10.dp),
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Image(
                                painter = painterResource(R.drawable.ic_launcher),
                                contentDescription = null,
                                modifier = Modifier.size(32.dp),
                            )
                            Text(text = stringResource(R.string.app_name))
                        }
                    },
                )
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
                    StatusGatewayCard(
                        connected = statusGatewayConnected,
                        onImportToken = onImportStatusGatewayToken,
                        onDisconnect = onDisconnectStatusGateway,
                    )
                    CurrentWatchCard(watchData = watchData)
                }
            }
        }
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun StatusGatewayCard(
    connected: Boolean,
    onImportToken: () -> Unit,
    onDisconnect: () -> Unit,
) {
    var showDisconnectAction by remember(connected) { mutableStateOf(false) }
    ElevatedCard(
        modifier = Modifier
            .fillMaxWidth()
            .combinedClickable(
                onClick = {},
                onLongClickLabel = stringResource(R.string.status_gateway_disconnect),
                onLongClick = {
                    if (connected) showDisconnectAction = !showDisconnectAction
                },
            ),
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            IconSectionHeader(
                title = stringResource(R.string.status_gateway_title),
                icon = if (connected) Icons.Outlined.CheckCircle else Icons.Outlined.LinkOff,
                contentDescription = if (connected) {
                    stringResource(R.string.status_gateway_connected)
                } else {
                    stringResource(R.string.status_gateway_missing)
                },
            )
            Text(
                text = if (connected) {
                    stringResource(R.string.status_gateway_connected_description)
                } else {
                    stringResource(R.string.status_gateway_missing_description)
                },
                style = MaterialTheme.typography.bodyMedium,
            )
            FlowRow(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                if (!connected) {
                    ActionIconButton(
                        icon = Icons.Outlined.ContentPaste,
                        contentDescription = stringResource(R.string.status_gateway_import_token),
                        onClick = onImportToken,
                    )
                }
                if (connected && showDisconnectAction) {
                    ActionIconButton(
                        icon = Icons.Outlined.Delete,
                        contentDescription = stringResource(R.string.status_gateway_disconnect),
                        onClick = onDisconnect,
                    )
                }
            }
        }
    }
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
) {
    FilledTonalIconButton(
        onClick = onClick,
        modifier = Modifier.size(56.dp),
    ) {
        Icon(
            imageVector = icon,
            contentDescription = contentDescription,
            modifier = Modifier.size(26.dp),
        )
    }
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
