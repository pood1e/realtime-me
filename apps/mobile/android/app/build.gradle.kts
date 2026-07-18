import java.io.File
import java.net.URI

plugins {
    id("com.android.application")
    id("dev.flutter.flutter-gradle-plugin")
}

fun String.asBuildConfigString(): String = replace("\\", "\\\\").replace("\"", "\\\"")

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

    buildFeatures {
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
}

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
