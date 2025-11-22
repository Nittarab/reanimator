# Incident Service

The Incident Service is the backend API that receives incident webhooks from observability platforms, manages incident state, and orchestrates GitHub Actions workflows for automated remediation.

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+

### Setup

1. Install dependencies:
```bash
go mod download
```

2. Start PostgreSQL and Redis (using Docker Compose):
```bash
docker-compose up -d postgres redis
```

3. Run database migrations:
```bash
go run cmd/migrate/main.go
```

4. Start the server:
```bash
go run cmd/server/main.go
```

The server will start on port 8080 by default.

## Testing

### Unit Tests

Run all tests:
```bash
go test -v ./...
```

### Property-Based Tests

The project includes property-based tests using `gopter`. These tests verify universal properties across randomly generated inputs (minimum 100 iterations).

To run property-based tests, you need a test database:

1. Start a test database:
```bash
./scripts/setup-test-db.sh
```

2. Set environment variables:
```bash
export TEST_DATABASE_HOST=localhost
export TEST_DATABASE_PORT=5433
export TEST_DATABASE_USER=postgres
export TEST_DATABASE_PASSWORD=postgres
export TEST_DATABASE_NAME=ai_sre_test
```

3. Run the tests:
```bash
cd incident-service
go test -v ./internal/database/...
```

4. Clean up:
```bash
docker stop ai-sre-test-db && docker rm ai-sre-test-db
```

### Test Coverage

Generate test coverage report:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Configuration

Configuration is loaded from `config.yaml` and environment variables. See `config.yaml` for available options.

Required environment variables:
- `GITHUB_TOKEN`: GitHub API token for workflow dispatch
- `DATABASE_HOST`: PostgreSQL host
- `DATABASE_PASSWORD`: PostgreSQL password
- `REDIS_HOST`: Redis host (optional)

## API Endpoints

- `GET /api/v1/health` - Health check endpoint
- `GET /api/v1/metrics` - Prometheus metrics
- `GET /api/v1/incidents` - List incidents
- `GET /api/v1/incidents/:id` - Get incident details
- `POST /api/v1/webhooks/incidents?provider={provider}` - Receive incident webhooks
- `POST /api/v1/incidents/:id/trigger` - Manually trigger remediation
- `POST /api/v1/webhooks/workflow-status` - Receive workflow status updates

## Architecture

The service follows a layered architecture:

- `cmd/`: Application entrypoints (server, migrate)
- `internal/api/`: HTTP handlers, middleware, logging, metrics
- `internal/config/`: Configuration management
- `internal/database/`: Database layer and repository pattern
- `internal/models/`: Data models
- `internal/adapters/`: Webhook adapters for observability platforms
- `internal/github/`: GitHub API client
- `migrations/`: Database schema migrations

## Observability

- **Metrics**: Prometheus metrics exposed on `/api/v1/metrics`
- **Logging**: Structured JSON logs to stdout
- **Health Checks**: `/api/v1/health` endpoint

## Docker

Build the Docker image:
```bash
docker build -t incident-service .
```

Run with Docker Compose:
```bash
docker-compose up incident-service
```
