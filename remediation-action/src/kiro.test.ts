/**
 * Tests for Kiro CLI management
 */

import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import { createIncidentContextFile } from './kiro';

describe('createIncidentContextFile', () => {
  let tempDir: string;

  beforeEach(async () => {
    tempDir = await fs.promises.mkdtemp(path.join(os.tmpdir(), 'kiro-test-'));
  });

  afterEach(async () => {
    await fs.promises.rm(tempDir, { recursive: true, force: true });
  });

  it('should create incident context file with all fields', async () => {
    const incidentData = {
      incident_id: 'inc_123',
      service_name: 'api-gateway',
      timestamp: '2024-01-15T10:00:00Z',
      error_message: 'NullPointerException',
      stack_trace: 'at Service.method(Service.java:42)',
    };

    const outputPath = path.join(tempDir, 'incident-context.md');
    await createIncidentContextFile(incidentData, outputPath);

    const content = await fs.promises.readFile(outputPath, 'utf-8');

    expect(content).toContain('inc_123');
    expect(content).toContain('api-gateway');
    expect(content).toContain('2024-01-15T10:00:00Z');
    expect(content).toContain('NullPointerException');
    expect(content).toContain('at Service.method(Service.java:42)');
  });

  it('should create incident context file without stack trace', async () => {
    const incidentData = {
      incident_id: 'inc_456',
      service_name: 'user-service',
      timestamp: '2024-01-15T11:00:00Z',
      error_message: 'Connection timeout',
    };

    const outputPath = path.join(tempDir, 'incident-context.md');
    await createIncidentContextFile(incidentData, outputPath);

    const content = await fs.promises.readFile(outputPath, 'utf-8');

    expect(content).toContain('inc_456');
    expect(content).toContain('user-service');
    expect(content).toContain('Connection timeout');
    expect(content).not.toContain('## Stack Trace');
  });

  it('should include remediation instructions', async () => {
    const incidentData = {
      incident_id: 'inc_789',
      service_name: 'payment-service',
      timestamp: '2024-01-15T12:00:00Z',
      error_message: 'Database error',
    };

    const outputPath = path.join(tempDir, 'incident-context.md');
    await createIncidentContextFile(incidentData, outputPath);

    const content = await fs.promises.readFile(outputPath, 'utf-8');

    expect(content).toContain('## Task');
    expect(content).toContain('AI SRE agent');
    expect(content).toContain('MCP servers');
    expect(content).toContain('root cause');
    expect(content).toContain('post-mortem');
  });

  it('should throw error if file write fails', async () => {
    const incidentData = {
      incident_id: 'inc_error',
      service_name: 'test-service',
      timestamp: '2024-01-15T13:00:00Z',
      error_message: 'Test error',
    };

    // Try to write to an invalid path
    const invalidPath = '/invalid/path/that/does/not/exist/file.md';

    await expect(
      createIncidentContextFile(incidentData, invalidPath)
    ).rejects.toThrow('Failed to create incident context file');
  });
});
