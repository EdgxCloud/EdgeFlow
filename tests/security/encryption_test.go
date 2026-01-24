package security

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/edgeflow/edgeflow/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	require.NoError(t, err)

	t.Run("Encrypt and decrypt simple text", func(t *testing.T) {
		plaintext := "Hello, EdgeFlow!"

		encrypted, err := security.Encrypt([]byte(plaintext), key)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)
		assert.NotEqual(t, plaintext, string(encrypted))

		decrypted, err := security.Decrypt(encrypted, key)
		require.NoError(t, err)
		assert.Equal(t, plaintext, string(decrypted))
	})

	t.Run("Encrypt and decrypt empty string", func(t *testing.T) {
		plaintext := ""

		encrypted, err := security.Encrypt([]byte(plaintext), key)
		require.NoError(t, err)

		decrypted, err := security.Decrypt(encrypted, key)
		require.NoError(t, err)
		assert.Equal(t, plaintext, string(decrypted))
	})

	t.Run("Encrypt and decrypt large data", func(t *testing.T) {
		plaintext := make([]byte, 1024*1024) // 1MB
		_, err := rand.Read(plaintext)
		require.NoError(t, err)

		encrypted, err := security.Encrypt(plaintext, key)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, encrypted)

		decrypted, err := security.Decrypt(encrypted, key)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(plaintext, decrypted))
	})

	t.Run("Encrypt and decrypt JSON data", func(t *testing.T) {
		jsonData := `{"username":"admin","password":"secret123","apiKey":"abc-def-ghi"}`

		encrypted, err := security.Encrypt([]byte(jsonData), key)
		require.NoError(t, err)

		decrypted, err := security.Decrypt(encrypted, key)
		require.NoError(t, err)
		assert.Equal(t, jsonData, string(decrypted))
	})

	t.Run("Encrypt produces different ciphertext each time", func(t *testing.T) {
		plaintext := "Same plaintext"

		encrypted1, err := security.Encrypt([]byte(plaintext), key)
		require.NoError(t, err)

		encrypted2, err := security.Encrypt([]byte(plaintext), key)
		require.NoError(t, err)

		// Ciphertexts should be different due to random nonce
		assert.NotEqual(t, encrypted1, encrypted2)

		// But both should decrypt to same plaintext
		decrypted1, err := security.Decrypt(encrypted1, key)
		require.NoError(t, err)

		decrypted2, err := security.Decrypt(encrypted2, key)
		require.NoError(t, err)

		assert.Equal(t, plaintext, string(decrypted1))
		assert.Equal(t, plaintext, string(decrypted2))
	})
}

func TestEncryptionKeyValidation(t *testing.T) {
	t.Run("Reject invalid key length", func(t *testing.T) {
		plaintext := "Test data"

		// Too short key
		shortKey := make([]byte, 16)
		_, err := security.Encrypt([]byte(plaintext), shortKey)
		assert.Error(t, err)

		// Too long key
		longKey := make([]byte, 64)
		_, err = security.Encrypt([]byte(plaintext), longKey)
		assert.Error(t, err)
	})

	t.Run("Require exact 32 byte key", func(t *testing.T) {
		plaintext := "Test data"
		validKey := make([]byte, 32)
		_, err := rand.Read(validKey)
		require.NoError(t, err)

		encrypted, err := security.Encrypt([]byte(plaintext), validKey)
		require.NoError(t, err)
		assert.NotEmpty(t, encrypted)
	})

	t.Run("Different keys produce different ciphertext", func(t *testing.T) {
		plaintext := "Secret data"

		key1 := make([]byte, 32)
		_, err := rand.Read(key1)
		require.NoError(t, err)

		key2 := make([]byte, 32)
		_, err = rand.Read(key2)
		require.NoError(t, err)

		encrypted1, err := security.Encrypt([]byte(plaintext), key1)
		require.NoError(t, err)

		encrypted2, err := security.Encrypt([]byte(plaintext), key2)
		require.NoError(t, err)

		assert.NotEqual(t, encrypted1, encrypted2)
	})
}

func TestDecryptionErrors(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	t.Run("Decrypt with wrong key", func(t *testing.T) {
		plaintext := "Secret message"

		encrypted, err := security.Encrypt([]byte(plaintext), key)
		require.NoError(t, err)

		wrongKey := make([]byte, 32)
		_, err = rand.Read(wrongKey)
		require.NoError(t, err)

		_, err = security.Decrypt(encrypted, wrongKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication")
	})

	t.Run("Decrypt tampered ciphertext", func(t *testing.T) {
		plaintext := "Original message"

		encrypted, err := security.Encrypt([]byte(plaintext), key)
		require.NoError(t, err)

		// Tamper with ciphertext
		if len(encrypted) > 20 {
			encrypted[20] ^= 0xFF
		}

		_, err = security.Decrypt(encrypted, key)
		assert.Error(t, err)
	})

	t.Run("Decrypt truncated ciphertext", func(t *testing.T) {
		plaintext := "Original message"

		encrypted, err := security.Encrypt([]byte(plaintext), key)
		require.NoError(t, err)

		// Truncate ciphertext
		truncated := encrypted[:len(encrypted)/2]

		_, err = security.Decrypt(truncated, key)
		assert.Error(t, err)
	})

	t.Run("Decrypt invalid data", func(t *testing.T) {
		invalidData := []byte("this is not encrypted data")

		_, err := security.Decrypt(invalidData, key)
		assert.Error(t, err)
	})
}

func TestKeyRotation(t *testing.T) {
	t.Run("Re-encrypt with new key", func(t *testing.T) {
		plaintext := "Sensitive data"

		// Encrypt with old key
		oldKey := make([]byte, 32)
		_, err := rand.Read(oldKey)
		require.NoError(t, err)

		encrypted, err := security.Encrypt([]byte(plaintext), oldKey)
		require.NoError(t, err)

		// Decrypt with old key
		decrypted, err := security.Decrypt(encrypted, oldKey)
		require.NoError(t, err)
		assert.Equal(t, plaintext, string(decrypted))

		// Re-encrypt with new key
		newKey := make([]byte, 32)
		_, err = rand.Read(newKey)
		require.NoError(t, err)

		reEncrypted, err := security.Encrypt(decrypted, newKey)
		require.NoError(t, err)

		// Verify old key no longer works
		_, err = security.Decrypt(reEncrypted, oldKey)
		assert.Error(t, err)

		// Verify new key works
		finalDecrypted, err := security.Decrypt(reEncrypted, newKey)
		require.NoError(t, err)
		assert.Equal(t, plaintext, string(finalDecrypted))
	})
}

func TestCredentialEncryption(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	t.Run("Encrypt database credentials", func(t *testing.T) {
		credentials := map[string]string{
			"host":     "localhost",
			"port":     "5432",
			"database": "edgeflow",
			"username": "admin",
			"password": "super-secret-password-123!@#",
		}

		// Encrypt password
		encrypted, err := security.Encrypt([]byte(credentials["password"]), key)
		require.NoError(t, err)

		// Store encrypted password
		credentials["password"] = string(encrypted)

		// Later, decrypt password
		decrypted, err := security.Decrypt([]byte(credentials["password"]), key)
		require.NoError(t, err)

		assert.Equal(t, "super-secret-password-123!@#", string(decrypted))
	})

	t.Run("Encrypt API keys", func(t *testing.T) {
		apiKey := "sk-1234567890abcdefghijklmnop"

		encrypted, err := security.Encrypt([]byte(apiKey), key)
		require.NoError(t, err)

		decrypted, err := security.Decrypt(encrypted, key)
		require.NoError(t, err)

		assert.Equal(t, apiKey, string(decrypted))
	})

	t.Run("Encrypt MQTT credentials", func(t *testing.T) {
		mqttPassword := "mqtt-broker-password-xyz"

		encrypted, err := security.Encrypt([]byte(mqttPassword), key)
		require.NoError(t, err)

		decrypted, err := security.Decrypt(encrypted, key)
		require.NoError(t, err)

		assert.Equal(t, mqttPassword, string(decrypted))
	})
}

func TestEncryptionPerformance(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)

	t.Run("Encrypt small data quickly", func(t *testing.T) {
		plaintext := "Small data"

		encrypted, err := security.Encrypt([]byte(plaintext), key)
		require.NoError(t, err)

		decrypted, err := security.Decrypt(encrypted, key)
		require.NoError(t, err)

		assert.Equal(t, plaintext, string(decrypted))
	})

	t.Run("Handle concurrent encryption", func(t *testing.T) {
		const numGoroutines = 100

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				plaintext := []byte("Concurrent test data")

				encrypted, err := security.Encrypt(plaintext, key)
				if err != nil {
					t.Errorf("Encryption failed: %v", err)
					done <- false
					return
				}

				decrypted, err := security.Decrypt(encrypted, key)
				if err != nil {
					t.Errorf("Decryption failed: %v", err)
					done <- false
					return
				}

				if !bytes.Equal(plaintext, decrypted) {
					t.Errorf("Decrypted data doesn't match")
					done <- false
					return
				}

				done <- true
			}(i)
		}

		for i := 0; i < numGoroutines; i++ {
			success := <-done
			assert.True(t, success)
		}
	})
}

// Benchmark tests
func BenchmarkEncrypt(b *testing.B) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	plaintext := []byte("Benchmark encryption data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = security.Encrypt(plaintext, key)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	plaintext := []byte("Benchmark decryption data")
	encrypted, _ := security.Encrypt(plaintext, key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = security.Decrypt(encrypted, key)
	}
}

func BenchmarkEncryptLargeData(b *testing.B) {
	key := make([]byte, 32)
	_, _ = rand.Read(key)

	plaintext := make([]byte, 1024*1024) // 1MB
	_, _ = rand.Read(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = security.Encrypt(plaintext, key)
	}
}
