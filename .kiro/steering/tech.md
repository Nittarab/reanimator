---
inclusion: always
---

# Technology Stack & Conventions

## Go (incident-service/)

**Stack**: Go 1.21+, Chi router, PostgreSQL 15+, Redis 7+, standard library preferred

### Critical Patterns (REQUIRED)

**Error Handling**: Always wrap errors with context using `%w` verb
```go
if err != nil {
    return fmt.Errorf("failed to parse webhook: %w", err)
}
```

**HTTP Response Order**: Header → Status → Body (this exact order)
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(response)
```

**Context Passing**: First parameter for all I/O operations
```go
func (r *Repository) GetIncident(ctx context.Context, id string) (*Incident, error)
```

**Repository Pattern**: ALL database queries through `internal/database/repository.go` interface. Never use direct SQL in handlers or business logic.

### Testing

- Unit: Standard `testing` package, naming `TestFunctionName_Scenario`
- Property: `gopter` library, suffix `*_property_test.go`, 100+ iterations, mark with `// Property: invariant description`
- Colocate tests with source files
- Always run with race detector: `go test -v -race ./...`

### Database

- PostgreSQL JSON columns for flexible metadata
- Migrations: Sequential `migrations/001_description.sql`
- Apply: `go run cmd/migrate/main.go`

## TypeScript/React (dashboard/)

**Stack**: React 18, TypeScript, Vite, TanStack Query, shadcn/ui, Tailwind CSS

### Critical Patterns (REQUIRED)

**Import Paths**: ALWAYS use `@/` alias, NEVER relative paths
```typescript
import { getIncidents } from '@/api/incidents'      // ✓ CORRECT
import { getIncidents } from '../api/incidents'     // ✗ WRONG
```

**Component Structure**: Typed props interface + named function export
```typescript
interface ComponentProps {
  incidentId: string
  onUpdate?: () => void
}

export function IncidentCard({ incidentId, onUpdate }: ComponentProps) {
  // Implementation
}
```

**Data Fetching**: ONLY TanStack Query, NEVER manual fetch() in components
```typescript
const { data, isLoading, error } = useQuery({
  queryKey: ['incidents', filters],
  queryFn: () => getIncidents(filters)
})
```

### Testing

- Vitest + React Testing Library
- Property: `fast-check`, suffix `*.property.test.ts`, 100+ iterations
- Naming: `describe('functionName', () => { it('should scenario', ...) })`
- Colocate: `incidents.ts` → `incidents.test.ts`

### UI

- Use shadcn/ui components from `src/components/ui/`

## Node.js (remediation-action/, demo-app/)

**Stack**: Node.js 20, TypeScript (remediation-action), JavaScript (demo-app), Express (demo-app)

### GitHub Action (remediation-action/)

- Minimal logic, single responsibility
- Use `@actions/core` for inputs/outputs
- Graceful error handling with actionable messages
- Export testable functions separately from entrypoint

### Testing

- Jest with `fast-check` for property tests (100+ iterations)

## API Conventions

### Existing Endpoints (incident-service)

- `POST /api/v1/webhooks/incidents?provider={provider}` - Webhook ingestion
- `GET /api/v1/incidents` - List (query: status, severity, service)
- `GET /api/v1/incidents/:id` - Get details
- `POST /api/v1/incidents/:id/trigger` - Trigger remediation
- `POST /api/v1/webhooks/workflow-status` - Status updates
- `GET /api/v1/health` - Health check
- `GET /api/v1/metrics` - Prometheus (port 9090)

### URL Conventions (REQUIRED)

- kebab-case paths: `/api/v1/workflow-status` NOT `/api/v1/workflowStatus`
- Always include `/api/v1/` prefix

### HTTP Status Codes (REQUIRED)

- 200: GET/PUT/PATCH success
- 201: POST success (created)
- 400: Validation error
- 404: Not found
- 500: Internal error

### Error Response (REQUIRED)

```json
{
  "error": "Human-readable message",
  "code": "ERROR_CODE",
  "details": {}
}
```

## Configuration

### config.yaml (committed, no secrets)

- Service-to-repository mappings
- MCP server configurations
- `max_concurrent_workflows`, deduplication windows

### .env (NEVER commit)

- Database URLs, API keys, tokens
- Use `.env.example` as template
- Load via environment variables

### GitHub Secrets

- Set via Settings → Secrets and variables → Actions
- Examples: `GITHUB_TOKEN`, `INCIDENT_SERVICE_URL`

## Testing Standards

### Coverage Requirements

- Unit: All public functions/methods
- Integration: End-to-end flows (webhook → database → GitHub Actions)
- Property: All invariants, 100+ iterations
- Overall: >80%, Critical paths: 100% (webhooks, remediation, database)

### Property-Based Testing

Test invariants that hold for all inputs:
- Go: `gopter`, suffix `*_property_test.go`
- TypeScript: `fast-check`, suffix `*.property.test.ts`
- Mark with: `// Property: invariant description`
- Minimum 100 iterations

### Organization

Always colocate tests with source:
- Go: `handlers.go` → `handlers_test.go`
- TypeScript: `incidents.ts` → `incidents.test.ts`

## Development Commands

### Quick Start

```bash
./scripts/dev.sh          # Start all services
./scripts/test.sh         # Run all tests
```

### Service-Specific

```bash
# Go (with race detector)
cd incident-service && go test -v -race ./...

# Dashboard
cd dashboard && npm test

# Remediation action
cd remediation-action && npm test

# Database migrations
cd incident-service && go run cmd/migrate/main.go
```

### Docker

```bash
docker-compose up                               # Dev
docker-compose -f docker-compose.prod.yml up -d # Prod
```

## Observability

**Logging Format** (REQUIRED):
```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "message": "Incident received",
  "context": {
    "incident_id": "inc_123",
    "provider": "datadog"
  }
}
```
- Levels: debug, info, warn, error
- Timestamp: ISO 8601 format
- Never log secrets, tokens, passwords, or PII

**Metrics** (Prometheus on port 9090):
- Counters: Use `*_total` suffix (e.g., `incidents_received_total`)
- Histograms: Use `*_duration_seconds` for latencies
- Gauges: Current state values (e.g., `active_incidents`)

**Health Checks**:
- Endpoint: `GET /api/v1/health`
- Returns 200 when all dependencies are healthy

## Security

**Secrets Management** (REQUIRED):
- NEVER log secrets, tokens, passwords, or PII
- Store secrets in environment variables only
- Use `.env` file locally (never commit)
- Use GitHub repository secrets for Actions

**Webhook Security**:
- Validate all webhook signatures before processing
- Implement rate limiting via Redis

**GitHub Tokens**:
- Use least-privilege, repo-scoped tokens
- Never use personal access tokens with broad permissions
