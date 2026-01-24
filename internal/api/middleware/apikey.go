package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// APIKey ساختار API Key
type APIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	KeyHash     string    `json:"key_hash"`      // Hashed version of the key
	Prefix      string    `json:"prefix"`        // First 8 chars for identification
	Permissions []string  `json:"permissions"`   // List of allowed permissions
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
	Active      bool      `json:"active"`
}

// APIKeyStore ذخیره API Keys
type APIKeyStore struct {
	keys map[string]*APIKey // key: hash of API key
	mu   sync.RWMutex
}

// NewAPIKeyStore ایجاد APIKeyStore
func NewAPIKeyStore() *APIKeyStore {
	return &APIKeyStore{
		keys: make(map[string]*APIKey),
	}
}

// GenerateAPIKey تولید API key جدید
func (s *APIKeyStore) GenerateAPIKey(name string, permissions []string, expiresIn time.Duration) (string, *APIKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate random key (32 bytes = 64 hex chars)
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", nil, err
	}
	key := "efk_" + hex.EncodeToString(keyBytes) // efk = EdgeFlow Key

	// Hash the key for storage
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	// Create APIKey
	apiKey := &APIKey{
		ID:          generateID(),
		Name:        name,
		KeyHash:     keyHash,
		Prefix:      key[:12], // Store prefix for identification
		Permissions: permissions,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(expiresIn),
		Active:      true,
	}

	// Store
	s.keys[keyHash] = apiKey

	return key, apiKey, nil
}

// ValidateAPIKey اعتبارسنجی API key
func (s *APIKeyStore) ValidateAPIKey(key string) (*APIKey, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Hash the provided key
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	// Find key
	apiKey, exists := s.keys[keyHash]
	if !exists {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check if active
	if !apiKey.Active {
		return nil, fmt.Errorf("API key is inactive")
	}

	// Check expiration
	if time.Now().After(apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key has expired")
	}

	// Update last used
	apiKey.LastUsedAt = time.Now()

	return apiKey, nil
}

// RevokeAPIKey لغو API key
func (s *APIKeyStore) RevokeAPIKey(keyHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	apiKey, exists := s.keys[keyHash]
	if !exists {
		return fmt.Errorf("API key not found")
	}

	apiKey.Active = false
	return nil
}

// ListAPIKeys لیست API keys
func (s *APIKeyStore) ListAPIKeys() []*APIKey {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]*APIKey, 0, len(s.keys))
	for _, key := range s.keys {
		keys = append(keys, key)
	}
	return keys
}

// APIKeyMiddleware میدلور API Key
func APIKeyMiddleware(store *APIKeyStore, requiredPermissions []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get API key from header
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			// Try query parameter
			apiKey = c.Query("api_key")
		}

		if apiKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing API key",
			})
		}

		// Validate API key
		key, err := store.ValidateAPIKey(apiKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// Check permissions
		if len(requiredPermissions) > 0 {
			hasPermission := false
			for _, required := range requiredPermissions {
				for _, permission := range key.Permissions {
					if permission == required || permission == "*" {
						hasPermission = true
						break
					}
				}
				if hasPermission {
					break
				}
			}
			if !hasPermission {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Insufficient permissions",
				})
			}
		}

		// Store API key info in context
		c.Locals("api_key_id", key.ID)
		c.Locals("api_key_name", key.Name)
		c.Locals("api_key_permissions", key.Permissions)

		return c.Next()
	}
}

// CombinedAuthMiddleware میدلور ترکیبی (JWT یا API Key)
func CombinedAuthMiddleware(jwtConfig JWTConfig, apiKeyStore *APIKeyStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if path should skip authentication
		path := c.Path()
		for _, skipPath := range jwtConfig.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return c.Next()
			}
		}

		// Try API Key first
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey != "" {
			// Validate API key
			key, err := apiKeyStore.ValidateAPIKey(apiKey)
			if err == nil {
				// API key is valid
				c.Locals("auth_type", "api_key")
				c.Locals("api_key_id", key.ID)
				c.Locals("api_key_name", key.Name)
				c.Locals("api_key_permissions", key.Permissions)
				return c.Next()
			}
		}

		// Try JWT
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString != authHeader {
				claims, err := ValidateToken(tokenString, jwtConfig)
				if err == nil {
					// JWT is valid
					c.Locals("auth_type", "jwt")
					c.Locals("user_id", claims.UserID)
					c.Locals("username", claims.Username)
					c.Locals("roles", claims.Roles)
					return c.Next()
				}
			}
		}

		// No valid authentication found
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}
}

// generateID تولید ID یکتا
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
