package middleware

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	config := JWTConfig{
		SecretKey:  "test-secret-key",
		Expiration: 1 * time.Hour,
		Issuer:     "test-issuer",
	}

	token, err := GenerateToken("user-123", "testuser", []string{"admin"}, config)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateToken_DefaultValues(t *testing.T) {
	config := JWTConfig{} // Empty config

	token, err := GenerateToken("user-123", "testuser", []string{"user"}, config)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestGenerateToken_MultipleRoles(t *testing.T) {
	config := JWTConfig{
		SecretKey: "test-secret-key",
	}

	roles := []string{"admin", "editor", "viewer"}
	token, err := GenerateToken("user-123", "testuser", roles, config)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate roles are in token
	claims, err := ValidateToken(token, config)
	require.NoError(t, err)
	assert.Equal(t, roles, claims.Roles)
}

func TestValidateToken(t *testing.T) {
	config := JWTConfig{
		SecretKey:  "test-secret-key",
		Expiration: 1 * time.Hour,
		Issuer:     "test-issuer",
	}

	// Generate token
	token, err := GenerateToken("user-123", "testuser", []string{"admin"}, config)
	require.NoError(t, err)

	// Validate token
	claims, err := ValidateToken(token, config)
	require.NoError(t, err)

	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, []string{"admin"}, claims.Roles)
	assert.Equal(t, "test-issuer", claims.Issuer)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	config := JWTConfig{
		SecretKey: "test-secret-key",
	}

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid format", "not.a.valid.token"},
		{"random string", "random-string-without-dots"},
		{"malformed jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateToken(tt.token, config)
			assert.Error(t, err)
		})
	}
}

func TestValidateToken_WrongKey(t *testing.T) {
	config1 := JWTConfig{SecretKey: "key-1"}
	config2 := JWTConfig{SecretKey: "key-2"}

	// Generate token with config1
	token, err := GenerateToken("user-123", "testuser", []string{"user"}, config1)
	require.NoError(t, err)

	// Validate with config1 should work
	claims, err := ValidateToken(token, config1)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)

	// Validate with config2 should fail
	_, err = ValidateToken(token, config2)
	assert.Error(t, err)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	config := JWTConfig{
		SecretKey:  "test-secret-key",
		Expiration: 1 * time.Nanosecond, // Extremely short expiration
	}

	token, err := GenerateToken("user-123", "testuser", []string{"user"}, config)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Validate should fail due to expiration
	_, err = ValidateToken(token, config)
	assert.Error(t, err)
}

func TestValidateToken_Claims(t *testing.T) {
	config := JWTConfig{
		SecretKey:  "test-secret-key",
		Expiration: 1 * time.Hour,
		Issuer:     "edgeflow",
	}

	token, err := GenerateToken("user-456", "admin_user", []string{"admin", "operator"}, config)
	require.NoError(t, err)

	claims, err := ValidateToken(token, config)
	require.NoError(t, err)

	// Verify all claims
	assert.Equal(t, "user-456", claims.UserID)
	assert.Equal(t, "admin_user", claims.Username)
	assert.Contains(t, claims.Roles, "admin")
	assert.Contains(t, claims.Roles, "operator")
	assert.Equal(t, "edgeflow", claims.Issuer)

	// Check timestamps
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)

	// ExpiresAt should be in the future
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestJWTConfig_Defaults(t *testing.T) {
	config := JWTConfig{}

	// GenerateToken should apply defaults
	token, err := GenerateToken("user-123", "testuser", []string{"user"}, config)
	require.NoError(t, err)

	// ValidateToken with empty config should also work (applies same defaults)
	claims, err := ValidateToken(token, JWTConfig{})
	require.NoError(t, err)

	assert.Equal(t, "edgeflow", claims.Issuer)
}

func TestGenerateValidateToken_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		username string
		roles    []string
	}{
		{
			name:     "single role",
			userID:   "user-1",
			username: "john",
			roles:    []string{"user"},
		},
		{
			name:     "multiple roles",
			userID:   "user-2",
			username: "jane",
			roles:    []string{"admin", "editor", "viewer"},
		},
		{
			name:     "empty roles",
			userID:   "user-3",
			username: "guest",
			roles:    []string{},
		},
		{
			name:     "special characters in username",
			userID:   "user-4",
			username: "user.name@example.com",
			roles:    []string{"user"},
		},
		{
			name:     "unicode username",
			userID:   "user-5",
			username: "کاربر",
			roles:    []string{"admin"},
		},
	}

	config := JWTConfig{
		SecretKey:  "test-secret-key",
		Expiration: 1 * time.Hour,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.username, tt.roles, config)
			require.NoError(t, err)

			claims, err := ValidateToken(token, config)
			require.NoError(t, err)

			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.username, claims.Username)
			assert.Equal(t, tt.roles, claims.Roles)
		})
	}
}

func TestTokenUniqueness(t *testing.T) {
	config := JWTConfig{
		SecretKey:  "test-secret-key",
		Expiration: 1 * time.Hour,
	}

	// Generate multiple tokens for the same user
	token1, _ := GenerateToken("user-123", "testuser", []string{"user"}, config)
	time.Sleep(1 * time.Millisecond) // Small delay to ensure different timestamp
	token2, _ := GenerateToken("user-123", "testuser", []string{"user"}, config)

	// Tokens should be different due to different timestamps
	// (IssuedAt and potentially ExpiresAt will differ)
	assert.NotEqual(t, token1, token2)
}

func BenchmarkGenerateToken(b *testing.B) {
	config := JWTConfig{
		SecretKey:  "benchmark-secret-key",
		Expiration: 1 * time.Hour,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateToken("user-123", "testuser", []string{"admin"}, config)
	}
}

func BenchmarkValidateToken(b *testing.B) {
	config := JWTConfig{
		SecretKey:  "benchmark-secret-key",
		Expiration: 1 * time.Hour,
	}

	token, _ := GenerateToken("user-123", "testuser", []string{"admin"}, config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateToken(token, config)
	}
}
