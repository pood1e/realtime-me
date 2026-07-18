package me.realtime.watch.health

import android.Manifest
import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import android.util.Log
import androidx.concurrent.futures.await
import androidx.health.services.client.HealthServices
import androidx.health.services.client.data.DataType
import androidx.health.services.client.data.PassiveListenerConfig
import androidx.work.BackoffPolicy
import androidx.work.CoroutineWorker
import androidx.work.ExistingWorkPolicy
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.WorkerParameters
import me.realtime.status.v1.ReportWatchSnapshotRequest
import me.realtime.watch.state.WatchSnapshotRepository
import me.realtime.watch.wear.SnapshotPublisher
import java.time.Duration

object PassiveHealthRegistration {
    private const val READ_HEART_RATE_PERMISSION = "android.permission.health.READ_HEART_RATE"
    private const val READ_HEALTH_DATA_IN_BACKGROUND_PERMISSION =
        "android.permission.health.READ_HEALTH_DATA_IN_BACKGROUND"
    private const val UNIQUE_WORK_NAME = "passive-health-registration"
    private val requestedDataTypes = setOf(DataType.HEART_RATE_BPM, DataType.STEPS_DAILY)

    fun requiredPermissions(): List<String> = primaryPermissions() + backgroundPermissions()

    fun nextPermissionRequest(context: Context): List<String> {
        val missingPrimaryPermissions = primaryPermissions().missingFrom(context)
        if (missingPrimaryPermissions.isNotEmpty()) return missingPrimaryPermissions
        return backgroundPermissions().missingFrom(context)
    }

    fun hasRequiredPermissions(context: Context): Boolean = requiredPermissions().missingFrom(context).isEmpty()

    private fun primaryPermissions(): List<String> = buildList {
        add(Manifest.permission.ACTIVITY_RECOGNITION)
        if (Build.VERSION.SDK_INT >= 36) {
            add(READ_HEART_RATE_PERMISSION)
        } else {
            add(Manifest.permission.BODY_SENSORS)
        }
    }

    private fun backgroundPermissions(): List<String> = buildList {
        if (Build.VERSION.SDK_INT >= 36) {
            add(READ_HEALTH_DATA_IN_BACKGROUND_PERMISSION)
        } else if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            add(Manifest.permission.BODY_SENSORS_BACKGROUND)
        }
    }

    private fun List<String>.missingFrom(context: Context): List<String> = filter { permission ->
        context.checkSelfPermission(permission) != PackageManager.PERMISSION_GRANTED
    }

    fun enqueue(context: Context) {
        val request = OneTimeWorkRequestBuilder<PassiveHealthRegistrationWorker>()
            .setBackoffCriteria(BackoffPolicy.EXPONENTIAL, Duration.ofMinutes(1))
            .build()
        WorkManager.getInstance(context.applicationContext)
            .enqueueUniqueWork(UNIQUE_WORK_NAME, ExistingWorkPolicy.REPLACE, request)
    }

    suspend fun register(context: Context): PassiveHealthRegistrationResult {
        if (!hasRequiredPermissions(context)) return PassiveHealthRegistrationResult.MissingPermission

        val appContext = context.applicationContext
        val passiveClient = HealthServices.getClient(appContext).passiveMonitoringClient
        val supportedDataTypes = runCatching {
            passiveClient.getCapabilitiesAsync().await().supportedDataTypesPassiveMonitoring
        }.getOrElse { error ->
            return PassiveHealthRegistrationResult.Failed(error)
        }
        val dataTypes = requestedDataTypes.intersect(supportedDataTypes)
        if (dataTypes.isEmpty()) return PassiveHealthRegistrationResult.NoSupportedDataType

        val config = PassiveListenerConfig.builder()
            .setDataTypes(dataTypes)
            .build()
        return runCatching {
            passiveClient.setPassiveListenerServiceAsync(PassiveHealthService::class.java, config).await()
            publishCurrentState(appContext)
            PassiveHealthRegistrationResult.Registered(dataTypes.map { it.name }.toSet())
        }.getOrElse { error ->
            PassiveHealthRegistrationResult.Failed(error)
        }
    }

    private fun publishCurrentState(context: Context) {
        val snapshot = WatchSnapshotRepository(context).refreshDeviceState()
        val payload = ReportWatchSnapshotRequest.newBuilder()
            .setWatchSnapshot(snapshot)
            .build()
        SnapshotPublisher(context).publishIfAllowed(payload)
    }
}

class PassiveHealthRegistrationWorker(
    appContext: Context,
    workerParameters: WorkerParameters,
) : CoroutineWorker(appContext, workerParameters) {
    override suspend fun doWork(): Result {
        return when (val result = PassiveHealthRegistration.register(applicationContext)) {
            PassiveHealthRegistrationResult.MissingPermission -> Result.failure()
            PassiveHealthRegistrationResult.NoSupportedDataType -> Result.failure()
            is PassiveHealthRegistrationResult.Registered -> {
                Log.i(TAG, "Passive Health Services registered for ${result.dataTypes.joinToString()}")
                Result.success()
            }
            is PassiveHealthRegistrationResult.Failed -> {
                Log.w(TAG, "Passive Health Services registration failed", result.error)
                Result.retry()
            }
        }
    }

    private companion object {
        const val TAG = "PassiveHealthRegister"
    }
}

sealed class PassiveHealthRegistrationResult {
    data class Registered(val dataTypes: Set<String>) : PassiveHealthRegistrationResult()
    data object MissingPermission : PassiveHealthRegistrationResult()
    data object NoSupportedDataType : PassiveHealthRegistrationResult()
    data class Failed(val error: Throwable) : PassiveHealthRegistrationResult()
}
