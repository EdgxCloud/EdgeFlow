package security

import (
	"testing"
	"time"

	"github.com/edgeflow/edgeflow/internal/api/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTGeneration(t *testing.T) {
	secretKey := []byte("test-secret-key-for-jwt-signing-12345")

	t.Run("Generate valid JWT token", func(t *testing.T) {
		username := "testuser"
		expiresAt := time.Now().Add(24 * time.Hour)

		token, err := middleware.GenerateJWT(username, expiresAt, secretKey)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// Token should be a valid JWT format (three parts separated by dots)
		assert.Contains(t, token, ".")
	})

	t.Run("Generate multiple different tokens", func(t *testing.T) {
		token1, err := middleware.GenerateJWT("user1", time.Now().Add(1*time.Hour), secretKey)
		require.NoError(t, err)

		token2, err := middleware.GenerateJWT("user2", time.Now().Add(1*time.Hour), secretKey)
		require.NoError(t, err)

		// Tokens for different users should be different
		assert.NotEqual(t, token1, token2)
	})
}

func TestJWTValidation(t *testing.T) {
	secretKey := []byte("test-secret-key-for-jwt-signing-12345")

	t.Run("Validate valid JWT token", func(t *testing.T) {
		username := "testuser"
		expiresAt := time.Now().Add(1 * time.Hour)

		token, err := middleware.GenerateJWT(username, expiresAt, secretKey)
		require.NoError(t, err)

		claims, err := middleware.ValidateJWT(token, secretKey)
		require.NoError(t, err)
		assert.Equal(t, username, claims.Username)
		assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	})

	t.Run("Reject invalid signature", func(t *testing.T) {
		token, err := middleware.GenerateJWT("testuser", time.Now().Add(1*time.Hour), secretKey)
		require.NoError(t, err)

		// Try to validate with different secret
		wrongSecret := []byte("wrong-secret-key")
		_, err = middleware.ValidateJWT(token, wrongSecret)
		assert.Error(t, err)
	})

	t.Run("Reject expired token", func(t *testing.T) {
		username := "testuser"
		expiresAt := time.Now().Add(-1 * time.Hour) // Already expired

		token, err := middleware.GenerateJWT(username, expiresAt, secretKey)
		require.NoError(t, err)

		_, err = middleware.ValidateJWT(token, secretKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("Reject malformed token", func(t *testing.T) {
		malformedTokens := []string{
			"",
			"not.a.token",
			"invalid-token",
			"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
		}

		for _, token := range malformedTokens {
			_, err := middleware.ValidateJWT(token, secretKey)
			assert.Error(t, err, "Should reject malformed token: %s", token)
		}
	})

	t.Run("Reject token with wrong algorithm", func(t *testing.T) {
		// Create a token with HS512 instead of HS256
		claims := jwt.MapClaims{
			"username": "testuser",
			"exp":      time.Now().Add(1 * time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		tokenString, err := token.SignedString(secretKey)
		require.NoError(t, err)

		_, err = middleware.ValidateJWT(tokenString, secretKey)
		assert.Error(t, err)
	})
}

func TestJWTExpiry(t *testing.T) {
	secretKey := []byte("test-secret-key-for-jwt-signing-12345")

	t.Run("Token expires correctly", func(t *testing.T) {
		username := "testuser"
		expiresAt := time.Now().Add(100 * time.Millisecond)

		token, err := middleware.GenerateJWT(username, expiresAt, secretKey)
		require.NoError(t, err)

		// Validate immediately - should work
		claims, err := middleware.ValidateJWT(token, secretKey)
		require.NoError(t, err)
		assert.Equal(t, username, claims.Username)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Validate after expiration - should fail
		_, err = middleware.ValidateJWT(token, secretKey)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("Extract remaining time", func(t *testing.T) {
		expiresAt := time.Now().Add(5 * time.Minute)
		token, err := middleware.GenerateJWT("testuser", expiresAt, secretKey)
		require.NoError(t, err)

		claims, err := middleware.ValidateJWT(token, secretKey)
		require.NoError(t, err)

		remaining := time.Until(claims.ExpiresAt.Time)
		assert.True(t, remaining > 4*time.Minute)
		assert.True(t, remaining < 6*time.Minute)
	})
}

func TestRefreshToken(t *testing.T) {
	secretKey := []byte("test-secret-key-for-jwt-signing-12345")

	t.Run("Refresh valid token", func(t *testing.T) {
		username := "testuser"
		expiresAt := time.Now().Add(1 * time.Hour)

		oldToken, err := middleware.GenerateJWT(username, expiresAt, secretKey)
		require.NoError(t, err)

		// Refresh token (extend expiration)
		newExpiresAt := time.Now().Add(24 * time.Hour)
		newToken, err := middleware.GenerateJWT(username, newExpiresAt, secretKey)
		require.NoError(t, err)

		// Both tokens should be valid but different
		assert.NotEqual(t, oldToken, newToken)

		oldClaims, err := middleware.ValidateJWT(oldToken, secretKey)
		require.NoError(t, err)

		newClaims, err := middleware.ValidateJWT(newToken, secretKey)
		require.NoError(t, err)

		assert.Equal(t, oldClaims.Username, newClaims.Username)
		assert.True(t, newClaims.ExpiresAt.Time.After(oldClaims.ExpiresAt.Time))
	})

	t.Run("Cannot refresh expired token", func(t *testing.T) {
		username := "testuser"
		expiresAt := time.Now().Add(-1 * time.Hour) // Already expired

		token, err := middleware.GenerateJWT(username, expiresAt, secretKey)
		require.NoError(t, err)

		// Try to validate expired token - should fail
		_, err = middleware.ValidateJWT(token, secretKey)
		assert.Error(t, err)
	})
}

func TestJWTClaims(t *testing.T) {
	secretKey := []byte("test-secret-key-for-jwt-signing-12345")

	t.Run("Extract username from token", func(t *testing.T) {
		username := "admin@edgeflow.io"
		expiresAt := time.Now().Add(1 * time.Hour)

		token, err := middleware.GenerateJWT(username, expiresAt, secretKey)
		require.NoError(t, err)

		claims, err := middleware.ValidateJWT(token, secretKey)
		require.NoError(t, err)

		assert.Equal(t, username, claims.Username)
	})

	t.Run("Claims contain correct fields", func(t *testing.T) {
		username := "testuser"
		expiresAt := time.Now().Add(1 * time.Hour)

		token, err := middleware.GenerateJWT(username, expiresAt, secretKey)
		require.NoError(t, err)

		claims, err := middleware.ValidateJWT(token, secretKey)
		require.NoError(t, err)

		assert.NotEmpty(t, claims.Username)
		assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
		assert.False(t, claims.ExpiresAt.Time.IsZero())
	})
}

func TestJWTSecurityScenarios(t *testing.T) {
	secretKey := []byte("test-secret-key-for-jwt-signing-12345")

	t.Run("Token reuse is allowed within validity", func(t *testing.T) {
		username := "testuser"
		expiresAt := time.Now().Add(1 * time.Hour)

		token, err := middleware.GenerateJWT(username, expiresAt, secretKey)
		require.NoError(t, err)

		// Use token multiple times
		for i := 0; i < 5; i++ {
			claims, err := middleware.ValidateJWT(token, secretKey)
			require.NoError(t, err)
			assert.Equal(t, username, claims.Username)
		}
	})

	t.Run("Different secrets produce different tokens", func(t *testing.T) {
		username := "testuser"
		expiresAt := time.Now().Add(1 * time.Hour)

		secret1 := []byte("secret-key-1")
		secret2 := []byte("secret-key-2")

		token1, err := middleware.GenerateJWT(username, expiresAt, secret1)
		require.NoError(t, err)

		token2, err := middleware.GenerateJWT(username, expiresAt, secret2)
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2)

		// Each token only validates with its own secret
		_, err = middleware.ValidateJWT(token1, secret1)
		assert.NoError(t, err)

		_, err = middleware.ValidateJWT(token1, secret2)
		assert.Error(t, err)
	})

	t.Run("Token tampering is detected", func(t *testing.T) {
		username := "testuser"
		expiresAt := time.Now().Add(1 * time.Hour)

		token, err := middleware.GenerateJWT(username, expiresAt, secretKey)
		require.NoError(t, err)

		// Try to tamper with token by changing a character
		if len(token) > 10 {
			tamperedToken := token[:10] + "X" + token[11:]

			_, err = middleware.ValidateJWT(tamperedToken, secretKey)
			assert.Error(t, err)
		}
	})
}

// Benchmark tests
func BenchmarkJWTGeneration(b *testing.B) {
	secretKey := []byte("test-secret-key-for-jwt-signing-12345")
	username := "testuser"
	expiresAt := time.Now().Add(24 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = middleware.GenerateJWT(username, expiresAt, secretKey)
	}
}

func BenchmarkJWTValidation(b *testing.B) {
	secretKey := []byte("test-secret-key-for-jwt-signing-12345")
	username := "testuser"
	expiresAt := time.Now().Add(24 * time.Hour)

	token, _ := middleware.GenerateJWT(username, expiresAt, secretKey)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = middleware.ValidateJWT(token, secretKey)
	}
}
