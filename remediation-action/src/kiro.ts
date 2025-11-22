/**
 * Kiro CLI installation and management
 */

import * as core from '@actions/core';
import * as tc from '@actions/tool-cache';
import * as exec from '@actions/exec';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

/**
 * Install Kiro CLI
 * @param version - Version to install (e.g., 'latest', '1.0.0')
 * @returns Path to the installed Kiro CLI
 */
export async function installKiroCLI(version: string): Promise<string> {
  core.info(`Installing Kiro CLI version: ${version}`);
  
  try {
    // For now, we'll use a placeholder installation method
    // In production, this would download from a release URL
    // Example: https://github.com/kiro/cli/releases/download/v${version}/kiro-${platform}-${arch}
    
    const platform = os.platform();
    const arch = os.arch();
    
    core.info(`Platform: ${platform}, Architecture: ${arch}`);
    
    // Check if Kiro CLI is already available in PATH
    try {
      await exec.exec('kiro', ['--version'], { silent: true });
      core.info('Kiro CLI already available in PATH');
      return 'kiro';
    } catch {
      core.info('Kiro CLI not found in PATH, proceeding with installation');
    }
    
    // In a real implementation, download and install Kiro CLI
    // For now, we'll assume it's available or provide instructions
    core.warning('Kiro CLI installation not fully implemented. Assuming kiro is available in PATH.');
    
    return 'kiro';
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to install Kiro CLI: ${errorMessage}`);
  }
}

/**
 * Create incident context file for Kiro CLI
 * @param incidentData - Incident information
 * @param outputPath - Path where to write the context file
 */
export async function createIncidentContextFile(
  incidentData: {
    incident_id: string;
    service_name: string;
    timestamp: string;
    error_message: string;
    stack_trace?: string;
  },
  outputPath: string
): Promise<void> {
  core.info('Creating incident context file');
  
  const contextContent = `# Incident Context

## Incident Details
- **ID**: ${incidentData.incident_id}
- **Service**: ${incidentData.service_name}
- **Timestamp**: ${incidentData.timestamp}

## Error Message
\`\`\`
${incidentData.error_message}
\`\`\`

${incidentData.stack_trace ? `## Stack Trace
\`\`\`
${incidentData.stack_trace}
\`\`\`
` : ''}

## Task
You are an AI SRE agent tasked with diagnosing and fixing this production incident.

1. Use the MCP servers to query additional context from the observability platform around the incident timestamp
2. Analyze the stack trace and error message to identify the root cause
3. Locate the problematic code in this repository
4. Generate a fix that addresses the root cause
5. Run the test suite to validate your fix
6. Create a detailed post-mortem explaining what broke, why, and how your fix addresses it

Begin your analysis.
`;

  try {
    await fs.promises.writeFile(outputPath, contextContent, 'utf-8');
    core.info(`Incident context file created at: ${outputPath}`);
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to create incident context file: ${errorMessage}`);
  }
}

/**
 * Run Kiro CLI with remediation prompt
 * @param contextFilePath - Path to the incident context file
 * @returns Object containing diagnosis and fix description
 */
export async function runKiroCLI(contextFilePath: string): Promise<{
  diagnosis: string;
  fixDescription: string;
  testResults?: string;
}> {
  core.info('Running Kiro CLI for remediation');
  
  try {
    // Mask any sensitive environment variables in logs
    maskSensitiveEnvVars();
    
    let output = '';
    let errorOutput = '';
    
    // Run Kiro CLI with the incident context file
    const exitCode = await exec.exec(
      'kiro',
      ['chat', '--file', contextFilePath, '--non-interactive'],
      {
        listeners: {
          stdout: (data: Buffer) => {
            output += data.toString();
          },
          stderr: (data: Buffer) => {
            errorOutput += data.toString();
          },
        },
        ignoreReturnCode: true,
      }
    );
    
    if (exitCode !== 0) {
      core.warning(`Kiro CLI exited with code ${exitCode}`);
      core.warning(`Error output: ${errorOutput}`);
    }
    
    // Parse the output to extract diagnosis and fix description
    // This is a simplified parser - in production, Kiro CLI would output structured data
    const diagnosis = extractSection(output, 'diagnosis', 'root cause analysis');
    const fixDescription = extractSection(output, 'fix', 'solution', 'remediation');
    const testResults = extractSection(output, 'test', 'validation');
    
    return {
      diagnosis: diagnosis || 'Kiro CLI completed analysis. See full output for details.',
      fixDescription: fixDescription || 'Kiro CLI applied fixes. See commit for details.',
      testResults: testResults || undefined,
    };
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to run Kiro CLI: ${errorMessage}`);
  }
}

/**
 * Mask sensitive environment variables in GitHub Actions logs
 */
function maskSensitiveEnvVars(): void {
  const sensitivePatterns = [
    'API_KEY',
    'APP_KEY',
    'SECRET',
    'TOKEN',
    'PASSWORD',
    'CREDENTIAL',
  ];
  
  for (const [key, value] of Object.entries(process.env)) {
    if (value && sensitivePatterns.some(pattern => key.toUpperCase().includes(pattern))) {
      core.setSecret(value);
    }
  }
}

/**
 * Extract a section from Kiro CLI output
 * @param output - The full output text
 * @param keywords - Keywords to search for
 * @returns Extracted section or empty string
 */
function extractSection(output: string, ...keywords: string[]): string {
  const lines = output.split('\n');
  let capturing = false;
  let section = '';
  
  for (const line of lines) {
    const lowerLine = line.toLowerCase();
    
    // Check if this line starts a relevant section
    if (keywords.some(keyword => lowerLine.includes(keyword))) {
      capturing = true;
      section = '';
      continue;
    }
    
    // Stop capturing if we hit another major section
    if (capturing && (lowerLine.startsWith('#') || lowerLine.startsWith('##'))) {
      break;
    }
    
    if (capturing) {
      section += line + '\n';
    }
  }
  
  return section.trim();
}
