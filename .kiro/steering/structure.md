# Project Structure

## Repository Layout

```
ai-sre-platform/
├── .github/
│   └── workflows/
│       ├── ci.yml                    # CI/CD pipeline
│       └── demo-remediate.yml        # Demo app remediation workflow
├── .kiro/
│   ├── specs/
│   │   └── ai-sre-platform/          # Feature specification
│   └── steering/                     # AI assistant guidance
├── incident-service/                 # Go backend service
│   ├── cmd/
│   │   ├── server/                   # HTTP server entrypoint
│   │   └── migrate/                  # Database migrations
│   ├── internal/
│   │   ├── adapters/                 # Webhook adapters (Datadog, PagerDuty, etc.)
│   │   ├── api/                      # HTTP handlers
│   │   ├── config/                   # Configuration parsing
│   │   ├── database/                 # Database layer
│   │   ├── github/                   # GitHub API client
│   │   └── models/                   # Data models
│   ├── migrations/                   # SQL migration files
│   ├── Dockerfile
│   └── go.mod
├── dashboard/                        # React web UI
│   ├── src/
│   │   ├── components/               # React components
│   │   ├── pages/                    # Page components
│   │   ├── api/                      # API client
│   │   └── lib/                      # Utilities
│   ├── public/                       # Static assets
│   ├── Dockerfile
│   └── package.json
├── demo-app/                         # Demo application
│   ├── src/
│   │   ├── routes/                   # API endpoints (buggy)
│   │   └── db/                       # Database layer
│   ├── public/                       # Demo UI
│   ├── Dockerfile
│   └── package.json
├── remediation-action/               # GitHub Action
│   ├── src/
│   │   ├── main.ts                   # Action entrypoint
│   │   ├── kiro.ts                   # Kiro CLI integration
│   │   └── github.ts                 # GitHub API client
│   ├── action.yml                    # Action metadata
│   └── package.json
├── scripts/
│   ├── dev.sh                        # Start development environment
│   ├── prod.sh                       # Deploy production
│   └── test.sh                       # Run all tests
├── docker-compose.yml                # Base compose config
├── docker-compose.dev.yml            # Development overrides
├── docker-compose.prod.yml           # Production config
├── .env.example                      # Environment template
└── README.md
```

## Component Organization

### Incident Service (Go)

- **cmd/**: Application entrypoints
- **internal/**: Private application code (not importable by other projects)
- **migrations/**: Database schema migrations (numbered sequentially)
- **Dockerfile**: Multi-stage build (golang:alpine → alpine)

### Dashboard (React)

- **src/components/**: Reusable UI components (shadcn/ui based)
- **src/pages/**: Page-level components (IncidentList, IncidentDetail, Config)
- **src/api/**: API client functions using TanStack Query
- **Dockerfile**: Multi-stage build (node:alpine → nginx:alpine)

### Demo App (Node.js)

- **src/routes/**: Express route handlers with intentional bugs
- **public/**: Static HTML/CSS/JS for demo UI
- **Dockerfile**: Single-stage Node.js runtime

### Remediation Action (TypeScript)

- **src/main.ts**: GitHub Action entrypoint
- **src/kiro.ts**: Kiro CLI installation and invocation
- **src/github.ts**: PR creation and status reporting
- **action.yml**: Action metadata (inputs, outputs, runs)

## Configuration Files

- **config.yaml**: Service mappings, MCP servers, deduplication settings, concurrency limits
- **.env**: Database URLs, API keys, GitHub tokens
- **Repository secrets**: Sensitive credentials for workflows

## Database Schema

Located in `incident-service/migrations/`:

- **001_create_incidents.sql**: Main incidents table
- **002_create_incident_events.sql**: Audit trail
- **003_create_service_mappings.sql**: Service-to-repo mappings
- **004_create_mcp_servers.sql**: MCP server configurations

## Testing Structure

Each component has tests colocated with source:

- Go: `*_test.go` files alongside source
- TypeScript: `*.test.ts` or `*.spec.ts` files
- Property-based tests: Clearly marked with `Property X:` comments

## Naming Conventions

- **Go**: PascalCase for exported, camelCase for unexported
- **TypeScript/JavaScript**: camelCase for variables/functions, PascalCase for components/classes
- **Files**: kebab-case for all files
- **Database**: snake_case for tables and columns
- **API endpoints**: kebab-case paths
- **Environment variables**: UPPER_SNAKE_CASE

## Import Organization

### Go

```go
import (
    // Standard library
    "context"
    "fmt"
    
    // External dependencies
    "github.com/go-chi/chi/v5"
    
    // Internal packages
    "github.com/your-org/ai-sre-platform/internal/models"
)
```

### TypeScript

```typescript
// External dependencies
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'

// Internal modules
import { getIncidents } from '@/api/incidents'
import { IncidentCard } from '@/components/IncidentCard'
```

## Key Architectural Patterns

- **Adapter Pattern**: Webhook adapters for different observability providers
- **Repository Pattern**: Database access layer abstraction
- **Event-Driven**: Incident state changes trigger events
- **Queue-Based**: Workflow dispatch with concurrency control
- **Property-Based Testing**: Universal properties verified across random inputs
