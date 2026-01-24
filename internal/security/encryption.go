package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptionService سرویس رمزنگاری
type EncryptionService struct {
	masterKey []byte
}

// NewEncryptionService ایجاد EncryptionService
func NewEncryptionService(password string) *EncryptionService {
	// Derive key from password using PBKDF2
	salt := []byte("edgeflow-salt-change-in-production") // Should be random and stored
	key := pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)

	return &EncryptionService{
		masterKey: key,
	}
}

// Encrypt رمزنگاری داده
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	// Create AES cipher
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt رمزگشایی داده
func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// Create AES cipher
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptCredentials رمزنگاری credentials
func (s *EncryptionService) EncryptCredentials(credentials map[string]string) (map[string]string, error) {
	encrypted := make(map[string]string)
	for key, value := range credentials {
		encryptedValue, err := s.Encrypt(value)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt %s: %w", key, err)
		}
		encrypted[key] = encryptedValue
	}
	return encrypted, nil
}

// DecryptCredentials رمزگشایی credentials
func (s *EncryptionService) DecryptCredentials(credentials map[string]string) (map[string]string, error) {
	decrypted := make(map[string]string)
	for key, value := range credentials {
		decryptedValue, err := s.Decrypt(value)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt %s: %w", key, err)
		}
		decrypted[key] = decryptedValue
	}
	return decrypted, nil
}

// HashPassword ایجاد hash برای password
func HashPassword(password string) string {
	salt := []byte("edgeflow-password-salt")
	hash := pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)
	return base64.StdEncoding.EncodeToString(hash)
}

// VerifyPassword تایید password
func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}
