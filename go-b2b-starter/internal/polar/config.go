package polar

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds configuration for the Polar client
type Config struct {
	// AccessToken is the Polar Organization Access Token (OAT)
	// Required for all API requests
	AccessToken string `mapstructure:"POLAR_ACCESS_TOKEN"`

	// BaseURL is the Polar API endpoint
	// Use "https://api.polar.sh" for production
	// Use "https://sandbox-api.polar.sh" for testing
	BaseURL string `mapstructure:"POLAR_BASE_URL"`

	// WebhookSecret is the secret used to verify webhook signatures
	// Get this from Polar Dashboard → Settings → Webhooks
	WebhookSecret string `mapstructure:"WEBHOOK_SECRET"`

	// Debug enables debug logging
	Debug bool `mapstructure:"POLAR_DEBUG"`
}

// LoadConfig reads configuration from file or environment variables
func LoadConfig() (Config, error) {
	var cfg Config

	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("POLAR_BASE_URL", "https://api.polar.sh")
	viper.SetDefault("POLAR_DEBUG", false)

	// Best-effort: ignore missing file, allow env-only usage
	if err := viper.ReadInConfig(); err == nil {
		_ = err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unable to decode polar config: %w", err)
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.AccessToken == "" {
		return fmt.Errorf("polar access token is required (POLAR_ACCESS_TOKEN)")
	}

	if c.BaseURL == "" {
		return fmt.Errorf("polar base URL is required (POLAR_BASE_URL)")
	}

	// WebhookSecret is optional - only needed for webhook verification
	// If not provided, webhook signature verification will be skipped (with warning)

	return nil
}

// DefaultConfig returns a configuration with sane defaults for production
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "https://api.polar.sh",
		Debug:   false,
	}
}

// SandboxConfig returns a configuration with defaults for sandbox environment
func SandboxConfig() *Config {
	return &Config{
		BaseURL: "https://sandbox-api.polar.sh",
		Debug:   true,
	}
}
