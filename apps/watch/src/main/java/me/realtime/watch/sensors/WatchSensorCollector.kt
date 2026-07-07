package me.realtime.watch.sensors

import android.Manifest
import android.content.Context
import android.content.pm.PackageManager
import android.hardware.Sensor
import android.hardware.SensorEvent
import android.hardware.SensorEventListener
import android.hardware.SensorManager
import android.os.Build
import android.os.SystemClock
import me.realtime.protocol.v1.ReportWatchSnapshotRequest
import me.realtime.protocol.v1.WatchSnapshot
import me.realtime.protocol.v1.WristState
import me.realtime.watch.state.WatchSnapshotRepository
import me.realtime.watch.wear.SnapshotPublisher
import java.time.Instant
import java.time.ZoneId
import java.util.concurrent.atomic.AtomicBoolean
import kotlin.math.roundToInt

object WatchSensorCollector {
    private const val HEART_RATE_EXTENDED_SENSOR = "com.google.sensor.heart_rate_extended"
    private const val BODY_SENSORS_BACKGROUND_PERMISSION = "android.permission.BODY_SENSORS_BACKGROUND"
    private const val READ_HEART_RATE_PERMISSION = "android.permission.health.READ_HEART_RATE"
    private const val READ_HEALTH_DATA_IN_BACKGROUND_PERMISSION = "android.permission.health.READ_HEALTH_DATA_IN_BACKGROUND"
    private const val STEP_BASELINE_PREFS = "step_counter_baseline"
    private const val BASELINE_DATE_KEY = "baseline_date"
    private const val BASELINE_VALUE_KEY = "baseline_value"
    private const val ON_BODY_VALUE = 1.0f
    private val started = AtomicBoolean(false)
    private var registeredSensorManager: SensorManager? = null
    private var registeredListener: SensorEventListener? = null

    fun requiredPermissions(): List<String> = buildList {
        if (Build.VERSION.SDK_INT >= 36) {
            add(READ_HEART_RATE_PERMISSION)
        } else {
            add(Manifest.permission.BODY_SENSORS)
        }
        add(Manifest.permission.ACTIVITY_RECOGNITION)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            add(Manifest.permission.POST_NOTIFICATIONS)
        }
        if (Build.VERSION.SDK_INT >= 36) {
            add(READ_HEALTH_DATA_IN_BACKGROUND_PERMISSION)
        } else if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            add(BODY_SENSORS_BACKGROUND_PERMISSION)
        }
    }

    fun hasRequiredPermissions(context: Context): Boolean = requiredPermissions().all { permission ->
        context.checkSelfPermission(permission) == PackageManager.PERMISSION_GRANTED
    }

    fun start(context: Context): SensorStartResult {
        if (!hasRequiredPermissions(context)) {
            return SensorStartResult.Failure("Sensor permissions are missing")
        }
        if (!started.compareAndSet(false, true)) {
            return SensorStartResult.Success
        }

        val appContext = context.applicationContext
        val sensorManager = appContext.getSystemService(SensorManager::class.java)
        if (sensorManager == null) {
            started.set(false)
            return SensorStartResult.Failure("Sensor manager is unavailable")
        }
        val listener = SnapshotSensorListener(appContext)
        val registrations = sensorManager.registerSnapshotSensors(listener)
        if (registrations.isEmpty()) {
            started.set(false)
            return SensorStartResult.Failure("No supported watch sensors are available")
        }
        registeredSensorManager = sensorManager
        registeredListener = listener

        val startupSnapshot = WatchSnapshotRepository(appContext)
            .refreshDeviceState(includeStepTotal = SensorKind.StepCounter in registrations)
        publish(startupSnapshot, appContext)
        return SensorStartResult.Success
    }

    // Unregisters the sensor listeners and clears the started latch so the
    // collection service releases sensors on destroy and can re-register on a
    // later restart. Called from the service's onDestroy.
    fun stop() {
        registeredListener?.let { listener ->
            registeredSensorManager?.unregisterListener(listener)
        }
        registeredSensorManager = null
        registeredListener = null
        started.set(false)
    }

    fun refreshCurrentState(context: Context): SensorStartResult {
        if (!hasRequiredPermissions(context)) {
            return SensorStartResult.Failure("Sensor permissions are missing")
        }

        val appContext = context.applicationContext
        val snapshot = WatchSnapshotRepository(appContext).refreshDeviceState()
        publish(snapshot, appContext)
        return SensorStartResult.Success
    }

    private fun SensorManager.registerSnapshotSensors(listener: SensorEventListener): Set<SensorKind> {
        val registrations = mutableSetOf<SensorKind>()
        registerSensor(registrations, SensorKind.HeartRate, listener, defaultSensor(Sensor.TYPE_HEART_RATE))
        registerSensor(registrations, SensorKind.HeartRate, listener, sensorByStringType(HEART_RATE_EXTENDED_SENSOR))
        registerSensor(registrations, SensorKind.StepCounter, listener, defaultSensor(Sensor.TYPE_STEP_COUNTER))
        registerSensor(registrations, SensorKind.WristState, listener, defaultSensor(Sensor.TYPE_LOW_LATENCY_OFFBODY_DETECT))
        return registrations
    }

    private fun SensorManager.registerSensor(
        registrations: MutableSet<SensorKind>,
        sensorKind: SensorKind,
        listener: SensorEventListener,
        sensor: Sensor?,
    ): Boolean {
        if (sensor == null) return false
        val registered = registerListener(listener, sensor, SensorManager.SENSOR_DELAY_NORMAL)
        if (registered) registrations += sensorKind
        return registered
    }

    private fun SensorManager.defaultSensor(sensorType: Int): Sensor? {
        return getDefaultSensor(sensorType, true) ?: getDefaultSensor(sensorType)
    }

    private fun SensorManager.sensorByStringType(stringType: String): Sensor? {
        return getSensorList(Sensor.TYPE_ALL).firstOrNull { sensor ->
            sensor.stringType == stringType && sensor.isWakeUpSensor
        } ?: getSensorList(Sensor.TYPE_ALL).firstOrNull { sensor ->
            sensor.stringType == stringType
        }
    }

    private fun publish(snapshot: WatchSnapshot, context: Context) {
        val payload = ReportWatchSnapshotRequest.newBuilder()
            .setWatchSnapshot(snapshot)
            .build()
        SnapshotPublisher(context).publishIfAllowed(payload)
    }

    private class SnapshotSensorListener(private val context: Context) : SensorEventListener {
        private val repository = WatchSnapshotRepository(context)

        override fun onSensorChanged(event: SensorEvent) {
            val snapshot = when (event.sensor.type) {
                Sensor.TYPE_HEART_RATE -> heartRateSnapshot(event)
                Sensor.TYPE_STEP_COUNTER -> stepSnapshot(event)
                Sensor.TYPE_LOW_LATENCY_OFFBODY_DETECT -> wristSnapshot(event)
                else -> when (event.sensor.stringType) {
                    HEART_RATE_EXTENDED_SENSOR -> heartRateSnapshot(event)
                    else -> null
                }
            } ?: return
            publish(snapshot, context)
        }

        override fun onAccuracyChanged(sensor: Sensor?, accuracy: Int) = Unit

        private fun heartRateSnapshot(event: SensorEvent): WatchSnapshot? {
            val beatsPerMinute = event.values.firstOrNull()?.roundToInt() ?: return null
            if (beatsPerMinute <= 0) return null
            return repository.updateHeartRate(beatsPerMinute, event.sampleInstant())
        }

        private fun stepSnapshot(event: SensorEvent): WatchSnapshot? {
            val rawSteps = event.values.firstOrNull()?.roundToInt() ?: return null
            if (rawSteps < 0) return null
            val sampleTime = event.sampleInstant()
            return repository.updateSteps(todaySteps(rawSteps, sampleTime), sampleTime)
        }

        private fun wristSnapshot(event: SensorEvent): WatchSnapshot {
            val wristState = if ((event.values.firstOrNull() ?: ON_BODY_VALUE) == ON_BODY_VALUE) {
                WristState.WRIST_STATE_ON_WRIST
            } else {
                WristState.WRIST_STATE_OFF_WRIST
            }
            return repository.updateWristState(wristState)
        }

        private fun todaySteps(rawSteps: Int, sampleTime: Instant): Int {
            val today = sampleTime.atZone(ZoneId.systemDefault()).toLocalDate().toString()
            val preferences = context.getSharedPreferences(STEP_BASELINE_PREFS, Context.MODE_PRIVATE)
            val baselineDate = preferences.getString(BASELINE_DATE_KEY, null)
            val baselineValue = preferences.getInt(BASELINE_VALUE_KEY, rawSteps)
            if (baselineDate != today || rawSteps < baselineValue) {
                preferences.edit()
                    .putString(BASELINE_DATE_KEY, today)
                    .putInt(BASELINE_VALUE_KEY, rawSteps)
                    .apply()
                return 0
            }
            return rawSteps - baselineValue
        }

        private fun SensorEvent.sampleInstant(): Instant {
            val bootEpochMillis = System.currentTimeMillis() - SystemClock.elapsedRealtime()
            return Instant.ofEpochMilli(bootEpochMillis + timestamp / NANOS_PER_MILLI)
        }
    }

    private const val NANOS_PER_MILLI = 1_000_000L

    private enum class SensorKind {
        HeartRate,
        StepCounter,
        WristState,
    }
}

sealed class SensorStartResult {
    data object Success : SensorStartResult()
    data class Failure(val message: String) : SensorStartResult()
}
