plugins {
    id("com.android.application")
    id("dev.flutter.flutter-gradle-plugin")
}

android {
    namespace = "me.realtime.mobile"
    compileSdk = 37
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    defaultConfig {
        applicationId = "me.realtime"
        minSdk = 26
        targetSdk = 37
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }
}

project.extensions.extraProperties["realtimeMe.signingPropertiesFile"] = rootProject.file("key.properties")
apply(from = rootProject.file("../../../gradle/release-signing.gradle"))

kotlin {
    compilerOptions {
        jvmTarget = org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17
    }
}

flutter {
    source = "../.."
}

dependencies {
    implementation(project(":status-protocol-android"))
    implementation(libs.coroutines.android)
    implementation(libs.play.services.wearable)
    implementation(libs.work.runtime.ktx)
}
