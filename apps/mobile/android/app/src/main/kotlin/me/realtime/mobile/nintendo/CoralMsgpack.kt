package me.realtime.mobile.nintendo

import java.io.ByteArrayOutputStream
import java.nio.ByteBuffer

internal object CoralMsgpack {
    fun pack(value: Any?): ByteArray = when (value) {
        null -> byteArrayOf(0xc0.toByte())
        is Boolean -> byteArrayOf(if (value) 0xc3.toByte() else 0xc2.toByte())
        is Int -> packLong(value.toLong())
        is Long -> packLong(value)
        is String -> packString(value)
        is ByteArray -> packBin(value)
        is List<*> -> packArray(value)
        is Map<*, *> -> packMap(value)
        else -> error("Unsupported msgpack type: ${value.javaClass.name}")
    }

    fun unpack(data: ByteArray): Any? = Reader(data).readRoot()

    fun packString(value: String): ByteArray {
        val data = value.toByteArray(Charsets.UTF_8)
        return when {
            data.size < 32 -> byteArrayOf((0xa0 + data.size).toByte()) + data
            data.size <= 0xff -> byteArrayOf(0xd9.toByte(), data.size.toByte()) + data
            data.size <= 0xffff -> byteArrayOf(0xda.toByte()) + shortBytes(data.size) + data
            else -> byteArrayOf(0xdb.toByte()) + intBytes(data.size) + data
        }
    }

    fun packBin(value: ByteArray): ByteArray = when {
        value.size <= 0xff -> byteArrayOf(0xc4.toByte(), value.size.toByte()) + value
        value.size <= 0xffff -> byteArrayOf(0xc5.toByte()) + shortBytes(value.size) + value
        else -> byteArrayOf(0xc6.toByte()) + intBytes(value.size) + value
    }

    private fun packLong(value: Long): ByteArray = when {
        value in 0..0x7f -> byteArrayOf(value.toByte())
        value in -32..-1 -> byteArrayOf(value.toByte())
        value in 0..0xff -> byteArrayOf(0xcc.toByte(), value.toByte())
        value in 0..0xffff -> byteArrayOf(0xcd.toByte()) + shortBytes(value.toInt())
        value in 0..0xffffffffL -> byteArrayOf(0xce.toByte()) + intBytes(value.toInt())
        value >= 0 -> byteArrayOf(0xcf.toByte()) + longBytes(value)
        value >= Byte.MIN_VALUE -> byteArrayOf(0xd0.toByte(), value.toByte())
        value >= Short.MIN_VALUE -> byteArrayOf(0xd1.toByte()) + shortBytes(value.toInt())
        value >= Int.MIN_VALUE -> byteArrayOf(0xd2.toByte()) + intBytes(value.toInt())
        else -> byteArrayOf(0xd3.toByte()) + longBytes(value)
    }

    private fun packArray(values: List<*>): ByteArray {
        val header = when {
            values.size < 16 -> byteArrayOf((0x90 + values.size).toByte())
            values.size <= 0xffff -> byteArrayOf(0xdc.toByte()) + shortBytes(values.size)
            else -> byteArrayOf(0xdd.toByte()) + intBytes(values.size)
        }
        return ByteArrayOutputStream().apply {
            write(header)
            values.forEach { write(pack(it)) }
        }.toByteArray()
    }

    private fun packMap(values: Map<*, *>): ByteArray {
        val header = when {
            values.size < 16 -> byteArrayOf((0x80 + values.size).toByte())
            values.size <= 0xffff -> byteArrayOf(0xde.toByte()) + shortBytes(values.size)
            else -> byteArrayOf(0xdf.toByte()) + intBytes(values.size)
        }
        return ByteArrayOutputStream().apply {
            write(header)
            values.forEach { (key, value) ->
                write(pack(key))
                write(pack(value))
            }
        }.toByteArray()
    }

    private fun shortBytes(value: Int): ByteArray = byteArrayOf((value ushr 8).toByte(), value.toByte())

    private fun intBytes(value: Int): ByteArray = ByteBuffer.allocate(Int.SIZE_BYTES).putInt(value).array()

    private fun longBytes(value: Long): ByteArray = ByteBuffer.allocate(Long.SIZE_BYTES).putLong(value).array()

    private class Reader(private val data: ByteArray) {
        private var offset = 0

        fun readRoot(): Any? {
            val value = readAny()
            check(offset == data.size) { "Trailing msgpack bytes" }
            return value
        }

        private fun readAny(): Any? {
            val code = u8()
            return when {
                code <= 0x7f -> code
                code >= 0xe0 -> code - 0x100
                code in 0x80..0x8f -> readMap(code and 0x0f)
                code in 0x90..0x9f -> readArray(code and 0x0f)
                code in 0xa0..0xbf -> readBytes(code and 0x1f).toString(Charsets.UTF_8)
                code == 0xc0 -> null
                code == 0xc2 -> false
                code == 0xc3 -> true
                code == 0xc4 -> readBytes(u8())
                code == 0xc5 -> readBytes(u16())
                code == 0xc6 -> readBytes(i32())
                code == 0xcc -> u8()
                code == 0xcd -> u16()
                code == 0xce -> i32().toLong() and 0xffffffffL
                code == 0xcf -> readLong()
                code == 0xd0 -> readBytes(1)[0].toInt()
                code == 0xd1 -> readBytes(2).let { ((it[0].toInt() and 0xff) shl 8 or (it[1].toInt() and 0xff)).toShort().toInt() }
                code == 0xd2 -> i32()
                code == 0xd3 -> readLong()
                code == 0xd9 -> readBytes(u8()).toString(Charsets.UTF_8)
                code == 0xda -> readBytes(u16()).toString(Charsets.UTF_8)
                code == 0xdb -> readBytes(i32()).toString(Charsets.UTF_8)
                else -> error("Unsupported msgpack code: 0x${code.toString(16)}")
            }
        }

        private fun readArray(size: Int): List<Any?> = List(size) { readAny() }

        private fun readMap(size: Int): Map<Any?, Any?> = buildMap {
            repeat(size) { put(readAny(), readAny()) }
        }

        private fun u8(): Int = data[offset++].toInt() and 0xff

        private fun u16(): Int = (u8() shl 8) or u8()

        private fun i32(): Int = (u8() shl 24) or (u8() shl 16) or (u8() shl 8) or u8()

        private fun readLong(): Long {
            var value = 0L
            repeat(Long.SIZE_BYTES) { value = (value shl 8) or u8().toLong() }
            return value
        }

        private fun readBytes(size: Int): ByteArray = data.copyOfRange(offset, offset + size).also { offset += size }
    }
}
