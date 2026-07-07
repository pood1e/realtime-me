package me.realtime.mobile.status

import android.Manifest
import android.bluetooth.BluetoothManager
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.content.pm.PackageManager
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
import android.os.BatteryManager
import android.os.Build
import android.provider.Settings
import me.realtime.mobile.state.StoredWatchSnapshot
import me.realtime.protocol.v1.ChargeState
import me.realtime.protocol.v1.WristState
import org.json.JSONObject
import java.time.Instant
import kotlin.math.roundToInt

class StatusGatewayPayloadBuilder(private val context: Context) {
    private val accessoryReader = BluetoothAudioAccessoryReader(context)

    fun build(storedWatchSnapshot: StoredWatchSnapshot?): JSONObject {
        return JSONObject()
            .put("device_id", StatusDeviceIdentity(context).id())
            .put("device_name", deviceName())
            .put("device_model", deviceModel())
            .put("updated_at", Instant.now().toString())
            .put("phone", phoneState())
            .also { payload ->
                storedWatchSnapshot?.let { payload.put("watch", watchState(it)) }
            }
    }

    private fun phoneState(): JSONObject {
        val batteryIntent = context.registerReceiver(null, IntentFilter(Intent.ACTION_BATTERY_CHANGED))
        return JSONObject()
            .put("battery_percent", batteryIntent?.batteryPercent() ?: 0)
            .put("charge_state", if (batteryIntent?.charging() == true) "charging" else "not_charging")
            .put("network", networkState())
            .also { phone ->
                accessoryReader.read().takeIf { it.length() > 0 }?.let { phone.put("accessories", it) }
            }
    }

    private fun watchState(storedWatchSnapshot: StoredWatchSnapshot): JSONObject {
        val snapshot = storedWatchSnapshot.snapshot
        val state = JSONObject()
            .put("steps", snapshot.activityTotals.steps)
            .put("battery_percent", snapshot.watchState.batteryPercent)
            .put("charge_state", snapshot.watchState.chargeState.toWireValue())
            .put("wrist_state", snapshot.watchState.wristState.toWireValue())
        if (snapshot.hasDeviceInfo()) {
            state.put("device_name", snapshot.deviceInfo.displayName)
            state.put("device_model", snapshot.deviceInfo.model)
        }
        if (snapshot.watchState.wristState != WristState.WRIST_STATE_OFF_WRIST && snapshot.heartRate.beatsPerMinute > 0) {
            state.put("heart_rate", snapshot.heartRate.beatsPerMinute)
        }
        return state
    }

    private fun deviceName(): String = firstNonBlank(
        bluetoothName(),
        globalDeviceName(),
        Build.MODEL,
    )

    private fun deviceModel(): String = listOf(Build.MANUFACTURER, Build.MODEL).joinToString(" ").trim()

    private fun bluetoothName(): String = runCatching {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S &&
            context.checkSelfPermission(Manifest.permission.BLUETOOTH_CONNECT) != PackageManager.PERMISSION_GRANTED
        ) {
            return@runCatching ""
        }
        context.getSystemService(BluetoothManager::class.java)?.adapter?.name.orEmpty()
    }.getOrDefault("")

    private fun globalDeviceName(): String = runCatching {
        Settings.Global.getString(context.contentResolver, Settings.Global.DEVICE_NAME)
    }.getOrDefault("")

    private fun firstNonBlank(vararg values: String?): String {
        return values.firstOrNull { !it.isNullOrBlank() && !it.equals("null", ignoreCase = true) }?.trim().orEmpty()
    }

    private fun networkState(): String {
        val manager = context.getSystemService(ConnectivityManager::class.java) ?: return "unknown"
        val network = manager.activeNetwork ?: return "offline"
        val capabilities = manager.getNetworkCapabilities(network) ?: return "unknown"
        return when {
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) -> "wifi"
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR) -> "cellular"
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_VPN) -> "vpn"
            capabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET) -> "online"
            else -> "offline"
        }
    }

    private fun Intent.batteryPercent(): Int {
        val level = getIntExtra(BatteryManager.EXTRA_LEVEL, -1)
        val scale = getIntExtra(BatteryManager.EXTRA_SCALE, -1)
        if (level < 0 || scale <= 0) return 0
        return ((level.toFloat() / scale.toFloat()) * 100).roundToInt().coerceIn(0, 100)
    }

    private fun Intent.charging(): Boolean {
        val status = getIntExtra(BatteryManager.EXTRA_STATUS, BatteryManager.BATTERY_STATUS_UNKNOWN)
        val plugged = getIntExtra(BatteryManager.EXTRA_PLUGGED, 0)
        return status == BatteryManager.BATTERY_STATUS_CHARGING ||
            status == BatteryManager.BATTERY_STATUS_FULL ||
            plugged != 0
    }

    private fun ChargeState.toWireValue(): String = when (this) {
        ChargeState.CHARGE_STATE_CHARGING -> "charging"
        ChargeState.CHARGE_STATE_NOT_CHARGING -> "not_charging"
        else -> "unknown"
    }

    private fun WristState.toWireValue(): String = when (this) {
        WristState.WRIST_STATE_ON_WRIST -> "on_wrist"
        WristState.WRIST_STATE_OFF_WRIST -> "off_wrist"
        else -> "unknown"
    }
}
