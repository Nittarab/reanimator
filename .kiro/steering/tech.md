---
inclusion: always
---

# Technology Stack & Conventions

## Go (incident-service/)

**Stack**: Go 1.21+, Chi router, PostgreSQL 15+, Redis 7+

**Critical Patterns**:
- Error wrapping: `fmt.Errorf("context: %w", err)` for all errors
- HTTP handlers: set Content-Type → write status → encode JSON
- Always pass `context.Context` for I/O operations (cancellation/timeouts)
- All database queries through `Repository` interface (enables mocking)
- Test with race detector: `go test -v -race ./...`

**Testing**:
- Unit tests: standard `testing` package, naming `TestFunctionName_Scenario`
- Property tests: `gopter` library, suffix `*_property_test.go`, min 100 iterations
- Mark properties: `// Property: description of invariant`
- Colocate tests with source files

**Database**:
- PostgreSQL JSON columns for flexible metadata
- Migrations: `migrations/001_description.sql` (sequential numbering)
- Run: `go run cmd/migrate/main.go`

**Dependencies**: Chi (routing), lib/pq (PostgreSQL), go-redis (Redis). Prefer stdlib over frameworks.

## TypeScript/React (dashboard/)

**Stack**: React 18, TypeScript, Vite, TanStack Query, shadcn/ui, Tailwind CSS

**Critical Patterns**:
- Imports: use `@/` alias for `src/` paths
- Components: functional with TypeScript interfaces for props
- Data fetching: TanStack Query only (no manual fetch)
- UI: shadcn/ui components in `src/components/ui/`

**Required Component Structure**:
```typescript
interface ComponentProps {
  prop: Type
}

export function Component({ prop }: ComponentProps) {
  // Logic here
}
```

**Required Data Fetching**:
```typescript
const { data, isLoading, error } = useQuery({
  queryKey: ['resource', params],
  queryFn: () => apiFunction(params)
})
```

**Testing**:
- Vitest + React Testing Library for components
- Property tests: `fast-check`, suffix `*.property.test.ts`
- Naming: `describe('functionName', () => { it('should scenario', ...) })`
- Colocate: `incidents.ts` → `incidents.test.ts`

## Node.js (remediation-action/, demo-app/)

**Stack**: Node.js 20, TypeScript (action), JavaScript (demo), Express (demo)

**Action Requirements**:
- Minimal, focused logic
- Use `@actions/core` for inputs/outputs
- Graceful error handling with clear messages

**Testing**: Jest with `fast-check` for property tests

## API Conventions

**Endpoints** (incident-service):
- `POST /api/v1/webhooks/incidents?provider={provider}` - Webhook ingestion
- `GET /api/v1/incidents` - List (filter: status, severity, service)
- `GET /api/v1/incidents/:id` - Details
- `POST /api/v1/incidents/:id/trigger` - Manual remediation
- `POST /api/v1/webhooks/workflow-status` - Status updates
- `GET /api/v1/health` - Health check
- `GET /api/v1/metrics` - Prometheus (port 9090)

**Rules**:
- kebab-case URLs, version prefix `/api/v1/`
- Proper status codes: 200 (OK), 201 (Created), 400 (Bad Request), 404 (Not Found), 500 (Server Error)
- Consistent JSON error format

## Configuration

**config.yaml**: Service→repo mappings, MCP servers, `max_concurrent_workflows`, deduplication windows

**.env**: Secrets (DB URLs, API keys, tokens). Never commit. Use `.env.example` as template.

**GitHub Secrets**: Set via repository settings UI for Actions credentials

## Testing Standards

**Required Coverage**:
- Unit tests for all functions
- Integration tests for end-to-end flows
- Property tests for invariants (min 100 iterations)
- >80% code coverage overall
- 100% coverage for critical paths (webhooks, remediation triggers)

**Property Test Files**:
- Go: `*_property_test.go`
- TypeScript: `*.property.test.ts`

## Development Commands

**Quick Start**:
```bash
./scripts/dev.sh          # Start all services
./scripts/test.sh         # Run all tests
```

**Service-Specific**:
```bash
cd incident-service && go test -v -race ./...  # Go tests
cd dashboard && npm test                        # Dashboard tests
cd incident-service && go run cmd/migrate/main.go  # Migrations
```

**Docker**:
```bash
docker-compose up                               # Dev mode
docker-compose -f docker-compose.prod.yml up -d # Production
```

## Observability

**Logging**: Structured JSON with `level` (debug/info/warn/error), `timestamp` (ISO 8601), `message`, `context`

**Metrics**: Prometheus on port 9090
- Counters: `*_total` suffix (e.g., `incidents_received_total`)
- Histograms: `*_duration_seconds` for latencies
- Gauges: current state values

**Health**: `/api/v1/health` returns 200 when healthy

## Security

- Never log secrets (tokens, passwords, PII)
- Secrets via environment variables only
- Validate all webhook signatures
- Rate limiting via Redis
- Least-privilege GitHub tokens (repo-scoped)
