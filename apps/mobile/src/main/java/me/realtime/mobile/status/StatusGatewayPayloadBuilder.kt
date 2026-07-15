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
import me.realtime.mobile.nintendo.NintendoSwitchPresenceReader
import me.realtime.protocol.v1.ChargeState
import me.realtime.protocol.v1.DeviceKind
import me.realtime.protocol.v1.EnrollDeviceRequest
import me.realtime.protocol.v1.NetworkState
import me.realtime.protocol.v1.PhoneState
import me.realtime.protocol.v1.ReportMobileStatusRequest
import kotlin.math.roundToInt

class StatusGatewayPayloadBuilder(private val context: Context) {
    private val accessoryReader = BluetoothAudioAccessoryReader(context)
    private val switchPresenceReader = NintendoSwitchPresenceReader()

    /** The one-time enrollment request describing this phone to the gateway. */
    fun enrollRequest(): EnrollDeviceRequest = EnrollDeviceRequest.newBuilder()
        .setKind(DeviceKind.DEVICE_KIND_PHONE)
        .setDisplayName(deviceName())
        .setModel(deviceModel())
        .build()

    fun build(deviceUid: String, storedWatchSnapshot: StoredWatchSnapshot?): ReportMobileStatusRequest {
        val builder = ReportMobileStatusRequest.newBuilder()
            .setDeviceUid(deviceUid)
            .setDisplayName(deviceName())
            .setModel(deviceModel())
            .setPhone(phoneState())
        // Forward the watch snapshot verbatim; the Data Layer contract and the
        // gateway's ingest contract share the same WatchSnapshot message.
        storedWatchSnapshot?.let { builder.setWatch(it.snapshot) }
        switchPresenceReader.read()?.let(builder::setSwitchPresence)
        return builder.build()
    }

    private fun phoneState(): PhoneState {
        val batteryIntent = context.registerReceiver(null, IntentFilter(Intent.ACTION_BATTERY_CHANGED))
        val builder = PhoneState.newBuilder()
            .setNetwork(networkState())
            .addAllAccessories(accessoryReader.read())
        if (batteryIntent != null) {
            builder.setBatteryPercent(batteryIntent.batteryPercent())
            builder.setChargeState(
                if (batteryIntent.charging()) {
                    ChargeState.CHARGE_STATE_CHARGING
                } else {
                    ChargeState.CHARGE_STATE_NOT_CHARGING
                },
            )
        }
        return builder.build()
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

    private fun networkState(): NetworkState {
        val manager = context.getSystemService(ConnectivityManager::class.java)
            ?: return NetworkState.NETWORK_STATE_UNSPECIFIED
        val network = manager.activeNetwork ?: return NetworkState.NETWORK_STATE_OFFLINE
        val capabilities = manager.getNetworkCapabilities(network)
            ?: return NetworkState.NETWORK_STATE_UNSPECIFIED
        return when {
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) -> NetworkState.NETWORK_STATE_WIFI
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR) -> NetworkState.NETWORK_STATE_CELLULAR
            capabilities.hasTransport(NetworkCapabilities.TRANSPORT_VPN) -> NetworkState.NETWORK_STATE_VPN
            capabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET) -> NetworkState.NETWORK_STATE_ONLINE
            else -> NetworkState.NETWORK_STATE_OFFLINE
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
}
