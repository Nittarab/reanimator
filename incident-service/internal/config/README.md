# Configuration Package

This package provides configuration management for the AI SRE Platform incident service, including:

- YAML configuration file parsing
- Environment variable expansion
- Service-to-repository mappings
- Custom rule engine for incident processing
- Hot-reloading of configuration files

## Basic Usage

### Loading Configuration

```go
cfg, err := config.Load("config.yaml")
if err != nil {
    log.Fatal(err)
}
```

### Using Configuration Watcher (Hot-Reload)

```go
// Create watcher
watcher, err := config.NewWatcher("config.yaml")
if err != nil {
    log.Fatal(err)
}
defer watcher.Stop()

// Register callback for config changes
watcher.OnReload(func(newCfg *config.Config) {
    log.Println("Configuration reloaded")
    // Update your services with new config
})

// Start watching (checks every 10 seconds)
go watcher.Start(10 * time.Second)

// Get current config (thread-safe)
cfg := watcher.Get()
```

## Custom Rules

Custom rules allow you to define incident detection and processing logic in the configuration file.

### Rule Structure

```yaml
custom_rules:
  - name: high-priority-payment-errors
    description: Escalate payment service errors to critical
    enabled: true
    conditions:
      service_name: payment-service
      error_pattern: ".*payment.*failed.*"
    actions:
      set_severity: critical
      add_metadata:
        priority: high
        team: payments
```

### Using the Rule Engine

```go
// Create rule engine from config
engine := config.NewRuleEngine(cfg.CustomRules)

// Evaluate rules against an incident
incident := &config.IncidentData{
    ServiceName:  "payment-service",
    ErrorMessage: "payment processing failed",
    Severity:     "high",
    Provider:     "datadog",
    Metadata:     map[string]string{},
}

matches := engine.Evaluate(incident)

// Apply rule actions
config.ApplyActions(incident, matches)

// Check if remediation should be skipped
if config.ShouldSkipRemediation(matches) {
    // Skip automated remediation
}

// Check for repository override
if repo := config.GetRepositoryOverride(matches); repo != nil {
    // Use custom repository instead of default mapping
}
```

## Rule Conditions

Rules support the following conditions (all must match for the rule to apply):

- `service_name`: Exact match on service name
- `error_pattern`: Regex pattern match on error message
- `severity`: Exact match on severity level (critical, high, medium, low)
- `provider`: Exact match on observability provider (datadog, pagerduty, grafana, sentry)
- `metadata`: Key-value pairs that must all match

## Rule Actions

Rules can perform the following actions:

- `set_severity`: Change the incident severity
- `add_metadata`: Add key-value pairs to incident metadata
- `set_repository`: Override the repository for remediation
- `skip_remediation`: Skip automated remediation for this incident

## Configuration File Format

```yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: ${DATABASE_HOST:-localhost}
  port: 5432
  database: ai_sre
  user: postgres
  password: ${DATABASE_PASSWORD}
  ssl_mode: disable

redis:
  host: localhost
  port: 6379
  password: ${REDIS_PASSWORD:-}
  db: 0

github:
  api_url: https://api.github.com
  token: ${GITHUB_TOKEN}
  workflow_name: remediate-incident.yml

service_mappings:
  - service_name: api-gateway
    repository: org/api-gateway
    branch: main
  - service_name: user-service
    repository: org/user-service
    branch: main

deduplication:
  time_window: 5m

concurrency:
  max_workflows_per_repo: 2

mcp_servers:
  - name: datadog
    type: datadog
    config:
      api_key: ${DATADOG_API_KEY}
      app_key: ${DATADOG_APP_KEY}

custom_rules:
  - name: escalate-payment-errors
    description: Escalate payment service errors
    enabled: true
    conditions:
      service_name: payment-service
      error_pattern: ".*payment.*"
    actions:
      set_severity: critical
      add_metadata:
        team: payments
```

## Environment Variables

The configuration file supports environment variable expansion using `${VAR_NAME}` syntax. You can also provide default values using `${VAR_NAME:-default}`.

## Validation

Configuration is automatically validated when loaded:

- Required fields are checked
- Custom rules are validated for syntax errors
- Regex patterns in rules are compiled to ensure validity
- Severity values are validated against allowed values

Invalid configurations will return an error with a descriptive message.
