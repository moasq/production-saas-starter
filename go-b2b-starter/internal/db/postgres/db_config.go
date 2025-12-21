package postgres

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Host         string `mapstructure:"POSTGRES_HOST"`
	Port         string `mapstructure:"POSTGRES_PORT"`
	User         string `mapstructure:"POSTGRES_USER"`
	Password     string `mapstructure:"POSTGRES_PASSWORD"`
	DBName       string `mapstructure:"POSTGRES_DB"`
	SSLMode      string `mapstructure:"DB_SSL_MODE"`
	MigrationURL string `mapstructure:"MIGRATION_URL"`
	SeedURL      string `mapstructure:"SEED_URL"`

	// Connection pool settings
	MaxConns          int           `mapstructure:"DB_MAX_CONNS"`
	MinConns          int           `mapstructure:"DB_MIN_CONNS"`
	ConnLifetime      time.Duration `mapstructure:"DB_CONN_LIFETIME"`
	ConnIdleTime      time.Duration `mapstructure:"DB_CONN_IDLE_TIME"`
	HealthCheckPeriod time.Duration `mapstructure:"DB_HEALTH_CHECK_PERIOD"`
}

// ConnectionString returns a formatted PostgreSQL connection string
func (c Config) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s application_name=nomadezy_api",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (Config, error) {
	var cfg Config

	viper.SetConfigName("app") // Name of the config file (without extension)
	viper.SetConfigType("env") // Set the type of the configuration files - .env
	viper.AddConfigPath(".")   // Optionally look for config in the working directory
	viper.AutomaticEnv()

	// Set default values, these are overridden if values are present in config or environment variables
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("DB_SSL_MODE", "disable") // Use "require" in production, "disable" for local dev
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_USER", "user")
	viper.SetDefault("POSTGRES_PASSWORD", "password")
	viper.SetDefault("POSTGRES_DB", "mydatabase")

	// Connection pool defaults
	viper.SetDefault("DB_MAX_CONNS", 20)
	viper.SetDefault("DB_MIN_CONNS", 5)
	viper.SetDefault("DB_CONN_LIFETIME", "1h")
	viper.SetDefault("DB_CONN_IDLE_TIME", "30m")
	viper.SetDefault("DB_HEALTH_CHECK_PERIOD", "1m")

	viper.SetDefault("MIGRATION_URL", "/migrations")
	viper.SetDefault("SEED_URL", "/seed")

	if err := viper.ReadInConfig(); err == nil {
		_ = err // Placeholder statement to avoid empty branch error
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
