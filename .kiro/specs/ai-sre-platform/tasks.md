# Implementation Plan

- [x] 1. Set up project structure and development environment
  - Create monorepo directory structure for incident-service, dashboard, demo-app, and remediation-action
  - Set up Docker Compose configuration for local development
  - Create environment configuration files (.env.example)
  - Set up CI/CD pipeline with GitHub Actions
  - Create quick start scripts (dev.sh, prod.sh, test.sh)
  - _Requirements: 9.1, 10.1, 10.3, 10.4_

- [x] 2. Implement Incident Service core infrastructure
  - Set up Go project with Chi router
  - Configure PostgreSQL database connection with migrations
  - Configure Redis connection for caching
  - Implement health check endpoint
  - Implement Prometheus metrics endpoint
  - Set up structured logging
  - _Requirements: 9.3, 9.4, 13.1, 13.3, 13.4, 13.5_

- [x] 2.1 Write property test for database connection resilience
  - **Property 1: Incident persistence round-trip**
  - **Validates: Requirements 1.5**

- [x] 3. Implement webhook adapter system
  - Create WebhookAdapter interface
  - Implement Datadog webhook adapter with signature validation
  - Implement PagerDuty webhook adapter with signature validation
  - Implement Grafana webhook adapter
  - Implement Sentry webhook adapter with signature validation
  - Create adapter registry and routing logic
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [x] 3.1 Write property test for provider transformation
  - **Property 2: Provider format transformation**
  - **Validates: Requirements 1.2**

- [x] 3.2 Write property test for malformed data handling
  - **Property 3: Malformed data error handling**
  - **Validates: Requirements 1.4**

- [x] 4. Implement incident management logic
  - Create Incident data model and database schema
  - Implement incident creation and persistence
  - Implement service-to-repository mapping lookup
  - Implement deduplication logic with time windows
  - Implement incident status state machine
  - _Requirements: 1.5, 2.1, 2.2, 2.3, 2.4, 2.5_

- [x] 4.1 Write property test for service lookup consistency
  - **Property 4: Service-to-repository lookup consistency**
  - **Validates: Requirements 2.2**

- [x] 4.2 Write property test for deduplication
  - **Property 5: Deduplication within time window**
  - **Validates: Requirements 2.3**

- [x] 5. Implement GitHub workflow dispatch integration
  - Create GitHub API client for workflow dispatch
  - Implement workflow trigger logic with incident context
  - Implement retry logic with exponential backoff
  - Implement concurrency limit tracking per repository
  - Implement incident queueing system
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 5.1 Write property test for workflow context completeness
  - **Property 6: Workflow dispatch includes required context**
  - **Validates: Requirements 3.2**

- [x] 5.2 Write property test for retry logic
  - **Property 7: Retry with exponential backoff**
  - **Validates: Requirements 3.3**

- [x] 5.3 Write property test for concurrency enforcement
  - **Property 8: Concurrency limit enforcement**
  - **Validates: Requirements 3.4, 12.2, 12.3**

- [x] 6. Implement configuration management
  - Create configuration file parser (YAML)
  - Implement service mapping configuration
  - Implement custom rule configuration
  - Implement configuration hot-reloading
  - _Requirements: 11.1, 11.2, 11.4, 16.1_

- [x] 6.1 Write property test for configuration parsing
  - **Property 11: Configuration parsing validity**
  - **Validates: Requirements 11.1, 11.2**

- [x] 6.2 Write property test for rule syntax validation
  - **Property 17: Rule syntax validation**
  - **Validates: Requirements 16.5**

- [x] 7. Implement custom rule engine
  - Create rule definition schema
  - Implement rule evaluation engine
  - Implement rule actions (severity adjustment, metadata enrichment)
  - Implement rule validation
  - _Requirements: 16.2, 16.3, 16.4, 16.5_

- [x] 7.1 Write property test for rule evaluation
  - **Property 16: Custom rule evaluation**
  - **Validates: Requirements 16.2, 16.3**

- [x] 8. Implement audit trail and incident history
  - Create incident_events table schema
  - Implement event logging for all state transitions
  - Implement audit trail query API with filtering
  - Implement statistics computation (success rate, MTTR)
  - Implement data retention policy
  - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 20.3_

- [x] 8.1 Write property test for audit completeness
  - **Property 13: Audit trail completeness**
  - **Validates: Requirements 14.1, 20.3**

- [x] 8.2 Write property test for filtering correctness
  - **Property 14: Incident filtering correctness**
  - **Validates: Requirements 14.3, 19.5**

- [x] 8.3 Write property test for statistics accuracy
  - **Property 15: Statistics computation accuracy**
  - **Validates: Requirements 14.4**

- [x] 9. Implement workflow status webhook handler
  - Create webhook endpoint for workflow completion
  - Implement incident status updates from workflow results
  - Implement queue processing on workflow completion
  - Store PR URL and diagnosis in incident record
  - _Requirements: 12.5_

- [x] 9.1 Write property test for queue processing
  - **Property 12: Workflow completion updates queue**
  - **Validates: Requirements 12.5**

- [x] 10. Checkpoint - Ensure all Incident Service tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 11. Build Dashboard React application
  - Set up React project with TypeScript and Vite
  - Set up TanStack Query for API state management
  - Set up shadcn/ui component library
  - Create API client for Incident Service
  - Implement routing with React Router
  - _Requirements: 19.1, 19.2, 19.3, 19.4, 19.5_

- [ ] 12. Implement Dashboard incident list view
  - Create incident list table component
  - Implement real-time updates with polling
  - Implement filtering by status, service, repository, time range
  - Implement sorting by timestamp (most recent first)
  - Display incident status, service, error message, repository
  - _Requirements: 19.1, 19.2, 19.5_

- [ ] 12.1 Write property test for incident ordering
  - **Property 19: Dashboard incident ordering**
  - **Validates: Requirements 19.1**

- [ ] 12.2 Write property test for incident display completeness
  - **Property 20: Dashboard incident display completeness**
  - **Validates: Requirements 19.2**

- [ ] 13. Implement Dashboard incident detail view
  - Create incident detail page component
  - Display full incident data including stack trace
  - Display incident timeline with all events
  - Display links to GitHub workflow and PR
  - Implement manual remediation trigger button
  - Handle trigger button state (disabled when workflow active)
  - _Requirements: 20.1, 20.2, 20.4, 20.5_

- [ ] 14. Implement Dashboard configuration view
  - Create configuration display page
  - Display service-to-repository mappings
  - _Requirements: 11.1, 11.2_

- [ ] 15. Checkpoint - Ensure all Dashboard tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 16. Build Remediation GitHub Action
  - Set up TypeScript project for GitHub Action
  - Define action.yml with inputs and outputs
  - Implement Kiro CLI installation logic
  - Implement MCP configuration reading from `.kiro/settings/mcp.json`
  - Implement MCP configuration generation from environment variables
  - Implement incident context file creation
  - _Requirements: 18.1, 18.2, 18.4, 22.1, 22.2, 22.3_

- [ ] 17. Implement remediation workflow logic
  - Implement Kiro CLI invocation with remediation prompt
  - Implement branch creation with incident ID in name
  - Implement git commit and push logic
  - Implement PR creation with GitHub API
  - Implement post-mortem generation
  - Implement secret masking for MCP credentials in logs
  - _Requirements: 4.4, 4.5, 8.1, 8.2, 8.3, 8.4, 8.5, 18.3, 18.5, 22.4_

- [ ] 17.1 Write property test for branch naming
  - **Property 9: Branch naming includes incident ID**
  - **Validates: Requirements 8.1**

- [ ] 17.2 Write property test for post-mortem completeness
  - **Property 10: Post-mortem completeness**
  - **Validates: Requirements 8.4, 8.5**

- [ ] 18. Implement notification system in action
  - Implement Slack notification integration
  - Implement custom webhook notifications
  - Implement notification error handling
  - Include incident severity, service, and PR link in notifications
  - _Requirements: 23.2, 23.3, 23.4, 23.5_

- [ ] 18.1 Write property test for notification content
  - **Property 18: Notification content completeness**
  - **Validates: Requirements 23.5**

- [ ] 19. Implement status reporting back to Incident Service
  - Implement webhook call to Incident Service on completion
  - Send PR URL and remediation status
  - Send diagnosis summary
  - Handle network failures gracefully
  - _Requirements: 18.5_

- [ ] 20. Checkpoint - Ensure all GitHub Action tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 21. Build Demo Application
  - Set up Node.js/Express project
  - Create buggy endpoints (division by zero, null pointer, array processing)
  - Integrate Sentry for error reporting
  - Implement health check endpoint
  - Create in-memory or SQLite database
  - _Requirements: Demo Application_

- [ ] 22. Build Demo UI
  - Create HTML/CSS/JS interface for triggering errors
  - Implement error trigger buttons for each scenario
  - Implement real-time incident status display
  - Display links to PRs and Sentry issues
  - Embed dashboard iframe or API integration
  - _Requirements: Demo Application_

- [ ] 23. Create Demo Kiro Specs and MCP configuration
  - Create .kiro/specs/demo-fixes directory
  - Write remediation strategies for division by zero
  - Write remediation strategies for null pointer errors
  - Write remediation strategies for array processing errors
  - Create .kiro/settings/mcp.json with Sentry MCP configuration
  - Document expected fixes for each scenario
  - _Requirements: Demo Application, 22.1_

- [ ] 24. Set up Demo workflow
  - Create .github/workflows/demo-remediate.yml
  - Configure workflow to use local remediation action
  - Set up workflow inputs for incident data
  - Configure GitHub secrets for Sentry credentials
  - Test end-to-end flow
  - _Requirements: Demo Application, 22.2_

- [ ] 25. Create Docker infrastructure
  - Write Dockerfile for Incident Service with multi-stage build
  - Write Dockerfile for Dashboard with multi-stage build
  - Write Dockerfile for Demo App
  - Create docker-compose.dev.yml for local development
  - Create docker-compose.prod.yml for production
  - Add health checks to all containers
  - _Requirements: 9.1, 10.1_

- [ ] 26. Create deployment scripts
  - Write scripts/dev.sh for local development
  - Write scripts/prod.sh for production deployment
  - Write scripts/test.sh for running all tests
  - Create .env.example with all required variables
  - Document environment variable requirements
  - _Requirements: 10.3, 10.4_

- [ ] 27. Set up CI/CD pipeline
  - Create .github/workflows/ci.yml for testing
  - Add test jobs for Incident Service (Go)
  - Add test jobs for Dashboard (TypeScript)
  - Add test jobs for Demo App (Node.js)
  - Add Docker image build and push jobs
  - Configure GitHub Container Registry
  - Add code coverage reporting
  - _Requirements: CI/CD_

- [ ] 28. Create documentation
  - Write README.md with project overview
  - Write CONTRIBUTING.md with development guide
  - Write docs/DEPLOYMENT.md with deployment instructions
  - Write docs/CONFIGURATION.md with configuration reference
  - Write docs/ADAPTERS.md with guide for adding new providers
  - Document API endpoints with OpenAPI spec
  - _Requirements: Documentation_

- [ ] 29. Final integration testing
  - Test complete flow: webhook → incident → workflow → PR
  - Test all four webhook adapters with real payloads
  - Test dashboard displays incidents correctly
  - Test manual remediation trigger
  - Test demo app error scenarios
  - Test MCP configuration from repository secrets
  - Test Docker Compose deployment
  - Verify all health checks work
  - Verify metrics are exposed correctly
  - _Requirements: All_

- [ ] 30. Final Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
