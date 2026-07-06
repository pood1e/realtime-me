package me.realtime.mobile.state

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import androidx.core.content.edit
import java.nio.charset.StandardCharsets
import java.security.KeyStore
import java.util.Base64
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

class EncryptedSecretStore(
    context: Context,
    private val prefsName: String,
    private val keyAlias: String,
    private val ivKey: String = DEFAULT_IV_KEY,
    private val ciphertextKey: String = DEFAULT_CIPHERTEXT_KEY,
) {
    private val preferences = context.getSharedPreferences(prefsName, Context.MODE_PRIVATE)

    fun save(value: String) {
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(Cipher.ENCRYPT_MODE, secretKey())
        val encrypted = cipher.doFinal(value.toByteArray(StandardCharsets.UTF_8))
        preferences.edit {
            putString(ivKey, Base64.getEncoder().encodeToString(cipher.iv))
            putString(ciphertextKey, Base64.getEncoder().encodeToString(encrypted))
        }
    }

    fun value(): String? {
        val encodedIv = preferences.getString(ivKey, null) ?: return null
        val encodedCiphertext = preferences.getString(ciphertextKey, null) ?: return null
        return runCatching {
            val cipher = Cipher.getInstance(TRANSFORMATION)
            cipher.init(
                Cipher.DECRYPT_MODE,
                secretKey(),
                GCMParameterSpec(GCM_TAG_LENGTH_BITS, Base64.getDecoder().decode(encodedIv)),
            )
            val decrypted = cipher.doFinal(Base64.getDecoder().decode(encodedCiphertext))
            String(decrypted, StandardCharsets.UTF_8).takeIf { it.isNotBlank() }
        }.getOrNull()
    }

    fun hasValue(): Boolean = value() != null

    fun clear() {
        preferences.edit {
            remove(ivKey)
            remove(ciphertextKey)
        }
    }

    private fun secretKey(): SecretKey {
        val keyStore = KeyStore.getInstance(ANDROID_KEYSTORE).apply { load(null) }
        (keyStore.getKey(keyAlias, null) as? SecretKey)?.let { return it }

        val generator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, ANDROID_KEYSTORE)
        val spec = KeyGenParameterSpec.Builder(
            keyAlias,
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT,
        )
            .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            .setKeySize(KEY_SIZE_BITS)
            .build()
        generator.init(spec)
        return generator.generateKey()
    }

    private companion object {
        const val DEFAULT_IV_KEY = "secret_iv"
        const val DEFAULT_CIPHERTEXT_KEY = "secret_ciphertext"
        const val ANDROID_KEYSTORE = "AndroidKeyStore"
        const val TRANSFORMATION = "AES/GCM/NoPadding"
        const val KEY_SIZE_BITS = 256
        const val GCM_TAG_LENGTH_BITS = 128
    }
}
