# Design Document

## Overview

The AI SRE Platform is a distributed system consisting of three main components:

1. **Incident Service** - A backend API service that receives incident webhooks from observability platforms, manages incident state, and triggers GitHub Actions workflows
2. **Dashboard** - A web application that provides real-time visibility into incidents and their remediation status
3. **Remediation GitHub Action** - A reusable GitHub Action that repositories install to enable automated fix generation using Kiro CLI with MCP integrations

The platform follows an event-driven architecture where incidents flow from observability platforms through the Incident Service, which orchestrates remediation by triggering workflows in the affected repositories. Kiro CLI, running within GitHub Actions, uses MCP to query observability platforms for additional context, performs root cause analysis, generates fixes, and creates pull requests.

## Architecture

### High-Level Architecture

```mermaid
graph TB
    subgraph "Observability Platforms"
        DD[Datadog]
        NR[New Relic]
        PROM[Prometheus]
    end
    
    subgraph "AI SRE Platform"
        IS[Incident Service]
        DB[(Database)]
        DASH[Dashboard]
    end
    
    subgraph "GitHub"
        REPO[Repository]
        GHA[GitHub Actions]
        KIRO[Kiro CLI + MCP]
    end
    
    DD -->|Webhook| IS
    NR -->|Webhook| IS
    PROM -->|Webhook| IS
    
    IS --> DB
    DASH --> IS
    IS -->|Workflow Dispatch| GHA
    
    GHA --> KIRO
    KIRO -->|MCP Query| DD
    KIRO -->|MCP Query| NR
    KIRO -->|Create PR| REPO
    
    GHA -->|Status Update| IS
```

### Component Interaction Flow

```mermaid
sequenceDiagram
    participant OBS as Observability Platform
    participant IS as Incident Service
    participant DB as Database
    participant DASH as Dashboard
    participant GH as GitHub Actions
    participant KIRO as Kiro CLI
    participant MCP as MCP Server
    
    OBS->>IS: POST /webhooks/incidents
    IS->>DB: Store incident
    IS->>IS: Map service to repository
    IS->>GH: Dispatch workflow
    IS->>DB: Update status: workflow_triggered
    
    DASH->>IS: GET /incidents
    IS->>DASH: Return incident list
    
    GH->>KIRO: Run remediation
    KIRO->>MCP: Query logs/traces
    MCP->>KIRO: Return context
    KIRO->>KIRO: Analyze & generate fix
    KIRO->>GH: Create PR
    
    GH->>IS: POST /webhooks/workflow-complete
    IS->>DB: Update status: pr_created
```

## Components and Interfaces

### 1. Incident Service

**Technology Stack:**
- **Language:** Go (for performance, concurrency, and operational simplicity)
- **Framework:** Chi router (lightweight, idiomatic HTTP routing)
- **Database:** PostgreSQL (ACID compliance, JSON support for flexible incident data)
- **Caching:** Redis (for rate limiting and deduplication)

**API Endpoints:**

```
POST   /api/v1/webhooks/incidents          # Receive incident from observability platform
GET    /api/v1/incidents                   # List incidents with filtering
GET    /api/v1/incidents/:id               # Get incident details
POST   /api/v1/incidents/:id/trigger       # Manually trigger remediation
POST   /api/v1/webhooks/workflow-status    # Receive workflow status updates
GET    /api/v1/health                      # Health check
GET    /api/v1/metrics                     # Prometheus metrics
```

**Core Responsibilities:**
- Receive and validate incident webhooks
- Transform provider-specific formats to internal schema using adapters
- Deduplicate incidents within time windows
- Map services to repositories using configuration
- Trigger GitHub workflow dispatch API
- Track workflow status and update incident records
- Provide query API for dashboard

**Webhook Adapter Architecture:**

The Incident Service uses a pluggable adapter pattern to support multiple observability platforms. Each adapter implements a common interface:

```go
type WebhookAdapter interface {
    // Validate checks if the webhook payload is valid and authentic
    Validate(r *http.Request) error
    
    // Parse transforms the provider-specific payload into our Incident struct
    Parse(body []byte) (*Incident, error)
    
    // ProviderName returns the name of the observability provider
    ProviderName() string
}
```

**Supported Adapters (MVP):**

1. **Datadog Adapter**
   - Validates webhook signature using shared secret
   - Extracts service name from tags (e.g., `service:api-gateway`)
   - Maps alert priority to severity
   - Extracts error message from alert body
   - Parses stack traces from alert details if present

2. **PagerDuty Adapter**
   - Validates webhook signature
   - Extracts service from incident.service.summary
   - Maps urgency to severity
   - Extracts error details from incident.body
   - Handles incident.triggered events

3. **Grafana Adapter**
   - Validates optional webhook secret
   - Extracts service from alert labels
   - Maps alert state to severity
   - Extracts error from alert annotations
   - Parses query results for stack traces

4. **Sentry Adapter**
   - Validates webhook signature using secret key
   - Extracts service from project name or tags
   - Maps event level to severity
   - Extracts exception message and stack trace
   - Handles error and issue events

**Adapter Registration:**

Adapters are registered at service startup:

```go
func NewIncidentService() *IncidentService {
    s := &IncidentService{
        adapters: make(map[string]WebhookAdapter),
    }
    
    // Register adapters
    s.RegisterAdapter(NewDatadogAdapter())
    s.RegisterAdapter(NewPagerDutyAdapter())
    s.RegisterAdapter(NewGrafanaAdapter())
    s.RegisterAdapter(NewSentryAdapter())
    
    return s
}
```

**Webhook Endpoint Routing:**

The provider is specified via query parameter:

```
POST /api/v1/webhooks/incidents?provider=datadog
POST /api/v1/webhooks/incidents?provider=pagerduty
POST /api/v1/webhooks/incidents?provider=grafana
POST /api/v1/webhooks/incidents?provider=sentry
```

The handler selects the appropriate adapter based on the provider parameter:

```go
func (s *IncidentService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    provider := r.URL.Query().Get("provider")
    
    adapter, ok := s.adapters[provider]
    if !ok {
        http.Error(w, "unsupported provider", http.StatusBadRequest)
        return
    }
    
    if err := adapter.Validate(r); err != nil {
        http.Error(w, "validation failed", http.StatusUnauthorized)
        return
    }
    
    body, _ := io.ReadAll(r.Body)
    incident, err := adapter.Parse(body)
    if err != nil {
        http.Error(w, "parse failed", http.StatusBadRequest)
        return
    }
    
    // Store incident and trigger workflow...
}
```

**Adding New Adapters:**

To support a new observability platform:

1. Implement the `WebhookAdapter` interface
2. Add provider-specific validation logic
3. Map provider fields to Incident struct
4. Register the adapter in `NewIncidentService()`
5. Document the webhook URL format for users

This design allows the platform to easily support new providers without modifying core business logic.

**Configuration Schema:**

```yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

database:
  host: postgres
  port: 5432
  database: ai_sre
  ssl_mode: require

github:
  api_url: https://api.github.com
  token: ${GITHUB_TOKEN}
  workflow_name: remediate-incident.yml

service_mappings:
  - service_name: api-gateway
    repository: org/api-gateway
    branch: main
  - service_name: user-service
    repository: org/user-service
    branch: main

deduplication:
  time_window: 5m
  
concurrency:
  max_workflows_per_repo: 2

mcp_servers:
  - name: datadog
    type: datadog
    api_key: ${DATADOG_API_KEY}
    app_key: ${DATADOG_APP_KEY}
```

### 2. Dashboard

**Technology Stack:**
- **Framework:** React with TypeScript
- **State Management:** TanStack Query (React Query) for server state
- **UI Library:** shadcn/ui (Tailwind-based components)
- **Build Tool:** Vite
- **Deployment:** Static files served via CDN or nginx

**Key Features:**
- Real-time incident list with auto-refresh
- Incident detail view with timeline
- Manual remediation trigger
- Filtering by status, service, repository, time range
- Pull request links and workflow status
- Incident history and statistics

**Pages:**

1. **Incidents List** (`/`)
   - Table view with columns: Status, Service, Repository, Error Message, Time, Actions
   - Filters: Status (all, pending, in_progress, pr_created, resolved, failed)
   - Search by service name or error message
   - Auto-refresh every 10 seconds

2. **Incident Detail** (`/incidents/:id`)
   - Full incident data including stack trace
   - Timeline of events (received, workflow triggered, PR created)
   - Link to GitHub workflow run
   - Link to pull request (if created)
   - Manual trigger button
   - Link to observability platform (if configured)

3. **Configuration** (`/config`)
   - View service-to-repository mappings
   - View MCP server configurations (read-only)

### 3. Remediation GitHub Action

**Action Metadata:**

```yaml
name: 'AI SRE Remediation'
description: 'Automatically diagnose and fix incidents using Kiro CLI'
author: 'AI SRE Platform'

inputs:
  incident_id:
    description: 'Unique incident identifier'
    required: true
  error_message:
    description: 'Error message from the incident'
    required: true
  stack_trace:
    description: 'Stack trace if available'
    required: false
  service_name:
    description: 'Name of the affected service'
    required: true
  timestamp:
    description: 'Incident timestamp'
    required: true
  kiro_version:
    description: 'Kiro CLI version to use'
    required: false
    default: 'latest'
  mcp_config:
    description: 'JSON string of MCP server configurations'
    required: false

outputs:
  pr_url:
    description: 'URL of the created pull request'
  status:
    description: 'Remediation status: success, failed, or no_fix_needed'
  diagnosis:
    description: 'Root cause diagnosis summary'

runs:
  using: 'composite'
```

**Workflow Steps:**

1. Checkout repository
2. Install Kiro CLI
3. Configure MCP servers from input
4. Create incident context file
5. Run Kiro CLI with remediation prompt
6. Commit changes if any
7. Create pull request
8. Report status back to Incident Service

**Example Usage in Repository:**

```yaml
name: Remediate Incident

on:
  workflow_dispatch:
    inputs:
      incident_id:
        required: true
      error_message:
        required: true
      stack_trace:
        required: false
      service_name:
        required: true
      timestamp:
        required: true

jobs:
  remediate:
    runs-on: ubuntu-latest
    steps:
      - uses: ai-sre-platform/remediation-action@v1
        with:
          incident_id: ${{ inputs.incident_id }}
          error_message: ${{ inputs.error_message }}
          stack_trace: ${{ inputs.stack_trace }}
          service_name: ${{ inputs.service_name }}
          timestamp: ${{ inputs.timestamp }}
          mcp_config: ${{ secrets.MCP_CONFIG }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 4. Kiro CLI Integration

**Remediation Prompt Template:**

The GitHub Action will invoke Kiro CLI with a structured prompt:

```
You are an AI SRE agent tasked with diagnosing and fixing a production incident.

## Incident Details
- ID: {incident_id}
- Service: {service_name}
- Timestamp: {timestamp}
- Error: {error_message}

## Stack Trace
{stack_trace}

## Your Task
1. Use the MCP servers to query additional context from the observability platform around the incident timestamp
2. Analyze the stack trace and error message to identify the root cause
3. Locate the problematic code in this repository
4. Generate a fix that addresses the root cause
5. Run the test suite to validate your fix
6. Create a detailed post-mortem explaining what broke, why, and how your fix addresses it

## MCP Servers Available
{mcp_server_list}

## Repository Context
You have access to the full repository. Check the .kiro directory for any custom remediation strategies or specs.

Begin your analysis.
```

**MCP Configuration:**

The action will create an `mcp.json` file for Kiro CLI:

```json
{
  "mcpServers": {
    "datadog": {
      "command": "npx",
      "args": ["-y", "@datadog/mcp-server"],
      "env": {
        "DATADOG_API_KEY": "${DATADOG_API_KEY}",
        "DATADOG_APP_KEY": "${DATADOG_APP_KEY}"
      }
    }
  }
}
```

## Data Models

### Incident

```go
type Incident struct {
    ID              string                 `json:"id" db:"id"`
    ServiceName     string                 `json:"service_name" db:"service_name"`
    Repository      string                 `json:"repository" db:"repository"`
    ErrorMessage    string                 `json:"error_message" db:"error_message"`
    StackTrace      *string                `json:"stack_trace,omitempty" db:"stack_trace"`
    Severity        string                 `json:"severity" db:"severity"`
    Status          IncidentStatus         `json:"status" db:"status"`
    Provider        string                 `json:"provider" db:"provider"`
    ProviderData    map[string]interface{} `json:"provider_data" db:"provider_data"`
    WorkflowRunID   *int64                 `json:"workflow_run_id,omitempty" db:"workflow_run_id"`
    PullRequestURL  *string                `json:"pull_request_url,omitempty" db:"pull_request_url"`
    Diagnosis       *string                `json:"diagnosis,omitempty" db:"diagnosis"`
    CreatedAt       time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
    TriggeredAt     *time.Time             `json:"triggered_at,omitempty" db:"triggered_at"`
    CompletedAt     *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

type IncidentStatus string

const (
    StatusPending          IncidentStatus = "pending"
    StatusWorkflowTriggered IncidentStatus = "workflow_triggered"
    StatusInProgress       IncidentStatus = "in_progress"
    StatusPRCreated        IncidentStatus = "pr_created"
    StatusResolved         IncidentStatus = "resolved"
    StatusFailed           IncidentStatus = "failed"
    StatusNoFixNeeded      IncidentStatus = "no_fix_needed"
)
```

### Service Mapping

```go
type ServiceMapping struct {
    ServiceName string `yaml:"service_name"`
    Repository  string `yaml:"repository"`
    Branch      string `yaml:"branch"`
}
```

### MCP Server Config

```go
type MCPServerConfig struct {
    Name   string            `yaml:"name" json:"name"`
    Type   string            `yaml:"type" json:"type"`
    Config map[string]string `yaml:"config" json:"config"`
}
```

### Webhook Payload Examples and Transformations

**Datadog Webhook Payload:**

```json
{
  "id": "1234567890",
  "title": "High error rate on api-gateway",
  "body": "Error rate exceeded threshold: NullPointerException in UserService.getUser()",
  "alert_type": "error",
  "priority": "normal",
  "tags": ["service:api-gateway", "env:production"],
  "date_happened": 1705320000,
  "aggregation_key": "api-gateway-errors",
  "source_type_name": "ALERT",
  "snapshot": "https://app.datadoghq.com/monitors/123456"
}
```

**Transformed to Internal Incident:**

```json
{
  "id": "inc_dd_1234567890",
  "service_name": "api-gateway",
  "repository": "",
  "error_message": "High error rate on api-gateway: Error rate exceeded threshold: NullPointerException in UserService.getUser()",
  "stack_trace": null,
  "severity": "high",
  "status": "pending",
  "provider": "datadog",
  "provider_data": {
    "alert_id": "1234567890",
    "snapshot_url": "https://app.datadoghq.com/monitors/123456",
    "tags": ["service:api-gateway", "env:production"]
  },
  "created_at": "2024-01-15T10:00:00Z"
}
```

**PagerDuty Webhook Payload:**

```json
{
  "event": {
    "id": "Q1234567890",
    "event_type": "incident.triggered",
    "resource_type": "incident",
    "occurred_at": "2024-01-15T10:00:00Z",
    "data": {
      "id": "PABCDEF",
      "type": "incident",
      "title": "Database connection timeout in payment-service",
      "service": {
        "id": "PSVC123",
        "summary": "payment-service"
      },
      "urgency": "high",
      "body": {
        "details": "Connection pool exhausted. Stack trace: at PaymentService.processPayment(PaymentService.java:45)"
      },
      "html_url": "https://mycompany.pagerduty.com/incidents/PABCDEF"
    }
  }
}
```

**Transformed to Internal Incident:**

```json
{
  "id": "inc_pd_PABCDEF",
  "service_name": "payment-service",
  "repository": "",
  "error_message": "Database connection timeout in payment-service",
  "stack_trace": "at PaymentService.processPayment(PaymentService.java:45)",
  "severity": "critical",
  "status": "pending",
  "provider": "pagerduty",
  "provider_data": {
    "incident_id": "PABCDEF",
    "incident_url": "https://mycompany.pagerduty.com/incidents/PABCDEF",
    "service_id": "PSVC123"
  },
  "created_at": "2024-01-15T10:00:00Z"
}
```

**Sentry Webhook Payload:**

```json
{
  "action": "created",
  "data": {
    "issue": {
      "id": "987654321",
      "title": "TypeError: Cannot read property 'id' of undefined",
      "culprit": "app/services/user-service.js in getUser",
      "level": "error",
      "platform": "javascript",
      "project": "user-service"
    },
    "event": {
      "event_id": "abc123def456",
      "timestamp": "2024-01-15T10:00:00Z",
      "exception": {
        "values": [
          {
            "type": "TypeError",
            "value": "Cannot read property 'id' of undefined",
            "stacktrace": {
              "frames": [
                {
                  "filename": "app/services/user-service.js",
                  "function": "getUser",
                  "lineno": 42
                }
              ]
            }
          }
        ]
      },
      "tags": [
        ["environment", "production"],
        ["service", "user-service"]
      ]
    }
  },
  "url": "https://sentry.io/organizations/myorg/issues/987654321/"
}
```

**Transformed to Internal Incident:**

```json
{
  "id": "inc_sentry_987654321",
  "service_name": "user-service",
  "repository": "",
  "error_message": "TypeError: Cannot read property 'id' of undefined",
  "stack_trace": "at getUser (app/services/user-service.js:42)",
  "severity": "high",
  "status": "pending",
  "provider": "sentry",
  "provider_data": {
    "issue_id": "987654321",
    "event_id": "abc123def456",
    "issue_url": "https://sentry.io/organizations/myorg/issues/987654321/",
    "platform": "javascript"
  },
  "created_at": "2024-01-15T10:00:00Z"
}
```

**Adapter Transformation Logic:**

Each adapter extracts the following fields:

| Field | Datadog Source | PagerDuty Source | Sentry Source |
|-------|---------------|------------------|---------------|
| service_name | tags (service:X) | service.summary | event.tags or project |
| error_message | title + body | title | exception.value |
| stack_trace | body (parsed) | body.details (parsed) | exception.stacktrace |
| severity | priority → severity map | urgency → severity map | level → severity map |
| provider_data | Original payload | Original payload | Original payload |

**Severity Mapping:**

```go
// Datadog priority → severity
var datadogSeverityMap = map[string]string{
    "P1": "critical",
    "P2": "high",
    "P3": "medium",
    "P4": "low",
}

// PagerDuty urgency → severity
var pagerDutySeverityMap = map[string]string{
    "high": "critical",
    "low": "medium",
}

// Sentry level → severity
var sentrySeverityMap = map[string]string{
    "fatal": "critical",
    "error": "high",
    "warning": "medium",
    "info": "low",
}
```


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Incident persistence round-trip
*For any* valid incident received via webhook, storing it to the database and then querying it by ID should return an equivalent incident with all required fields (unique identifier, timestamp, status) populated.
**Validates: Requirements 1.5**

### Property 2: Provider format transformation
*For any* valid provider-specific incident payload, transforming it to the internal Incident structure should produce a valid Incident with all required fields populated.
**Validates: Requirements 1.2**

### Property 3: Malformed data error handling
*For any* malformed or invalid webhook payload, the Incident Service should return an HTTP error code in the 4xx range and log the error without crashing.
**Validates: Requirements 1.4**

### Property 4: Service-to-repository lookup consistency
*For any* configured service name, looking up the associated repository should always return the same repository configuration.
**Validates: Requirements 2.2**

### Property 5: Deduplication within time window
*For any* two incidents with the same service and error signature sent within the configured time window, only one incident record should exist in the database.
**Validates: Requirements 2.3**

### Property 6: Workflow dispatch includes required context
*For any* incident ready for remediation, the workflow dispatch API call should include all required fields: incident_id, error_message, service_name, and timestamp.
**Validates: Requirements 3.2**

### Property 7: Retry with exponential backoff
*For any* failed workflow dispatch API call, the service should retry up to three times with exponentially increasing delays between attempts.
**Validates: Requirements 3.3**

### Property 8: Concurrency limit enforcement
*For any* repository with N concurrent workflows configured, dispatching more than N incidents should result in the excess incidents being queued.
**Validates: Requirements 3.4, 12.2, 12.3**

### Property 9: Branch naming includes incident ID
*For any* incident being remediated, the created branch name should contain the incident ID.
**Validates: Requirements 8.1**

### Property 10: Post-mortem completeness
*For any* generated post-mortem document, it should contain sections for: what broke, why it broke, how the fix addresses it, and test results.
**Validates: Requirements 8.4, 8.5**

### Property 11: Configuration parsing validity
*For any* valid YAML configuration file, parsing it should produce a valid configuration object with all service mappings and MCP servers accessible.
**Validates: Requirements 11.1, 11.2, 22.1**

### Property 12: Workflow completion updates queue
*For any* workflow completion webhook received, the service should decrement the active workflow count and process the next queued incident if any exist.
**Validates: Requirements 12.5**

### Property 13: Audit trail completeness
*For any* incident processed through the system, the audit trail should contain records for all state transitions (received, workflow_triggered, in_progress, pr_created, etc.).
**Validates: Requirements 14.1, 20.3**

### Property 14: Incident filtering correctness
*For any* filter criteria (status, service, repository, time range), the returned incidents should all match the filter conditions.
**Validates: Requirements 14.3, 19.5**

### Property 15: Statistics computation accuracy
*For any* set of incidents, computing success rate should equal (successful incidents / total incidents) and mean time to resolution should equal the average of (completed_at - created_at) for resolved incidents.
**Validates: Requirements 14.4**

### Property 16: Custom rule evaluation
*For any* incident and custom rule, if the incident matches the rule conditions, then the rule's actions (severity adjustment, metadata enrichment) should be applied to the incident.
**Validates: Requirements 16.2, 16.3**

### Property 17: Rule syntax validation
*For any* custom rule with syntax errors, attempting to load it should fail with a descriptive error message and not crash the service.
**Validates: Requirements 16.5**

### Property 18: MCP configuration merging
*For any* platform-level MCP configuration and repository-specific MCP configuration, merging them should result in repository-specific values overriding platform-level values for the same keys.
**Validates: Requirements 22.3**

### Property 19: MCP credential encryption
*For any* MCP credentials stored by the service, they should be encrypted at rest and only decrypted when passed to workflows.
**Validates: Requirements 22.4**

### Property 20: Notification content completeness
*For any* notification sent by the workflow, it should include incident severity, affected service, and pull request link.
**Validates: Requirements 23.5**

### Property 21: Dashboard incident ordering
*For any* list of incidents displayed in the dashboard, they should be ordered by timestamp with the most recent first.
**Validates: Requirements 19.1**

### Property 22: Dashboard incident display completeness
*For any* incident displayed in the dashboard, it should show status, affected service, error message, and associated repository.
**Validates: Requirements 19.2**

## Error Handling

### Incident Service Error Scenarios

1. **Invalid Webhook Payload**
   - Return 400 Bad Request with validation errors
   - Log the malformed payload for debugging
   - Do not create an incident record

2. **Database Connection Failure**
   - Return 503 Service Unavailable
   - Retry database operations with exponential backoff
   - Alert operators via metrics/logs

3. **GitHub API Failure**
   - Retry up to 3 times with exponential backoff
   - Mark incident as failed after exhausting retries
   - Store error details in incident record

4. **Unmapped Service**
   - Create incident with status "requires_manual_setup"
   - Send notification to platform operators
   - Do not trigger workflow

5. **Concurrency Limit Reached**
   - Queue incident for later processing
   - Update incident status to "queued"
   - Process when workflow completes

### GitHub Action Error Scenarios

1. **Kiro CLI Installation Failure**
   - Exit with non-zero status code
   - Report failure to Incident Service
   - Include installation logs in output

2. **MCP Server Connection Failure**
   - Log warning but continue with available data
   - Include MCP failure in post-mortem
   - Do not fail the entire remediation

3. **No Fix Generated**
   - Create incident report explaining why
   - Mark incident as "no_fix_needed" or "requires_human"
   - Do not create pull request

4. **Test Failures After Fix**
   - Include test output in post-mortem
   - Create pull request marked as draft
   - Clearly indicate tests are failing

5. **Pull Request Creation Failure**
   - Retry once
   - Report failure to Incident Service
   - Store branch name for manual PR creation

### Dashboard Error Scenarios

1. **API Request Failure**
   - Display user-friendly error message
   - Retry automatically after delay
   - Provide manual refresh button

2. **WebSocket Connection Loss**
   - Attempt reconnection with exponential backoff
   - Fall back to polling if WebSocket unavailable
   - Show connection status indicator

3. **Manual Trigger Failure**
   - Display specific error message
   - Allow retry
   - Log error for debugging

## Testing Strategy

### Unit Testing

**Incident Service:**
- HTTP handler functions with mock database
- Incident transformation logic for each provider
- Service-to-repository mapping logic
- Deduplication algorithm
- Configuration parsing
- GitHub API client with mock responses

**Dashboard:**
- React component rendering
- Filter logic
- API client functions
- State management

**GitHub Action:**
- Branch naming logic
- Post-mortem generation
- MCP configuration file creation
- Input validation

### Property-Based Testing

The platform will use property-based testing to verify universal properties across all inputs. Each property-based test will run a minimum of 100 iterations with randomly generated inputs.

**Testing Framework:** 
- Go: `gopter` (property-based testing library)
- TypeScript: `fast-check` (property-based testing library)

**Property Test Examples:**

1. **Incident Persistence Round-Trip** (Property 1)
   - Generate random valid incidents
   - Store to database
   - Query by ID
   - Assert retrieved incident equals original

2. **Provider Transformation** (Property 2)
   - Generate random provider payloads
   - Transform to internal format
   - Assert all required fields present and valid

3. **Deduplication** (Property 5)
   - Generate pairs of similar incidents
   - Send within time window
   - Assert only one record exists

4. **Concurrency Enforcement** (Property 8)
   - Generate N+M incidents for same repository (N = limit)
   - Dispatch all
   - Assert N active, M queued

5. **Configuration Merging** (Property 18)
   - Generate random platform and repo configs
   - Merge them
   - Assert repo values override platform values

### Integration Testing

**End-to-End Flow:**
1. Send incident webhook to Incident Service
2. Verify incident stored in database
3. Verify workflow dispatch API called
4. Mock workflow completion webhook
5. Verify incident status updated
6. Verify dashboard displays updated status

**GitHub Action Testing:**
1. Create test repository with sample code containing a bug
2. Trigger action with incident data
3. Verify Kiro CLI installed
4. Verify MCP configuration created
5. Verify branch created and PR opened
6. Verify post-mortem includes required sections

### Performance Testing

**Incident Service:**
- Webhook endpoint should respond within 500ms (p99)
- Support 100 concurrent webhook requests
- Process 1000 incidents per minute

**Dashboard:**
- Initial page load under 2 seconds
- Incident list rendering under 500ms for 1000 incidents
- Real-time updates within 1 second of status change

## Deployment Architecture

### Docker Architecture

The platform uses a multi-container architecture with Docker Compose for local development and production deployment.

**Container Structure:**

```
ai-sre-platform/
├── incident-service/
│   ├── Dockerfile
│   └── .dockerignore
├── dashboard/
│   ├── Dockerfile
│   └── .dockerignore
├── demo-app/
│   ├── Dockerfile
│   └── .dockerignore
├── docker-compose.yml
├── docker-compose.dev.yml
└── docker-compose.prod.yml
```

### Dockerfiles

**1. Incident Service Dockerfile**

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o incident-service ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/incident-service .

# Copy config template
COPY --from=builder /app/config.example.yaml ./config.yaml

EXPOSE 8080 9090

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

CMD ["./incident-service"]
```

**2. Dashboard Dockerfile**

```dockerfile
# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source code
COPY . .

# Build for production
RUN npm run build

# Runtime stage
FROM nginx:alpine

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
```

**3. Demo App Dockerfile**

```dockerfile
FROM node:20-alpine

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci --only=production

# Copy source code
COPY . .

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

CMD ["node", "server.js"]
```

### Docker Compose Configurations

**Development (`docker-compose.dev.yml`):**

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ai_sre_dev
      POSTGRES_USER: ai_sre
      POSTGRES_PASSWORD: dev_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./incident-service/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ai_sre"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  incident-service:
    build:
      context: ./incident-service
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - DATABASE_URL=postgres://ai_sre:dev_password@postgres:5432/ai_sre_dev?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - GITHUB_TOKEN=${GITHUB_TOKEN}
      - LOG_LEVEL=debug
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./incident-service/config.yaml:/root/config.yaml
    restart: unless-stopped

  dashboard:
    build:
      context: ./dashboard
      dockerfile: Dockerfile
    ports:
      - "3001:80"
    environment:
      - VITE_API_URL=http://localhost:8080
    depends_on:
      - incident-service
    restart: unless-stopped

  demo-app:
    build:
      context: ./demo-app
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - SENTRY_DSN=${SENTRY_DSN}
      - NODE_ENV=development
    restart: unless-stopped

volumes:
  postgres_data:
```

**Production (`docker-compose.prod.yml`):**

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: always

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "--pass", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: always

  incident-service:
    image: ghcr.io/your-org/ai-sre-incident-service:latest
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=require
      - REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379
      - GITHUB_TOKEN=${GITHUB_TOKEN}
      - LOG_LEVEL=info
      - SENTRY_DSN=${SENTRY_DSN}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./config.prod.yaml:/root/config.yaml:ro
    restart: always
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M

  dashboard:
    image: ghcr.io/your-org/ai-sre-dashboard:latest
    ports:
      - "80:80"
      - "443:443"
    environment:
      - VITE_API_URL=${API_URL}
    volumes:
      - ./nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - incident-service
    restart: always

  demo-app:
    image: ghcr.io/your-org/ai-sre-demo-app:latest
    ports:
      - "3000:3000"
    environment:
      - SENTRY_DSN=${DEMO_SENTRY_DSN}
      - NODE_ENV=production
    restart: always

volumes:
  postgres_data:
  redis_data:
```

### Quick Start Scripts

**`scripts/dev.sh`:**

```bash
#!/bin/bash
set -e

echo "🚀 Starting AI SRE Platform (Development)"

# Check if .env exists
if [ ! -f .env ]; then
    echo "📝 Creating .env from .env.example"
    cp .env.example .env
    echo "⚠️  Please edit .env with your configuration"
    exit 1
fi

# Load environment variables
export $(cat .env | xargs)

# Start services
docker-compose -f docker-compose.dev.yml up --build

echo "✅ Platform is running!"
echo "📊 Dashboard: http://localhost:3001"
echo "🔧 API: http://localhost:8080"
echo "🎯 Demo App: http://localhost:3000"
echo "📈 Metrics: http://localhost:9090/metrics"
```

**`scripts/prod.sh`:**

```bash
#!/bin/bash
set -e

echo "🚀 Deploying AI SRE Platform (Production)"

# Check if .env.prod exists
if [ ! -f .env.prod ]; then
    echo "❌ .env.prod not found"
    exit 1
fi

# Load environment variables
export $(cat .env.prod | xargs)

# Pull latest images
docker-compose -f docker-compose.prod.yml pull

# Start services
docker-compose -f docker-compose.prod.yml up -d

echo "✅ Platform deployed!"
echo "🔍 Check status: docker-compose -f docker-compose.prod.yml ps"
echo "📋 View logs: docker-compose -f docker-compose.prod.yml logs -f"
```

**`scripts/test.sh`:**

```bash
#!/bin/bash
set -e

echo "🧪 Running tests"

# Run incident service tests
echo "Testing incident service..."
cd incident-service
go test -v -race -coverprofile=coverage.out ./...
cd ..

# Run dashboard tests
echo "Testing dashboard..."
cd dashboard
npm test -- --coverage
cd ..

# Run demo app tests
echo "Testing demo app..."
cd demo-app
npm test
cd ..

echo "✅ All tests passed!"
```

### CI/CD Pipeline

**`.github/workflows/ci.yml`:**

```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test-incident-service:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        working-directory: ./incident-service
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./incident-service/coverage.out

  test-dashboard:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      
      - name: Install dependencies
        working-directory: ./dashboard
        run: npm ci
      
      - name: Run tests
        working-directory: ./dashboard
        run: npm test -- --coverage
      
      - name: Build
        working-directory: ./dashboard
        run: npm run build

  test-demo-app:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      
      - name: Install dependencies
        working-directory: ./demo-app
        run: npm ci
      
      - name: Run tests
        working-directory: ./demo-app
        run: npm test

  build-images:
    needs: [test-incident-service, test-dashboard, test-demo-app]
    runs-on: ubuntu-latest
    if: github.event_name == 'push'
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Build and push incident-service
        uses: docker/build-push-action@v5
        with:
          context: ./incident-service
          push: true
          tags: ghcr.io/${{ github.repository }}/incident-service:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
      
      - name: Build and push dashboard
        uses: docker/build-push-action@v5
        with:
          context: ./dashboard
          push: true
          tags: ghcr.io/${{ github.repository }}/dashboard:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
      
      - name: Build and push demo-app
        uses: docker/build-push-action@v5
        with:
          context: ./demo-app
          push: true
          tags: ghcr.io/${{ github.repository }}/demo-app:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

### Environment Configuration

**`.env.example`:**

```bash
# Database
POSTGRES_DB=ai_sre_dev
POSTGRES_USER=ai_sre
POSTGRES_PASSWORD=dev_password

# Redis
REDIS_PASSWORD=dev_redis_password

# GitHub
GITHUB_TOKEN=ghp_your_token_here

# Sentry (for demo app)
SENTRY_DSN=https://xxx@sentry.io/xxx
DEMO_SENTRY_DSN=https://xxx@sentry.io/xxx

# API URLs
API_URL=http://localhost:8080

# Observability Platform Secrets
DATADOG_API_KEY=your_datadog_api_key
DATADOG_APP_KEY=your_datadog_app_key
```

### Container Images

1. **Incident Service**
   - Base: `golang:1.21-alpine` (build), `alpine:latest` (runtime)
   - Size: ~20MB
   - Exposed ports: 8080 (HTTP), 9090 (metrics)
   - Health check: `GET /api/v1/health`

2. **Dashboard**
   - Base: `node:20-alpine` (build), `nginx:alpine` (runtime)
   - Size: ~25MB
   - Exposed ports: 80 (HTTP)
   - Static files served from `/usr/share/nginx/html`

3. **Demo App**
   - Base: `node:20-alpine`
   - Size: ~150MB
   - Exposed ports: 3000 (HTTP)
   - Health check: `GET /health`

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: incident-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: incident-service
  template:
    metadata:
      labels:
        app: incident-service
    spec:
      containers:
      - name: incident-service
        image: ai-sre-platform/incident-service:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: incident-service-secrets
              key: database-url
        - name: GITHUB_TOKEN
          valueFrom:
            secretKeyRef:
              name: incident-service-secrets
              key: github-token
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### Database Schema

```sql
CREATE TABLE incidents (
    id VARCHAR(255) PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    repository VARCHAR(255),
    error_message TEXT NOT NULL,
    stack_trace TEXT,
    severity VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    provider VARCHAR(100) NOT NULL,
    provider_data JSONB,
    workflow_run_id BIGINT,
    pull_request_url TEXT,
    diagnosis TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    triggered_at TIMESTAMP,
    completed_at TIMESTAMP,
    INDEX idx_service_name (service_name),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at DESC)
);

CREATE TABLE incident_events (
    id SERIAL PRIMARY KEY,
    incident_id VARCHAR(255) NOT NULL REFERENCES incidents(id),
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    INDEX idx_incident_id (incident_id),
    INDEX idx_created_at (created_at DESC)
);

CREATE TABLE service_mappings (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(255) UNIQUE NOT NULL,
    repository VARCHAR(255) NOT NULL,
    branch VARCHAR(255) NOT NULL DEFAULT 'main',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE mcp_servers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(100) NOT NULL,
    config JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Security Considerations

### Authentication & Authorization

1. **Incident Service**
   - Webhook endpoints: Validate HMAC signatures from observability providers
   - Dashboard API: JWT-based authentication
   - GitHub API: Use fine-grained personal access tokens with minimal scopes

2. **GitHub Action**
   - Use `GITHUB_TOKEN` with repository-scoped permissions
   - Read MCP credentials from repository secrets
   - Never log sensitive credentials

### Data Protection

1. **Encryption at Rest**
   - Database: Enable PostgreSQL encryption
   - Secrets: Use operator-provided encryption keys
   - MCP credentials: Encrypt before storing

2. **Encryption in Transit**
   - All HTTP communication over TLS 1.3
   - GitHub API calls over HTTPS
   - Database connections over SSL

### Network Security

1. **Incident Service**
   - Expose only necessary ports (8080, 9090)
   - Use network policies to restrict egress
   - Rate limit webhook endpoints

2. **Dashboard**
   - Serve over HTTPS only
   - Implement CSP headers
   - Use SameSite cookies

### Secrets Management

1. **Supported Systems**
   - Environment variables
   - Kubernetes secrets
   - HashiCorp Vault (optional)
   - AWS Secrets Manager (optional)

2. **Best Practices**
   - Rotate credentials regularly
   - Use least-privilege access
   - Audit secret access

## Monitoring & Observability

### Metrics (Prometheus Format)

```
# Incident ingestion
ai_sre_incidents_received_total{provider="datadog",status="success"}
ai_sre_incidents_received_total{provider="datadog",status="error"}

# Workflow dispatch
ai_sre_workflows_dispatched_total{repository="org/repo",status="success"}
ai_sre_workflows_dispatched_total{repository="org/repo",status="error"}
ai_sre_workflows_active{repository="org/repo"}
ai_sre_workflows_queued{repository="org/repo"}

# Incident resolution
ai_sre_incidents_resolved_total{repository="org/repo",outcome="pr_created"}
ai_sre_incidents_resolved_total{repository="org/repo",outcome="failed"}
ai_sre_incident_resolution_duration_seconds{repository="org/repo"}

# System health
ai_sre_database_connections_active
ai_sre_http_requests_total{endpoint="/api/v1/incidents",method="GET",status="200"}
ai_sre_http_request_duration_seconds{endpoint="/api/v1/incidents",method="GET"}
```

### Structured Logging

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "message": "Incident received",
  "incident_id": "inc_123456",
  "service_name": "api-gateway",
  "provider": "datadog",
  "severity": "high"
}
```

### Alerting Rules

1. **High Error Rate**
   - Alert if incident ingestion error rate > 10% over 5 minutes
   - Severity: Warning

2. **Workflow Dispatch Failures**
   - Alert if workflow dispatch failure rate > 20% over 5 minutes
   - Severity: Critical

3. **Database Connection Issues**
   - Alert if database connection failures > 5 in 1 minute
   - Severity: Critical

4. **High Queue Depth**
   - Alert if queued incidents > 50 for any repository
   - Severity: Warning

## Technology Stack Summary

### Incident Service
- **Language:** Go 1.21+
- **Framework:** Chi (HTTP router)
- **Database:** PostgreSQL 15+
- **Cache:** Redis 7+
- **Testing:** `testing`, `gopter` (PBT)

### Dashboard
- **Framework:** React 18 with TypeScript
- **State Management:** TanStack Query
- **UI Library:** shadcn/ui (Tailwind CSS)
- **Build Tool:** Vite
- **Testing:** Vitest, React Testing Library, `fast-check` (PBT)

### GitHub Action
- **Runtime:** Node.js 20
- **Language:** TypeScript
- **Testing:** Jest, `fast-check` (PBT)

### Infrastructure
- **Container Runtime:** Docker
- **Orchestration:** Kubernetes
- **Package Manager:** Helm
- **CI/CD:** GitHub Actions

## Demo Application

To demonstrate the platform's capabilities and provide a testing environment, we'll include a demo application within the repository.

### Demo App Architecture

```mermaid
graph TB
    USER[User/Demo UI]
    DEMO[Demo App API]
    SENTRY[Sentry]
    INCIDENT[Incident Service]
    GHA[GitHub Actions]
    
    USER -->|Trigger Error| DEMO
    DEMO -->|Report Error| SENTRY
    SENTRY -->|Webhook| INCIDENT
    INCIDENT -->|Dispatch Workflow| GHA
    GHA -->|Create PR| DEMO
    
    USER -->|View Status| INCIDENT
```

### Demo Application Components

**1. Demo Service (`/demo-app`)**

A simple Node.js/Express API with intentionally buggy endpoints:

```javascript
// Buggy endpoints that can be triggered
app.get('/api/divide', (req, res) => {
  const { a, b } = req.query;
  // Bug: No validation, will crash on b=0
  const result = a / b;
  res.json({ result });
});

app.get('/api/user/:id', async (req, res) => {
  const user = await db.getUser(req.params.id);
  // Bug: No null check, will crash if user not found
  res.json({ name: user.name });
});

app.post('/api/process', (req, res) => {
  const data = req.body.data;
  // Bug: No validation, will crash on undefined
  const processed = data.map(item => item.value * 2);
  res.json({ processed });
});
```

**Features:**
- Multiple buggy endpoints with different error types
- Integrated with Sentry for error reporting
- Health check endpoint
- Metrics endpoint
- Simple in-memory database (or SQLite)

**2. Demo UI (`/demo-app/public`)**

A simple web interface for triggering errors and viewing results:

```
┌─────────────────────────────────────┐
│  AI SRE Platform Demo               │
├─────────────────────────────────────┤
│                                     │
│  Trigger Test Incidents:            │
│  ┌─────────────────────────────┐   │
│  │ [Division by Zero]          │   │
│  │ [Null Pointer Error]        │   │
│  │ [Array Processing Error]    │   │
│  └─────────────────────────────┘   │
│                                     │
│  Recent Incidents:                  │
│  ┌─────────────────────────────┐   │
│  │ ⏱️  2 min ago - Division     │   │
│  │ 🔄 Workflow Running...       │   │
│  │                              │   │
│  │ ✅ 5 min ago - Null Pointer  │   │
│  │ 🔗 PR #123 Created           │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

**Features:**
- Buttons to trigger each type of error
- Real-time incident status updates
- Links to GitHub PRs when created
- Links to Sentry issues
- Embedded dashboard showing incident flow

**3. Demo Workflow (`.github/workflows/demo-remediate.yml`)**

The demo app has the remediation workflow installed:

```yaml
name: Demo Remediation

on:
  workflow_dispatch:
    inputs:
      incident_id:
        required: true
      error_message:
        required: true
      stack_trace:
        required: false
      service_name:
        required: true
      timestamp:
        required: true

jobs:
  remediate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: ./remediation-action  # Use local action for testing
        with:
          incident_id: ${{ inputs.incident_id }}
          error_message: ${{ inputs.error_message }}
          stack_trace: ${{ inputs.stack_trace }}
          service_name: ${{ inputs.service_name }}
          timestamp: ${{ inputs.timestamp }}
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SENTRY_DSN: ${{ secrets.SENTRY_DSN }}
```

**4. Demo Kiro Specs (`.kiro/specs/demo-fixes`)**

Pre-configured remediation strategies for common demo errors:

```markdown
# Demo Fix Strategies

## Division by Zero
When you encounter a division by zero error:
1. Add input validation to check if divisor is zero
2. Return appropriate error response
3. Add unit test for zero divisor case

## Null Pointer Errors
When you encounter null/undefined access:
1. Add null check before accessing properties
2. Return 404 if resource not found
3. Add unit test for missing resource case

## Array Processing Errors
When you encounter array processing errors:
1. Add validation that data exists and is an array
2. Return 400 Bad Request for invalid input
3. Add unit test for invalid input case
```

### Demo Deployment

**Local Development:**
```bash
# Start the demo app
cd demo-app
npm install
npm run dev

# Start the incident service
cd incident-service
go run main.go

# Start the dashboard
cd dashboard
npm install
npm run dev
```

**Public Demo (Hosted):**
- Deploy demo app to Vercel/Railway
- Deploy incident service to Fly.io/Railway
- Deploy dashboard to Vercel
- Configure Sentry project for demo app
- Public URL: `https://demo.ai-sre-platform.dev`

### Demo Flow

1. **User visits demo site**
   - Sees the demo UI with error trigger buttons
   - Sees real-time incident dashboard

2. **User clicks "Trigger Division by Zero"**
   - Demo app makes request to `/api/divide?a=10&b=0`
   - App crashes and reports to Sentry
   - Sentry sends webhook to Incident Service

3. **Incident Service processes webhook**
   - Creates incident record
   - Maps to demo-app repository
   - Triggers GitHub Actions workflow

4. **GitHub Actions runs remediation**
   - Kiro CLI analyzes the error
   - Generates fix (adds validation)
   - Runs tests
   - Creates PR

5. **User sees results**
   - Dashboard updates to show "PR Created"
   - Link to PR appears in demo UI
   - User can click to see the generated fix

### Demo Configuration

**Environment Variables:**
```bash
# Demo App
SENTRY_DSN=https://xxx@sentry.io/xxx
PORT=3000

# Incident Service
DEMO_MODE=true
GITHUB_TOKEN=ghp_xxx
SENTRY_WEBHOOK_SECRET=xxx
DATABASE_URL=postgres://localhost/ai_sre_demo

# Dashboard
VITE_API_URL=http://localhost:8080
VITE_DEMO_MODE=true
```

### Testing Scenarios

The demo app includes these testable scenarios:

1. **Division by Zero** - Arithmetic error
2. **Null Pointer** - Missing data access
3. **Array Processing** - Invalid input handling
4. **Database Connection** - Simulated connection failure
5. **Memory Leak** - Simulated resource exhaustion
6. **Timeout** - Simulated slow operation

Each scenario:
- Has a known bug
- Has a clear fix
- Can be triggered via UI
- Reports to Sentry
- Results in a PR with the fix

### Demo Benefits

**For Development:**
- End-to-end testing without external dependencies
- Rapid iteration on remediation logic
- Validate webhook integrations
- Test GitHub Actions locally

**For Demos/Hackathon:**
- Live demonstration of full flow
- Audience can trigger incidents themselves
- Shows real PRs being created
- Proves the concept works end-to-end

**For Documentation:**
- Concrete examples of fixes
- Reference implementation for users
- Tutorial for setting up their own services

## Future Enhancements

1. **Multi-Repository Fixes**
   - Support incidents that require changes across multiple repositories
   - Coordinate PRs and ensure they're merged in correct order

2. **Learning from Resolutions**
   - Track which types of incidents are successfully auto-remediated
   - Use historical data to improve fix generation

3. **Custom Remediation Strategies**
   - Allow repositories to define custom Kiro specs for common failure patterns
   - Template library for common issues (OOM, deadlocks, etc.)

4. **Incident Correlation**
   - Detect cascading failures across services
   - Group related incidents into a single remediation effort

5. **Rollback Detection**
   - Monitor if a PR causes new incidents
   - Automatically create revert PR if issues detected

6. **Cost Tracking**
   - Track GitHub Actions minutes consumed
   - Track AI API costs per incident
   - Provide cost optimization recommendations
