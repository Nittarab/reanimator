---
inclusion: always
---

# Project Structure & Code Organization

## Monorepo Components

- `incident-service/` - Go backend (Chi, PostgreSQL, Redis)
- `dashboard/` - React frontend (Vite, shadcn/ui, TanStack Query)
- `demo-app/` - Node.js demo service
- `remediation-action/` - GitHub Action for remediation

## File Placement

### incident-service/ (Go)

- Handlers: `internal/api/handlers.go`
- Adapters: `internal/adapters/{provider}.go` (implement `Adapter` interface)
- Database: `internal/database/repository.go` (implement `Repository` interface)
- Models: `internal/models/`
- Tests: colocate as `*_test.go`, property tests as `*_property_test.go`
- Migrations: `migrations/001_description.sql` (sequential numbering)

### dashboard/ (React)

- UI components: `src/components/ui/` (shadcn/ui)
- Pages: `src/pages/{Name}Page.tsx`
- API clients: `src/api/{resource}.ts`
- Tests: colocate as `*.test.ts` or `*.test.tsx`, property tests as `*.property.test.ts`
- Always use `@/` alias for imports from `src/`

### remediation-action/ (TypeScript)

- Action logic: `src/main.ts`
- GitHub API: `src/github.ts`
- Kiro CLI: `src/kiro.ts`

## Naming Conventions

| Context | Convention | Example |
|---------|-----------|---------|
| Go exported | PascalCase | `CreateIncident` |
| Go unexported | camelCase | `parseWebhook` |
| Go interfaces | Noun/adjective | `Repository`, `Adapter` |
| TS/JS variables/functions | camelCase | `getIncidents` |
| TS/JS components/classes | PascalCase | `IncidentCard` |
| TS/JS constants | UPPER_SNAKE_CASE | `API_BASE_URL` |
| Files | kebab-case | `incident-detail-page.tsx` |
| React components (exception) | PascalCase.tsx | `IncidentCard.tsx` |
| Database tables/columns | snake_case | `incident_events` |
| API endpoints | kebab-case | `/api/v1/workflow-status` |

## Import Organization

**Go** - Three groups with blank line separators:
```go
import (
    // Standard library
    "context"
    
    // External dependencies
    "github.com/go-chi/chi/v5"
    
    // Internal (alphabetical)
    "github.com/your-org/ai-sre-platform/internal/models"
)
```

**TypeScript** - Two groups:
```typescript
// External
import { useState } from 'react'

// Internal (use @/ alias)
import { getIncidents } from '@/api/incidents'
```

## Testing Requirements

- Colocate tests with source files
- Property tests: minimum 100 iterations, mark with `// Property: description`
- Go naming: `func TestFunctionName_Scenario(t *testing.T)`
- TypeScript naming: `describe('functionName', () => { it('should scenario', ...) })`

## Architecture Patterns

**Adapter Pattern**: Normalize provider webhooks
- Each provider: `internal/adapters/{provider}.go`
- Implements `Adapter` interface
- Converts to internal `Incident` model

**Repository Pattern**: Abstract database access
- All queries through `Repository` interface
- Enables mock testing

**Event-Driven**: Incident state changes
- Events in `incident_events` table
- Provides audit trail

**Queue-Based Dispatch**: Control workflow concurrency
- Redis-backed queue
- Respects `max_concurrent_workflows`

## Configuration Files

- `config.yaml`: Service mappings, MCP servers, concurrency limits
- `.env`: Secrets (never commit)
- Repository secrets: GitHub Actions credentials (set via UI)

## Database Migrations

- Location: `incident-service/migrations/`
- Format: `001_description.sql` (sequential)
- Run: `go run cmd/migrate/main.go`
- Always include up and down migrations

## Code Patterns

**Go error handling:**
```go
if err != nil {
    return fmt.Errorf("context: %w", err)
}
```

**Go HTTP responses:**
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(response)
```

**TypeScript data fetching:**
```typescript
const { data, isLoading, error } = useQuery({
  queryKey: ['incidents', filters],
  queryFn: () => getIncidents(filters)
})
```

**React components:**
```typescript
interface ComponentProps {
  prop: Type
}

export function Component({ prop }: ComponentProps) {
  // Logic
}
```

## Adding Features

- **Webhook provider**: Add `internal/adapters/{provider}.go` implementing `Adapter`
- **API endpoint**: Add to `internal/api/handlers.go`, register in `cmd/server/main.go`
- **Database table**: Create migration, update `repository.go`
- **UI page**: Create `src/pages/{Name}Page.tsx`, add route in `App.tsx`
- **Config option**: Update `config.yaml` and `internal/config/config.go`
