// Package local provides a self-hosted JWT authentication adapter.
package local

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config holds the configuration for the local auth adapter.
type Config struct {
	// JWT settings
	Issuer                 string
	Audience               string
	AccessTokenDuration    time.Duration
	RefreshTokenDuration   time.Duration
	PrivateKey             *rsa.PrivateKey
	PublicKey              *rsa.PublicKey

	// Password policy
	PasswordMinLength       int
	PasswordRequireUpper    bool
	PasswordRequireLower    bool
	PasswordRequireDigit    bool
	PasswordRequireSpecial  bool

	// Security settings
	MaxFailedAttempts int
	LockoutDuration   time.Duration
}

// LoadConfig loads the local auth configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Issuer:                 getEnv("JWT_ISSUER", "go-b2b-starter"),
		Audience:               getEnv("JWT_AUDIENCE", "go-b2b-api"),
		AccessTokenDuration:    parseDuration("JWT_ACCESS_TOKEN_DURATION", 15*time.Minute),
		RefreshTokenDuration:   parseDuration("JWT_REFRESH_TOKEN_DURATION", 7*24*time.Hour),
		PasswordMinLength:      parseInt("PASSWORD_MIN_LENGTH", 8),
		PasswordRequireUpper:   parseBool("PASSWORD_REQUIRE_UPPERCASE", true),
		PasswordRequireLower:   parseBool("PASSWORD_REQUIRE_LOWERCASE", true),
		PasswordRequireDigit:   parseBool("PASSWORD_REQUIRE_DIGIT", true),
		PasswordRequireSpecial: parseBool("PASSWORD_REQUIRE_SPECIAL", false),
		MaxFailedAttempts:      parseInt("AUTH_MAX_FAILED_ATTEMPTS", 5),
		LockoutDuration:        parseDuration("AUTH_LOCKOUT_DURATION", 15*time.Minute),
	}

	// Load RSA keys
	privateKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
	publicKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")

	if privateKeyPath != "" && publicKeyPath != "" {
		// Load keys from files
		privateKeyData, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}
		cfg.PrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		publicKeyData, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key: %w", err)
		}
		cfg.PublicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	} else {
		// Generate ephemeral keys for development
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
		cfg.PrivateKey = privateKey
		cfg.PublicKey = &privateKey.PublicKey
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}

func parseInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func parseBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
