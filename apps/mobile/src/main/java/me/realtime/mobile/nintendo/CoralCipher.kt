package me.realtime.mobile.nintendo

import java.io.ByteArrayInputStream
import java.io.ByteArrayOutputStream
import java.security.MessageDigest
import java.security.SecureRandom
import java.util.Locale
import java.util.zip.GZIPInputStream
import java.util.zip.GZIPOutputStream
import javax.crypto.Cipher
import javax.crypto.spec.IvParameterSpec
import javax.crypto.spec.SecretKeySpec

internal class CoralCipher {
    private val random = SecureRandom()

    fun encryptRequest(appVersion: String, url: String, bodyText: String): ByteArray {
        val timestampMs = System.currentTimeMillis()
        val r = randomBytes(16).joinToString(separator = "") { "%02X".format(Locale.US, it.toInt() and 0xff) }
        val iv = randomBytes(16)
        val plaintext = CoralMsgpack.pack(
            linkedMapOf(
                "b" to bodyText,
                "e" to requestExtension(timestampMs, url, randomBytes(16)),
            ),
        )
        return CoralMsgpack.pack(
            linkedMapOf(
                "v" to appVersion,
                "r" to r,
                "t" to timestampMs,
                "p" to iv + encryptPayload(r, iv, plaintext),
            ),
        )
    }

    fun decryptBody(envelope: ByteArray): String {
        val outer = CoralMsgpack.unpack(envelope) as? Map<*, *> ?: error("Invalid Coral envelope")
        val r = outer["r"] as? String ?: error("Missing Coral r")
        val p = outer["p"] as? ByteArray ?: error("Missing Coral payload")
        val plaintext = decryptPayload(r, p.copyOfRange(0, 16), p.copyOfRange(16, p.size))
        val inner = CoralMsgpack.unpack(plaintext) as? Map<*, *> ?: error("Invalid Coral plaintext")
        return inner["b"] as? String ?: error("Missing Coral body")
    }

    private fun encryptPayload(r: String, iv: ByteArray, plaintext: ByteArray): ByteArray =
        aes(Cipher.ENCRYPT_MODE, r, iv).doFinal(gzip(plaintext))

    private fun decryptPayload(r: String, iv: ByteArray, ciphertext: ByteArray): ByteArray =
        gunzip(aes(Cipher.DECRYPT_MODE, r, iv).doFinal(ciphertext))

    private fun aes(mode: Int, r: String, iv: ByteArray): Cipher = Cipher.getInstance("AES/CBC/PKCS5Padding").apply {
        init(mode, SecretKeySpec(key(r), "AES"), IvParameterSpec(iv))
    }

    private fun key(r: String): ByteArray = MessageDigest.getInstance("SHA-256")
        .digest(r.toByteArray(Charsets.US_ASCII) + KEY_SALT)

    private fun requestExtension(timestampMs: Long, url: String, nonce: ByteArray): ByteArray {
        val metadata = ByteArrayOutputStream().apply {
            write(0x96)
            write(0xcf)
            write(longBytes(timestampMs))
            write(0xd0)
            write(0x00)
            write(CoralMsgpack.packString(url))
            write(0x00)
            write(0x00)
            write(0xd0)
            write(0x01)
        }.toByteArray()
        return ByteArrayOutputStream().apply {
            write(0x92)
            write(CoralMsgpack.packBin(nonce))
            write(0x95)
            repeat(4) { write(CoralMsgpack.packBin(ByteArray(0))) }
            write(CoralMsgpack.packBin(metadata))
        }.toByteArray()
    }

    private fun randomBytes(size: Int): ByteArray = ByteArray(size).also(random::nextBytes)

    private fun gzip(data: ByteArray): ByteArray = ByteArrayOutputStream().use { output ->
        GZIPOutputStream(output).use { it.write(data) }
        output.toByteArray()
    }

    private fun gunzip(data: ByteArray): ByteArray = GZIPInputStream(ByteArrayInputStream(data)).use { it.readBytes() }

    private fun longBytes(value: Long): ByteArray = byteArrayOf(
        (value ushr 56).toByte(),
        (value ushr 48).toByte(),
        (value ushr 40).toByte(),
        (value ushr 32).toByte(),
        (value ushr 24).toByte(),
        (value ushr 16).toByte(),
        (value ushr 8).toByte(),
        value.toByte(),
    )

    private companion object {
        val KEY_SALT: ByteArray = "b16a654fabf9857cd8bf91abd4ee86e5a7706dba04bf10a40b0c4b0cec425208"
            .chunked(2)
            .map { it.toInt(16).toByte() }
            .toByteArray()
    }
}
