package me.realtime.mobile.status

import android.Manifest
import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothClass
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothManager
import android.bluetooth.BluetoothProfile
import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import me.realtime.status.v1.Accessory
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit

class BluetoothAudioAccessoryReader(private val context: Context) {
    fun read(): List<Accessory> {
        if (!hasBluetoothPermission()) return emptyList()
        val adapter = context.getSystemService(BluetoothManager::class.java)?.adapter ?: return emptyList()
        val devices = linkedMapOf<String, BluetoothDevice>()
        for (profile in audioProfiles()) {
            for (device in connectedDevices(adapter, profile)) {
                if (device.isAudioAccessory()) {
                    devices[device.address] = device
                }
            }
        }

        return devices.values.mapNotNull { it.toAccessory() }
    }

    private fun hasBluetoothPermission(): Boolean {
        return Build.VERSION.SDK_INT < Build.VERSION_CODES.S ||
            context.checkSelfPermission(Manifest.permission.BLUETOOTH_CONNECT) == PackageManager.PERMISSION_GRANTED
    }

    private fun audioProfiles(): List<Int> = buildList {
        add(BluetoothProfile.A2DP)
        add(BluetoothProfile.HEADSET)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) add(BluetoothProfile.HEARING_AID)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) add(BluetoothProfile.LE_AUDIO)
    }

    private fun connectedDevices(adapter: BluetoothAdapter, profileId: Int): List<BluetoothDevice> {
        var proxy: BluetoothProfile? = null
        val latch = CountDownLatch(1)
        val listener = object : BluetoothProfile.ServiceListener {
            override fun onServiceConnected(profile: Int, serviceProxy: BluetoothProfile) {
                if (profile == profileId) proxy = serviceProxy
                latch.countDown()
            }

            override fun onServiceDisconnected(profile: Int) {
                if (profile == profileId) latch.countDown()
            }
        }

        return try {
            if (!adapter.getProfileProxy(context, listener, profileId)) return emptyList()
            latch.await(PROFILE_TIMEOUT_MS, TimeUnit.MILLISECONDS)
            proxy?.connectedDevices.orEmpty()
        } catch (_: RuntimeException) {
            emptyList()
        } finally {
            proxy?.let { adapter.closeProfileProxy(profileId, it) }
        }
    }

    private fun BluetoothDevice.toAccessory(): Accessory? {
        val name = displayName()
        if (name.isBlank()) return null
        val builder = Accessory.newBuilder()
            .setKind(ACCESSORY_KIND)
            .setDisplayName(name)
        batteryPercent()?.let { builder.setBatteryPercent(it) }
        return builder.build()
    }

    private fun BluetoothDevice.displayName(): String {
        return firstNonBlank(
            runCatching {
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) alias else ""
            }.getOrDefault(""),
            runCatching { name }.getOrDefault(""),
        )
    }

    private fun BluetoothDevice.isAudioAccessory(): Boolean {
        val bluetoothClass = runCatching { bluetoothClass }.getOrNull() ?: return true
        if (bluetoothClass.majorDeviceClass == BluetoothClass.Device.Major.AUDIO_VIDEO) return true
        return bluetoothClass.deviceClass == BluetoothClass.Device.AUDIO_VIDEO_HEADPHONES ||
            bluetoothClass.deviceClass == BluetoothClass.Device.AUDIO_VIDEO_WEARABLE_HEADSET
    }

    private fun BluetoothDevice.batteryPercent(): Int? {
        val value = runCatching {
            javaClass.getMethod("getBatteryLevel").invoke(this) as? Int
        }.getOrNull() ?: return null
        return value.takeIf { it in 0..100 }
    }

    private fun firstNonBlank(vararg values: String?): String {
        return values.firstOrNull { !it.isNullOrBlank() && !it.equals("null", ignoreCase = true) }?.trim().orEmpty()
    }

    private companion object {
        const val ACCESSORY_KIND = "bluetooth_audio"
        const val PROFILE_TIMEOUT_MS = 800L
    }
}
