package polar

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config deliberately separates provider environment from the application's build mode.
type Config struct {
	Enabled     bool
	ProductID   string
	AccessToken string
	BaseURL     string
}

func LoadConfig() (Config, error) {
	raw := strings.TrimSpace(os.Getenv("BILLING_ENABLED"))
	enabled := false
	if raw != "" {
		var err error
		enabled, err = strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("BILLING_ENABLED must be true or false")
		}
	}
	cfg := Config{Enabled: enabled, AccessToken: strings.TrimSpace(os.Getenv("POLAR_ACCESS_TOKEN")), ProductID: strings.TrimSpace(os.Getenv("POLAR_PRODUCT_ID"))}
	environment := strings.TrimSpace(os.Getenv("POLAR_ENVIRONMENT"))
	switch environment {
	case "", "sandbox":
		cfg.BaseURL = "https://sandbox-api.polar.sh"
	case "production":
		cfg.BaseURL = "https://api.polar.sh"
	default:
		return cfg, fmt.Errorf("POLAR_ENVIRONMENT must be sandbox or production")
	}
	return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
	if c.Enabled && c.ProductID == "" {
		return fmt.Errorf("POLAR_PRODUCT_ID is required when BILLING_ENABLED=true")
	}
	if c.Enabled && c.AccessToken == "" {
		return fmt.Errorf("POLAR_ACCESS_TOKEN is required when BILLING_ENABLED=true")
	}
	return nil
}
