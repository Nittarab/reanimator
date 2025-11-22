package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadRootConfig tests loading the actual config.yaml from the repository root
func TestLoadRootConfig(t *testing.T) {
	// Set required environment variables for validation
	os.Setenv("GITHUB_TOKEN", "test_token_for_validation")
	os.Setenv("DATABASE_HOST", "localhost")
	os.Setenv("DATABASE_NAME", "test_db")
	defer func() {
		os.Unsetenv("GITHUB_TOKEN")
		os.Unsetenv("DATABASE_HOST")
		os.Unsetenv("DATABASE_NAME")
	}()

	// Find the config.yaml in the repository root
	configPath := filepath.Join("..", "..", "..", "config.yaml")

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load root config.yaml: %v", err)
	}

	// Verify basic structure
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", cfg.Server.Port)
	}

	if len(cfg.ServiceMappings) == 0 {
		t.Error("Expected at least one service mapping")
	}

	// Verify custom rules are valid
	for i, rule := range cfg.CustomRules {
		if err := ValidateRule(&rule); err != nil {
			t.Errorf("Invalid custom rule at index %d (%s): %v", i, rule.Name, err)
		}
	}

	t.Logf("Successfully loaded config with:")
	t.Logf("  - %d service mappings", len(cfg.ServiceMappings))
	t.Logf("  - %d custom rules", len(cfg.CustomRules))
	t.Logf("  - %d MCP servers", len(cfg.MCPServers))
}
