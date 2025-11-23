# CI/CD Pipeline Documentation

This directory contains GitHub Actions workflows for the AI SRE Platform.

## Workflows

### CI Pipeline (`ci.yml`)

The main continuous integration pipeline that runs on every push and pull request.

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches
- Manual workflow dispatch

**Jobs:**

#### 1. Lint Go Code (`lint-go`)
- Runs `golangci-lint` on the incident-service
- Checks code quality, formatting, and common issues
- Timeout: 5 minutes

#### 2. Lint TypeScript/JavaScript (`lint-typescript`)
- Runs ESLint on dashboard, demo-app, and remediation-action
- Uses matrix strategy to run in parallel
- Checks code quality and formatting

#### 3. Test Incident Service (`test-incident-service`)
- Depends on: `lint-go`
- Services: PostgreSQL 15, Redis 7
- Runs Go tests with race detector
- Generates code coverage report
- Uploads coverage to Codecov

**Environment:**
```
DATABASE_URL: postgresql://postgres:postgres@localhost:5432/ai_sre_test?sslmode=disable
REDIS_URL: redis://localhost:6379
```

#### 4. Test Dashboard (`test-dashboard`)
- Depends on: `lint-typescript`
- Runs Vitest tests with coverage
- Uploads coverage to Codecov

#### 5. Test Demo App (`test-demo-app`)
- Depends on: `lint-typescript`
- Runs Vitest tests with coverage
- Uploads coverage to Codecov

#### 6. Test Remediation Action (`test-remediation-action`)
- Depends on: `lint-typescript`
- Builds the action
- Runs Jest unit tests
- Runs integration tests with test fixtures
- Uploads coverage to Codecov

#### 7. Build Docker Images (`build-images`)
- Depends on: All test jobs
- Only runs on: Push to `main` branch
- Builds and pushes images to GitHub Container Registry (ghcr.io)
- Uses Docker Buildx for multi-platform builds
- Implements layer caching for faster builds

**Images Built:**
- `ghcr.io/<owner>/<repo>/incident-service:latest`
- `ghcr.io/<owner>/<repo>/dashboard:latest`
- `ghcr.io/<owner>/<repo>/demo-app:latest`

**Image Tags:**
- `latest` - Latest build from main branch
- `main-<sha>` - Specific commit SHA
- `main` - Branch name

#### 8. CI Success (`ci-success`)
- Depends on: All lint and test jobs
- Always runs (even if previous jobs fail)
- Checks status of all jobs
- Fails if any job failed
- Provides clear success/failure message

### Demo Remediation Workflow (`demo-remediate.yml`)

Workflow for demonstrating the automated remediation process in the demo application.

**Triggers:**
- Workflow dispatch with incident inputs

**Purpose:**
- Demonstrates end-to-end remediation flow
- Used by the demo application
- Tests the remediation-action in a real scenario

## Environment Variables

The CI pipeline uses the following environment variables:

```yaml
GO_VERSION: '1.21'        # Go version for incident-service
NODE_VERSION: '20'        # Node.js version for TypeScript/JavaScript projects
REGISTRY: ghcr.io         # Container registry
IMAGE_NAME: ${{ github.repository }}  # Repository name for image tagging
```

## Secrets Required

For the CI pipeline to work correctly, the following secrets should be configured:

### Repository Secrets

- `GITHUB_TOKEN` - Automatically provided by GitHub Actions
  - Used for: Pushing Docker images to GHCR, creating releases

### Optional Secrets

- `CODECOV_TOKEN` - For uploading coverage reports to Codecov
  - Not required if repository is public

## Permissions

The workflows require the following permissions:

**CI Pipeline:**
- `contents: read` - Read repository contents
- `packages: write` - Push Docker images to GHCR (build-images job only)

## Caching Strategy

The pipeline implements caching to speed up builds:

### Go Modules Cache
```yaml
path: ~/go/pkg/mod
key: ${{ runner.os }}-go-${{ hashFiles('incident-service/go.sum') }}
```

### NPM Cache
```yaml
cache: 'npm'
cache-dependency-path: <project>/package-lock.json
```

### Docker Layer Cache
```yaml
cache-from: type=gha
cache-to: type=gha,mode=max
```

## Running Workflows Locally

### Using Act (nektos/act)

You can run GitHub Actions workflows locally using [act](https://github.com/nektos/act):

```bash
# Install act
brew install act  # macOS
# or
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash  # Linux

# List available workflows
act -l

# Run the CI workflow
act push

# Run a specific job
act -j test-incident-service

# Run with secrets
act -s GITHUB_TOKEN=<token>
```

**Note:** Some jobs (like those requiring PostgreSQL/Redis services) may not work perfectly with act due to service container limitations.

### Manual Testing

For more reliable local testing:

```bash
# Lint
cd incident-service && golangci-lint run
cd dashboard && npm run lint

# Test
./scripts/test.sh

# Build Docker images
docker build -t incident-service:local ./incident-service
docker build -t dashboard:local ./dashboard
docker build -t demo-app:local ./demo-app
```

## Monitoring and Debugging

### Viewing Workflow Runs

1. Go to the "Actions" tab in GitHub
2. Select the workflow run you want to inspect
3. Click on individual jobs to see logs

### Common Issues

#### Linting Failures

**Symptom:** `lint-go` or `lint-typescript` job fails

**Solution:**
```bash
# Fix Go linting issues
cd incident-service
golangci-lint run --fix

# Fix TypeScript linting issues
cd dashboard
npm run lint -- --fix
```

#### Test Failures

**Symptom:** Test jobs fail with assertion errors

**Solution:**
1. Check the test logs in the Actions tab
2. Run tests locally: `./scripts/test.sh`
3. Fix the failing tests
4. Ensure all property-based tests run 100+ iterations

#### Docker Build Failures

**Symptom:** `build-images` job fails

**Solution:**
1. Check Dockerfile syntax
2. Verify all dependencies are available
3. Test build locally:
   ```bash
   docker build -t test ./incident-service
   ```
4. Check for missing files or incorrect COPY paths

#### Coverage Drops

**Symptom:** Coverage report shows decreased coverage

**Solution:**
1. Add tests for new code
2. Ensure property-based tests are included
3. Check coverage locally:
   ```bash
   cd incident-service
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

#### Service Container Issues

**Symptom:** Tests fail with database connection errors

**Solution:**
1. Verify service container health checks are passing
2. Check service ports are correctly mapped
3. Ensure DATABASE_URL and REDIS_URL are correct
4. Wait for services to be healthy before running tests

## Best Practices

### Writing CI-Friendly Code

1. **Fast Tests:** Keep unit tests fast (<1s each)
2. **Isolated Tests:** Don't depend on external services in unit tests
3. **Property Tests:** Use property-based testing for universal properties
4. **Coverage:** Aim for >80% overall, 100% for critical paths
5. **Linting:** Fix linting issues before committing

### Optimizing CI Performance

1. **Caching:** Use caching for dependencies
2. **Parallelization:** Run independent jobs in parallel
3. **Minimal Builds:** Only build what changed
4. **Layer Caching:** Use Docker layer caching for images

### Security

1. **Secrets:** Never commit secrets to the repository
2. **Permissions:** Use minimal required permissions
3. **Dependencies:** Keep dependencies up to date
4. **Scanning:** Consider adding security scanning jobs

## Future Enhancements

Potential improvements to the CI/CD pipeline:

- [ ] Add security scanning (Snyk, Trivy)
- [ ] Add performance benchmarking
- [ ] Add deployment to staging environment
- [ ] Add smoke tests after deployment
- [ ] Add notification on failures (Slack, email)
- [ ] Add automatic dependency updates (Dependabot)
- [ ] Add release automation
- [ ] Add changelog generation

## Support

For issues with the CI/CD pipeline:

1. Check this documentation
2. Review workflow logs in the Actions tab
3. Open an issue with the `ci/cd` label
4. Contact the maintainers

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Codecov Action](https://github.com/codecov/codecov-action)
- [golangci-lint](https://golangci-lint.run/)
- [nektos/act](https://github.com/nektos/act)
