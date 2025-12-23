package stytch

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Environment constants supported by the Stytch B2B API.
const (
	EnvTest = "test"
	EnvLive = "live"
)

// Config captures the runtime configuration for Stytch authentication.
//
// All configuration values can be set via environment variables with the
// STYTCH_ prefix (e.g., STYTCH_PROJECT_ID, STYTCH_SECRET).
type Config struct {
	// ProjectID is the Stytch project identifier (required)
	ProjectID string `mapstructure:"STYTCH_PROJECT_ID"`

	// Secret is the Stytch API secret (required)
	Secret string `mapstructure:"STYTCH_SECRET"`

	// Env is the Stytch environment: "test" or "live"
	Env string `mapstructure:"STYTCH_ENV"`

	// BaseURL is the Stytch API base URL (derived from Env if not set)
	BaseURL string `mapstructure:"STYTCH_BASE_URL"`

	// CustomDomain is an optional custom domain for Stytch
	CustomDomain string `mapstructure:"STYTCH_CUSTOM_DOMAIN"`

	// JWKSURL is the JWKS endpoint URL (derived from BaseURL if not set)
	JWKSURL string `mapstructure:"STYTCH_JWKS_URL"`

	// SessionDurationMinutes is how long sessions should last
	SessionDurationMinutes int32 `mapstructure:"STYTCH_SESSION_DURATION_MINUTES"`

	// DisableSessionVerification disables JWT signature verification (testing only!)
	DisableSessionVerification bool `mapstructure:"STYTCH_DISABLE_SESSION_VERIFICATION"`

	// OwnerRoleSlug is the role slug for organization owners
	OwnerRoleSlug string `mapstructure:"STYTCH_OWNER_ROLE_SLUG"`

	// InviteRedirectURL is where to redirect after invitation acceptance
	InviteRedirectURL string `mapstructure:"STYTCH_INVITE_REDIRECT_URL"`

	// LoginRedirectURL is where to redirect after login
	LoginRedirectURL string `mapstructure:"STYTCH_LOGIN_REDIRECT_URL"`

	// APITimeout is the timeout for Stytch API calls
	APITimeout time.Duration `mapstructure:"STYTCH_API_TIMEOUT"`
}

// LoadConfig loads the Stytch configuration from environment variables and app.env file.
//
// Configuration priority:
//  1. Environment variables (highest)
//  2. app.env file
//  3. Default values (lowest)
func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("app")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AutomaticEnv()

	// Set defaults
	v.SetDefault("STYTCH_ENV", EnvTest)
	v.SetDefault("STYTCH_SESSION_DURATION_MINUTES", 1440) // 24 hours
	v.SetDefault("STYTCH_API_TIMEOUT", "15s")
	v.SetDefault("STYTCH_DISABLE_SESSION_VERIFICATION", false)

	// Try to read config file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode stytch config: %w", err)
	}

	// Normalize environment
	cfg.Env = strings.ToLower(strings.TrimSpace(cfg.Env))
	if cfg.Env == "" {
		cfg.Env = EnvTest
	}

	// Validate required fields
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("stytch configuration invalid: STYTCH_PROJECT_ID is required")
	}
	if cfg.Secret == "" {
		return nil, fmt.Errorf("stytch configuration invalid: STYTCH_SECRET is required")
	}

	// Normalize timeout
	if cfg.APITimeout <= 0 {
		cfg.APITimeout = 15 * time.Second
	}

	// Derive base URL if not set
	if cfg.BaseURL == "" {
		switch cfg.Env {
		case EnvLive:
			cfg.BaseURL = "https://api.stytch.com"
		default:
			cfg.BaseURL = "https://test.stytch.com"
		}
	}

	// Custom domain overrides base URL
	if cfg.CustomDomain != "" {
		cfg.BaseURL = fmt.Sprintf("https://%s", strings.TrimSuffix(cfg.CustomDomain, "/"))
	}

	// Derive JWKS URL if not set
	if cfg.JWKSURL == "" {
		if cfg.CustomDomain != "" {
			cfg.JWKSURL = fmt.Sprintf("https://%s/.well-known/jwks.json", strings.TrimSuffix(cfg.CustomDomain, "/"))
		} else {
			cfg.JWKSURL = fmt.Sprintf("%s/v1/b2b/sessions/jwks/%s", strings.TrimSuffix(cfg.BaseURL, "/"), cfg.ProjectID)
		}
	}

	return &cfg, nil
}

// Validate checks that the configuration has all required fields.
func (c *Config) Validate() error {
	if c.ProjectID == "" {
		return fmt.Errorf("stytch configuration invalid: ProjectID is required")
	}
	if c.Secret == "" {
		return fmt.Errorf("stytch configuration invalid: Secret is required")
	}
	return nil
}

// This allows gradual migration from the old config type.
func NewConfigFromExisting(projectID, secret, env, baseURL, jwksURL string, sessionDurationMinutes int32, disableVerification bool, apiTimeout time.Duration) *Config {
	return &Config{
		ProjectID:                  projectID,
		Secret:                     secret,
		Env:                        env,
		BaseURL:                    baseURL,
		JWKSURL:                    jwksURL,
		SessionDurationMinutes:     sessionDurationMinutes,
		DisableSessionVerification: disableVerification,
		APITimeout:                 apiTimeout,
	}
}
