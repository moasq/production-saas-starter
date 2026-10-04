package config

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
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
	Env                Environment `mapstructure:"ENV"`
	AuthInternalSecret string      `mapstructure:"AUTH_INTERNAL_SECRET"`

	// Server settings
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`

	// Security settings (cannot be disabled in production)
	TLSTerminatedByProxy bool   `mapstructure:"TLS_TERMINATED_BY_PROXY"`
	EnableTLS            bool   `mapstructure:"ENABLE_TLS"`    // Must be true in production
	TLSCertPath          string `mapstructure:"TLS_CERT_PATH"` // Required in production
	TLSKeyPath           string `mapstructure:"TLS_KEY_PATH"`  // Required in production

	// Rate limiting (cannot be disabled in production)
	RateLimitPerSecond int `mapstructure:"RATE_LIMIT_PER_SECOND"`

	// CORS settings (more restrictive in production)
	AllowedOrigins []string `mapstructure:"ALLOWED_ORIGINS"`

	// Optional security features
	TrustedProxies []string `mapstructure:"TRUSTED_PROXIES"`
	MaxRequestSize int      `mapstructure:"MAX_REQUEST_SIZE"`
}

func (c *Config) IsProd() bool {
	return c.Env == PROD
}

// LoadConfig reads configuration from environment variables or .env files.
func LoadConfig() (*Config, error) {
	var cfg *Config
	v := viper.New()

	v.SetConfigName("app")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	typ := reflect.TypeOf(Config{})
	for i := 0; i < typ.NumField(); i++ {
		if err := v.BindEnv(typ.Field(i).Tag.Get("mapstructure")); err != nil {
			return nil, err
		}
	}

	// Set default values
	v.SetDefault("SERVER_ADDRESS", ":8080")
	v.SetDefault("RATE_LIMIT_PER_SECOND", 100)
	v.SetDefault("MAX_REQUEST_SIZE", 1024*1024*10) // 10MB

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.Env != DEV && cfg.Env != PROD {
		return nil, fmt.Errorf("ENV must be explicitly set to DEV or PROD")
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

	// TLS must be enabled or terminated by a trusted reverse proxy in production
	if !cfg.EnableTLS && !cfg.TLSTerminatedByProxy {
		errors = append(errors, "TLS must be enabled or terminated by a trusted reverse proxy in production")
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

	if !validProductionSecret(cfg.AuthInternalSecret) {
		errors = append(errors, "AUTH_INTERNAL_SECRET must contain at least 32 characters and must not be a placeholder or repeated character")
	}

	// Allowed origins must be set in production
	if len(cfg.AllowedOrigins) == 0 {
		errors = append(errors, "Allowed origins must be set in production")
	}

	for _, origin := range cfg.AllowedOrigins {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || (u.Path != "" && u.Path != "/") || strings.ContainsAny(origin, "?#") || strings.TrimSpace(origin) != origin {
			errors = append(errors, "Allowed origins must be HTTPS origins without credentials, paths, query strings or fragments in production")
			break
		}

		if port := u.Port(); port != "" {
			number, err := strconv.Atoi(port)
			if err != nil || number < 1 || number > 65535 {
				errors = append(errors, "Allowed origins must use valid TCP ports")
				break
			}
		}
	}

	// Rate limiting must be reasonable in production
	if cfg.RateLimitPerSecond < 1 || cfg.RateLimitPerSecond > 1000 {
		errors = append(errors, "Rate limit per second cannot exceed 1000 in production")
	}

	if len(errors) > 0 {
		return fmt.Errorf("invalid production configuration: %s", strings.Join(errors, "; "))
	}

	return nil
}

// This catches common copied examples, not the entropy of an arbitrary secret.
var placeholderSecret = regexp.MustCompile(`(?i)(change[-_]?me|replace[-_]?me|placeholder|your[-_]secret|test[-_]only)`)

func validProductionSecret(value string) bool {
	if len(value) < 32 || strings.TrimSpace(value) != value || placeholderSecret.MatchString(value) {
		return false
	}
	var first rune
	for index, character := range value {
		if index == 0 {
			first = character
		} else if character != first {
			return true
		}
	}
	return false
}
