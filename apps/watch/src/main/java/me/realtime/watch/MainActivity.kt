package me.realtime.watch

import android.app.Activity
import android.os.Bundle
import android.widget.Toast
import me.realtime.watch.health.PassiveHealthRegistration

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

        if (PassiveHealthRegistration.hasRequiredPermissions(this)) {
            PassiveHealthRegistration.enqueue(this)
            Toast.makeText(this, R.string.passive_sync_enabled, Toast.LENGTH_SHORT).show()
            finish()
            return
        }

        val nextPermissions = PassiveHealthRegistration.nextPermissionRequest(this).toTypedArray()
        if (!permissions.contentEquals(nextPermissions) && nextPermissions.isNotEmpty()) {
            requestPermissions(nextPermissions, PERMISSIONS_REQUEST_CODE)
            return
        }

        Toast.makeText(this, R.string.health_permissions_required, Toast.LENGTH_LONG).show()
        finish()
    }

    private fun requestPermissionsOrStart() {
        val missingPermissions = PassiveHealthRegistration.nextPermissionRequest(this).toTypedArray()
        if (missingPermissions.isEmpty()) {
            PassiveHealthRegistration.enqueue(this)
            Toast.makeText(this, R.string.passive_sync_enabled, Toast.LENGTH_SHORT).show()
            finish()
            return
        }

        requestPermissions(missingPermissions, PERMISSIONS_REQUEST_CODE)
    }

    private companion object {
        const val PERMISSIONS_REQUEST_CODE = 1001
    }
}
