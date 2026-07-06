plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.compose)
}

fun String.asBuildConfigString(): String = replace("\\", "\\\\").replace("\"", "\\\"")

android {
    namespace = "me.realtime.mobile"
    compileSdk = 37

    defaultConfig {
        applicationId = "me.realtime"
        minSdk = 26
        targetSdk = 37
        versionCode = 1
        versionName = "1.0"
    }

    buildFeatures {
        compose = true
        buildConfig = true
    }

    buildTypes {
        all {
            buildConfigField(
                "String",
                "STATUS_GATEWAY_LAN_URL",
                "\"${providers.gradleProperty("statusGatewayLanUrl").orElse("").get().asBuildConfigString()}\"",
            )
            buildConfigField(
                "String",
                "STATUS_GATEWAY_PUBLIC_URL",
                "\"${providers.gradleProperty("statusGatewayPublicUrl").orElse("").get().asBuildConfigString()}\"",
            )
            manifestPlaceholders["usesCleartextTraffic"] = providers.gradleProperty("statusGatewayAllowCleartext")
                .orElse("false")
                .get()
        }
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}

dependencies {
    implementation(dependencies.project(mapOf("path" to ":libs:protocol")))
    implementation(platform(libs.compose.bom))
    implementation(libs.activity.compose)
    implementation(libs.compose.ui)
    implementation(libs.compose.ui.tooling.preview)
    implementation(libs.compose.material3)
    implementation(libs.compose.material.icons.extended)
    debugImplementation(libs.compose.ui.tooling)
    implementation(libs.coroutines.android)
    implementation(libs.play.services.wearable)
    implementation(libs.work.runtime.ktx)
}
