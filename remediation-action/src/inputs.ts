/**
 * GitHub Actions input handling
 */

import * as core from '@actions/core';
import { ActionInputs } from './types';

/**
 * Get all action inputs
 * @returns Action inputs object
 */
export function getInputs(): ActionInputs {
  // Get repository from environment (format: owner/repo)
  const repository = process.env.GITHUB_REPOSITORY || '';
  
  return {
    incidentId: core.getInput('incident_id', { required: true }),
    errorMessage: core.getInput('error_message', { required: true }),
    stackTrace: core.getInput('stack_trace', { required: false }) || '',
    serviceName: core.getInput('service_name', { required: true }),
    timestamp: core.getInput('timestamp', { required: true }),
    severity: core.getInput('severity', { required: false }) || 'medium',
    kiroVersion: core.getInput('kiro_version', { required: false }) || 'latest',
    mcpConfig: core.getInput('mcp_config', { required: false }) || '{}',
    incidentServiceUrl: core.getInput('incident_service_url', { required: false }) || '',
    repository,
  };
}

/**
 * Set action outputs
 * @param outputs - Output values to set
 */
export function setOutputs(outputs: {
  prUrl?: string;
  status: string;
  diagnosis?: string;
}): void {
  if (outputs.prUrl) {
    core.setOutput('pr_url', outputs.prUrl);
  }
  core.setOutput('status', outputs.status);
  if (outputs.diagnosis) {
    core.setOutput('diagnosis', outputs.diagnosis);
  }
}
