package me.realtime.mobile

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch
import me.realtime.mobile.state.StatusRepository
import me.realtime.mobile.state.StoredWatchSnapshot
import me.realtime.protocol.v1.ChargeState
import me.realtime.protocol.v1.WatchSnapshot
import java.text.NumberFormat
import java.time.Instant
import java.time.LocalDate
import java.time.ZoneId
import java.time.format.DateTimeFormatter

/**
 * Holds the home screen's state and business logic. The activity only observes
 * [uiState], forwards user intents, and reacts to one-off [uiEvents]; there is
 * no polling loop — the watch data flows in reactively from the repository.
 */
class MainViewModel(application: Application) : AndroidViewModel(application) {
    private val repository = StatusRepository(application)
    private val numberFormat = NumberFormat.getIntegerInstance()
    private val zoneId = ZoneId.systemDefault()
    private val timeFormatter = DateTimeFormatter.ofPattern("HH:mm:ss").withZone(zoneId)
    private val dateTimeFormatter = DateTimeFormatter.ofPattern("MMM d, HH:mm:ss").withZone(zoneId)

    private val connected = MutableStateFlow(repository.hasToken())
    private val events = Channel<MainEvent>(Channel.BUFFERED)
    val uiEvents = events.receiveAsFlow()

    val uiState: StateFlow<MainUiState> =
        combine(connected, repository.watchSnapshots()) { isConnected, stored ->
            MainUiState(
                statusGatewayConnected = isConnected,
                watchData = stored?.toUiState() ?: WatchDataUiState.Empty,
            )
        }.stateIn(
            scope = viewModelScope,
            started = SharingStarted.WhileSubscribed(STOP_TIMEOUT_MS),
            initialValue = MainUiState(connected.value, WatchDataUiState.Empty),
        )

    fun onStart() {
        viewModelScope.launch { repository.ensureSyncActive() }
        refresh()
    }

    fun refresh() {
        viewModelScope.launch { repository.refreshFromWatch() }
    }

    fun importToken(rawToken: String?) {
        val token = rawToken?.trim().orEmpty()
        if (token.length < MIN_TOKEN_LENGTH) {
            events.trySend(MainEvent.TokenMissing)
            return
        }
        viewModelScope.launch {
            if (!repository.saveToken(token)) {
                events.trySend(MainEvent.TokenSaveFailed)
                return@launch
            }
            connected.value = true
            events.trySend(MainEvent.TokenSaved)
            repository.refreshFromWatch()
        }
    }

    fun disconnect() {
        viewModelScope.launch {
            repository.clearToken()
            connected.value = false
            events.trySend(MainEvent.Disconnected)
        }
    }

    private fun StoredWatchSnapshot.toUiState(): WatchDataUiState.Loaded = WatchDataUiState.Loaded(
        heartRate = formatHeartRate(snapshot),
        steps = formatSteps(snapshot),
        battery = formatBattery(snapshot),
        isCharging = isCharging(snapshot),
        updatedAt = formatReceivedAt(receivedAt),
    )

    private fun formatHeartRate(snapshot: WatchSnapshot): String {
        if (!snapshot.hasHeartRate() || snapshot.heartRate.beatsPerMinute <= 0) return MISSING_VALUE
        return getApplication<Application>().getString(R.string.watch_metric_heart_rate_value, snapshot.heartRate.beatsPerMinute)
    }

    private fun formatSteps(snapshot: WatchSnapshot): String {
        if (!snapshot.hasActivityTotals() || snapshot.activityTotals.steps < 0) return MISSING_VALUE
        return numberFormat.format(snapshot.activityTotals.steps)
    }

    private fun formatBattery(snapshot: WatchSnapshot): String {
        if (!snapshot.hasWatchState()) return MISSING_VALUE
        return getApplication<Application>().getString(R.string.watch_metric_battery_value, snapshot.watchState.batteryPercent)
    }

    private fun isCharging(snapshot: WatchSnapshot): Boolean =
        snapshot.hasWatchState() && snapshot.watchState.chargeState == ChargeState.CHARGE_STATE_CHARGING

    private fun formatReceivedAt(receivedAt: Instant): String {
        val receivedDate = receivedAt.atZone(zoneId).toLocalDate()
        return if (receivedDate == LocalDate.now(zoneId)) {
            timeFormatter.format(receivedAt)
        } else {
            dateTimeFormatter.format(receivedAt)
        }
    }

    private companion object {
        const val MIN_TOKEN_LENGTH = 16
        const val STOP_TIMEOUT_MS = 5_000L
        const val MISSING_VALUE = "—"
    }
}

data class MainUiState(
    val statusGatewayConnected: Boolean,
    val watchData: WatchDataUiState,
)

sealed interface WatchDataUiState {
    data object Empty : WatchDataUiState

    data class Loaded(
        val heartRate: String,
        val steps: String,
        val battery: String,
        val isCharging: Boolean,
        val updatedAt: String,
    ) : WatchDataUiState
}

sealed interface MainEvent {
    data object TokenSaved : MainEvent
    data object TokenMissing : MainEvent
    data object TokenSaveFailed : MainEvent
    data object Disconnected : MainEvent
}
