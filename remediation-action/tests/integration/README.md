# Integration Tests for Remediation Action

This directory contains integration tests for the AI SRE Remediation GitHub Action using [nektos/act](https://github.com/nektos/act).

## Overview

These tests verify the remediation workflow executes correctly by:
- Testing workflow structure and configuration
- Validating incident context file creation
- Verifying branch naming conventions
- Testing MCP configuration handling
- Running workflows locally with act (dry-run mode)

## Prerequisites

### Install nektos/act

**macOS (Homebrew):**
```bash
brew install act
```

**Linux:**
```bash
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

**Manual Installation:**
See [act installation guide](https://github.com/nektos/act#installation)

### Install jq (for JSON processing)

**macOS:**
```bash
brew install jq
```

**Linux:**
```bash
sudo apt-get install jq  # Debian/Ubuntu
sudo yum install jq      # RHEL/CentOS
```

### Build the Action

Before running tests, build the action:
```bash
cd remediation-action
npm install
npm run build
```

## Running Tests

### Run All Integration Tests

```bash
cd remediation-action/tests/integration
./run-act-tests.sh
```

### Run Workflow Execution Tests

These tests verify the workflow structure and configuration without requiring act:

```bash
cd remediation-action/tests/integration
./test-workflow-execution.sh
```

## Test Structure

```
tests/
├── fixtures/
│   ├── incident-events/          # Sample incident payloads
│   │   ├── simple-error.json
│   │   ├── division-by-zero.json
│   │   └── null-pointer.json
│   └── test-repo/                # Test repository with buggy code
│       ├── src/
│       │   ├── users.js          # Bug: null pointer errors
│       │   └── math.js           # Bug: division by zero
│       ├── .github/workflows/
│       │   └── remediate.yml     # Test workflow
│       ├── .kiro/settings/
│       │   └── mcp.json          # MCP configuration
│       └── package.json
└── integration/
    ├── run-act-tests.sh          # Main test runner using act
    ├── test-workflow-execution.sh # Workflow structure tests
    └── README.md                 # This file
```

## Test Scenarios

### 1. Simple TypeError Remediation
- **Incident:** `simple-error.json`
- **Bug:** Null pointer in `users.js`
- **Expected:** Workflow creates context file and prepares branch

### 2. Division by Zero
- **Incident:** `division-by-zero.json`
- **Bug:** Empty array handling in `math.js`
- **Expected:** Workflow identifies math error

### 3. Null Pointer Exception
- **Incident:** `null-pointer.json`
- **Bug:** Missing null check in user lookup
- **Expected:** Workflow handles null reference error

### 4. Error Handling Tests
- Missing Kiro CLI installation
- Git operation failures
- Invalid MCP configuration

## Test Output

Successful test run:
```
==========================================
AI SRE Remediation Action - Integration Tests
==========================================

✓ nektos/act is installed

Running integration tests...

----------------------------------------
Test: Simple TypeError remediation
----------------------------------------
Incident ID: inc_test_001
Service: user-service
Severity: high

Running workflow with act (dry-run)...
✓ PASSED: Simple TypeError remediation

==========================================
Test Summary
==========================================
Passed: 3
Failed: 0
Total:  3

All tests passed!
```

## Continuous Integration

These tests are integrated into the CI/CD pipeline in `.github/workflows/ci.yml`:

```yaml
- name: Run integration tests
  run: |
    cd remediation-action/tests/integration
    ./test-workflow-execution.sh
```

Note: Full act-based tests with workflow execution are run separately due to Docker requirements.

## Troubleshooting

### act not found
Install act using the instructions above.

### jq not found
Install jq using your package manager.

### Action not built
Run `npm run build` in the `remediation-action` directory.

### Docker errors with act
act requires Docker to run workflows. Ensure Docker is installed and running:
```bash
docker --version
docker ps
```

### Permission denied
Make test scripts executable:
```bash
chmod +x remediation-action/tests/integration/*.sh
```

## Adding New Tests

To add a new test scenario:

1. Create incident fixture in `fixtures/incident-events/`:
```json
{
  "incident_id": "inc_test_004",
  "error_message": "Your error message",
  "stack_trace": "Stack trace here",
  "service_name": "your-service",
  "timestamp": "2024-01-15T10:00:00Z",
  "severity": "high"
}
```

2. Add buggy code to `fixtures/test-repo/src/` if needed

3. Add test case to `run-act-tests.sh`:
```bash
run_test \
    "Your test name" \
    "$FIXTURES_DIR/incident-events/your-incident.json" \
    "Expected behavior description"
```

## Limitations

- **Dry-run mode:** Tests use act's dry-run mode to avoid actual GitHub API calls
- **Kiro CLI:** Tests don't execute actual Kiro CLI (would require full environment)
- **PR creation:** Tests verify workflow structure but don't create real PRs
- **MCP servers:** Tests use mock MCP configuration

For full end-to-end testing, use the demo application with real GitHub Actions.

## References

- [nektos/act documentation](https://github.com/nektos/act)
- [GitHub Actions workflow syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [AI SRE Platform documentation](../../README.md)
