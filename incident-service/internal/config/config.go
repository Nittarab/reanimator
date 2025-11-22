package config

import (
	"fmt"
	"os"
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

// Load reads configuration from a YAML file and environment variables
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in the YAML content
	expanded := os.ExpandEnv(string(data))

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
