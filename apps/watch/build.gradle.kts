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
        // Wear artifacts require a version code distinct from every phone artifact.
        versionCode = 37_000_001
        versionName = "0.1.0"
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}

project.extensions.extraProperties["realtimeMe.signingPropertiesFile"] =
    rootProject.file("apps/mobile/android/key.properties")
apply(from = rootProject.file("gradle/release-signing.gradle"))

dependencies {
    implementation(project(":packages:status-protocol-android"))
    implementation(libs.coroutines.android)
    implementation(libs.play.services.wearable)
    implementation(libs.work.runtime.ktx)
    implementation(libs.health.services.client)
    implementation(libs.concurrent.futures.ktx)
}
