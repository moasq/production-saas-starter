package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	R2 R2Config
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set environment variable overrides
	viper.SetEnvPrefix("APCASH")
	viper.AutomaticEnv()

	// Set default values for R2
	viper.SetDefault("r2.region", "auto")
	viper.SetDefault("r2.bucketName", "invoices")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Bind environment variables to viper keys for R2
	viper.BindEnv("r2.accountID", "R2_ACCOUNT_ID")
	viper.BindEnv("r2.accessKeyID", "R2_ACCESS_KEY_ID")
	viper.BindEnv("r2.secretAccessKey", "R2_SECRET_ACCESS_KEY")
	viper.BindEnv("r2.bucketName", "R2_BUCKET")
	viper.BindEnv("r2.region", "R2_REGION")

	config := &Config{
		R2: R2Config{
			AccountID:       viper.GetString("r2.accountID"),
			AccessKeyID:     viper.GetString("r2.accessKeyID"),
			SecretAccessKey: viper.GetString("r2.secretAccessKey"),
			BucketName:      viper.GetString("r2.bucketName"),
			Region:          viper.GetString("r2.region"),
		},
	}

	return config, nil
}