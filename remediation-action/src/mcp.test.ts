/**
 * Tests for MCP configuration management
 */

import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import {
  generateMCPConfigFromEnv,
  substituteEnvVars,
  mergeMCPConfigs,
  readMCPConfigFromRepo,
  writeMCPConfig,
} from './mcp';
import { MCPConfiguration } from './types';

describe('generateMCPConfigFromEnv', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    // Reset environment
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it('should generate Datadog MCP config when credentials are present', () => {
    process.env.DATADOG_API_KEY = 'test-api-key';
    process.env.DATADOG_APP_KEY = 'test-app-key';

    const config = generateMCPConfigFromEnv();

    expect(config.mcpServers.datadog).toBeDefined();
    expect(config.mcpServers.datadog.command).toBe('npx');
    expect(config.mcpServers.datadog.args).toEqual(['-y', '@datadog/mcp-server']);
    expect(config.mcpServers.datadog.env?.DATADOG_API_KEY).toBe('test-api-key');
    expect(config.mcpServers.datadog.env?.DATADOG_APP_KEY).toBe('test-app-key');
  });

  it('should generate Sentry MCP config when credentials are present', () => {
    process.env.SENTRY_AUTH_TOKEN = 'test-token';
    process.env.SENTRY_ORG = 'test-org';

    const config = generateMCPConfigFromEnv();

    expect(config.mcpServers.sentry).toBeDefined();
    expect(config.mcpServers.sentry.env?.SENTRY_AUTH_TOKEN).toBe('test-token');
    expect(config.mcpServers.sentry.env?.SENTRY_ORG).toBe('test-org');
  });

  it('should return empty config when no credentials are present', () => {
    const config = generateMCPConfigFromEnv();

    expect(Object.keys(config.mcpServers)).toHaveLength(0);
  });

  it('should generate multiple MCP configs when multiple credentials are present', () => {
    process.env.DATADOG_API_KEY = 'test-api-key';
    process.env.DATADOG_APP_KEY = 'test-app-key';
    process.env.SENTRY_AUTH_TOKEN = 'test-token';

    const config = generateMCPConfigFromEnv();

    expect(Object.keys(config.mcpServers)).toHaveLength(2);
    expect(config.mcpServers.datadog).toBeDefined();
    expect(config.mcpServers.sentry).toBeDefined();
  });
});

describe('substituteEnvVars', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it('should substitute environment variables in config', () => {
    process.env.TEST_VAR = 'test-value';

    const config: MCPConfiguration = {
      mcpServers: {
        test: {
          command: 'test',
          args: [],
          env: {
            KEY: '${TEST_VAR}',
          },
        },
      },
    };

    const result = substituteEnvVars(config);

    expect(result.mcpServers.test.env?.KEY).toBe('test-value');
  });

  it('should keep original value if environment variable not found', () => {
    const config: MCPConfiguration = {
      mcpServers: {
        test: {
          command: 'test',
          args: [],
          env: {
            KEY: '${NONEXISTENT_VAR}',
          },
        },
      },
    };

    const result = substituteEnvVars(config);

    expect(result.mcpServers.test.env?.KEY).toBe('${NONEXISTENT_VAR}');
  });

  it('should not substitute non-placeholder values', () => {
    const config: MCPConfiguration = {
      mcpServers: {
        test: {
          command: 'test',
          args: [],
          env: {
            KEY: 'literal-value',
          },
        },
      },
    };

    const result = substituteEnvVars(config);

    expect(result.mcpServers.test.env?.KEY).toBe('literal-value');
  });
});

describe('mergeMCPConfigs', () => {
  it('should return env config when repo config is null', () => {
    const envConfig: MCPConfiguration = {
      mcpServers: {
        datadog: {
          command: 'npx',
          args: ['-y', '@datadog/mcp-server'],
        },
      },
    };

    const result = mergeMCPConfigs(null, envConfig);

    expect(result).toEqual(envConfig);
  });

  it('should merge configs with repo config taking precedence', () => {
    const envConfig: MCPConfiguration = {
      mcpServers: {
        datadog: {
          command: 'npx',
          args: ['-y', '@datadog/mcp-server'],
          env: { KEY: 'env-value' },
        },
      },
    };

    const repoConfig: MCPConfiguration = {
      mcpServers: {
        datadog: {
          command: 'custom',
          args: ['custom-arg'],
          env: { KEY: 'repo-value' },
        },
      },
    };

    const result = mergeMCPConfigs(repoConfig, envConfig);

    expect(result.mcpServers.datadog.command).toBe('custom');
    expect(result.mcpServers.datadog.env?.KEY).toBe('repo-value');
  });

  it('should include servers from both configs', () => {
    const envConfig: MCPConfiguration = {
      mcpServers: {
        datadog: {
          command: 'npx',
          args: ['-y', '@datadog/mcp-server'],
        },
      },
    };

    const repoConfig: MCPConfiguration = {
      mcpServers: {
        sentry: {
          command: 'npx',
          args: ['-y', '@sentry/mcp-server'],
        },
      },
    };

    const result = mergeMCPConfigs(repoConfig, envConfig);

    expect(Object.keys(result.mcpServers)).toHaveLength(2);
    expect(result.mcpServers.datadog).toBeDefined();
    expect(result.mcpServers.sentry).toBeDefined();
  });
});

describe('readMCPConfigFromRepo', () => {
  let tempDir: string;

  beforeEach(async () => {
    tempDir = await fs.promises.mkdtemp(path.join(os.tmpdir(), 'mcp-test-'));
  });

  afterEach(async () => {
    await fs.promises.rm(tempDir, { recursive: true, force: true });
  });

  it('should return null when config file does not exist', async () => {
    const result = await readMCPConfigFromRepo(tempDir);
    expect(result).toBeNull();
  });

  it('should read and parse valid config file', async () => {
    const configDir = path.join(tempDir, '.kiro', 'settings');
    await fs.promises.mkdir(configDir, { recursive: true });

    const config: MCPConfiguration = {
      mcpServers: {
        test: {
          command: 'test',
          args: ['arg1'],
        },
      },
    };

    await fs.promises.writeFile(
      path.join(configDir, 'mcp.json'),
      JSON.stringify(config),
      'utf-8'
    );

    const result = await readMCPConfigFromRepo(tempDir);

    expect(result).toEqual(config);
  });

  it('should return null on invalid JSON', async () => {
    const configDir = path.join(tempDir, '.kiro', 'settings');
    await fs.promises.mkdir(configDir, { recursive: true });

    await fs.promises.writeFile(
      path.join(configDir, 'mcp.json'),
      'invalid json',
      'utf-8'
    );

    const result = await readMCPConfigFromRepo(tempDir);

    expect(result).toBeNull();
  });
});

describe('writeMCPConfig', () => {
  let tempDir: string;

  beforeEach(async () => {
    tempDir = await fs.promises.mkdtemp(path.join(os.tmpdir(), 'mcp-test-'));
  });

  afterEach(async () => {
    await fs.promises.rm(tempDir, { recursive: true, force: true });
  });

  it('should write config to file', async () => {
    const config: MCPConfiguration = {
      mcpServers: {
        test: {
          command: 'test',
          args: ['arg1'],
        },
      },
    };

    const outputPath = path.join(tempDir, 'config.json');
    await writeMCPConfig(config, outputPath);

    const content = await fs.promises.readFile(outputPath, 'utf-8');
    const parsed = JSON.parse(content);

    expect(parsed).toEqual(config);
  });

  it('should create directory if it does not exist', async () => {
    const config: MCPConfiguration = {
      mcpServers: {},
    };

    const outputPath = path.join(tempDir, 'nested', 'dir', 'config.json');
    await writeMCPConfig(config, outputPath);

    expect(fs.existsSync(outputPath)).toBe(true);
  });
});
