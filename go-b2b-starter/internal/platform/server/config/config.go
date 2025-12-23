package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Environment string

const (
	DEV  Environment = "DEV"
	PROD Environment = "PROD"
)

type Config struct {
	// Environment (cannot be disabled in production)
	Env Environment `mapstructure:"ENV"`

	// Server settings
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`

	// Security settings (cannot be disabled in production)
	EnableTLS   bool   `mapstructure:"ENABLE_TLS"`    // Must be true in production
	TLSCertPath string `mapstructure:"TLS_CERT_PATH"` // Required in production
	TLSKeyPath  string `mapstructure:"TLS_KEY_PATH"`  // Required in production

	// Rate limiting (cannot be disabled in production)
	RateLimitPerSecond int `mapstructure:"RATE_LIMIT_PER_SECOND"`

	// CORS settings (more restrictive in production)
	AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`

	// Logging (always enabled in production)
	LogLevel string `mapstructure:"LOG_LEVEL"`

	// Optional security features
	TrustedProxies []string `mapstructure:"TRUSTED_PROXIES"`
	MaxRequestSize int      `mapstructure:"MAX_REQUEST_SIZE"`

	// IP Protection Settings
	IPWhitelist       []string `mapstructure:"IP_WHITELIST"`
	IPBlacklist       []string `mapstructure:"IP_BLACKLIST"`
	MaxFailedAttempts int      `mapstructure:"MAX_FAILED_ATTEMPTS"`
	BlockDuration     string   `mapstructure:"BLOCK_DURATION"`

	// Request Sanitization
	DisableXSS           bool `mapstructure:"DISABLE_XSS"`
	DisableSQLInjection  bool `mapstructure:"DISABLE_SQL_INJECTION"`
	DisablePathTraversal bool `mapstructure:"DISABLE_PATH_TRAVERSAL"`

	// Security Logging
	SecurityLogPath  string `mapstructure:"SECURITY_LOG_PATH"`
	LogRetentionDays int    `mapstructure:"LOG_RETENTION_DAYS"`

	// Processing Settings
	ExtractionTimeoutSeconds int `mapstructure:"EXTRACTION_TIMEOUT_SECONDS"`
	
	// Duplicate Detection Settings
	DuplicateSimilarityThreshold float64 `mapstructure:"DUPLICATE_SIMILARITY_THRESHOLD"`
	DuplicateSearchLimit         int32   `mapstructure:"DUPLICATE_SEARCH_LIMIT"`
}

// SanitizationConfig represents security sanitization settings
type SanitizationConfig struct {
	DisableXSS           bool
	DisableSQLInjection  bool
	DisablePathTraversal bool
}

// GetSanitizationConfig returns sanitization configuration
func (c *Config) GetSanitizationConfig() SanitizationConfig {
	return SanitizationConfig{
		DisableXSS:           c.DisableXSS,
		DisableSQLInjection:  c.DisableSQLInjection,
		DisablePathTraversal: c.DisablePathTraversal,
	}
}

func (c *Config) IsProd() bool {
	return c.Env == PROD
}

// LoadConfig reads configuration from environment variables or .env files.
func LoadConfig() (*Config, error) {
	var cfg *Config

	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("ENV", "DEV")
	viper.SetDefault("SERVER_ADDRESS", ":8080")
	viper.SetDefault("RATE_LIMIT_PER_SECOND", 100)
	viper.SetDefault("MAX_REQUEST_SIZE", 1024*1024*10) // 10MB
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("MAX_FAILED_ATTEMPTS", 5)
	viper.SetDefault("BLOCK_DURATION", "15m")
	viper.SetDefault("DISABLE_XSS", false)
	viper.SetDefault("DISABLE_SQL_INJECTION", false)
	viper.SetDefault("DISABLE_PATH_TRAVERSAL", false)
	viper.SetDefault("SECURITY_LOG_PATH", "logs/security.log")
	viper.SetDefault("LOG_RETENTION_DAYS", 30)
	viper.SetDefault("EXTRACTION_TIMEOUT_SECONDS", 60)
	viper.SetDefault("DUPLICATE_SIMILARITY_THRESHOLD", 0.85)
	viper.SetDefault("DUPLICATE_SEARCH_LIMIT", 10)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Validate production configuration
	if cfg.Env == PROD {
		if err := validateProductionConfig(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func validateProductionConfig(cfg *Config) error {
	var errors []string

	// TLS must be enabled in production
	if !cfg.EnableTLS {
		errors = append(errors, "TLS must be enabled in production")
	}

	// TLS certificates must be provided in production
	if cfg.EnableTLS {
		if cfg.TLSCertPath == "" {
			errors = append(errors, "TLS certificate path must be provided in production")
		}
		if cfg.TLSKeyPath == "" {
			errors = append(errors, "TLS key path must be provided in production")
		}
	}

	// Allowed origins must be set in production
	if len(cfg.AllowedOrigins) == 0 {
		errors = append(errors, "Allowed origins must be set in production")
	}

	// Rate limiting must be reasonable in production
	if cfg.RateLimitPerSecond > 1000 {
		errors = append(errors, "Rate limit per second cannot exceed 1000 in production")
	}

	if len(errors) > 0 {
		return fmt.Errorf("invalid production configuration: %s", strings.Join(errors, "; "))
	}

	// if cfg.DisableXSS || cfg.DisableSQLInjection || cfg.DisablePathTraversal {
	// 	errors = append(errors, "Security sanitization cannot be disabled in production")
	// }

	// if cfg.MaxFailedAttempts < 3 {
	// 	errors = append(errors, "MaxFailedAttempts must be at least 3 in production")
	// }

	// if cfg.SecurityLogPath == "" {
	// 	errors = append(errors, "SecurityLogPath must be set in production")
	// }

	return nil
}
