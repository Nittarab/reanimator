# AI SRE Platform

An autonomous infrastructure remediation system that acts as a "digital immune system" for software services. The platform receives incident notifications from observability platforms, triggers GitHub Actions workflows that use Kiro CLI to diagnose root causes and generate code fixes, and provides a dashboard for on-call engineers to monitor remediation progress.

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Node.js 20+ (for local development)
- GitHub Personal Access Token with `workflow` scope

### Development Setup

1. Clone the repository:
```bash
git clone https://github.com/your-org/ai-sre-platform.git
cd ai-sre-platform
```

2. Copy the environment template:
```bash
cp .env.example .env
```

3. Edit `.env` with your credentials:
```bash
# Required: GitHub token for workflow dispatch
GITHUB_TOKEN=your_github_token_here

# Required: Encryption key (32 bytes)
ENCRYPTION_KEY=your_32_byte_encryption_key_here

# Optional: Observability platform credentials
DATADOG_API_KEY=your_datadog_api_key
DATADOG_APP_KEY=your_datadog_app_key
SENTRY_DSN=your_sentry_dsn_here
```

4. Start the development environment:
```bash
./scripts/dev.sh
```

5. Access the services:
- **Dashboard**: http://localhost:3000
- **Incident Service API**: http://localhost:8080
- **Demo App**: http://localhost:3001
- **Metrics**: http://localhost:9090/metrics

## 📦 Architecture

The platform consists of four main components:

### 1. Incident Service (Go)
Backend API service that:
- Receives incident webhooks from observability platforms (Datadog, PagerDuty, Grafana, Sentry)
- Manages incident state and deduplication
- Triggers GitHub Actions workflows
- Provides query API for the dashboard

### 2. Dashboard (React + TypeScript)
Web application that provides:
- Real-time incident list with filtering
- Incident detail view with timeline
- Manual remediation trigger
- Configuration display

### 3. Remediation GitHub Action (TypeScript)
Reusable GitHub Action that:
- Installs Kiro CLI
- Configures MCP servers for observability platforms
- Runs automated diagnosis and fix generation
- Creates pull requests with post-mortems

### 4. Demo Application (Node.js + Express)
Sample buggy service demonstrating:
- Integration with Sentry for error tracking
- End-to-end remediation flow
- Common error scenarios (division by zero, null pointer, etc.)

## 🔧 Configuration

### Service Mappings

Edit `config.yaml` to map services to repositories:

```yaml
service_mappings:
  - service_name: api-gateway
    repository: org/api-gateway
    branch: main
  - service_name: user-service
    repository: org/user-service
    branch: main
```

### MCP Servers

Configure observability platform integrations:

```yaml
mcp_servers:
  - name: datadog
    type: datadog
    config:
      api_key_env: DATADOG_API_KEY
      app_key_env: DATADOG_APP_KEY
```

### Webhook Setup

Configure your observability platform to send webhooks to:

```
POST http://your-domain:8080/api/v1/webhooks/incidents?provider=datadog
POST http://your-domain:8080/api/v1/webhooks/incidents?provider=pagerduty
POST http://your-domain:8080/api/v1/webhooks/incidents?provider=grafana
POST http://your-domain:8080/api/v1/webhooks/incidents?provider=sentry
```

## 🧪 Testing

Run all tests:
```bash
./scripts/test.sh
```

Run tests for specific components:
```bash
# Incident Service
cd incident-service && go test -v ./...

# Dashboard
cd dashboard && npm test -- --run

# Demo App
cd demo-app && npm test -- --run

# Remediation Action
cd remediation-action && npm test
```

## 🚢 Production Deployment

1. Build and deploy:
```bash
./scripts/prod.sh
```

2. View logs:
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f
```

3. Check service health:
```bash
curl http://localhost:8080/api/v1/health
```

## 📊 Monitoring

The platform exposes Prometheus metrics at:
```
http://localhost:9090/metrics
```

Key metrics:
- `incidents_received_total` - Total incidents received
- `workflows_dispatched_total` - Total workflows triggered
- `workflow_dispatch_errors_total` - Failed workflow dispatches
- `incidents_by_status` - Current incidents by status
- `queue_depth` - Number of queued incidents

## 🔐 Security

- All credentials are encrypted at rest using AES-256
- GitHub tokens use least-privilege scopes (workflow dispatch only)
- Webhook signatures are validated for supported providers
- TLS 1.3 required for all external communication

## 📚 Documentation

- [Deployment Guide](docs/DEPLOYMENT.md)
- [Configuration Reference](docs/CONFIGURATION.md)
- [Adding New Adapters](docs/ADAPTERS.md)
- [API Documentation](docs/API.md)

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🆘 Support

- GitHub Issues: https://github.com/your-org/ai-sre-platform/issues
- Documentation: https://docs.your-domain.com
