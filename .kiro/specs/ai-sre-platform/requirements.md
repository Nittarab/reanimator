# Requirements Document

## Introduction

The AI SRE Platform is an autonomous infrastructure remediation system that acts as a "digital immune system" for software services. The platform receives incident and error notifications from existing observability platforms, triggers GitHub Actions workflows that use Kiro CLI to diagnose root causes and generate code fixes, and provides a dashboard for on-call engineers to monitor remediation progress. The system leverages existing observability infrastructure, GitHub Actions, and Kiro's agent capabilities with MCP integrations to transform on-call engineering from reactive firefighting to proactive code review.

## Glossary

- **Platform**: The AI SRE Platform system as a whole
- **Incident Service**: The backend service responsible for receiving, storing, and managing incident notifications
- **Dashboard**: The web UI where on-call engineers monitor incidents and remediation progress
- **Incident**: A notification of a failure or error from an observability platform
- **Remediation Workflow**: A GitHub Actions workflow that uses Kiro CLI to diagnose and fix incidents
- **Kiro CLI**: The command-line interface for Kiro that executes AI agent tasks
- **Kiro Spec**: Configuration files in the .kiro folder that define system prompts and remediation strategies
- **MCP Server**: Model Context Protocol server that Kiro CLI uses to query observability platforms for additional context
- **Resolution Loop**: The workflow from incident receipt through GitHub Actions trigger to pull request creation
- **Observability Provider**: External monitoring and logging systems (e.g., Datadog, New Relic, Prometheus)
- **Post-Mortem**: A detailed explanation document describing what broke, why, and how the fix addresses it
- **GitHub Action**: A reusable workflow component that can be installed in repositories
- **Workflow Dispatch**: GitHub's API for triggering workflows programmatically

## Requirements

### Requirement 1

**User Story:** As a platform operator, I want to receive incident notifications from my observability platform, so that the system can trigger automated remediation.

#### Acceptance Criteria

1. WHEN an Observability Provider sends an incident notification via webhook, THEN the Incident Service SHALL receive and acknowledge the notification within 500 milliseconds
2. WHEN the Incident Service receives a notification in provider-specific format, THEN the Incident Service SHALL transform it into a standardized Incident structure
3. WHERE multiple Observability Providers are configured, the Incident Service SHALL process notifications from all providers concurrently
4. WHEN the Incident Service receives malformed data, THEN the Incident Service SHALL log the error and return an appropriate HTTP error code
5. WHEN an Incident is created, THEN the Incident Service SHALL persist it to the database with a unique identifier, timestamp, and status

### Requirement 2

**User Story:** As a platform operator, I want the system to route incidents to the correct repository, so that remediation workflows are triggered appropriately.

#### Acceptance Criteria

1. WHEN the Incident Service receives an incident, THEN the Incident Service SHALL extract service identifiers from the incident metadata
2. WHEN a service identifier is extracted, THEN the Incident Service SHALL look up the associated repository from the configuration
3. WHEN multiple related incidents occur within a configurable time window, THEN the Incident Service SHALL deduplicate them and update the existing incident record
4. WHEN an incident is mapped to a repository, THEN the Incident Service SHALL mark it as ready for remediation
5. WHEN an incident cannot be mapped to a repository, THEN the Incident Service SHALL mark it as requiring manual configuration and send an alert

### Requirement 3

**User Story:** As a platform operator, I want to trigger GitHub Actions workflows for each incident, so that remediation happens in the repository's native CI/CD environment.

#### Acceptance Criteria

1. WHEN an incident is ready for remediation, THEN the Incident Service SHALL invoke the GitHub workflow dispatch API for the associated repository
2. WHEN the Incident Service triggers a workflow, THEN the Incident Service SHALL pass incident context including error messages, stack traces, and service metadata as workflow inputs
3. WHEN the workflow dispatch API call fails, THEN the Incident Service SHALL retry with exponential backoff up to three attempts
4. WHEN multiple incidents are pending for the same repository, THEN the Incident Service SHALL queue them to respect configured concurrency limits
5. WHERE a repository does not have the remediation workflow installed, the Incident Service SHALL mark the incident as requiring manual setup and send a notification

### Requirement 4

**User Story:** As a remediation workflow, I want to use Kiro CLI to analyze the repository code, so that I can diagnose the failure cause.

#### Acceptance Criteria

1. WHEN the Remediation Workflow starts, THEN the workflow SHALL checkout the repository at the commit or branch specified in the incident metadata
2. WHEN the repository is checked out, THEN the workflow SHALL install the Kiro CLI using the specified version
3. WHEN Kiro CLI installation fails, THEN the workflow SHALL report the failure and exit with a non-zero status code
4. WHEN the workflow needs to access Kiro Specs, THEN the workflow SHALL read configuration files from the .kiro directory
5. WHEN the .kiro directory does not exist, THEN the workflow SHALL use default remediation prompts and log a warning

### Requirement 5

**User Story:** As a remediation workflow, I want Kiro CLI to analyze stack traces and error logs, so that the root cause can be identified.

#### Acceptance Criteria

1. WHEN the workflow invokes Kiro CLI with incident data, THEN Kiro CLI SHALL parse the stack trace to extract file paths, line numbers, and function names
2. WHEN Kiro CLI identifies relevant code locations, THEN Kiro CLI SHALL read the source files and analyze the surrounding context
3. WHEN Kiro CLI analyzes code, THEN Kiro CLI SHALL use the repository's existing codebase context to understand the failure
4. WHEN Kiro CLI cannot locate the error source in the repository, THEN Kiro CLI SHALL output a diagnostic report indicating insufficient information
5. WHEN Kiro CLI completes root cause analysis, THEN Kiro CLI SHALL generate a structured diagnosis document in the workflow output

### Requirement 6

**User Story:** As a remediation workflow, I want Kiro CLI to validate fixes by running tests, so that I can verify the solution works.

#### Acceptance Criteria

1. WHEN Kiro CLI generates a code fix, THEN Kiro CLI SHALL identify relevant test commands from the repository configuration
2. WHEN the repository has a test suite, THEN the workflow SHALL execute the tests after applying the fix
3. WHEN tests pass after applying the fix, THEN Kiro CLI SHALL mark the fix as validated in the output
4. WHEN tests fail after applying the fix, THEN Kiro CLI SHALL include test failure details in the diagnostic output
5. WHERE the repository does not have automated tests, the workflow SHALL skip validation and document this limitation in the pull request

### Requirement 7

**User Story:** As a remediation workflow, I want Kiro CLI to generate a code fix for the identified issue, so that the problem can be resolved.

#### Acceptance Criteria

1. WHEN Kiro CLI completes diagnosis, THEN Kiro CLI SHALL generate a code fix based on the root cause analysis
2. WHEN a code fix is generated, THEN Kiro CLI SHALL apply the changes to the working directory
3. WHEN Kiro CLI applies changes, THEN Kiro CLI SHALL output a summary of modified files and changes made
4. WHEN Kiro CLI cannot generate a fix, THEN Kiro CLI SHALL output a diagnostic report explaining why automated remediation is not possible
5. WHERE Kiro Specs define custom remediation strategies, Kiro CLI SHALL follow those strategies when generating fixes

### Requirement 8

**User Story:** As a remediation workflow, I want to create a pull request with the fix, so that human engineers can review and merge the solution.

#### Acceptance Criteria

1. WHEN Kiro CLI has generated a fix, THEN the workflow SHALL commit the changes to a new branch with a descriptive name including the incident ID
2. WHEN changes are committed, THEN the workflow SHALL push the branch to the GitHub repository
3. WHEN the branch is pushed, THEN the workflow SHALL create a pull request targeting the original branch using the GitHub API
4. WHEN the pull request is created, THEN the workflow SHALL include a Post-Mortem document in the pull request description
5. WHEN the Post-Mortem is generated, THEN the workflow SHALL include sections for what broke, why it broke, how the fix addresses it, and test results

### Requirement 9

**User Story:** As a platform operator, I want the incident service to be self-hosted, so that I maintain control over incident data.

#### Acceptance Criteria

1. THE Incident Service SHALL be deployable as a containerized application
2. WHEN the Incident Service processes incidents, THEN the service SHALL NOT transmit data to external services except GitHub API for workflow dispatch
3. WHEN the Incident Service stores data, THEN the service SHALL use operator-configured database systems
4. WHEN the Incident Service communicates with GitHub, THEN the service SHALL use operator-provided GitHub tokens
5. WHERE the operator requires air-gapped deployment, the Incident Service SHALL support GitHub Enterprise Server endpoints

### Requirement 10

**User Story:** As a platform operator, I want to deploy the platform on any infrastructure, so that I am not locked into a specific vendor.

#### Acceptance Criteria

1. THE Incident Service and Dashboard SHALL be packaged as container images compatible with Docker and Kubernetes
2. WHEN the services are deployed, THEN the services SHALL use standard protocols (HTTP, HTTPS) for all external communication
3. WHEN the services are configured, THEN the services SHALL accept configuration through environment variables or configuration files
4. WHEN the services start, THEN the services SHALL validate required configuration and report any missing parameters
5. WHERE the operator uses Kubernetes, the Platform SHALL provide Helm charts for simplified deployment

### Requirement 11

**User Story:** As a platform operator, I want to configure which repositories and services are monitored, so that I can control the scope of automated remediation.

#### Acceptance Criteria

1. WHEN the Platform is configured, THEN the Platform SHALL accept a list of repository URLs and access credentials
2. WHEN the Platform is configured, THEN the Platform SHALL accept service-to-repository mappings for incident routing
3. WHEN an incident occurs for an unconfigured service, THEN the Platform SHALL log the incident but NOT create a Repair Agent
4. WHEN repository credentials are updated, THEN the Platform SHALL reload the configuration without requiring a restart
5. WHERE multiple teams use the Platform, the Platform SHALL support namespace isolation for configurations

### Requirement 12

**User Story:** As a platform operator, I want to configure workflow concurrency limits, so that the platform does not overwhelm GitHub Actions.

#### Acceptance Criteria

1. WHEN the Platform is configured, THEN the Platform SHALL accept limits for maximum concurrent workflow dispatches per repository
2. WHEN the Platform dispatches workflows, THEN the Platform SHALL track active workflows and respect the concurrency limit
3. WHEN the concurrency limit is reached, THEN the Platform SHALL queue additional incidents until active workflows complete
4. WHEN a workflow exceeds a configured timeout, THEN the Platform SHALL mark the incident as requiring human intervention
5. WHEN the Platform receives workflow completion webhooks, THEN the Platform SHALL update the active workflow count and process queued incidents

### Requirement 13

**User Story:** As a platform operator, I want to monitor the incident service's health and performance, so that I can ensure it is operating correctly.

#### Acceptance Criteria

1. THE Incident Service SHALL expose metrics for incident ingestion rate, workflow dispatch success rate, and queue depth
2. WHEN the service processes incidents, THEN the service SHALL record timing metrics for each stage of the Resolution Loop
3. WHEN the service encounters errors, THEN the service SHALL emit structured logs with severity levels
4. THE Incident Service SHALL expose a health check endpoint that responds within 100 milliseconds
5. WHEN the service starts, THEN the service SHALL log the version, configuration summary, and startup time

### Requirement 14

**User Story:** As a platform operator, I want to review the history of incidents and resolutions, so that I can understand patterns and improve the system.

#### Acceptance Criteria

1. WHEN an incident is processed, THEN the Platform SHALL persist a complete audit trail including all agent actions and decisions
2. WHEN a Repair Agent completes, THEN the Platform SHALL store the diagnosis, fix attempts, and final outcome
3. WHEN the operator queries the audit trail, THEN the Platform SHALL provide filtering by time range, service, repository, and outcome
4. WHEN the operator requests incident statistics, THEN the Platform SHALL compute aggregates for success rate, mean time to resolution, and common failure patterns
5. WHEN audit data exceeds the configured retention period, THEN the Platform SHALL archive or delete old records according to the retention policy

### Requirement 15

**User Story:** As a repository owner, I want to configure AI model providers in my workflow, so that I can control which AI service Kiro CLI uses.

#### Acceptance Criteria

1. WHEN the Remediation Workflow is configured, THEN the workflow SHALL accept AI model provider settings as workflow inputs or repository secrets
2. WHEN Kiro CLI is invoked, THEN the workflow SHALL pass AI model configuration to Kiro CLI via environment variables
3. WHEN the AI model provider is unavailable, THEN Kiro CLI SHALL retry with exponential backoff and eventually fail gracefully
4. WHERE the repository does not specify a model provider, Kiro CLI SHALL use the default provider configured in the workflow
5. WHEN Kiro CLI completes, THEN the workflow SHALL log which AI model provider was used for audit purposes

### Requirement 16

**User Story:** As a platform operator, I want to define custom incident detection rules, so that I can tailor the system to my specific infrastructure patterns.

#### Acceptance Criteria

1. WHEN the Platform is configured, THEN the Platform SHALL accept custom rule definitions in a declarative format
2. WHEN a custom rule is defined, THEN the Platform SHALL evaluate incoming Incident Events against the rule conditions
3. WHEN a custom rule matches, THEN the Platform SHALL apply the rule's actions including severity adjustment and metadata enrichment
4. WHEN custom rules are updated, THEN the Platform SHALL reload them without requiring a restart
5. WHEN a custom rule contains syntax errors, THEN the Platform SHALL reject the rule and log a detailed error message

### Requirement 17

**User Story:** As a security engineer, I want the platform to operate with least-privilege access, so that the security blast radius is minimized.

#### Acceptance Criteria

1. WHEN the Incident Service triggers workflows, THEN the service SHALL use GitHub tokens with workflow dispatch permissions only
2. WHEN the Remediation Workflow executes, THEN the workflow SHALL use GitHub's GITHUB_TOKEN with repository-scoped permissions
3. WHEN the Incident Service stores credentials, THEN the service SHALL encrypt them at rest using operator-provided encryption keys
4. WHEN the service transmits credentials, THEN the service SHALL use secure channels with TLS 1.3 or higher
5. WHERE the operator uses secrets management systems, the Incident Service SHALL support reading credentials from environment variables or mounted secret files

### Requirement 18

**User Story:** As a repository owner, I want to install a reusable GitHub Action, so that I can enable automated remediation without writing custom workflow code.

#### Acceptance Criteria

1. THE Platform SHALL provide a reusable GitHub Action published to the GitHub Marketplace
2. WHEN a repository owner installs the action, THEN the action SHALL accept incident data as workflow inputs
3. WHEN the action executes, THEN the action SHALL install Kiro CLI, run the remediation process, and create a pull request
4. WHEN the action is configured, THEN the action SHALL accept optional parameters for Kiro CLI version, AI model provider, and custom Kiro Specs path
5. WHEN the action completes, THEN the action SHALL output the pull request URL and remediation status

### Requirement 19

**User Story:** As an on-call engineer, I want to view a dashboard of active incidents and their remediation status, so that I can monitor the platform's progress.

#### Acceptance Criteria

1. THE Dashboard SHALL display a list of incidents ordered by timestamp with the most recent first
2. WHEN an incident is displayed, THEN the Dashboard SHALL show the incident status, affected service, error message, and associated repository
3. WHEN a remediation workflow is triggered, THEN the Dashboard SHALL update the incident status to show workflow progress
4. WHEN a pull request is created, THEN the Dashboard SHALL display a link to the pull request
5. WHEN an on-call engineer filters incidents, THEN the Dashboard SHALL support filtering by status, service, repository, and time range

### Requirement 20

**User Story:** As an on-call engineer, I want to manually trigger remediation for an incident, so that I can retry failed attempts or handle edge cases.

#### Acceptance Criteria

1. WHEN an on-call engineer views an incident in the Dashboard, THEN the Dashboard SHALL provide a button to trigger remediation
2. WHEN the engineer clicks the trigger button, THEN the Dashboard SHALL invoke the Incident Service to dispatch the workflow
3. WHEN a workflow is manually triggered, THEN the Incident Service SHALL record the manual trigger event in the incident history
4. WHEN an incident already has an active workflow, THEN the Dashboard SHALL disable the trigger button and show the active workflow status
5. WHEN manual triggering fails, THEN the Dashboard SHALL display an error message with details

### Requirement 21

**User Story:** As a remediation workflow, I want to use MCP to query the observability platform for additional context, so that I can gather detailed information about the incident.

#### Acceptance Criteria

1. WHEN the Remediation Workflow starts, THEN the workflow SHALL configure Kiro CLI with MCP server settings for the observability platform
2. WHEN Kiro CLI analyzes an incident, THEN Kiro CLI SHALL use the MCP server to query logs, traces, and metrics around the incident timeframe
3. WHEN the MCP server returns additional context, THEN Kiro CLI SHALL incorporate it into the root cause analysis
4. WHEN the MCP server is unavailable, THEN Kiro CLI SHALL proceed with the incident data provided in the workflow inputs
5. WHERE multiple observability platforms are configured, the workflow SHALL configure multiple MCP servers for Kiro CLI to query

### Requirement 22

**User Story:** As a repository owner, I want to configure MCP servers using GitHub secrets and repository configuration, so that Kiro CLI can query additional context during remediation.

#### Acceptance Criteria

1. WHEN a remediation workflow starts, THEN the workflow SHALL read MCP server configurations from the repository's `.kiro/settings/mcp.json` file if it exists
2. WHEN the workflow reads MCP configuration, THEN the workflow SHALL substitute environment variables from GitHub secrets into the MCP server configuration
3. WHERE no `.kiro/settings/mcp.json` exists, the workflow SHALL create a default MCP configuration based on available environment variables (e.g., DATADOG_API_KEY)
4. WHEN MCP credentials are passed to Kiro CLI, THEN the workflow SHALL use GitHub's secret masking to prevent credential exposure in logs
5. WHERE multiple MCP servers are configured, the workflow SHALL configure all of them for Kiro CLI to use

### Requirement 23

**User Story:** As a developer, I want to receive notifications when the platform creates a pull request for my service, so that I can review and merge fixes promptly.

#### Acceptance Criteria

1. WHEN the Remediation Workflow creates a pull request, THEN GitHub SHALL send native pull request notifications to repository watchers
2. WHERE Slack integration is configured in the repository, the workflow SHALL post a message with incident summary and pull request link
3. WHERE custom notification webhooks are configured, the workflow SHALL send incident data to the configured endpoints
4. WHEN a notification fails to send, THEN the workflow SHALL log the failure but continue with pull request creation
5. WHEN the workflow sends notifications, THEN the workflow SHALL include incident severity, affected service, and pull request link
