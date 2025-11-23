# Kiro Settings - MCP Configuration

This directory contains configuration for Kiro CLI's Model Context Protocol (MCP) servers. MCP allows Kiro to query external systems for additional context during incident remediation.

## Overview

The `mcp.json` file configures which MCP servers Kiro CLI should use when analyzing incidents. For the demo application, we configure a Sentry MCP server to query error details, stack traces, and related context from Sentry.

## MCP Configuration File

**File**: `mcp.json`

This file defines MCP servers that Kiro CLI will use during remediation workflows.

### Structure

```json
{
  "mcpServers": {
    "server-name": {
      "command": "command-to-run",
      "args": ["array", "of", "arguments"],
      "env": {
        "ENV_VAR": "${GITHUB_SECRET_NAME}"
      },
      "disabled": false,
      "autoApprove": []
    }
  }
}
```

### Fields

- **command**: The command to execute the MCP server (e.g., `npx`, `uvx`, `node`)
- **args**: Array of arguments to pass to the command
- **env**: Environment variables for the MCP server (use `${VAR}` for substitution)
- **disabled**: Set to `true` to disable the server without removing configuration
- **autoApprove**: Array of tool names to auto-approve (empty = require approval for all)

## Sentry MCP Server Configuration

The demo application uses Sentry for error tracking. The Sentry MCP server allows Kiro to:

1. Query error details from Sentry issues
2. Retrieve full stack traces
3. Get breadcrumbs and context
4. Access related errors and patterns

### Configuration

```json
{
  "mcpServers": {
    "sentry": {
      "command": "npx",
      "args": ["-y", "@sentry/mcp-server"],
      "env": {
        "SENTRY_AUTH_TOKEN": "${SENTRY_AUTH_TOKEN}",
        "SENTRY_ORG": "${SENTRY_ORG}",
        "SENTRY_PROJECT": "${SENTRY_PROJECT}"
      },
      "disabled": false,
      "autoApprove": []
    }
  }
}
```

### Required GitHub Secrets

For the Sentry MCP server to work, configure these secrets in your GitHub repository:

1. **SENTRY_AUTH_TOKEN**
   - Create in Sentry: Settings → Account → API → Auth Tokens
   - Scopes needed: `project:read`, `event:read`
   - Add to GitHub: Settings → Secrets and variables → Actions → New repository secret

2. **SENTRY_ORG**
   - Your Sentry organization slug (e.g., `my-company`)
   - Found in Sentry URL: `https://sentry.io/organizations/{org-slug}/`

3. **SENTRY_PROJECT**
   - Your Sentry project slug (e.g., `demo-app`)
   - Found in Sentry URL: `https://sentry.io/organizations/{org}/projects/{project}/`

### Setting Up GitHub Secrets

```bash
# Navigate to your repository on GitHub
# Go to: Settings → Secrets and variables → Actions

# Add each secret:
# 1. Click "New repository secret"
# 2. Name: SENTRY_AUTH_TOKEN
#    Value: your-sentry-auth-token
# 3. Click "Add secret"

# Repeat for SENTRY_ORG and SENTRY_PROJECT
```

## How It Works

### During Remediation Workflow

1. **Workflow Triggered**: GitHub Actions workflow starts for an incident
2. **MCP Configuration**: Remediation action reads `mcp.json`
3. **Secret Substitution**: GitHub secrets replace `${VAR}` placeholders
4. **Kiro CLI Invocation**: Kiro starts with MCP servers configured
5. **Context Queries**: Kiro queries Sentry for additional incident context
6. **Fix Generation**: Kiro uses all available context to generate fix

### Example Kiro Queries

When analyzing an incident, Kiro might query Sentry for:

```
# Get full error details
sentry.get_issue(issue_id)

# Get recent occurrences
sentry.get_events(issue_id, limit=10)

# Get stack trace
sentry.get_event_stacktrace(event_id)

# Get breadcrumbs (user actions leading to error)
sentry.get_event_breadcrumbs(event_id)

# Get related errors
sentry.get_related_issues(issue_id)
```

## Environment Variable Substitution

The remediation action automatically substitutes environment variables:

**In mcp.json**:
```json
"env": {
  "SENTRY_AUTH_TOKEN": "${SENTRY_AUTH_TOKEN}"
}
```

**At runtime** (in GitHub Actions):
```json
"env": {
  "SENTRY_AUTH_TOKEN": "actual-token-value-from-github-secret"
}
```

This allows:
- Secrets to stay secure in GitHub
- Same configuration across repositories
- Easy credential rotation

## Adding Additional MCP Servers

To add more MCP servers (e.g., Datadog, PagerDuty):

### 1. Update mcp.json

```json
{
  "mcpServers": {
    "sentry": {
      "command": "npx",
      "args": ["-y", "@sentry/mcp-server"],
      "env": {
        "SENTRY_AUTH_TOKEN": "${SENTRY_AUTH_TOKEN}",
        "SENTRY_ORG": "${SENTRY_ORG}",
        "SENTRY_PROJECT": "${SENTRY_PROJECT}"
      }
    },
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

### 2. Add GitHub Secrets

Add the required secrets for the new MCP server to your repository.

### 3. Test

Trigger an incident and verify Kiro can query the new MCP server.

## Troubleshooting

### MCP Server Not Working

**Check GitHub Actions logs**:
```
# Look for MCP configuration output
# Should show: "Configured MCP servers: sentry"
```

**Verify secrets are set**:
```
# In GitHub: Settings → Secrets and variables → Actions
# Ensure all required secrets exist
```

**Test MCP server locally**:
```bash
# Install the MCP server
npm install -g @sentry/mcp-server

# Set environment variables
export SENTRY_AUTH_TOKEN=your-token
export SENTRY_ORG=your-org
export SENTRY_PROJECT=your-project

# Test the server
sentry-mcp-server
```

### Secret Substitution Not Working

**Check variable names match**:
- In `mcp.json`: `"SENTRY_AUTH_TOKEN": "${SENTRY_AUTH_TOKEN}"`
- In GitHub Secrets: Name must be exactly `SENTRY_AUTH_TOKEN`

**Check for typos**:
- Variable names are case-sensitive
- Must use `${VAR}` syntax, not `$VAR`

### Kiro Not Using MCP Server

**Verify mcp.json location**:
- Must be at `.kiro/settings/mcp.json` in repository root
- Check file is committed to git

**Check disabled flag**:
```json
"disabled": false  // Must be false or omitted
```

**Review Kiro logs**:
- Look for "Configured MCP servers" message
- Check for MCP query attempts

## Security Best Practices

### Never Commit Secrets

❌ **WRONG**:
```json
{
  "env": {
    "SENTRY_AUTH_TOKEN": "actual-token-here"
  }
}
```

✅ **CORRECT**:
```json
{
  "env": {
    "SENTRY_AUTH_TOKEN": "${SENTRY_AUTH_TOKEN}"
  }
}
```

### Use Least-Privilege Tokens

- Only grant necessary scopes
- Create separate tokens for CI/CD
- Rotate tokens regularly

### Monitor Token Usage

- Review Sentry audit logs
- Set up alerts for unusual API usage
- Revoke tokens if compromised

## Related Files

- **Remediation Action**: `remediation-action/src/mcp.ts`
- **Workflow Configuration**: `.github/workflows/demo-remediate.yml`
- **Demo App**: `demo-app/src/index.js` (Sentry initialization)

## Resources

- [Sentry MCP Server Documentation](https://docs.sentry.io/mcp)
- [Model Context Protocol Specification](https://modelcontextprotocol.io)
- [Kiro CLI MCP Guide](https://docs.kiro.ai/mcp)
- [GitHub Actions Secrets](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
