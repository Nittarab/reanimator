# Environment Variables Reference

This document provides a comprehensive reference for all environment variables used by the AI SRE Platform.

## Table of Contents

- [Quick Start](#quick-start)
- [Required Variables](#required-variables)
- [Optional Variables](#optional-variables)
- [Component-Specific Variables](#component-specific-variables)
- [Environment-Specific Configuration](#environment-specific-configuration)
- [Security Best Practices](#security-best-practices)

## Quick Start

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Set the required variables (see [Required Variables](#required-variables))

3. Start the platform:
   ```bash
   ./scripts/dev.sh
   ```

## Required Variables

These variables MUST be set for the platform to function correctly.

### GITHUB_TOKEN

**Description:** GitHub Personal Access Token for triggering workflows

**Required:** Yes

**Format:** `ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx`

**Permissions:** `workflow` scope (to dispatch GitHub Actions workflows)

**How to create:**
1. Go to https://github.com/settings/tokens/new
2. Select `workflow` scope
3. Generate token and copy it

**Example:**
```bash
GITHUB_TOKEN=ghp_1234567890abcdefghijklmnopqrstuvwxyz
```

**Validation:** The Incident Service will fail to start if this is not set or is invalid.

---

### DATABASE_PASSWORD

**Description:** Password for PostgreSQL database connection

**Required:** Yes (in production)

**Format:** String (recommend 16+ characters)

**Default:** `postgres` (development only)

**Example:**
```bash
DATABASE_PASSWORD=MySecureP@ssw0rd123!
```

**Security:** Use a strong, unique password in production. Never use the default password.

---

### POSTGRES_PASSWORD

**Description:** Password for PostgreSQL Docker container

**Required:** Yes (when using Docker Compose)

**Format:** String (recommend 16+ characters)

**Default:** None

**Example:**
```bash
POSTGRES_PASSWORD=MySecureP@ssw0rd123!
```

**Note:** This should match `DATABASE_PASSWORD` when using Docker Compose.

## Optional Variables

These variables have sensible defaults but can be customized.

### Database Configuration

#### DATABASE_HOST
- **Description:** PostgreSQL server hostname
- **Default:** `localhost`
- **Docker Compose:** `postgres` (service name)
- **Example:** `DATABASE_HOST=postgres`

#### DATABASE_PORT
- **Description:** PostgreSQL server port
- **Default:** `5432`
- **Example:** `DATABASE_PORT=5432`

#### DATABASE_NAME
- **Description:** PostgreSQL database name
- **Default:** `ai_sre`
- **Example:** `DATABASE_NAME=ai_sre`

#### DATABASE_USER
- **Description:** PostgreSQL username
- **Default:** `postgres`
- **Example:** `DATABASE_USER=postgres`

#### DATABASE_SSL_MODE
- **Description:** PostgreSQL SSL mode
- **Default:** `disable`
- **Production:** `require`
- **Options:** `disable`, `require`, `verify-ca`, `verify-full`
- **Example:** `DATABASE_SSL_MODE=require`

### Redis Configuration

#### REDIS_HOST
- **Description:** Redis server hostname
- **Default:** `localhost`
- **Docker Compose:** `redis` (service name)
- **Example:** `REDIS_HOST=redis`

#### REDIS_PORT
- **Description:** Redis server port
- **Default:** `6379`
- **Example:** `REDIS_PORT=6379`

#### REDIS_PASSWORD
- **Description:** Redis authentication password
- **Default:** Empty (no authentication)
- **Example:** `REDIS_PASSWORD=myredispassword`

#### REDIS_DB
- **Description:** Redis database number
- **Default:** `0`
- **Example:** `REDIS_DB=0`

### GitHub Configuration

#### GITHUB_API_URL
- **Description:** GitHub API base URL
- **Default:** `https://api.github.com`
- **GitHub Enterprise:** `https://github.your-company.com/api/v3`
- **Example:** `GITHUB_API_URL=https://api.github.com`

### Dashboard Configuration

#### VITE_API_BASE_URL
- **Description:** Base URL for the Incident Service API
- **Default:** `/api/v1` (relative URL)
- **Development:** `http://localhost:8080/api/v1`
- **Production:** `https://your-domain.com/api/v1`
- **Example:** `VITE_API_BASE_URL=http://localhost:8080/api/v1`

**Note:** This variable is used at build time by Vite. Changes require rebuilding the dashboard.

### Demo App Configuration

#### PORT
- **Description:** Port for the demo application server
- **Default:** `3000`
- **Docker Compose:** `3002` (to avoid conflicts)
- **Example:** `PORT=3002`

#### NODE_ENV
- **Description:** Node.js environment
- **Default:** `development`
- **Options:** `development`, `production`, `test`
- **Example:** `NODE_ENV=production`

#### SENTRY_DSN
- **Description:** Sentry Data Source Name for error tracking
- **Default:** Empty (Sentry disabled)
- **Format:** `https://[key]@[organization].ingest.sentry.io/[project]`
- **Example:** `SENTRY_DSN=https://abc123@o123456.ingest.sentry.io/789012`
- **How to get:** Create a project at https://sentry.io and copy the DSN

### MCP Server Configuration

These variables are used by Kiro CLI to query observability platforms for additional context during incident remediation.

#### DATADOG_API_KEY
- **Description:** Datadog API key
- **Required:** Only if using Datadog MCP integration
- **How to get:** https://app.datadoghq.com/organization-settings/api-keys
- **Example:** `DATADOG_API_KEY=1234567890abcdef1234567890abcdef`

#### DATADOG_APP_KEY
- **Description:** Datadog application key
- **Required:** Only if using Datadog MCP integration
- **How to get:** https://app.datadoghq.com/organization-settings/application-keys
- **Example:** `DATADOG_APP_KEY=abcdef1234567890abcdef1234567890abcdef12`

#### PAGERDUTY_API_KEY
- **Description:** PagerDuty API key
- **Required:** Only if using PagerDuty MCP integration
- **Example:** `PAGERDUTY_API_KEY=u+abcdefghijklmnopqrstuvwx`

#### GRAFANA_API_KEY
- **Description:** Grafana API key
- **Required:** Only if using Grafana MCP integration
- **Example:** `GRAFANA_API_KEY=eyJrIjoiYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoifQ==`

#### GRAFANA_URL
- **Description:** Grafana instance URL
- **Required:** Only if using Grafana MCP integration
- **Example:** `GRAFANA_URL=https://your-company.grafana.net`

## Component-Specific Variables

### Incident Service (Go)

The Incident Service reads configuration from `config.yaml`, which supports environment variable expansion using `${VAR}` or `${VAR:-default}` syntax.

**Configuration file:** `config.yaml`

**Environment variables used:**
- `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_NAME`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_SSL_MODE`
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`
- `GITHUB_TOKEN`, `GITHUB_API_URL`

**Example config.yaml snippet:**
```yaml
database:
  host: ${DATABASE_HOST:-localhost}
  port: ${DATABASE_PORT:-5432}
  password: ${DATABASE_PASSWORD}
```

### Dashboard (React + Vite)

The Dashboard uses Vite's environment variable system. Only variables prefixed with `VITE_` are exposed to the client-side code.

**Build-time variables:**
- `VITE_API_BASE_URL`

**Note:** Changes to these variables require rebuilding the dashboard:
```bash
cd dashboard && npm run build
```

### Demo App (Node.js + Express)

The Demo App uses standard Node.js environment variables via `process.env`.

**Runtime variables:**
- `PORT`
- `NODE_ENV`
- `SENTRY_DSN`

### Remediation Action (GitHub Action)

The Remediation Action receives environment variables from GitHub Actions workflow files and repository secrets.

**Configured in workflow files:**
- `GITHUB_TOKEN` (automatically provided by GitHub Actions)
- `DATADOG_API_KEY`, `DATADOG_APP_KEY` (from repository secrets)
- Any other MCP server credentials (from repository secrets)

**Example workflow configuration:**
```yaml
env:
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  DATADOG_API_KEY: ${{ secrets.DATADOG_API_KEY }}
  DATADOG_APP_KEY: ${{ secrets.DATADOG_APP_KEY }}
```

## Environment-Specific Configuration

### Development

For local development, use the defaults in `.env.example`:

```bash
cp .env.example .env
# Edit .env with your GitHub token
./scripts/dev.sh
```

**Key settings:**
- `DATABASE_HOST=postgres` (Docker service name)
- `DATABASE_SSL_MODE=disable`
- `REDIS_HOST=redis` (Docker service name)
- `VITE_API_BASE_URL=http://localhost:8080/api/v1`

### Production

For production deployment, update these settings:

```bash
# Use strong passwords
DATABASE_PASSWORD=<strong-password>
POSTGRES_PASSWORD=<strong-password>

# Enable SSL
DATABASE_SSL_MODE=require

# Use production URLs
VITE_API_BASE_URL=https://your-domain.com/api/v1
GITHUB_API_URL=https://api.github.com

# Set production environment
NODE_ENV=production
```

**Deploy with:**
```bash
./scripts/prod.sh
```

### Testing

The test script (`scripts/test.sh`) automatically sets up test-specific environment variables:

```bash
TEST_DATABASE_HOST=localhost
TEST_DATABASE_PORT=5434
TEST_DATABASE_NAME=ai_sre_test
TEST_REDIS_HOST=localhost
TEST_REDIS_PORT=6380
```

**Run tests:**
```bash
./scripts/test.sh
```

## Security Best Practices

### 1. Never Commit Secrets

- ✅ DO: Use `.env` for local secrets (already in `.gitignore`)
- ❌ DON'T: Commit `.env` or hardcode secrets in code

### 2. Use Strong Passwords

- ✅ DO: Generate random passwords (16+ characters)
- ❌ DON'T: Use default passwords in production

**Generate a secure password:**
```bash
openssl rand -base64 32
```

### 3. Rotate Credentials Regularly

- Rotate GitHub tokens every 90 days
- Rotate database passwords quarterly
- Rotate API keys when team members leave

### 4. Use Least-Privilege Access

- GitHub token: Only `workflow` scope
- Database user: Only necessary permissions
- API keys: Read-only when possible

### 5. Enable SSL in Production

```bash
DATABASE_SSL_MODE=require
```

### 6. Use Secrets Management

For production, consider using a secrets manager:

- **AWS:** AWS Secrets Manager
- **GCP:** Google Secret Manager
- **Azure:** Azure Key Vault
- **HashiCorp:** Vault
- **Kubernetes:** Kubernetes Secrets

**Example with Kubernetes Secrets:**
```yaml
env:
  - name: DATABASE_PASSWORD
    valueFrom:
      secretKeyRef:
        name: ai-sre-secrets
        key: database-password
```

### 7. Audit Access

- Log all secret access
- Monitor for unauthorized access
- Review access logs regularly

### 8. Separate Environments

Use different credentials for each environment:

- Development: Separate database, test tokens
- Staging: Separate credentials, limited access
- Production: Strongest security, minimal access

## Troubleshooting

### "github.token is required" Error

**Problem:** Incident Service fails to start

**Solution:** Set `GITHUB_TOKEN` in `.env`:
```bash
GITHUB_TOKEN=ghp_your_token_here
```

### "failed to connect to database" Error

**Problem:** Cannot connect to PostgreSQL

**Solutions:**
1. Check database is running: `docker ps | grep postgres`
2. Verify credentials in `.env` match `POSTGRES_PASSWORD`
3. Check `DATABASE_HOST` is correct (`postgres` for Docker, `localhost` for local)

### Dashboard Shows "Network Error"

**Problem:** Dashboard cannot reach Incident Service API

**Solutions:**
1. Verify Incident Service is running: `curl http://localhost:8080/api/v1/health`
2. Check `VITE_API_BASE_URL` in `.env`
3. Rebuild dashboard if you changed `VITE_API_BASE_URL`: `cd dashboard && npm run build`

### Sentry Not Tracking Errors

**Problem:** Demo app errors not appearing in Sentry

**Solutions:**
1. Verify `SENTRY_DSN` is set correctly
2. Check Sentry project is active
3. Verify DSN format: `https://[key]@[org].ingest.sentry.io/[project]`

### MCP Server Connection Failed

**Problem:** Kiro CLI cannot query observability platform

**Solutions:**
1. Verify API keys are set: `DATADOG_API_KEY`, `DATADOG_APP_KEY`
2. Check API keys have correct permissions
3. Verify API keys are not expired
4. Check network connectivity to observability platform

## Validation

The Incident Service validates required configuration on startup. If validation fails, you'll see an error message indicating which variable is missing or invalid.

**Example validation error:**
```
failed to load config: invalid configuration: github.token is required
```

**To validate configuration without starting services:**
```bash
cd incident-service
go run cmd/server/main.go --validate-config
```

## Reference

### All Variables Summary

| Variable | Required | Default | Component |
|----------|----------|---------|-----------|
| `GITHUB_TOKEN` | Yes | - | Incident Service |
| `DATABASE_HOST` | No | `localhost` | Incident Service |
| `DATABASE_PORT` | No | `5432` | Incident Service |
| `DATABASE_NAME` | No | `ai_sre` | Incident Service |
| `DATABASE_USER` | No | `postgres` | Incident Service |
| `DATABASE_PASSWORD` | Yes* | `postgres` | Incident Service |
| `DATABASE_SSL_MODE` | No | `disable` | Incident Service |
| `REDIS_HOST` | No | `localhost` | Incident Service |
| `REDIS_PORT` | No | `6379` | Incident Service |
| `REDIS_PASSWORD` | No | - | Incident Service |
| `REDIS_DB` | No | `0` | Incident Service |
| `GITHUB_API_URL` | No | `https://api.github.com` | Incident Service |
| `VITE_API_BASE_URL` | No | `/api/v1` | Dashboard |
| `PORT` | No | `3000` | Demo App |
| `NODE_ENV` | No | `development` | Demo App |
| `SENTRY_DSN` | No | - | Demo App |
| `DATADOG_API_KEY` | No | - | MCP Integration |
| `DATADOG_APP_KEY` | No | - | MCP Integration |
| `PAGERDUTY_API_KEY` | No | - | MCP Integration |
| `GRAFANA_API_KEY` | No | - | MCP Integration |
| `GRAFANA_URL` | No | - | MCP Integration |

\* Required in production, has default for development

## Additional Resources

- [Configuration Reference](docs/CONFIGURATION.md) - Detailed `config.yaml` documentation
- [Deployment Guide](docs/DEPLOYMENT.md) - Production deployment instructions
- [Security Guide](docs/SECURITY.md) - Security best practices
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md) - Common issues and solutions
