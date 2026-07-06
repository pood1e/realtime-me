package me.realtime.watch.state

import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.os.BatteryManager
import me.realtime.protocol.toProtoTimestamp
import me.realtime.protocol.v1.ChargeState
import me.realtime.protocol.v1.WatchState
import me.realtime.protocol.v1.WristState
import java.time.Instant
import kotlin.math.roundToInt

object WatchStateReader {
    fun read(
        context: Context,
        wristState: WristState = WristState.WRIST_STATE_UNSPECIFIED,
    ): WatchState {
        val batteryIntent = context.registerReceiver(null, IntentFilter(Intent.ACTION_BATTERY_CHANGED))
        val batteryPercent = batteryIntent?.batteryPercent() ?: 0
        val chargeState = if (batteryIntent?.charging() == true) {
            ChargeState.CHARGE_STATE_CHARGING
        } else {
            ChargeState.CHARGE_STATE_NOT_CHARGING
        }

        return WatchState.newBuilder()
            .setBatteryPercent(batteryPercent)
            .setChargeState(chargeState)
            .setWristState(wristState)
            .setSampleTime(Instant.now().toProtoTimestamp())
            .build()
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
