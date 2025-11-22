# AI SRE Remediation GitHub Action

Automatically diagnose and fix production incidents using Kiro CLI with MCP integrations.

## Features

- **Kiro CLI Installation**: Automatically installs the specified version of Kiro CLI
- **MCP Configuration**: Reads MCP server configuration from repository or generates from environment variables
- **Incident Context**: Creates structured incident context files for Kiro CLI analysis
- **Environment Variable Substitution**: Securely handles credentials via GitHub secrets

## Usage

```yaml
name: Remediate Incident

on:
  workflow_dispatch:
    inputs:
      incident_id:
        required: true
      error_message:
        required: true
      stack_trace:
        required: false
      service_name:
        required: true
      timestamp:
        required: true

jobs:
  remediate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: ./remediation-action
        with:
          incident_id: ${{ inputs.incident_id }}
          error_message: ${{ inputs.error_message }}
          stack_trace: ${{ inputs.stack_trace }}
          service_name: ${{ inputs.service_name }}
          timestamp: ${{ inputs.timestamp }}
          kiro_version: 'latest'
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          DATADOG_API_KEY: ${{ secrets.DATADOG_API_KEY }}
          DATADOG_APP_KEY: ${{ secrets.DATADOG_APP_KEY }}
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `incident_id` | Unique incident identifier | Yes | - |
| `error_message` | Error message from the incident | Yes | - |
| `stack_trace` | Stack trace if available | No | '' |
| `service_name` | Name of the affected service | Yes | - |
| `timestamp` | Incident timestamp | Yes | - |
| `severity` | Incident severity (critical, high, medium, low) | No | 'medium' |
| `kiro_version` | Kiro CLI version to use | No | 'latest' |
| `incident_service_url` | URL of the incident service for status updates | No | '' |

## Outputs

| Output | Description |
|--------|-------------|
| `pr_url` | URL of the created pull request |
| `status` | Remediation status: success, failed, or no_fix_needed |
| `diagnosis` | Root cause diagnosis summary |

## MCP Configuration

The action supports two ways to configure MCP servers:

### 1. Repository Configuration (Recommended)

Create a `.kiro/settings/mcp.json` file in your repository:

```json
{
  "mcpServers": {
    "datadog": {
      "command": "npx",
      "args": ["-y", "@datadog/mcp-server"],
      "env": {
        "DATADOG_API_KEY": "${DATADOG_API_KEY}",
        "DATADOG_APP_KEY": "${DATADOG_APP_KEY}"
      }
    }
  }
}
```

The action will substitute `${VAR_NAME}` patterns with values from GitHub secrets.

### 2. Environment Variables (Auto-generated)

If no `.kiro/settings/mcp.json` exists, the action will automatically generate MCP configuration based on available environment variables:

- **Datadog**: Requires `DATADOG_API_KEY` and `DATADOG_APP_KEY`
- **Sentry**: Requires `SENTRY_AUTH_TOKEN` and optionally `SENTRY_ORG`
- **PagerDuty**: Requires `PAGERDUTY_API_KEY`
- **Grafana**: Requires `GRAFANA_API_KEY` and `GRAFANA_URL`

## Notifications

The action supports sending notifications when a pull request is created:

### Slack Notifications

Set the `SLACK_WEBHOOK_URL` environment variable to enable Slack notifications:

```yaml
env:
  SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

Slack notifications include:
- Incident ID and severity
- Affected service
- Error message
- Pull request link
- Timestamp

### Custom Webhook Notifications

Set the `CUSTOM_WEBHOOK_URL` environment variable to send notifications to a custom endpoint:

```yaml
env:
  CUSTOM_WEBHOOK_URL: ${{ secrets.CUSTOM_WEBHOOK_URL }}
```

The custom webhook receives a JSON payload with all incident details:

```json
{
  "incidentId": "inc_123",
  "serviceName": "api-gateway",
  "severity": "high",
  "errorMessage": "NullPointerException in UserService",
  "prUrl": "https://github.com/org/repo/pull/42",
  "timestamp": "2024-01-15T10:00:00Z"
}
```

**Note**: Notification failures are logged as warnings but do not fail the action. The pull request will still be created even if notifications fail.

## Development

```bash
# Install dependencies
npm install

# Build the action
npm run build

# Run tests
npm test

# Lint code
npm run lint
```

## Requirements

- Node.js 20+
- GitHub Actions environment
- Kiro CLI (installed automatically)

## License

See the main project LICENSE file.
