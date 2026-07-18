plugins {
    alias(libs.plugins.android.application)
}

android {
    namespace = "me.realtime.watch"
    compileSdk = 37

    defaultConfig {
        applicationId = "me.realtime"
        minSdk = 30
        targetSdk = 37
        versionCode = 1
        versionName = "1.0"
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}

dependencies {
    implementation(dependencies.project(mapOf("path" to ":packages:status-protocol-android")))
    implementation(libs.coroutines.android)
    implementation(libs.play.services.wearable)
    implementation(libs.work.runtime.ktx)
    implementation(libs.health.services.client)
    implementation(libs.concurrent.futures.ktx)
}
