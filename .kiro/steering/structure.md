---
inclusion: always
---

# Project Structure & Code Organization

## Monorepo Layout

```
incident-service/     Go backend (Chi, PostgreSQL, Redis)
├── cmd/             Entrypoints (server, migrate)
├── internal/        Private application code
│   ├── adapters/    Provider webhook normalization (Adapter pattern)
│   ├── api/         HTTP handlers, middleware
│   ├── config/      Configuration management
│   ├── database/    Repository pattern, Redis
│   ├── github/      GitHub API client
│   └── models/      Domain models
└── migrations/      Sequential SQL migrations

dashboard/           React frontend (Vite, TypeScript, TanStack Query)
├── src/
│   ├── api/         API client functions
│   ├── components/  React components (shadcn/ui)
│   ├── pages/       Page components
│   └── lib/         Utilities

remediation-action/  GitHub Action (TypeScript)
├── src/             Action implementation
└── tests/           Integration tests

demo-app/            Demo Node.js service
```

## File Placement

### Go (incident-service/)

**Webhook provider**: `internal/adapters/{provider}.go`
- Implement `Adapter` interface: `ParseWebhook([]byte) (*models.Incident, error)`
- Colocate test: `{provider}_test.go`

**API endpoint**: `internal/api/handlers.go`
- Add handler function
- Register route in `cmd/server/main.go`

**Model**: `internal/models/{entity}.go`
- Include validation methods
- Colocate test: `{entity}_test.go`

**Database changes**:
- Migration: `migrations/00X_description.sql` (next sequential number)
- Update: `internal/database/repository.go` interface and implementation

**Tests**: Always colocate with source
- Unit: `*_test.go`
- Property: `*_property_test.go`

### TypeScript/React (dashboard/)

**Page**: `src/pages/{Name}Page.tsx`
- Add route in `src/App.tsx`
- Colocate test: `{Name}Page.test.tsx`

**API function**: `src/api/{resource}.ts`
- Export typed functions
- Colocate test: `{resource}.test.ts`

**UI component**: `src/components/ui/{name}.tsx`
- Use shadcn/ui patterns
- Export as named function

**CRITICAL**: Always use `@/` alias for internal imports
```typescript
import { getIncidents } from '@/api/incidents'  // ✓ CORRECT
import { getIncidents } from '../api/incidents' // ✗ WRONG
```

### TypeScript (remediation-action/)

- Entrypoint: `src/index.ts`
- Integrations: `src/github.ts`, `src/kiro.ts`, `src/mcp.ts`
- Tests: `*.test.ts`, property tests: `*.property.test.ts`

## Naming Conventions

**Go**:
- Exported: `PascalCase` (functions, types, constants)
- Unexported: `camelCase` (private functions, variables)
- Interfaces: Noun/adjective (`Repository`, `Adapter`)
- Tests: `TestFunctionName_Scenario` (e.g., `TestParseWebhook_InvalidJSON`)

**TypeScript/JavaScript**:
- Variables/functions: `camelCase` (`getIncidents`, `userId`)
- Components/classes: `PascalCase` (`IncidentCard`, `ApiClient`)
- Constants: `UPPER_SNAKE_CASE` (`API_BASE_URL`, `MAX_RETRIES`)
- Types/interfaces: `PascalCase` (`IncidentResponse`, `WebhookPayload`)

**Files**:
- General: `kebab-case` (`incident-detail.ts`, `webhook-handler.go`)
- React components: `PascalCase.tsx` (`IncidentCard.tsx`)
- Tests: Match source + suffix (`handlers_test.go`, `incidents.test.ts`)
- Property tests: Add `.property` before `.test` (`rules_property_test.go`)

**Database**: `snake_case` (`incident_events`, `created_at`)

**API**: `kebab-case` with `/api/v1/` prefix (`/api/v1/workflow-status`)

## Import Organization

**Go** (3 groups, blank line separated, alphabetical):
```go
import (
    // 1. Standard library
    "context"
    "fmt"
    
    // 2. External dependencies
    "github.com/go-chi/chi/v5"
    
    // 3. Internal packages
    "github.com/your-org/ai-sre-platform/internal/models"
)
```

**TypeScript/React** (2 groups, blank line separated):
```typescript
// 1. External packages
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'

// 2. Internal modules (MUST use @/ alias)
import { getIncidents } from '@/api/incidents'
import { Button } from '@/components/ui/button'
```

## Testing

**Placement**: Always colocate tests with source
- `handlers.go` → `handlers_test.go`
- `incidents.ts` → `incidents.test.ts`
- Property: `rules_property_test.go`, `notifications.property.test.ts`

**Go**:
- Naming: `func TestFunctionName_Scenario(t *testing.T)`
- Property tests: `gopter`, 100+ iterations, mark with `// Property: invariant description`
- Run: `go test -v -race ./...`

**TypeScript**:
- Structure: `describe('functionName', () => { it('should scenario', ...) })`
- Property tests: `fast-check`, 100+ iterations
- React: Vitest + React Testing Library

**Coverage**: >80% overall, 100% for critical paths (webhooks, remediation)

## Architecture Patterns

**Adapter Pattern** (webhook normalization):
- Location: `internal/adapters/{provider}.go`
- Interface: `ParseWebhook([]byte) (*models.Incident, error)`
- Converts provider payloads to internal `Incident` model
- Examples: `datadog.go`, `pagerduty.go`, `sentry.go`

**Repository Pattern** (database abstraction):
- Location: `internal/database/repository.go`
- ALL database queries MUST go through `Repository` interface
- Enables test mocking without real database
- Always pass `context.Context` for cancellation/timeouts

**Event-Driven Architecture** (audit trail):
- State changes recorded in `incident_events` table
- Events: `created`, `workflow_triggered`, `pr_opened`, `resolved`

**Queue-Based Dispatch** (concurrency control):
- Redis-backed queue for workflow triggers
- Respects `max_concurrent_workflows` from config

## Configuration

**config.yaml** (committed to repo):
- Service-to-repository mappings
- MCP server configurations
- `max_concurrent_workflows`, deduplication windows

**.env** (NEVER commit):
- Database URLs, API keys, tokens
- Use `.env.example` as template

**GitHub Secrets** (Settings → Secrets and variables → Actions):
- Used by remediation-action workflows
- Examples: `GITHUB_TOKEN`, `INCIDENT_SERVICE_URL`

## Database Migrations

**Location**: `incident-service/migrations/`

**Naming**: `{number}_{description}.sql` (sequential, zero-padded)
- Examples: `001_create_incidents.sql`, `002_create_incident_events.sql`

**Apply**: `go run cmd/migrate/main.go`

## Required Patterns

### Go

**Error handling** (ALWAYS wrap with %w):
```go
if err != nil {
    return fmt.Errorf("failed to parse webhook: %w", err)
}
```

**HTTP responses** (REQUIRED order: header → status → body):
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(response)
```

**Context passing** (REQUIRED for all I/O):
```go
func (r *Repository) GetIncident(ctx context.Context, id string) (*Incident, error)
```

### TypeScript/React

**Data fetching** (ONLY use TanStack Query):
```typescript
const { data, isLoading, error } = useQuery({
  queryKey: ['incidents', filters],
  queryFn: () => getIncidents(filters)
})
// NEVER use manual fetch() in components
```

**Component structure** (typed props, named export):
```typescript
interface ComponentProps {
  incidentId: string
  onUpdate?: () => void
}

export function IncidentCard({ incidentId, onUpdate }: ComponentProps) {
  // Implementation
}
```

**API functions** (typed, exported):
```typescript
export async function getIncidents(filters?: IncidentFilters): Promise<Incident[]> {
  const response = await apiClient.get('/incidents', { params: filters })
  return response.data
}
```

## Quick Reference: Adding Features

| Feature | Files to Create/Update |
|---------|------------------------|
| Webhook provider | `internal/adapters/{provider}.go` (implement `Adapter`), `{provider}_test.go` |
| API endpoint | `internal/api/handlers.go` (add handler), `cmd/server/main.go` (register route) |
| Database table | `migrations/00X_description.sql`, `internal/database/repository.go` |
| UI page | `src/pages/{Name}Page.tsx`, `src/App.tsx` (add route), `{Name}Page.test.tsx` |
| Config option | `config.yaml`, `internal/config/config.go` |
