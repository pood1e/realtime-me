package me.realtime.mobile

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import androidx.core.content.ContextCompat
import io.flutter.embedding.android.FlutterActivity
import io.flutter.embedding.engine.FlutterEngine
import me.realtime.mobile.background.StatusBackgroundSync
import me.realtime.mobile.platform.SnapshotsStreamHandler
import me.realtime.mobile.platform.StatusBridgeHost
import me.realtime.mobile.platform.StatusHostApi
import me.realtime.mobile.platform.StatusSnapshotEvents

class MainActivity : FlutterActivity() {
    private var permissionResult: ((Result<Boolean>) -> Unit)? = null
    private var statusHost: StatusBridgeHost? = null
    private var statusEvents: StatusSnapshotEvents? = null

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)
        val messenger = flutterEngine.dartExecutor.binaryMessenger
        statusHost = StatusBridgeHost(
            context = applicationContext,
            hasRequiredPermissions = ::hasRequiredStatusPermissions,
            requestPermissions = ::requestStatusPermissions,
        ).also { StatusHostApi.setUp(messenger, it) }
        statusEvents = StatusSnapshotEvents(applicationContext).also {
            SnapshotsStreamHandler.register(messenger, it)
        }
        StatusBackgroundSync.ensureActive(applicationContext)
    }

    override fun cleanUpFlutterEngine(flutterEngine: FlutterEngine) {
        StatusHostApi.setUp(flutterEngine.dartExecutor.binaryMessenger, null)
        statusHost?.close()
        statusHost = null
        statusEvents?.close()
        statusEvents = null
        permissionResult?.invoke(Result.failure(IllegalStateException("Activity detached")))
        permissionResult = null
        super.cleanUpFlutterEngine(flutterEngine)
    }

    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray,
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode != STATUS_PERMISSION_REQUEST_CODE) return
        permissionResult?.invoke(Result.success(hasRequiredStatusPermissions()))
        permissionResult = null
    }

    private fun hasRequiredStatusPermissions(): Boolean = requiredStatusPermissions().all { permission ->
        ContextCompat.checkSelfPermission(this, permission) == PackageManager.PERMISSION_GRANTED
    }

    private fun requestStatusPermissions(callback: (Result<Boolean>) -> Unit) {
        if (hasRequiredStatusPermissions()) {
            callback(Result.success(true))
            return
        }
        if (permissionResult != null) {
            callback(Result.failure(IllegalStateException("A permission request is already active")))
            return
        }
        permissionResult = callback
        requestPermissions(requiredStatusPermissions().toTypedArray(), STATUS_PERMISSION_REQUEST_CODE)
    }

    private fun requiredStatusPermissions(): List<String> = buildList {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            add(Manifest.permission.BLUETOOTH_CONNECT)
        }
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            add(Manifest.permission.POST_NOTIFICATIONS)
        }
    }

    private companion object {
        const val STATUS_PERMISSION_REQUEST_CODE = 1001
    }
}
