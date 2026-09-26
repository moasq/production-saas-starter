package polar

import (
	"encoding/json"
	"os"
	"testing"
)

func TestSharedBillingEnvironmentContract(t *testing.T) {
	data, err := os.ReadFile("testdata/config.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name        string
		Env         map[string]string
		Error       string
		Enabled     bool
		BaseURL     string
		AccessToken string
		ProductID   string
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			for _, key := range []string{"BILLING_ENABLED", "POLAR_ENVIRONMENT", "POLAR_ACCESS_TOKEN", "POLAR_PRODUCT_ID", "NODE_ENV", "ENV"} {
				t.Setenv(key, tc.Env[key])
			}
			cfg, err := LoadConfig()
			if tc.Error != "" {
				if err == nil || err.Error() != tc.Error {
					t.Fatalf("expected configuration error %q, got %v", tc.Error, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.Enabled != tc.Enabled || cfg.BaseURL != tc.BaseURL || cfg.AccessToken != tc.AccessToken || cfg.ProductID != tc.ProductID {
				t.Fatal("configuration differs from shared contract")
			}
		})
	}
}
