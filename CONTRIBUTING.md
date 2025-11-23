# Contributing to AI SRE Platform

Thank you for your interest in contributing to the AI SRE Platform! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.21+
- Node.js 20+
- Docker and Docker Compose
- Git

### Getting Started

1. Fork the repository
2. Clone your fork:
```bash
git clone https://github.com/your-username/ai-sre-platform.git
cd ai-sre-platform
```

3. Set up your development environment:
```bash
cp .env.example .env
# Edit .env with your credentials
./scripts/dev.sh
```

## Project Structure

```
ai-sre-platform/
├── incident-service/     # Go backend service
├── dashboard/            # React web UI
├── demo-app/            # Demo application
├── remediation-action/  # GitHub Action
├── scripts/             # Development scripts
└── .github/workflows/   # CI/CD pipelines
```

## Development Workflow

### Making Changes

1. Create a new branch:
```bash
git checkout -b feature/your-feature-name
```

2. Make your changes following the coding standards below

3. Run tests:
```bash
./scripts/test.sh
```

4. Commit your changes:
```bash
git add .
git commit -m "feat: add your feature description"
```

5. Push to your fork:
```bash
git push origin feature/your-feature-name
```

6. Create a Pull Request

### Commit Message Format

We follow the Conventional Commits specification:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `test:` - Adding or updating tests
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks

## Coding Standards

### Go (Incident Service)

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Add comments for exported functions
- Write table-driven tests
- Use property-based testing for universal properties

Example:
```go
// ProcessIncident transforms and stores an incident
func ProcessIncident(ctx context.Context, incident *Incident) error {
    // Implementation
}
```

### TypeScript/JavaScript

- Use TypeScript for type safety
- Follow ESLint rules
- Use functional components in React
- Write descriptive test names

Example:
```typescript
export function IncidentCard({ incident }: IncidentCardProps) {
  // Implementation
}
```

### Testing

- Write unit tests for all new functions
- Write property-based tests for universal properties
- Aim for >80% code coverage
- Test error cases and edge cases

Property-based test example:
```go
// Property 1: Incident persistence round-trip
// For any valid incident, storing and retrieving should return equivalent data
func TestIncidentPersistenceRoundTrip(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("round trip preserves incident data", 
        prop.ForAll(
            func(incident *Incident) bool {
                stored := store.Save(incident)
                retrieved := store.Get(stored.ID)
                return reflect.DeepEqual(stored, retrieved)
            },
            genIncident(),
        ),
    )
    
    properties.TestingRun(t, gopter.ConsoleReporter(t))
}
```

## Adding New Features

### Adding a New Webhook Adapter

1. Create adapter in `incident-service/internal/adapters/`
2. Implement the `WebhookAdapter` interface
3. Add validation logic
4. Map provider fields to internal Incident struct
5. Register in `NewIncidentService()`
6. Add tests
7. Update documentation

### Adding New Dashboard Pages

1. Create page component in `dashboard/src/pages/`
2. Add route in router configuration
3. Create API client functions in `dashboard/src/api/`
4. Add tests
5. Update navigation

### Modifying the Demo UI

The demo application includes an interactive web UI at `demo-app/public/index.html`.

**Key Components:**

- **View Switching**: Three-tab interface (Bugs, Incidents, Dashboard)
- **Auto-Refresh**: 5-second polling of incident service API
- **Real-Time Updates**: Incident status updates without page reload
- **Dynamic Links**: PR and observability platform links rendered based on incident data

**Adding a New Bug Scenario:**

1. Add buggy endpoint to `demo-app/src/routes/buggy.js`
2. Add trigger button to `demo-app/public/index.html`:
```html
<div class="card">
  <span class="bug-type">Your Bug Type</span>
  <h2>Bug Name</h2>
  <p>Description of the bug</p>
  <button onclick="triggerBugN()">Trigger Bug</button>
  <div id="responseN" class="response"></div>
</div>
```
3. Add trigger function in the `<script>` section:
```javascript
async function triggerBugN() {
  try {
    const response = await fetch('/api/buggy/your-endpoint');
    const data = await response.json();
    showResponse('responseN', data, !response.ok);
  } catch (error) {
    showResponse('responseN', { error: error.message }, true);
  }
}
```
4. Update `demo-app/README.md` with bug details
5. Test the end-to-end flow

**Testing the Demo UI:**

1. Start the full stack: `./scripts/dev.sh`
2. Open http://localhost:3000
3. Trigger bugs and verify:
   - Error responses display correctly
   - Incidents appear in the sidebar within 5 seconds
   - Status updates reflect in real-time
   - Links to PRs appear when available
   - Dashboard iframe loads correctly

## Running Tests

### All Tests
```bash
./scripts/test.sh
```

### Incident Service
```bash
cd incident-service
go test -v -race ./...
```

### Dashboard
```bash
cd dashboard
npm test -- --run
```

### Demo App
```bash
cd demo-app
npm test -- --run
```

### Remediation Action

#### Unit Tests
```bash
cd remediation-action
npm test
```

#### Integration Tests

The remediation action includes integration tests that verify the GitHub Actions workflow executes correctly using [nektos/act](https://github.com/nektos/act).

**Prerequisites:**

Install nektos/act:
```bash
# macOS
brew install act

# Linux
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

Install jq (JSON processor):
```bash
# macOS
brew install jq

# Linux (Debian/Ubuntu)
sudo apt-get install jq
```

**Running Integration Tests:**

```bash
cd remediation-action

# Build the action first
npm run build

# Run workflow structure tests (no act required)
cd tests/integration
./test-workflow-execution.sh

# Run full integration tests with act (requires Docker)
./run-act-tests.sh
```

**What the Integration Tests Verify:**

- Workflow file structure and configuration
- Incident context file creation
- Branch naming conventions (includes incident ID)
- MCP configuration handling
- Test repository fixtures with buggy code
- Error handling scenarios

**Test Fixtures:**

The integration tests use sample incident events and a test repository with intentional bugs:

- `tests/fixtures/incident-events/` - Sample incident payloads
- `tests/fixtures/test-repo/` - Test repository with buggy code (null pointers, division by zero)

See [remediation-action/tests/integration/README.md](remediation-action/tests/integration/README.md) for detailed documentation.

## CI/CD Pipeline

The project uses GitHub Actions for continuous integration and deployment. The pipeline automatically runs on every push and pull request.

### Pipeline Overview

The CI pipeline consists of the following stages:

1. **Linting** - Code quality checks
   - Go: `golangci-lint` for incident-service
   - TypeScript/JavaScript: ESLint for dashboard, demo-app, and remediation-action

2. **Testing** - Automated test execution
   - Incident Service: Go tests with race detector, PostgreSQL and Redis services
   - Dashboard: Vitest tests with coverage
   - Demo App: Vitest tests with coverage
   - Remediation Action: Jest tests + integration tests

3. **Coverage Reporting** - Code coverage uploaded to Codecov
   - Target: >80% overall coverage
   - Critical paths: 100% coverage (webhooks, remediation, database)

4. **Docker Image Building** - Multi-stage builds (main branch only)
   - Incident Service image
   - Dashboard image
   - Demo App image
   - Images pushed to GitHub Container Registry (ghcr.io)

### Workflow File

The main CI workflow is defined in `.github/workflows/ci.yml`.

### Running CI Locally

**Lint Go Code:**
```bash
cd incident-service
golangci-lint run --timeout=5m
```

**Lint TypeScript/JavaScript:**
```bash
# Dashboard
cd dashboard && npm run lint

# Demo App
cd demo-app && npm run lint

# Remediation Action
cd remediation-action && npm run lint
```

**Run All Tests:**
```bash
./scripts/test.sh
```

**Build Docker Images Locally:**
```bash
# Incident Service
docker build -t incident-service:local ./incident-service

# Dashboard
docker build -t dashboard:local ./dashboard

# Demo App
docker build -t demo-app:local ./demo-app
```

### CI Requirements for Pull Requests

Before your PR can be merged:

1. ✅ All linting checks must pass
2. ✅ All tests must pass (unit, property-based, integration)
3. ✅ Code coverage must not decrease
4. ✅ Docker images must build successfully (for main branch)
5. ✅ At least one code review approval

### Viewing CI Results

- Check the "Actions" tab in GitHub to see workflow runs
- Click on a specific run to see detailed logs
- Failed jobs will show error messages and stack traces
- Coverage reports are available in Codecov

### Manual Workflow Triggering

You can manually trigger the CI pipeline:

1. Go to the "Actions" tab
2. Select "CI Pipeline"
3. Click "Run workflow"
4. Choose the branch and click "Run workflow"

### Troubleshooting CI Failures

**Linting Failures:**
- Run linters locally to see the issues
- Fix formatting and code quality issues
- Commit and push the fixes

**Test Failures:**
- Check the test logs in the Actions tab
- Run tests locally to reproduce: `./scripts/test.sh`
- Fix the failing tests and push

**Docker Build Failures:**
- Verify Dockerfiles are correct
- Test builds locally: `docker build -t test ./service-name`
- Check for missing dependencies or incorrect paths

**Coverage Drops:**
- Add tests for new code
- Ensure property-based tests run 100+ iterations
- Check coverage locally: `go test -coverprofile=coverage.out ./...`

## Code Review Process

1. All changes require a pull request
2. At least one approval required
3. All CI checks must pass
4. Code coverage should not decrease
5. Documentation must be updated

## Reporting Issues

When reporting issues, please include:

- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, versions)
- Relevant logs or error messages

## Questions?

- Open a GitHub Discussion
- Check existing issues and PRs
- Review the documentation

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
