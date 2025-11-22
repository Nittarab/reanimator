/**
 * Type definitions for the remediation action
 */

export interface ActionInputs {
  incidentId: string;
  errorMessage: string;
  stackTrace: string;
  serviceName: string;
  timestamp: string;
  severity: string;
  kiroVersion: string;
  mcpConfig: string;
  incidentServiceUrl: string;
  repository: string;
}

export interface MCPServerConfig {
  command: string;
  args: string[];
  env?: Record<string, string>;
  disabled?: boolean;
  autoApprove?: string[];
}

export interface MCPConfiguration {
  mcpServers: Record<string, MCPServerConfig>;
}

export interface IncidentContext {
  incident_id: string;
  service_name: string;
  timestamp: string;
  error_message: string;
  stack_trace?: string;
}

export interface ActionOutputs {
  prUrl?: string;
  status: 'success' | 'failed' | 'no_fix_needed';
  diagnosis?: string;
}

export interface KiroResult {
  diagnosis: string;
  fixDescription: string;
  testResults?: string;
}
