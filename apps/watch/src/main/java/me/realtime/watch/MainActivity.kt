package me.realtime.watch

import android.app.Activity
import android.content.pm.PackageManager
import android.os.Bundle
import android.widget.Toast
import me.realtime.watch.service.SensorCollectionService
import me.realtime.watch.sensors.WatchSensorCollector

class MainActivity : Activity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        requestPermissionsOrStart()
    }

    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray,
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode != PERMISSIONS_REQUEST_CODE) return finish()

        if (WatchSensorCollector.hasRequiredPermissions(this)) {
            SensorCollectionService.start(this)
            Toast.makeText(this, R.string.sensor_sync_started, Toast.LENGTH_SHORT).show()
        } else {
            Toast.makeText(this, R.string.sensor_permissions_required, Toast.LENGTH_LONG).show()
        }
        finish()
    }

    private fun requestPermissionsOrStart() {
        val missingPermissions = WatchSensorCollector.requiredPermissions()
            .filter { checkSelfPermission(it) != PackageManager.PERMISSION_GRANTED }
            .toTypedArray()
        if (missingPermissions.isEmpty()) {
            SensorCollectionService.start(this)
            Toast.makeText(this, R.string.sensor_sync_started, Toast.LENGTH_SHORT).show()
            finish()
            return
        }

        requestPermissions(missingPermissions, PERMISSIONS_REQUEST_CODE)
    }

    private companion object {
        const val PERMISSIONS_REQUEST_CODE = 1001
    }
}
