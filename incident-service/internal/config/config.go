package config

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server          ServerConfig        `yaml:"server"`
	Database        DatabaseConfig      `yaml:"database"`
	Redis           RedisConfig         `yaml:"redis"`
	GitHub          GitHubConfig        `yaml:"github"`
	ServiceMappings []ServiceMapping    `yaml:"service_mappings"`
	Deduplication   DeduplicationConfig `yaml:"deduplication"`
	Concurrency     ConcurrencyConfig   `yaml:"concurrency"`
	MCPServers      []MCPServerConfig   `yaml:"mcp_servers"`
	CustomRules     []CustomRule        `yaml:"custom_rules"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// DatabaseConfig contains PostgreSQL connection settings
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

// RedisConfig contains Redis connection settings
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// GitHubConfig contains GitHub API settings
type GitHubConfig struct {
	APIURL       string `yaml:"api_url"`
	Token        string `yaml:"token"`
	WorkflowName string `yaml:"workflow_name"`
}

// DeduplicationConfig contains incident deduplication settings
type DeduplicationConfig struct {
	TimeWindow time.Duration `yaml:"time_window"`
}

// ConcurrencyConfig contains workflow concurrency settings
type ConcurrencyConfig struct {
	MaxWorkflowsPerRepo int `yaml:"max_workflows_per_repo"`
}

// ServiceMapping maps a service name to a repository
type ServiceMapping struct {
	ServiceName string `yaml:"service_name"`
	Repository  string `yaml:"repository"`
	Branch      string `yaml:"branch"`
}

// MCPServerConfig contains MCP server configuration
type MCPServerConfig struct {
	Name   string            `yaml:"name"`
	Type   string            `yaml:"type"`
	Config map[string]string `yaml:"config"`
}

// CustomRule represents a custom incident detection rule
type CustomRule struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Conditions  RuleConditions    `yaml:"conditions"`
	Actions     RuleActions       `yaml:"actions"`
	Enabled     bool              `yaml:"enabled"`
}

// RuleConditions defines the conditions that must be met for a rule to match
type RuleConditions struct {
	ServiceName  *string            `yaml:"service_name"`
	ErrorPattern *string            `yaml:"error_pattern"`
	Severity     *string            `yaml:"severity"`
	Provider     *string            `yaml:"provider"`
	Metadata     map[string]string  `yaml:"metadata"`
}

// RuleActions defines the actions to take when a rule matches
type RuleActions struct {
	SetSeverity     *string            `yaml:"set_severity"`
	AddMetadata     map[string]string  `yaml:"add_metadata"`
	SetRepository   *string            `yaml:"set_repository"`
	SkipRemediation bool               `yaml:"skip_remediation"`
}

// expandEnvWithDefaults expands environment variables with support for default values
// Supports both ${VAR} and ${VAR:-default} syntax
func expandEnvWithDefaults(s string) string {
	// Pattern matches ${VAR} or ${VAR:-default}
	pattern := regexp.MustCompile(`\$\{([^}:]+)(:-([^}]*))?\}`)
	
	return pattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name and default value
		parts := pattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}
		
		varName := parts[1]
		defaultValue := ""
		if len(parts) >= 4 {
			defaultValue = parts[3]
		}
		
		// Get environment variable value
		if value := os.Getenv(varName); value != "" {
			return value
		}
		
		return defaultValue
	})
}

// Load reads configuration from a YAML file and environment variables
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in the YAML content with default value support
	expanded := expandEnvWithDefaults(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks that required configuration fields are set
func (c *Config) Validate() error {
	if c.Server.Port == 0 {
		return fmt.Errorf("server.port is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Database == "" {
		return fmt.Errorf("database.database is required")
	}
	if c.GitHub.Token == "" {
		return fmt.Errorf("github.token is required")
	}

	// Validate custom rules
	for i, rule := range c.CustomRules {
		if err := ValidateRule(&rule); err != nil {
			return fmt.Errorf("invalid custom rule at index %d: %w", i, err)
		}
	}

	return nil
}

// DatabaseDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

// RedisAddr returns the Redis connection address
func (c *RedisConfig) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ValidateRule validates a custom rule's syntax and structure
func ValidateRule(rule *CustomRule) error {
	if rule.Name == "" {
		return fmt.Errorf("rule name is required")
	}

	// Validate error pattern is a valid regex if provided
	if rule.Conditions.ErrorPattern != nil && *rule.Conditions.ErrorPattern != "" {
		if _, err := regexp.Compile(*rule.Conditions.ErrorPattern); err != nil {
			return fmt.Errorf("invalid error_pattern regex in rule '%s': %w", rule.Name, err)
		}
	}

	// Validate severity values
	validSeverities := map[string]bool{
		"critical": true,
		"high":     true,
		"medium":   true,
		"low":      true,
	}

	if rule.Conditions.Severity != nil && !validSeverities[*rule.Conditions.Severity] {
		return fmt.Errorf("invalid severity in rule '%s': must be one of critical, high, medium, low", rule.Name)
	}

	if rule.Actions.SetSeverity != nil && !validSeverities[*rule.Actions.SetSeverity] {
		return fmt.Errorf("invalid set_severity in rule '%s': must be one of critical, high, medium, low", rule.Name)
	}

	// Validate that at least one condition is specified
	if rule.Conditions.ServiceName == nil &&
		rule.Conditions.ErrorPattern == nil &&
		rule.Conditions.Severity == nil &&
		rule.Conditions.Provider == nil &&
		len(rule.Conditions.Metadata) == 0 {
		return fmt.Errorf("rule '%s' must have at least one condition", rule.Name)
	}

	// Validate that at least one action is specified
	if rule.Actions.SetSeverity == nil &&
		len(rule.Actions.AddMetadata) == 0 &&
		rule.Actions.SetRepository == nil &&
		!rule.Actions.SkipRemediation {
		return fmt.Errorf("rule '%s' must have at least one action", rule.Name)
	}

	return nil
}

// Watcher watches a configuration file for changes and reloads it
type Watcher struct {
	path       string
	config     *Config
	mu         sync.RWMutex
	lastModTime time.Time
	stopCh     chan struct{}
	callbacks  []func(*Config)
}

// NewWatcher creates a new configuration watcher
func NewWatcher(path string) (*Watcher, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}

	return &Watcher{
		path:        path,
		config:      cfg,
		lastModTime: info.ModTime(),
		stopCh:      make(chan struct{}),
		callbacks:   make([]func(*Config), 0),
	}, nil
}

// Get returns the current configuration (thread-safe)
func (w *Watcher) Get() *Config {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.config
}

// OnReload registers a callback to be called when configuration is reloaded
func (w *Watcher) OnReload(callback func(*Config)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, callback)
}

// Start begins watching the configuration file for changes
func (w *Watcher) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.checkAndReload(); err != nil {
				// Log error but don't stop watching
				fmt.Fprintf(os.Stderr, "failed to reload config: %v\n", err)
			}
		case <-w.stopCh:
			return
		}
	}
}

// Stop stops watching the configuration file
func (w *Watcher) Stop() {
	close(w.stopCh)
}

// checkAndReload checks if the config file has changed and reloads it
func (w *Watcher) checkAndReload() error {
	info, err := os.Stat(w.path)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	// Check if file has been modified
	if !info.ModTime().After(w.lastModTime) {
		return nil
	}

	// Load new configuration
	newCfg, err := Load(w.path)
	if err != nil {
		return fmt.Errorf("failed to load new config: %w", err)
	}

	// Update configuration atomically
	w.mu.Lock()
	w.config = newCfg
	w.lastModTime = info.ModTime()
	callbacks := make([]func(*Config), len(w.callbacks))
	copy(callbacks, w.callbacks)
	w.mu.Unlock()

	// Call callbacks outside of lock
	for _, callback := range callbacks {
		callback(newCfg)
	}

	return nil
}
