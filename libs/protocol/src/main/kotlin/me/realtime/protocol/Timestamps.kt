package me.realtime.protocol

import com.google.protobuf.Timestamp
import java.time.Instant

fun Instant.toProtoTimestamp(): Timestamp =
    Timestamp.newBuilder()
        .setSeconds(epochSecond)
        .setNanos(nano)
        .build()

fun Timestamp.toJavaInstant(): Instant = Instant.ofEpochSecond(seconds, nanos.toLong())
