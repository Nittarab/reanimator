---
inclusion: always
---

# Project Structure & Code Organization

## Directory Layout

The project is a monorepo with four main components:

- `incident-service/` - Go backend (Chi router, PostgreSQL, Redis)
- `dashboard/` - React frontend (Vite, shadcn/ui, TanStack Query)
- `demo-app/` - Node.js demo service with intentional bugs
- `remediation-action/` - GitHub Action for automated remediation

## File Placement Rules

### Go (incident-service/)

- Place new handlers in `internal/api/handlers.go`
- Add new adapters in `internal/adapters/{provider}.go`
- Database queries go in `internal/database/repository.go`
- Models belong in `internal/models/`
- Tests colocate with source: `*_test.go` files
- Property tests use suffix: `*_property_test.go`
- Migrations are numbered sequentially: `migrations/00X_description.sql`

### React (dashboard/)

- UI components go in `src/components/ui/` (shadcn/ui pattern)
- Page components go in `src/pages/{PageName}Page.tsx`
- API client functions go in `src/api/{resource}.ts`
- Tests colocate with source: `*.test.ts` or `*.test.tsx`
- Use `@/` path alias for imports from `src/`

### TypeScript (remediation-action/)

- Keep action logic in `src/main.ts`
- GitHub API interactions in `src/github.ts`
- Kiro CLI integration in `src/kiro.ts`

## Naming Conventions

**Go:**
- Exported: `PascalCase` (e.g., `CreateIncident`)
- Unexported: `camelCase` (e.g., `parseWebhook`)
- Interfaces: noun or adjective (e.g., `Repository`, `Adapter`)

**TypeScript/JavaScript:**
- Variables/functions: `camelCase` (e.g., `getIncidents`)
- Components/classes: `PascalCase` (e.g., `IncidentCard`)
- Constants: `UPPER_SNAKE_CASE` (e.g., `API_BASE_URL`)

**Files:**
- All files: `kebab-case` (e.g., `incident-detail-page.tsx`)
- Exception: React components may use `PascalCase.tsx`

**Database:**
- Tables/columns: `snake_case` (e.g., `incident_events`)

**API:**
- Endpoints: `kebab-case` (e.g., `/api/v1/workflow-status`)

## Import Organization

**Go - Three groups separated by blank lines:**
```go
import (
    // 1. Standard library
    "context"
    "fmt"
    
    // 2. External dependencies
    "github.com/go-chi/chi/v5"
    
    // 3. Internal packages (alphabetical)
    "github.com/your-org/ai-sre-platform/internal/models"
)
```

**TypeScript - Two groups:**
```typescript
// 1. External dependencies
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'

// 2. Internal modules (use @/ alias)
import { getIncidents } from '@/api/incidents'
import { IncidentCard } from '@/components/IncidentCard'
```

## Testing Patterns

**Colocate tests with source code:**
- Go: `handler.go` → `handler_test.go`
- TypeScript: `incidents.ts` → `incidents.test.ts`

**Property-based tests:**
- Use suffix `*_property_test.go` or `*.property.test.ts`
- Mark properties with comments: `// Property: description`
- Minimum 100 iterations for property tests

**Test naming:**
- Go: `func TestFunctionName_Scenario(t *testing.T)`
- TypeScript: `describe('functionName', () => { it('should scenario', ...) })`

## Architectural Patterns

**Adapter Pattern** - Normalize webhooks from different providers:
- Each provider has its own adapter in `internal/adapters/{provider}.go`
- All implement the `Adapter` interface
- Convert provider-specific formats to internal `Incident` model

**Repository Pattern** - Abstract database access:
- All database queries go through `Repository` interface
- Implementation in `internal/database/repository.go`
- Enables testing with mock repositories

**Event-Driven** - Incident state changes emit events:
- Events stored in `incident_events` table
- Provides audit trail and enables future event sourcing

**Queue-Based Workflow Dispatch** - Control concurrency:
- Redis-backed queue for GitHub workflow triggers
- Respects `max_concurrent_workflows` from config
- Prevents overwhelming GitHub Actions

## Configuration Management

**config.yaml** - Service mappings and platform settings:
- Maps service names to GitHub repositories
- Defines MCP servers for Kiro CLI
- Sets concurrency limits and deduplication windows

**.env** - Secrets and environment-specific values:
- Database connection strings
- API keys and tokens
- Never commit to version control

**Repository secrets** - GitHub Actions credentials:
- Set via GitHub UI or API
- Used by workflows for authentication

## Database Migrations

- Located in `incident-service/migrations/`
- Numbered sequentially: `001_`, `002_`, etc.
- Run via `go run cmd/migrate/main.go`
- Always include both `up` and `down` migrations
- Test migrations on a copy of production data before deploying

## Code Style Specifics

**Error handling (Go):**
```go
if err != nil {
    return fmt.Errorf("context: %w", err)  // Wrap errors with context
}
```

**HTTP responses (Go):**
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(response)
```

**API calls (TypeScript):**
```typescript
// Use TanStack Query for data fetching
const { data, isLoading, error } = useQuery({
  queryKey: ['incidents', filters],
  queryFn: () => getIncidents(filters)
})
```

**Component structure (React):**
```typescript
// Props interface, component, export
interface IncidentCardProps {
  incident: Incident
}

export function IncidentCard({ incident }: IncidentCardProps) {
  // Component logic
}
```

## When Adding New Features

1. **New webhook provider**: Add adapter in `internal/adapters/{provider}.go`
2. **New API endpoint**: Add handler in `internal/api/handlers.go`, register route in `cmd/server/main.go`
3. **New database table**: Create migration in `migrations/`, update repository
4. **New UI page**: Create in `src/pages/`, add route in `App.tsx`
5. **New configuration option**: Update `config.yaml` schema and `internal/config/config.go`
