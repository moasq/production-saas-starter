package config

import (
	"strings"
	"testing"
)

func TestProductionBehindTLSProxyWithoutEnvFile(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ENV", "PROD")
	t.Setenv("AUTH_INTERNAL_SECRET", "9a429cbf2dd80e61423bd94aa7bfe501")
	t.Setenv("TLS_TERMINATED_BY_PROXY", "true")
	t.Setenv("ALLOWED_ORIGINS", "https://app.example")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.IsProd() || len(cfg.AllowedOrigins) != 1 {
		t.Fatal("environment not loaded")
	}
	t.Setenv("TLS_TERMINATED_BY_PROXY", "false")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("production without TLS accepted")
	}
}

func TestEnvironmentMustBeExplicit(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, mode := range []string{"", "prod", "production", "staging", "PROD "} {
		t.Setenv("ENV", mode)
		if _, err := LoadConfig(); err == nil || !strings.Contains(err.Error(), "ENV") {
			t.Fatalf("accepted invalid mode %q: %v", mode, err)
		}
	}
	t.Setenv("ENV", "DEV")
	if cfg, err := LoadConfig(); err != nil || cfg.IsProd() {
		t.Fatalf("explicit local mode failed: %v", err)
	}
}

func TestProductionRejectsUnsafeOriginsAndBridgeSecrets(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ENV", "PROD")
	t.Setenv("TLS_TERMINATED_BY_PROXY", "true")
	t.Setenv("ALLOWED_ORIGINS", "https://app.example.com")
	for _, value := range []string{"", "short", strings.Repeat("a", 40), strings.Repeat("é", 40), "change-me-before-production-12345678", "test-only-internal-secret-000000000000", "placeholder-secret-000000000000000", " 9a429cbf2dd80e61423bd94aa7bfe501"} {
		t.Setenv("AUTH_INTERNAL_SECRET", value)
		_, err := LoadConfig()
		if err == nil || !strings.Contains(err.Error(), "AUTH_INTERNAL_SECRET") {
			t.Fatal("unsafe bridge secret accepted")
		}
		if value != "" && strings.Contains(err.Error(), value) {
			t.Fatal("secret leaked in error")
		}
	}
	t.Setenv("AUTH_INTERNAL_SECRET", "9a429cbf2dd80e61423bd94aa7bfe501")
	for _, origin := range []string{"http://app.example.com", "https://user:password@app.example.com", "https://app.example.com/path", "https://app.example.com?x=1", "https://app.example.com#x", "https://app.example.com?", "https://app.example.com#", "https://app.example.com:99999", "*"} {
		t.Setenv("ALLOWED_ORIGINS", origin)
		if _, err := LoadConfig(); err == nil {
			t.Fatalf("accepted unsafe production origin %q", origin)
		}
	}
}
