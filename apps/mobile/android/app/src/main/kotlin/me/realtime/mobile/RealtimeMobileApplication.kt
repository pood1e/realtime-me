package me.realtime.mobile

import android.app.Application
import me.realtime.mobile.background.StatusBackgroundSync

class RealtimeMobileApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        StatusBackgroundSync.ensureScheduled(this)
    }
}
