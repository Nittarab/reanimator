/**
 * Main entry point for the AI SRE Remediation GitHub Action
 */

import * as core from '@actions/core';
import * as path from 'path';
import { getInputs, setOutputs } from './inputs';
import { installKiroCLI, createIncidentContextFile } from './kiro';
import { getFinalMCPConfig, writeMCPConfig } from './mcp';
import { reportStatus, getWorkflowRunId } from './status-reporter';
import { IncidentContext } from './types';

/**
 * Main action execution
 */
async function run(): Promise<void> {
  try {
    core.info('Starting AI SRE Remediation Action');
    
    // Get action inputs
    const inputs = getInputs();
    core.info(`Processing incident: ${inputs.incidentId}`);
    core.info(`Service: ${inputs.serviceName}`);
    core.info(`Timestamp: ${inputs.timestamp}`);
    
    // Step 1: Install Kiro CLI
    core.startGroup('Installing Kiro CLI');
    try {
      const kiroPath = await installKiroCLI(inputs.kiroVersion);
      core.info(`Kiro CLI installed at: ${kiroPath}`);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      const diagnosis = `Kiro CLI installation failed: ${errorMessage}`;
      
      // Report failure to Incident Service
      if (inputs.incidentServiceUrl) {
        await reportStatus(inputs.incidentServiceUrl, {
          incident_id: inputs.incidentId,
          status: 'failed',
          diagnosis,
          repository: inputs.repository,
          workflow_run_id: getWorkflowRunId(),
        });
      }
      
      core.setFailed(`Failed to install Kiro CLI: ${errorMessage}`);
      setOutputs({ status: 'failed', diagnosis });
      return;
    }
    core.endGroup();
    
    // Step 2: Get repository path (current working directory)
    const repoPath = process.cwd();
    core.info(`Repository path: ${repoPath}`);
    
    // Step 3: Configure MCP servers
    core.startGroup('Configuring MCP servers');
    try {
      const mcpConfig = await getFinalMCPConfig(repoPath);
      
      const serverCount = Object.keys(mcpConfig.mcpServers).length;
      if (serverCount === 0) {
        core.warning('No MCP servers configured. Kiro CLI will proceed with incident data only.');
      } else {
        core.info(`Configured ${serverCount} MCP server(s): ${Object.keys(mcpConfig.mcpServers).join(', ')}`);
      }
      
      // Write MCP configuration to .kiro/settings/mcp.json if it doesn't exist
      // or if we generated it from environment variables
      const mcpConfigPath = path.join(repoPath, '.kiro', 'settings', 'mcp.json');
      await writeMCPConfig(mcpConfig, mcpConfigPath);
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      core.warning(`Failed to configure MCP servers: ${errorMessage}`);
      core.info('Continuing without MCP configuration');
    }
    core.endGroup();
    
    // Step 4: Create incident context file
    core.startGroup('Creating incident context file');
    let contextFilePath: string;
    try {
      const incidentContext: IncidentContext = {
        incident_id: inputs.incidentId,
        service_name: inputs.serviceName,
        timestamp: inputs.timestamp,
        error_message: inputs.errorMessage,
      };
      
      if (inputs.stackTrace) {
        incidentContext.stack_trace = inputs.stackTrace;
      }
      
      contextFilePath = path.join(repoPath, 'incident-context.md');
      await createIncidentContextFile(incidentContext, contextFilePath);
      
      core.info('Incident context file created successfully');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      const diagnosis = `Context file creation failed: ${errorMessage}`;
      
      // Report failure to Incident Service
      if (inputs.incidentServiceUrl) {
        await reportStatus(inputs.incidentServiceUrl, {
          incident_id: inputs.incidentId,
          status: 'failed',
          diagnosis,
          repository: inputs.repository,
          workflow_run_id: getWorkflowRunId(),
        });
      }
      
      core.setFailed(`Failed to create incident context file: ${errorMessage}`);
      setOutputs({ status: 'failed', diagnosis });
      return;
    }
    core.endGroup();
    
    // Step 5: Run Kiro CLI for remediation
    core.startGroup('Running Kiro CLI for remediation');
    let kiroResult: { diagnosis: string; fixDescription: string; testResults?: string };
    try {
      const { runKiroCLI } = await import('./kiro');
      kiroResult = await runKiroCLI(contextFilePath);
      
      core.info('Kiro CLI completed successfully');
      core.info(`Diagnosis: ${kiroResult.diagnosis.substring(0, 200)}...`);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      core.warning(`Kiro CLI execution encountered issues: ${errorMessage}`);
      
      // Set default values if Kiro CLI fails
      kiroResult = {
        diagnosis: `Kiro CLI execution failed: ${errorMessage}`,
        fixDescription: 'No automated fix could be generated.',
      };
    }
    core.endGroup();
    
    // Step 6: Check if there are changes to commit
    core.startGroup('Checking for changes');
    const { hasChanges } = await import('./github');
    const changesExist = await hasChanges();
    
    if (!changesExist) {
      core.info('No changes detected. No fix was needed or could be generated.');
      
      // Report no_fix_needed status to Incident Service
      if (inputs.incidentServiceUrl) {
        await reportStatus(inputs.incidentServiceUrl, {
          incident_id: inputs.incidentId,
          status: 'no_fix_needed',
          diagnosis: kiroResult.diagnosis,
          repository: inputs.repository,
          workflow_run_id: getWorkflowRunId(),
        });
      }
      
      setOutputs({
        status: 'no_fix_needed',
        diagnosis: kiroResult.diagnosis,
      });
      return;
    }
    core.info('Changes detected, proceeding with branch creation and PR');
    core.endGroup();
    
    // Step 7: Create branch with incident ID
    core.startGroup('Creating branch');
    let branchName: string;
    try {
      const { createBranch } = await import('./github');
      branchName = await createBranch(inputs.incidentId);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      const diagnosis = `Branch creation failed: ${errorMessage}`;
      
      // Report failure to Incident Service
      if (inputs.incidentServiceUrl) {
        await reportStatus(inputs.incidentServiceUrl, {
          incident_id: inputs.incidentId,
          status: 'failed',
          diagnosis,
          repository: inputs.repository,
          workflow_run_id: getWorkflowRunId(),
        });
      }
      
      core.setFailed(`Failed to create branch: ${errorMessage}`);
      setOutputs({ status: 'failed', diagnosis });
      return;
    }
    core.endGroup();
    
    // Step 8: Commit changes
    core.startGroup('Committing changes');
    try {
      const { commitChanges } = await import('./github');
      await commitChanges(inputs.incidentId, inputs.serviceName, inputs.errorMessage);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      const diagnosis = `Commit failed: ${errorMessage}`;
      
      // Report failure to Incident Service
      if (inputs.incidentServiceUrl) {
        await reportStatus(inputs.incidentServiceUrl, {
          incident_id: inputs.incidentId,
          status: 'failed',
          diagnosis,
          repository: inputs.repository,
          workflow_run_id: getWorkflowRunId(),
        });
      }
      
      core.setFailed(`Failed to commit changes: ${errorMessage}`);
      setOutputs({ status: 'failed', diagnosis });
      return;
    }
    core.endGroup();
    
    // Step 9: Push branch
    core.startGroup('Pushing branch');
    try {
      const { pushBranch } = await import('./github');
      await pushBranch(branchName);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      const diagnosis = `Push failed: ${errorMessage}`;
      
      // Report failure to Incident Service
      if (inputs.incidentServiceUrl) {
        await reportStatus(inputs.incidentServiceUrl, {
          incident_id: inputs.incidentId,
          status: 'failed',
          diagnosis,
          repository: inputs.repository,
          workflow_run_id: getWorkflowRunId(),
        });
      }
      
      core.setFailed(`Failed to push branch: ${errorMessage}`);
      setOutputs({ status: 'failed', diagnosis });
      return;
    }
    core.endGroup();
    
    // Step 10: Generate post-mortem
    core.startGroup('Generating post-mortem');
    let postMortem: string;
    try {
      const { generatePostMortem } = await import('./postmortem');
      postMortem = generatePostMortem({
        incidentId: inputs.incidentId,
        serviceName: inputs.serviceName,
        timestamp: inputs.timestamp,
        errorMessage: inputs.errorMessage,
        stackTrace: inputs.stackTrace || undefined,
        diagnosis: kiroResult.diagnosis,
        fixDescription: kiroResult.fixDescription,
        testResults: kiroResult.testResults,
      });
      
      core.info('Post-mortem generated successfully');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      core.warning(`Failed to generate post-mortem: ${errorMessage}`);
      postMortem = `# Post-Mortem: Incident ${inputs.incidentId}\n\nPost-mortem generation failed: ${errorMessage}`;
    }
    core.endGroup();
    
    // Step 11: Create pull request
    core.startGroup('Creating pull request');
    let prUrl: string;
    try {
      const { createPullRequest } = await import('./github');
      prUrl = await createPullRequest(
        branchName,
        inputs.incidentId,
        inputs.serviceName,
        postMortem
      );
      
      core.info(`Pull request created: ${prUrl}`);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      const diagnosis = `PR creation failed: ${errorMessage}`;
      
      // Report failure to Incident Service
      if (inputs.incidentServiceUrl) {
        await reportStatus(inputs.incidentServiceUrl, {
          incident_id: inputs.incidentId,
          status: 'failed',
          diagnosis,
          repository: inputs.repository,
          workflow_run_id: getWorkflowRunId(),
        });
      }
      
      core.setFailed(`Failed to create pull request: ${errorMessage}`);
      setOutputs({ status: 'failed', diagnosis });
      return;
    }
    core.endGroup();
    
    // Step 12: Send notifications
    core.startGroup('Sending notifications');
    try {
      const { sendNotifications } = await import('./notifications');
      await sendNotifications({
        incidentId: inputs.incidentId,
        serviceName: inputs.serviceName,
        severity: inputs.severity,
        errorMessage: inputs.errorMessage,
        prUrl,
        timestamp: inputs.timestamp,
      });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      core.warning(`Failed to send notifications: ${errorMessage}`);
      // Don't fail the action if notifications fail
    }
    core.endGroup();
    
    // Step 13: Report status back to Incident Service (if configured)
    if (inputs.incidentServiceUrl) {
      core.startGroup('Reporting status to Incident Service');
      await reportStatus(inputs.incidentServiceUrl, {
        incident_id: inputs.incidentId,
        status: 'success',
        pr_url: prUrl,
        diagnosis: kiroResult.diagnosis,
        repository: inputs.repository,
        workflow_run_id: getWorkflowRunId(),
      });
      core.endGroup();
    }
    
    // Step 14: Set outputs
    core.info('Remediation workflow completed successfully');
    setOutputs({
      prUrl,
      status: 'success',
      diagnosis: kiroResult.diagnosis,
    });
    
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    
    // Try to report failure to Incident Service
    try {
      const inputs = getInputs();
      if (inputs.incidentServiceUrl) {
        await reportStatus(inputs.incidentServiceUrl, {
          incident_id: inputs.incidentId,
          status: 'failed',
          diagnosis: errorMessage,
          repository: inputs.repository,
          workflow_run_id: getWorkflowRunId(),
        });
      }
    } catch (reportError) {
      // Ignore errors when reporting the failure
      core.debug(`Failed to report error status: ${reportError}`);
    }
    
    core.setFailed(`Action failed: ${errorMessage}`);
    setOutputs({ status: 'failed', diagnosis: errorMessage });
  }
}

// Run the action
run();
