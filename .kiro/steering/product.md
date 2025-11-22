# Product Overview

The AI SRE Platform is an autonomous infrastructure remediation system that acts as a "digital immune system" for software services. It receives incident notifications from observability platforms (Datadog, PagerDuty, Grafana, Sentry), triggers GitHub Actions workflows that use Kiro CLI to diagnose root causes and generate code fixes, and provides a dashboard for on-call engineers to monitor remediation progress.

## Core Value Proposition

Transform on-call engineering from reactive firefighting to proactive code review by automating the diagnosis and fix generation for production incidents.

## Key Components

- **Incident Service**: Backend API that receives webhooks, manages incident state, and orchestrates GitHub Actions workflows
- **Dashboard**: Web UI for real-time visibility into incidents and remediation status
- **Remediation GitHub Action**: Reusable action that repositories install to enable automated fix generation using Kiro CLI with MCP integrations
- **Demo Application**: Sample buggy service demonstrating the full remediation flow

## Architecture Principles

- **Self-hosted**: Platform operators maintain control over incident data
- **Vendor-agnostic**: Deploy on any infrastructure (Docker, Kubernetes)
- **Least-privilege**: Minimal security blast radius with scoped permissions
- **Event-driven**: Incidents flow from observability platforms through the system to pull requests
