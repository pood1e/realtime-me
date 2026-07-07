plugins {
    alias(libs.plugins.android.library)
    alias(libs.plugins.protobuf)
}

android {
    namespace = "me.realtime.protocol"
    compileSdk = 37

    defaultConfig {
        minSdk = 26
    }

    // Generate from the single shared contract in /proto so the Android client
    // and the Go gateway never drift; there is no per-module copy of the schema.
    sourceSets {
        getByName("main") {
            proto {
                srcDir("$rootDir/proto")
            }
        }
    }
}

protobuf {
    protoc {
        artifact = "com.google.protobuf:protoc:${libs.versions.protobuf.get()}"
    }
    generateProtoTasks {
        all().configureEach {
            builtins {
                create("java") {
                    option("lite")
                }
            }
        }
    }
}


dependencies {
    api(libs.protobuf.javalite)
}
