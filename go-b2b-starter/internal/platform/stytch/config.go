package stytch

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Environment constants supported by the Stytch B2B API.
const (
	EnvTest = "test"
	EnvLive = "live"
)

// Config captures the runtime knobs required to communicate with Stytch.
type Config struct {
	ProjectID                  string        `mapstructure:"STYTCH_PROJECT_ID"`
	Secret                     string        `mapstructure:"STYTCH_SECRET"`
	Env                        string        `mapstructure:"STYTCH_ENV"`
	BaseURL                    string        `mapstructure:"STYTCH_BASE_URL"`
	CustomDomain               string        `mapstructure:"STYTCH_CUSTOM_DOMAIN"`
	JWKSURL                    string        `mapstructure:"STYTCH_JWKS_URL"`
	SessionDurationMinutes     int32         `mapstructure:"STYTCH_SESSION_DURATION_MINUTES"`
	DisableSessionVerification bool          `mapstructure:"STYTCH_DISABLE_SESSION_VERIFICATION"`
	OwnerRoleSlug              string        `mapstructure:"STYTCH_OWNER_ROLE_SLUG"`
	InviteRedirectURL          string        `mapstructure:"STYTCH_INVITE_REDIRECT_URL"`
	LoginRedirectURL           string        `mapstructure:"STYTCH_LOGIN_REDIRECT_URL"`
	APITimeout                 time.Duration `mapstructure:"STYTCH_API_TIMEOUT"`
}

// LoadConfig hydrates the Stytch configuration from app.env + process environment.
func LoadConfig() (*Config, error) {
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

	// Defaults mirror the Auth0 integration to minimize surprises.
	v.SetDefault("STYTCH_ENV", EnvTest)
	v.SetDefault("STYTCH_SESSION_DURATION_MINUTES", 1440) // 24 hours (previously 60 minutes)
	v.SetDefault("STYTCH_API_TIMEOUT", "15s")
	v.SetDefault("STYTCH_DISABLE_SESSION_VERIFICATION", false)

	// Best-effort: ignore missing file, allow env-only usage
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg *Config
	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("unable to decode stytch config: %w", err)
	}

	cfg.Env = strings.ToLower(strings.TrimSpace(cfg.Env))
	if cfg.Env == "" {
		cfg.Env = EnvTest
	}

	if cfg.Env != EnvTest && cfg.Env != EnvLive {
		return nil, fmt.Errorf("STYTCH_ENV must be test or live")
	}
	if cfg.DisableSessionVerification {
		return nil, fmt.Errorf("STYTCH_DISABLE_SESSION_VERIFICATION is no longer supported")
	}
	if (cfg.ProjectID == "") != (cfg.Secret == "") {
		return nil, fmt.Errorf("set both STYTCH_PROJECT_ID and STYTCH_SECRET")
	}
	if cfg.Configured() && ((cfg.Env == EnvTest && !strings.HasPrefix(cfg.ProjectID, "project-test-")) || (cfg.Env == EnvLive && !strings.HasPrefix(cfg.ProjectID, "project-live-"))) {
		return nil, fmt.Errorf("STYTCH_PROJECT_ID does not match STYTCH_ENV")
	}

	// Normalize timeout (viper unmarshals duration strings automatically).
	if cfg.APITimeout <= 0 {
		cfg.APITimeout = 15 * time.Second
	}

	// Derive base URL if none supplied.
	if cfg.BaseURL == "" {
		switch cfg.Env {
		case EnvLive:
			cfg.BaseURL = "https://api.stytch.com"
		default:
			cfg.BaseURL = "https://test.stytch.com"
		}
	}

	// Custom domain overrides base URL for API + JWKS.
	if cfg.CustomDomain != "" {
		cfg.BaseURL = fmt.Sprintf("https://%s", strings.TrimSuffix(cfg.CustomDomain, "/"))
	}

	// Derive JWKS URL if absent.
	if cfg.JWKSURL == "" {
		if cfg.CustomDomain != "" {
			cfg.JWKSURL = fmt.Sprintf("https://%s/.well-known/jwks.json", strings.TrimSuffix(cfg.CustomDomain, "/"))
		} else {
			cfg.JWKSURL = fmt.Sprintf("%s/v1/b2b/sessions/jwks/%s", strings.TrimSuffix(cfg.BaseURL, "/"), cfg.ProjectID)
		}
	}

	return cfg, nil
}

// Configured distinguishes setup mode from an enabled provider.
func (c *Config) Configured() bool {
	return c.ProjectID != "" && c.Secret != "" && !strings.Contains(strings.ToUpper(c.ProjectID+c.Secret), "REPLACE")
}
func (c *Config) Validate() error {
	if !c.Configured() {
		return fmt.Errorf("Stytch is not configured")
	}
	if c.DisableSessionVerification {
		return fmt.Errorf("session verification cannot be disabled")
	}
	return nil
}
