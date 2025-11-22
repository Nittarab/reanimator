---
inclusion: always
---

# Technology Stack & Conventions

## Go (incident-service/)

**Stack**: Go 1.21+, Chi router, PostgreSQL 15+, Redis 7+

**Code Style**:
- Always use `fmt.Errorf("context: %w", err)` to wrap errors with context
- HTTP handlers: set Content-Type, write status code, then encode JSON response
- Use `context.Context` for cancellation and timeouts in all I/O operations
- Run tests with race detector: `go test -v -race ./...`

**Testing**:
- Standard `testing` package for unit tests
- `gopter` for property-based tests (suffix: `*_property_test.go`)
- Property tests must run minimum 100 iterations
- Mark properties with comments: `// Property: description`
- Test naming: `func TestFunctionName_Scenario(t *testing.T)`

**Database**:
- Use PostgreSQL JSON columns for flexible data (e.g., incident metadata)
- All queries through Repository interface for testability
- Migrations in `migrations/` numbered sequentially: `001_`, `002_`, etc.
- Run migrations: `go run cmd/migrate/main.go`

**Dependencies**:
- Chi for routing (lightweight, idiomatic)
- `lib/pq` for PostgreSQL driver
- `go-redis` for Redis client
- Avoid heavy frameworks; prefer standard library

## TypeScript/React (dashboard/)

**Stack**: React 18, TypeScript, Vite, TanStack Query, shadcn/ui, Tailwind CSS

**Code Style**:
- Use `@/` path alias for imports from `src/`
- Functional components with TypeScript interfaces for props
- TanStack Query for all data fetching (no manual fetch in components)
- shadcn/ui components for consistent UI (in `src/components/ui/`)

**Testing**:
- Vitest + React Testing Library for component tests
- `fast-check` for property-based tests (suffix: `*.property.test.ts`)
- Test naming: `describe('functionName', () => { it('should scenario', ...) })`
- Colocate tests: `incidents.ts` → `incidents.test.ts`

**Data Fetching Pattern**:
```typescript
const { data, isLoading, error } = useQuery({
  queryKey: ['resource', params],
  queryFn: () => apiFunction(params)
})
```

**Component Structure**:
```typescript
interface ComponentProps {
  prop: Type
}

export function Component({ prop }: ComponentProps) {
  // Component logic
}
```

## Node.js (remediation-action/, demo-app/)

**Stack**: Node.js 20, TypeScript (action) / JavaScript (demo), Express (demo)

**Testing**: Jest with `fast-check` for property-based tests

**Action Conventions**:
- Keep action logic minimal and focused
- Use `@actions/core` for inputs/outputs
- Handle errors gracefully with clear messages

## API Design

**Endpoints** (incident-service):
- `POST /api/v1/webhooks/incidents?provider={provider}` - Receive webhooks
- `GET /api/v1/incidents` - List with filtering (status, severity, service)
- `GET /api/v1/incidents/:id` - Get details
- `POST /api/v1/incidents/:id/trigger` - Manual remediation trigger
- `POST /api/v1/webhooks/workflow-status` - Workflow status updates
- `GET /api/v1/health` - Health check
- `GET /api/v1/metrics` - Prometheus metrics (port 9090)

**Conventions**:
- Use kebab-case for URL paths
- Version all APIs: `/api/v1/`
- Return proper HTTP status codes (200, 201, 400, 404, 500)
- JSON responses with consistent error format

## Configuration Management

**config.yaml**: Service mappings, MCP servers, concurrency limits
- Maps service names to GitHub repositories
- Defines Kiro CLI MCP server configurations
- Sets `max_concurrent_workflows` and deduplication windows

**.env**: Secrets and environment-specific values
- Database URLs, API keys, tokens
- Never commit to version control
- Use `.env.example` as template

**Repository Secrets**: GitHub Actions credentials (set via GitHub UI)

## Testing Requirements

**All Code**:
- Unit tests for individual functions
- Integration tests for end-to-end flows
- Property-based tests for invariants (min 100 iterations)

**Property Test Naming**:
- Go: `*_property_test.go`
- TypeScript: `*.property.test.ts`

**Coverage Goals**:
- Aim for >80% code coverage
- 100% coverage for critical paths (webhook processing, remediation triggers)

## Development Workflow

**Local Development**:
```bash
./scripts/dev.sh          # Start all services
./scripts/test.sh         # Run all tests
```

**Service-Specific**:
```bash
# Go tests with race detector
cd incident-service && go test -v -race ./...

# Dashboard tests
cd dashboard && npm test

# Database migrations
cd incident-service && go run cmd/migrate/main.go
```

**Docker**:
```bash
docker-compose build                    # Build all images
docker-compose up                       # Start services
docker-compose -f docker-compose.prod.yml up -d  # Production
```

## Observability Standards

**Logging**: Structured JSON with fields:
- `level`: debug, info, warn, error
- `timestamp`: ISO 8601
- `message`: Human-readable description
- `context`: Additional structured data

**Metrics**: Prometheus format on port 9090
- Counter: `_total` suffix (e.g., `incidents_received_total`)
- Histogram: `_duration_seconds` for latencies
- Gauge: Current state values

**Health Checks**: `/api/v1/health` returns 200 when healthy

## Security Conventions

- Never log sensitive data (tokens, passwords, PII)
- Use environment variables for secrets
- Validate all webhook signatures
- Rate limit API endpoints via Redis
- Use least-privilege GitHub tokens (scoped to specific repos)
