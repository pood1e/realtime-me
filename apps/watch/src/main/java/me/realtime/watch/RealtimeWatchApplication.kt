package me.realtime.watch

import android.app.Application
import me.realtime.watch.service.SensorCollectionService
import me.realtime.watch.sensors.WatchSensorCollector

class RealtimeWatchApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        if (WatchSensorCollector.hasRequiredPermissions(this)) {
            runCatching { SensorCollectionService.start(this) }
        }
    }
}
