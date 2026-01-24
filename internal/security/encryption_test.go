package security

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEncryptionService(t *testing.T) {
	service := NewEncryptionService("test-password")
	assert.NotNil(t, service)
	assert.NotEmpty(t, service.masterKey)
	assert.Equal(t, 32, len(service.masterKey)) // AES-256 requires 32-byte key
}

func TestEncryptionService_EncryptDecrypt(t *testing.T) {
	service := NewEncryptionService("test-password")

	tests := []struct {
		name      string
		plaintext string
	}{
		{"simple text", "Hello, World!"},
		{"empty string", ""},
		{"unicode text", "Hello, 世界! مرحبا!"},
		{"long text", strings.Repeat("This is a long text. ", 100)},
		{"special characters", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"json", `{"key": "value", "number": 123}`},
		{"multiline", "Line 1\nLine 2\nLine 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := service.Encrypt(tt.plaintext)
			require.NoError(t, err)
			assert.NotEqual(t, tt.plaintext, encrypted)

			decrypted, err := service.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, decrypted)
		})
	}
}

func TestEncryptionService_UniqueNonce(t *testing.T) {
	service := NewEncryptionService("test-password")
	plaintext := "Test message"

	// Encrypt same message multiple times
	encrypted1, err := service.Encrypt(plaintext)
	require.NoError(t, err)

	encrypted2, err := service.Encrypt(plaintext)
	require.NoError(t, err)

	encrypted3, err := service.Encrypt(plaintext)
	require.NoError(t, err)

	// Each encryption should produce different ciphertext due to random nonce
	assert.NotEqual(t, encrypted1, encrypted2)
	assert.NotEqual(t, encrypted1, encrypted3)
	assert.NotEqual(t, encrypted2, encrypted3)

	// But all should decrypt to the same plaintext
	decrypted1, _ := service.Decrypt(encrypted1)
	decrypted2, _ := service.Decrypt(encrypted2)
	decrypted3, _ := service.Decrypt(encrypted3)

	assert.Equal(t, plaintext, decrypted1)
	assert.Equal(t, plaintext, decrypted2)
	assert.Equal(t, plaintext, decrypted3)
}

func TestEncryptionService_DifferentKeys(t *testing.T) {
	service1 := NewEncryptionService("password1")
	service2 := NewEncryptionService("password2")

	plaintext := "Secret message"

	// Encrypt with service1
	encrypted, err := service1.Encrypt(plaintext)
	require.NoError(t, err)

	// Decrypt with service1 should work
	decrypted, err := service1.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)

	// Decrypt with service2 should fail
	_, err = service2.Decrypt(encrypted)
	assert.Error(t, err)
}

func TestEncryptionService_Decrypt_InvalidCiphertext(t *testing.T) {
	service := NewEncryptionService("test-password")

	tests := []struct {
		name       string
		ciphertext string
	}{
		{"invalid base64", "not-valid-base64!@#"},
		{"too short", "YWJj"}, // "abc" in base64
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Decrypt(tt.ciphertext)
			assert.Error(t, err)
		})
	}
}

func TestEncryptionService_EncryptCredentials(t *testing.T) {
	service := NewEncryptionService("test-password")

	credentials := map[string]string{
		"username": "admin",
		"password": "secret123",
		"apiKey":   "sk-test-12345",
	}

	encrypted, err := service.EncryptCredentials(credentials)
	require.NoError(t, err)
	assert.Len(t, encrypted, 3)

	// All values should be encrypted (different from original)
	for key, original := range credentials {
		assert.NotEqual(t, original, encrypted[key])
	}
}

func TestEncryptionService_DecryptCredentials(t *testing.T) {
	service := NewEncryptionService("test-password")

	originalCredentials := map[string]string{
		"username": "admin",
		"password": "secret123",
		"apiKey":   "sk-test-12345",
	}

	// Encrypt first
	encrypted, err := service.EncryptCredentials(originalCredentials)
	require.NoError(t, err)

	// Then decrypt
	decrypted, err := service.DecryptCredentials(encrypted)
	require.NoError(t, err)

	assert.Equal(t, originalCredentials, decrypted)
}

func TestEncryptionService_EncryptDecryptCredentials_RoundTrip(t *testing.T) {
	service := NewEncryptionService("test-password")

	original := map[string]string{
		"dbHost":     "localhost",
		"dbPassword": "P@ssw0rd!",
		"mqttBroker": "tcp://mqtt.example.com:1883",
		"mqttUser":   "sensor",
		"mqttPass":   "secret",
	}

	encrypted, err := service.EncryptCredentials(original)
	require.NoError(t, err)

	decrypted, err := service.DecryptCredentials(encrypted)
	require.NoError(t, err)

	assert.Equal(t, original, decrypted)
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"simple password", "password123"},
		{"complex password", "P@ssw0rd!@#$%"},
		{"unicode password", "密码123"},
		{"empty password", ""},
		{"long password", strings.Repeat("a", 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashPassword(tt.password)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)
		})
	}
}

func TestHashPassword_Consistency(t *testing.T) {
	password := "test-password"

	hash1 := HashPassword(password)
	hash2 := HashPassword(password)

	// Same password should produce same hash
	assert.Equal(t, hash1, hash2)
}

func TestHashPassword_Uniqueness(t *testing.T) {
	passwords := []string{
		"password1",
		"password2",
		"Password1",
		"PASSWORD1",
	}

	hashes := make(map[string]bool)
	for _, password := range passwords {
		hash := HashPassword(password)
		assert.False(t, hashes[hash], "Hash collision detected")
		hashes[hash] = true
	}
}

func TestVerifyPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		check    string
		expected bool
	}{
		{"correct password", "mypassword", "mypassword", true},
		{"wrong password", "mypassword", "wrongpassword", false},
		{"empty password", "", "", true},
		{"case sensitive", "Password", "password", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashPassword(tt.password)
			result := VerifyPassword(tt.check, hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVerifyPassword_Integration(t *testing.T) {
	// Simulate user registration and login
	originalPassword := "SecureP@ssw0rd!"

	// Registration: store hash
	storedHash := HashPassword(originalPassword)

	// Login: verify password
	loginPassword := "SecureP@ssw0rd!"
	assert.True(t, VerifyPassword(loginPassword, storedHash))

	// Wrong password
	wrongPassword := "WrongPassword"
	assert.False(t, VerifyPassword(wrongPassword, storedHash))
}

func BenchmarkEncrypt(b *testing.B) {
	service := NewEncryptionService("benchmark-password")
	plaintext := "Benchmark test message for encryption"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Encrypt(plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	service := NewEncryptionService("benchmark-password")
	plaintext := "Benchmark test message for encryption"
	encrypted, _ := service.Encrypt(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.Decrypt(encrypted)
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmark-password"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}
