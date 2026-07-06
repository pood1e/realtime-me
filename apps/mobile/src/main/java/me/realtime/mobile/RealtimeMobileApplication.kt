package me.realtime.mobile

import android.app.Application
import me.realtime.mobile.background.GitHubBackgroundSync

class RealtimeMobileApplication : Application() {
    override fun onCreate() {
        super.onCreate()
        GitHubBackgroundSync.ensureScheduled(this)
    }
}
