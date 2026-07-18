package me.realtime.watch.state

import android.os.Build
import me.realtime.status.v1.DeviceInfo

object DeviceInfoReader {
    fun read(): DeviceInfo = DeviceInfo.newBuilder()
        .setDisplayName(Build.MODEL)
        .setModel(listOf(Build.MANUFACTURER, Build.MODEL).joinToString(" ").trim())
        .build()
}
