package infra

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	MistralAPIKey string
	APIEndpoint   string
	TimeoutSec    int
}

func (c Config) Validate() error {
	if c.MistralAPIKey == "" {
		return fmt.Errorf("Mistral API key is required")
	}
	if c.APIEndpoint == "" {
		return fmt.Errorf("API endpoint is required")
	}
	return nil
}

func NewOCRConfig() Config {
	timeoutSec, _ := strconv.Atoi(getEnvOrDefault("OCR_TIMEOUT_SEC", "120"))

	return Config{
		MistralAPIKey: os.Getenv("MISTRAL_API_KEY"),
		APIEndpoint:   getEnvOrDefault("MISTRAL_OCR_ENDPOINT", "https://api.mistral.ai/v1/ocr"),
		TimeoutSec:    timeoutSec,
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}