package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig تنظیمات JWT
type JWTConfig struct {
	SecretKey     string
	Expiration    time.Duration
	Issuer        string
	SkipPaths     []string // Paths that don't require authentication
	AllowedRoles  []string // Empty = all roles allowed
}

// Claims JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// JWTMiddleware میدلور JWT
func JWTMiddleware(config JWTConfig) fiber.Handler {
	// Default values
	if config.Expiration == 0 {
		config.Expiration = 24 * time.Hour
	}
	if config.Issuer == "" {
		config.Issuer = "edgeflow"
	}
	if config.SecretKey == "" {
		config.SecretKey = "edgeflow-secret-key-change-in-production"
	}

	return func(c *fiber.Ctx) error {
		// Check if path should skip authentication
		path := c.Path()
		for _, skipPath := range config.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return c.Next()
			}
		}

		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.SecretKey), nil
		})

		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token: " + err.Error(),
			})
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Check roles if required
		if len(config.AllowedRoles) > 0 {
			hasRole := false
			for _, allowedRole := range config.AllowedRoles {
				for _, userRole := range claims.Roles {
					if userRole == allowedRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}
			if !hasRole {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Insufficient permissions",
				})
			}
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("roles", claims.Roles)

		return c.Next()
	}
}

// GenerateToken تولید JWT token
func GenerateToken(userID, username string, roles []string, config JWTConfig) (string, error) {
	// Default values
	if config.Expiration == 0 {
		config.Expiration = 24 * time.Hour
	}
	if config.Issuer == "" {
		config.Issuer = "edgeflow"
	}
	if config.SecretKey == "" {
		config.SecretKey = "edgeflow-secret-key-change-in-production"
	}

	// Create claims
	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(config.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    config.Issuer,
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(config.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken اعتبارسنجی token
func ValidateToken(tokenString string, config JWTConfig) (*Claims, error) {
	if config.SecretKey == "" {
		config.SecretKey = "edgeflow-secret-key-change-in-production"
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
