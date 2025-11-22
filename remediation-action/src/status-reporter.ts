/**
 * Status reporting to Incident Service
 */

import * as core from '@actions/core';

export interface StatusReportPayload {
  incident_id: string;
  status: 'success' | 'failed' | 'no_fix_needed';
  pr_url?: string;
  diagnosis?: string;
  repository: string;
  workflow_run_id?: number;
}

/**
 * Report status back to Incident Service with retry logic
 * @param incidentServiceUrl - Base URL of the Incident Service
 * @param payload - Status report payload
 * @param maxRetries - Maximum number of retry attempts (default: 3)
 * @returns Promise that resolves when status is reported successfully
 */
export async function reportStatus(
  incidentServiceUrl: string,
  payload: StatusReportPayload,
  maxRetries: number = 3
): Promise<void> {
  if (!incidentServiceUrl) {
    core.debug('No incident service URL configured, skipping status report');
    return;
  }

  const url = `${incidentServiceUrl}/api/v1/webhooks/workflow-status`;
  
  // Limit diagnosis length to avoid payload size issues
  const sanitizedPayload = {
    ...payload,
    diagnosis: payload.diagnosis ? payload.diagnosis.substring(0, 500) : undefined,
  };

  let lastError: Error | null = null;
  
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      core.debug(`Reporting status to ${url} (attempt ${attempt}/${maxRetries})`);
      
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(sanitizedPayload),
        signal: AbortSignal.timeout(10000), // 10 second timeout
      });

      if (!response.ok) {
        const errorText = await response.text().catch(() => 'Unable to read response');
        throw new Error(`HTTP ${response.status}: ${response.statusText} - ${errorText}`);
      }

      core.info(`Status reported successfully to Incident Service: ${payload.status}`);
      return;
      
    } catch (error) {
      lastError = error instanceof Error ? error : new Error(String(error));
      
      if (attempt < maxRetries) {
        // Exponential backoff: 1s, 2s, 4s
        const delayMs = Math.pow(2, attempt - 1) * 1000;
        core.warning(`Failed to report status (attempt ${attempt}/${maxRetries}): ${lastError.message}`);
        core.debug(`Retrying in ${delayMs}ms...`);
        await sleep(delayMs);
      }
    }
  }

  // All retries exhausted
  const errorMessage = lastError ? lastError.message : 'Unknown error';
  core.warning(`Failed to report status to Incident Service after ${maxRetries} attempts: ${errorMessage}`);
  core.warning('Continuing workflow execution despite status reporting failure');
}

/**
 * Sleep for specified milliseconds
 * @param ms - Milliseconds to sleep
 */
function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Get workflow run ID from environment
 * @returns Workflow run ID or undefined
 */
export function getWorkflowRunId(): number | undefined {
  const runId = process.env.GITHUB_RUN_ID;
  return runId ? parseInt(runId, 10) : undefined;
}
