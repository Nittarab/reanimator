package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: localhost
  port: 5432
  database: test_db
  user: test_user
  password: test_pass
  ssl_mode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

github:
  api_url: https://api.github.com
  token: test_token
  workflow_name: test.yml

service_mappings:
  - service_name: test-service
    repository: org/test-repo
    branch: main

deduplication:
  time_window: 5m

concurrency:
  max_workflows_per_repo: 2

mcp_servers: []

custom_rules: []
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want 8080", cfg.Server.Port)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %s, want localhost", cfg.Database.Host)
	}

	if len(cfg.ServiceMappings) != 1 {
		t.Errorf("len(ServiceMappings) = %d, want 1", len(cfg.ServiceMappings))
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("TEST_DB_HOST", "envhost")
	os.Setenv("TEST_GITHUB_TOKEN", "env_token")
	defer os.Unsetenv("TEST_DB_HOST")
	defer os.Unsetenv("TEST_GITHUB_TOKEN")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: ${TEST_DB_HOST}
  port: 5432
  database: test_db
  user: test_user
  password: test_pass
  ssl_mode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

github:
  api_url: https://api.github.com
  token: ${TEST_GITHUB_TOKEN}
  workflow_name: test.yml

service_mappings: []
deduplication:
  time_window: 5m
concurrency:
  max_workflows_per_repo: 2
mcp_servers: []
custom_rules: []
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Database.Host != "envhost" {
		t.Errorf("Database.Host = %s, want envhost", cfg.Database.Host)
	}

	if cfg.GitHub.Token != "env_token" {
		t.Errorf("GitHub.Token = %s, want env_token", cfg.GitHub.Token)
	}
}

func TestValidateRule(t *testing.T) {
	tests := []struct {
		name    string
		rule    CustomRule
		wantErr bool
	}{
		{
			name: "valid rule with all fields",
			rule: CustomRule{
				Name:        "test-rule",
				Description: "Test rule",
				Enabled:     true,
				Conditions: RuleConditions{
					ServiceName: stringPtr("test-service"),
				},
				Actions: RuleActions{
					SetSeverity: stringPtr("high"),
				},
			},
			wantErr: false,
		},
		{
			name: "valid rule with error pattern",
			rule: CustomRule{
				Name:    "pattern-rule",
				Enabled: true,
				Conditions: RuleConditions{
					ErrorPattern: stringPtr(".*error.*"),
				},
				Actions: RuleActions{
					SetSeverity: stringPtr("critical"),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid rule - no name",
			rule: CustomRule{
				Conditions: RuleConditions{
					ServiceName: stringPtr("test"),
				},
				Actions: RuleActions{
					SetSeverity: stringPtr("high"),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid rule - bad regex",
			rule: CustomRule{
				Name: "bad-regex",
				Conditions: RuleConditions{
					ErrorPattern: stringPtr("[invalid"),
				},
				Actions: RuleActions{
					SetSeverity: stringPtr("high"),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid rule - bad severity in condition",
			rule: CustomRule{
				Name: "bad-severity",
				Conditions: RuleConditions{
					Severity: stringPtr("invalid"),
				},
				Actions: RuleActions{
					SetSeverity: stringPtr("high"),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid rule - bad severity in action",
			rule: CustomRule{
				Name: "bad-action-severity",
				Conditions: RuleConditions{
					ServiceName: stringPtr("test"),
				},
				Actions: RuleActions{
					SetSeverity: stringPtr("invalid"),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid rule - no conditions",
			rule: CustomRule{
				Name:       "no-conditions",
				Conditions: RuleConditions{},
				Actions: RuleActions{
					SetSeverity: stringPtr("high"),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid rule - no actions",
			rule: CustomRule{
				Name: "no-actions",
				Conditions: RuleConditions{
					ServiceName: stringPtr("test"),
				},
				Actions: RuleActions{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRule(&tt.rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Host: "localhost", Database: "test"},
				GitHub:   GitHubConfig{Token: "token"},
			},
			wantErr: false,
		},
		{
			name: "missing server port",
			config: Config{
				Database: DatabaseConfig{Host: "localhost", Database: "test"},
				GitHub:   GitHubConfig{Token: "token"},
			},
			wantErr: true,
		},
		{
			name: "missing database host",
			config: Config{
				Server: ServerConfig{Port: 8080},
				Database: DatabaseConfig{Database: "test"},
				GitHub:   GitHubConfig{Token: "token"},
			},
			wantErr: true,
		},
		{
			name: "missing github token",
			config: Config{
				Server:   ServerConfig{Port: 8080},
				Database: DatabaseConfig{Host: "localhost", Database: "test"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWatcher(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	initialConfig := `
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: localhost
  port: 5432
  database: test_db
  user: test_user
  password: test_pass
  ssl_mode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

github:
  api_url: https://api.github.com
  token: test_token
  workflow_name: test.yml

service_mappings:
  - service_name: test-service
    repository: org/test-repo
    branch: main

deduplication:
  time_window: 5m

concurrency:
  max_workflows_per_repo: 2

mcp_servers: []
custom_rules: []
`

	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	watcher, err := NewWatcher(configPath)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}
	defer watcher.Stop()

	// Check initial config
	cfg := watcher.Get()
	if cfg.Server.Port != 8080 {
		t.Errorf("initial Server.Port = %d, want 8080", cfg.Server.Port)
	}

	// Set up callback
	reloaded := make(chan bool, 1)
	watcher.OnReload(func(newCfg *Config) {
		reloaded <- true
	})

	// Start watcher in background
	go watcher.Start(100 * time.Millisecond)

	// Wait a bit to ensure watcher is running
	time.Sleep(150 * time.Millisecond)

	// Update config file
	updatedConfig := `
server:
  port: 9090
  read_timeout: 30s
  write_timeout: 30s

database:
  host: localhost
  port: 5432
  database: test_db
  user: test_user
  password: test_pass
  ssl_mode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

github:
  api_url: https://api.github.com
  token: test_token
  workflow_name: test.yml

service_mappings:
  - service_name: test-service
    repository: org/test-repo
    branch: main

deduplication:
  time_window: 5m

concurrency:
  max_workflows_per_repo: 2

mcp_servers: []
custom_rules: []
`

	if err := os.WriteFile(configPath, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("failed to update config file: %v", err)
	}

	// Wait for reload
	select {
	case <-reloaded:
		// Config was reloaded
	case <-time.After(1 * time.Second):
		t.Fatal("config was not reloaded within timeout")
	}

	// Check updated config
	cfg = watcher.Get()
	if cfg.Server.Port != 9090 {
		t.Errorf("updated Server.Port = %d, want 9090", cfg.Server.Port)
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

// Property-Based Tests

// **Feature: ai-sre-platform, Property 11: Configuration parsing validity**
// **Validates: Requirements 11.1, 11.2**
func TestProperty_ConfigurationParsingValidity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid service name (alphanumeric with hyphens)
	genServiceName := gen.Identifier().SuchThat(func(s string) bool { 
		return len(s) > 0 && len(s) < 50 
	})

	// Generator for valid repository (org/repo format)
	genRepository := gopter.CombineGens(
		gen.Identifier(),
		gen.Identifier(),
	).Map(func(vals []interface{}) string {
		return vals[0].(string) + "/" + vals[1].(string)
	})

	properties.Property("service mappings should be accessible after parsing", prop.ForAll(
		func(serviceName string, repository string) bool {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			yamlContent := fmt.Sprintf(`server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: localhost
  port: 5432
  database: test_db
  user: test_user
  password: test_pass
  ssl_mode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

github:
  api_url: https://api.github.com
  token: test_token
  workflow_name: test.yml

service_mappings:
  - service_name: %s
    repository: %s
    branch: main

deduplication:
  time_window: 5m

concurrency:
  max_workflows_per_repo: 2

mcp_servers: []
custom_rules: []
`, serviceName, repository)

			if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
				return false
			}

			cfg, err := Load(configPath)
			if err != nil {
				return false
			}

			// Verify service mapping is accessible
			if len(cfg.ServiceMappings) != 1 {
				return false
			}

			if cfg.ServiceMappings[0].ServiceName != serviceName {
				return false
			}

			if cfg.ServiceMappings[0].Repository != repository {
				return false
			}

			return true
		},
		genServiceName,
		genRepository,
	))

	properties.Property("MCP servers should be accessible after parsing", prop.ForAll(
		func(mcpName string, mcpType string) bool {
			// Skip empty strings
			if mcpName == "" || mcpType == "" {
				return true
			}

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")

			yamlContent := fmt.Sprintf(`server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: localhost
  port: 5432
  database: test_db
  user: test_user
  password: test_pass
  ssl_mode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

github:
  api_url: https://api.github.com
  token: test_token
  workflow_name: test.yml

service_mappings: []

deduplication:
  time_window: 5m

concurrency:
  max_workflows_per_repo: 2

mcp_servers:
  - name: %s
    type: %s
    config:
      key: value

custom_rules: []
`, mcpName, mcpType)

			if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
				return false
			}

			cfg, err := Load(configPath)
			if err != nil {
				return false
			}

			// Verify MCP server is accessible
			if len(cfg.MCPServers) != 1 {
				return false
			}

			if cfg.MCPServers[0].Name != mcpName {
				return false
			}

			if cfg.MCPServers[0].Type != mcpType {
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// **Feature: ai-sre-platform, Property 17: Rule syntax validation**
// **Validates: Requirements 16.5**
func TestProperty_RuleSyntaxValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid regex patterns
	genValidRegex := gen.OneConstOf(
		".*error.*",
		"^test.*",
		"[a-z]+",
		"\\d{3}",
		"(foo|bar)",
	)

	// Generator for invalid regex patterns
	genInvalidRegex := gen.OneConstOf(
		"[invalid",
		"(unclosed",
		"*invalid",
		"(?P<invalid",
		"[z-a]",
	)

	// Generator for valid severities
	genValidSeverity := gen.OneConstOf("critical", "high", "medium", "low")

	// Generator for invalid severities
	genInvalidSeverity := gen.OneConstOf("urgent", "extreme", "minor", "trivial", "")

	properties.Property("rules with syntax errors should fail validation", prop.ForAll(
		func(invalidRegex string) bool {
			rule := CustomRule{
				Name: "test-rule",
				Conditions: RuleConditions{
					ErrorPattern: &invalidRegex,
				},
				Actions: RuleActions{
					SetSeverity: stringPtr("high"),
				},
			}

			err := ValidateRule(&rule)
			// Should return an error for invalid regex
			return err != nil
		},
		genInvalidRegex,
	))

	properties.Property("rules with valid regex should pass validation", prop.ForAll(
		func(validRegex string, severity string) bool {
			rule := CustomRule{
				Name: "test-rule",
				Conditions: RuleConditions{
					ErrorPattern: &validRegex,
				},
				Actions: RuleActions{
					SetSeverity: &severity,
				},
			}

			err := ValidateRule(&rule)
			// Should not return an error for valid regex and severity
			return err == nil
		},
		genValidRegex,
		genValidSeverity,
	))

	properties.Property("rules with invalid severity should fail validation", prop.ForAll(
		func(invalidSeverity string) bool {
			rule := CustomRule{
				Name: "test-rule",
				Conditions: RuleConditions{
					ServiceName: stringPtr("test-service"),
				},
				Actions: RuleActions{
					SetSeverity: &invalidSeverity,
				},
			}

			err := ValidateRule(&rule)
			// Should return an error for invalid severity
			return err != nil
		},
		genInvalidSeverity,
	))

	properties.Property("rules without name should fail validation", prop.ForAll(
		func(serviceName string) bool {
			rule := CustomRule{
				Name: "", // Empty name
				Conditions: RuleConditions{
					ServiceName: &serviceName,
				},
				Actions: RuleActions{
					SetSeverity: stringPtr("high"),
				},
			}

			err := ValidateRule(&rule)
			// Should return an error for missing name
			return err != nil
		},
		gen.AlphaString(),
	))

	properties.Property("rules without conditions should fail validation", prop.ForAll(
		func(name string) bool {
			if name == "" {
				return true // Skip empty names
			}

			rule := CustomRule{
				Name:       name,
				Conditions: RuleConditions{}, // No conditions
				Actions: RuleActions{
					SetSeverity: stringPtr("high"),
				},
			}

			err := ValidateRule(&rule)
			// Should return an error for missing conditions
			return err != nil
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.Property("rules without actions should fail validation", prop.ForAll(
		func(name string, serviceName string) bool {
			if name == "" {
				return true // Skip empty names
			}

			rule := CustomRule{
				Name: name,
				Conditions: RuleConditions{
					ServiceName: &serviceName,
				},
				Actions: RuleActions{}, // No actions
			}

			err := ValidateRule(&rule)
			// Should return an error for missing actions
			return err != nil
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString(),
	))

	properties.Property("valid regex patterns should compile successfully", prop.ForAll(
		func(pattern string) bool {
			rule := CustomRule{
				Name: "test-rule",
				Conditions: RuleConditions{
					ErrorPattern: &pattern,
				},
				Actions: RuleActions{
					SetSeverity: stringPtr("high"),
				},
			}

			err := ValidateRule(&rule)
			
			// Check if pattern is valid by trying to compile it
			_, compileErr := regexp.Compile(pattern)
			
			// If pattern compiles, validation should pass
			// If pattern doesn't compile, validation should fail
			if compileErr == nil {
				return err == nil
			}
			return err != nil
		},
		genValidRegex,
	))

	properties.TestingRun(t)
}
