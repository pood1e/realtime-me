import java.io.File
import java.net.URI

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.compose)
}

fun String.asBuildConfigString(): String = replace("\\", "\\\\").replace("\"", "\\\"")

// Generate a network-security-config that permits cleartext for ONLY the
// configured LAN gateway host (if any), keeping the rest of the app HTTPS-only.
// This replaces the app-wide usesCleartextTraffic flag.
val networkSecurityResDir = layout.buildDirectory.dir("generated/networkSecurity/res").get().asFile
run {
    val lanHost = providers.gradleProperty("statusGatewayLanUrl").orElse("").get()
        .let { runCatching { URI(it).host }.getOrNull().orEmpty() }
    val configFile = File(networkSecurityResDir, "xml/network_security_config.xml")
    configFile.parentFile.mkdirs()
    configFile.writeText(
        buildString {
            append("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
            append("<network-security-config>\n")
            append("    <base-config cleartextTrafficPermitted=\"false\" />\n")
            if (lanHost.isNotEmpty()) {
                append("    <domain-config cleartextTrafficPermitted=\"true\">\n")
                append("        <domain includeSubdomains=\"false\">$lanHost</domain>\n")
                append("    </domain-config>\n")
            }
            append("</network-security-config>\n")
        },
    )
}

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

    sourceSets {
        getByName("main").res.srcDir(networkSecurityResDir)
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
