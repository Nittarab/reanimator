# Technology Stack

## Incident Service

- **Language**: Go 1.21+
- **Framework**: Chi (lightweight HTTP router)
- **Database**: PostgreSQL 15+ (ACID compliance, JSON support)
- **Cache**: Redis 7+ (rate limiting, deduplication)
- **Testing**: `testing` package, `gopter` for property-based testing

## Dashboard

- **Framework**: React 18 with TypeScript
- **State Management**: TanStack Query (React Query)
- **UI Library**: shadcn/ui (Tailwind CSS components)
- **Build Tool**: Vite
- **Testing**: Vitest, React Testing Library, `fast-check` for property-based testing

## Remediation GitHub Action

- **Runtime**: Node.js 20
- **Language**: TypeScript
- **Testing**: Jest, `fast-check` for property-based testing

## Demo Application

- **Runtime**: Node.js 20
- **Framework**: Express
- **Database**: SQLite or in-memory
- **Error Tracking**: Sentry integration

## Infrastructure

- **Containerization**: Docker with multi-stage builds
- **Orchestration**: Docker Compose (dev/prod), Kubernetes (optional)
- **CI/CD**: GitHub Actions
- **Container Registry**: GitHub Container Registry (ghcr.io)

## Common Commands

### Development

```bash
# Start all services locally
./scripts/dev.sh

# Run all tests
./scripts/test.sh

# Test incident service
cd incident-service && go test -v -race ./...

# Test dashboard
cd dashboard && npm test

# Test demo app
cd demo-app && npm test
```

### Production

```bash
# Deploy to production
./scripts/prod.sh

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Check service status
docker-compose -f docker-compose.prod.yml ps
```

### Database

```bash
# Run migrations
cd incident-service && go run cmd/migrate/main.go

# Connect to database
psql $DATABASE_URL
```

### Docker

```bash
# Build all images
docker-compose build

# Build specific service
docker build -t incident-service ./incident-service

# Push to registry
docker push ghcr.io/your-org/incident-service:latest
```

## API Endpoints

### Incident Service

- `POST /api/v1/webhooks/incidents?provider={provider}` - Receive incident webhooks
- `GET /api/v1/incidents` - List incidents with filtering
- `GET /api/v1/incidents/:id` - Get incident details
- `POST /api/v1/incidents/:id/trigger` - Manually trigger remediation
- `POST /api/v1/webhooks/workflow-status` - Receive workflow status updates
- `GET /api/v1/health` - Health check
- `GET /api/v1/metrics` - Prometheus metrics

## Configuration

Configuration is managed via YAML files and environment variables:

- `config.yaml` - Service mappings, MCP servers, concurrency limits
- `.env` - Secrets and environment-specific settings
- Repository secrets - GitHub tokens, API keys

## Testing Strategy

- **Unit tests**: Test individual functions and components
- **Property-based tests**: Verify universal properties across random inputs (minimum 100 iterations)
- **Integration tests**: Test end-to-end flows with real services
- **Performance tests**: Validate response times and throughput

## Observability

- **Metrics**: Prometheus format exposed on port 9090
- **Logging**: Structured JSON logs with severity levels
- **Tracing**: Optional OpenTelemetry integration
- **Health checks**: HTTP endpoints for liveness/readiness probes
