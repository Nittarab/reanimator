/**
 * MCP configuration management
 */

import * as core from '@actions/core';
import * as fs from 'fs';
import * as path from 'path';
import { MCPConfiguration, MCPServerConfig } from './types';

/**
 * Read MCP configuration from repository's .kiro/settings/mcp.json
 * @param repoPath - Path to the repository root
 * @returns MCP configuration object or null if file doesn't exist
 */
export async function readMCPConfigFromRepo(repoPath: string): Promise<MCPConfiguration | null> {
  const mcpConfigPath = path.join(repoPath, '.kiro', 'settings', 'mcp.json');
  
  core.info(`Looking for MCP configuration at: ${mcpConfigPath}`);
  
  try {
    if (!fs.existsSync(mcpConfigPath)) {
      core.info('No .kiro/settings/mcp.json found in repository');
      return null;
    }
    
    const configContent = await fs.promises.readFile(mcpConfigPath, 'utf-8');
    const config = JSON.parse(configContent) as MCPConfiguration;
    
    core.info(`Loaded MCP configuration with ${Object.keys(config.mcpServers || {}).length} servers`);
    return config;
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    core.warning(`Failed to read MCP configuration from repository: ${errorMessage}`);
    return null;
  }
}

/**
 * Generate default MCP configuration from environment variables
 * Supports common observability platforms like Datadog, Sentry, etc.
 * @returns MCP configuration object
 */
export function generateMCPConfigFromEnv(): MCPConfiguration {
  core.info('Generating MCP configuration from environment variables');
  
  const mcpServers: Record<string, MCPServerConfig> = {};
  
  // Datadog MCP server
  if (process.env.DATADOG_API_KEY && process.env.DATADOG_APP_KEY) {
    core.info('Found Datadog credentials, adding Datadog MCP server');
    mcpServers.datadog = {
      command: 'npx',
      args: ['-y', '@datadog/mcp-server'],
      env: {
        DATADOG_API_KEY: process.env.DATADOG_API_KEY,
        DATADOG_APP_KEY: process.env.DATADOG_APP_KEY,
      },
    };
  }
  
  // Sentry MCP server
  if (process.env.SENTRY_AUTH_TOKEN) {
    core.info('Found Sentry credentials, adding Sentry MCP server');
    mcpServers.sentry = {
      command: 'npx',
      args: ['-y', '@sentry/mcp-server'],
      env: {
        SENTRY_AUTH_TOKEN: process.env.SENTRY_AUTH_TOKEN,
        SENTRY_ORG: process.env.SENTRY_ORG || '',
      },
    };
  }
  
  // PagerDuty MCP server
  if (process.env.PAGERDUTY_API_KEY) {
    core.info('Found PagerDuty credentials, adding PagerDuty MCP server');
    mcpServers.pagerduty = {
      command: 'npx',
      args: ['-y', '@pagerduty/mcp-server'],
      env: {
        PAGERDUTY_API_KEY: process.env.PAGERDUTY_API_KEY,
      },
    };
  }
  
  // Grafana MCP server
  if (process.env.GRAFANA_API_KEY && process.env.GRAFANA_URL) {
    core.info('Found Grafana credentials, adding Grafana MCP server');
    mcpServers.grafana = {
      command: 'npx',
      args: ['-y', '@grafana/mcp-server'],
      env: {
        GRAFANA_API_KEY: process.env.GRAFANA_API_KEY,
        GRAFANA_URL: process.env.GRAFANA_URL,
      },
    };
  }
  
  return { mcpServers };
}

/**
 * Substitute environment variables in MCP configuration
 * Replaces ${VAR_NAME} patterns with actual environment variable values
 * @param config - MCP configuration with potential env var placeholders
 * @returns Configuration with substituted values
 */
export function substituteEnvVars(config: MCPConfiguration): MCPConfiguration {
  const substituted: MCPConfiguration = {
    mcpServers: {},
  };
  
  for (const [serverName, serverConfig] of Object.entries(config.mcpServers || {})) {
    const newConfig: MCPServerConfig = {
      ...serverConfig,
      env: {},
    };
    
    // Substitute environment variables in env section
    if (serverConfig.env) {
      for (const [key, value] of Object.entries(serverConfig.env)) {
        // Match ${VAR_NAME} pattern
        const envVarMatch = value.match(/^\$\{([^}]+)\}$/);
        if (envVarMatch) {
          const envVarName = envVarMatch[1];
          const envValue = process.env[envVarName];
          
          if (envValue) {
            newConfig.env![key] = envValue;
            // Mask the value in logs for security
            core.setSecret(envValue);
          } else {
            core.warning(`Environment variable ${envVarName} not found for ${serverName}.${key}`);
            newConfig.env![key] = value; // Keep original if not found
          }
        } else {
          newConfig.env![key] = value;
        }
      }
    }
    
    substituted.mcpServers[serverName] = newConfig;
  }
  
  return substituted;
}

/**
 * Merge MCP configurations with repository config taking precedence
 * @param repoConfig - Configuration from repository (higher priority)
 * @param envConfig - Configuration from environment variables (lower priority)
 * @returns Merged configuration
 */
export function mergeMCPConfigs(
  repoConfig: MCPConfiguration | null,
  envConfig: MCPConfiguration
): MCPConfiguration {
  if (!repoConfig) {
    return envConfig;
  }
  
  const merged: MCPConfiguration = {
    mcpServers: {
      ...envConfig.mcpServers,
      ...repoConfig.mcpServers,
    },
  };
  
  core.info(`Merged MCP configuration with ${Object.keys(merged.mcpServers).length} total servers`);
  return merged;
}

/**
 * Write MCP configuration to a file
 * @param config - MCP configuration to write
 * @param outputPath - Path where to write the configuration
 */
export async function writeMCPConfig(config: MCPConfiguration, outputPath: string): Promise<void> {
  try {
    // Ensure directory exists
    const dir = path.dirname(outputPath);
    await fs.promises.mkdir(dir, { recursive: true });
    
    // Write configuration
    const configJson = JSON.stringify(config, null, 2);
    await fs.promises.writeFile(outputPath, configJson, 'utf-8');
    
    core.info(`MCP configuration written to: ${outputPath}`);
  } catch (error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    throw new Error(`Failed to write MCP configuration: ${errorMessage}`);
  }
}

/**
 * Get final MCP configuration by reading from repo and merging with env vars
 * @param repoPath - Path to the repository root
 * @returns Final MCP configuration
 */
export async function getFinalMCPConfig(repoPath: string): Promise<MCPConfiguration> {
  // Read from repository
  const repoConfig = await readMCPConfigFromRepo(repoPath);
  
  // Generate from environment
  const envConfig = generateMCPConfigFromEnv();
  
  // Merge configurations
  let merged = mergeMCPConfigs(repoConfig, envConfig);
  
  // Substitute environment variables
  merged = substituteEnvVars(merged);
  
  return merged;
}
