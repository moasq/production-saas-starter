package config

import "testing"

func TestProductionBehindTLSProxyWithoutEnvFile(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("ENV", "PROD")
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
