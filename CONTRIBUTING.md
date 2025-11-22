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

## Code Review Process

1. All changes require a pull request
2. At least one approval required
3. All tests must pass
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
