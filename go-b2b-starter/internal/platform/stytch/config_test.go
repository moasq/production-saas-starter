package stytch

import "testing"

func TestEnvironmentOnlyConfiguration(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("STYTCH_PROJECT_ID", "project-test-example")
	t.Setenv("STYTCH_SECRET", "secret-example")
	t.Setenv("STYTCH_ENV", "test")
	t.Setenv("STYTCH_DISABLE_SESSION_VERIFICATION", "false")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Configured() || cfg.ProjectID != "project-test-example" || cfg.BaseURL != "https://test.stytch.com" {
		t.Fatalf("environment configuration was not loaded correctly")
	}
	t.Setenv("STYTCH_ENV", "live")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("mixed environment accepted")
	}
	t.Setenv("STYTCH_ENV", "test")
	t.Setenv("STYTCH_DISABLE_SESSION_VERIFICATION", "true")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("verification bypass accepted")
	}
}
func TestMissingCredentialsEnterSetupMode(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("STYTCH_PROJECT_ID", "")
	t.Setenv("STYTCH_SECRET", "")
	t.Setenv("STYTCH_ENV", "test")
	t.Setenv("STYTCH_DISABLE_SESSION_VERIFICATION", "false")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Configured() {
		t.Fatal("empty provider configured")
	}
	t.Setenv("STYTCH_PROJECT_ID", "project-test-example")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("partial credentials accepted")
	}
}
