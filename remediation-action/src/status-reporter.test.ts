/**
 * Tests for status reporter
 */

import { it } from 'node:test';
import { it } from 'node:test';
import { it } from 'node:test';
import { afterEach } from 'node:test';
import { beforeEach } from 'node:test';
import { describe } from 'node:test';
import { it } from 'node:test';
import { it } from 'node:test';
import { it } from 'node:test';
import { it } from 'node:test';
import { it } from 'node:test';
import { it } from 'node:test';
import { it } from 'node:test';
import { it } from 'node:test';
import { it } from 'node:test';
import { afterEach } from 'node:test';
import { beforeEach } from 'node:test';
import { describe } from 'node:test';
import { reportStatus, getWorkflowRunId } from './status-reporter';

// Mock @actions/core
jest.mock('@actions/core');

describe('reportStatus', () => {
  let originalFetch: typeof global.fetch;
  
  beforeEach(() => {
    originalFetch = global.fetch;
    jest.clearAllMocks();
  });
  
  afterEach(() => {
    global.fetch = originalFetch;
  });

  it('should successfully report status on first attempt', async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      status: 200,
    });
    global.fetch = mockFetch as any;

    await reportStatus('http://localhost:8080', {
      incident_id: 'inc_123',
      status: 'success',
      pr_url: 'https://github.com/org/repo/pull/1',
      diagnosis: 'Fixed null pointer exception',
      repository: 'org/repo',
    });

    expect(mockFetch).toHaveBeenCalledTimes(1);
    expect(mockFetch).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/webhooks/workflow-status',
      expect.objectContaining({
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      })
    );

    const callArgs = mockFetch.mock.calls[0];
    const body = JSON.parse(callArgs[1].body);
    expect(body).toEqual({
      incident_id: 'inc_123',
      status: 'success',
      pr_url: 'https://github.com/org/repo/pull/1',
      diagnosis: 'Fixed null pointer exception',
      repository: 'org/repo',
    });
  });

  it('should truncate long diagnosis to 500 characters', async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      status: 200,
    });
    global.fetch = mockFetch as any;

    const longDiagnosis = 'a'.repeat(1000);
    
    await reportStatus('http://localhost:8080', {
      incident_id: 'inc_123',
      status: 'failed',
      diagnosis: longDiagnosis,
      repository: 'org/repo',
    });

    const callArgs = mockFetch.mock.calls[0];
    const body = JSON.parse(callArgs[1].body);
    expect(body.diagnosis).toHaveLength(500);
    expect(body.diagnosis).toBe('a'.repeat(500));
  });

  it('should retry on network failure with exponential backoff', async () => {
    const mockFetch = jest.fn()
      .mockRejectedValueOnce(new Error('Network error'))
      .mockRejectedValueOnce(new Error('Network error'))
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
      });
    global.fetch = mockFetch as any;

    const startTime = Date.now();
    await reportStatus('http://localhost:8080', {
      incident_id: 'inc_123',
      status: 'success',
      repository: 'org/repo',
    });
    const duration = Date.now() - startTime;

    expect(mockFetch).toHaveBeenCalledTimes(3);
    // Should have delays of 1s and 2s = 3s total minimum
    expect(duration).toBeGreaterThanOrEqual(3000);
  });

  it('should retry on HTTP error response', async () => {
    const mockFetch = jest.fn()
      .mockResolvedValueOnce({
        ok: false,
        status: 500,
        statusText: 'Internal Server Error',
        text: jest.fn().mockResolvedValue('Server error details'),
      })
      .mockResolvedValueOnce({
        ok: true,
        status: 200,
      });
    global.fetch = mockFetch as any;

    await reportStatus('http://localhost:8080', {
      incident_id: 'inc_123',
      status: 'success',
      repository: 'org/repo',
    });

    expect(mockFetch).toHaveBeenCalledTimes(2);
  });

  it('should not throw after exhausting retries', async () => {
    const mockFetch = jest.fn().mockRejectedValue(new Error('Network error'));
    global.fetch = mockFetch as any;

    // Should not throw
    await expect(
      reportStatus('http://localhost:8080', {
        incident_id: 'inc_123',
        status: 'success',
        repository: 'org/repo',
      })
    ).resolves.toBeUndefined();

    expect(mockFetch).toHaveBeenCalledTimes(3);
  });

  it('should skip reporting if no URL provided', async () => {
    const mockFetch = jest.fn();
    global.fetch = mockFetch as any;

    await reportStatus('', {
      incident_id: 'inc_123',
      status: 'success',
      repository: 'org/repo',
    });

    expect(mockFetch).not.toHaveBeenCalled();
  });

  it('should include workflow_run_id in payload', async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      status: 200,
    });
    global.fetch = mockFetch as any;

    await reportStatus('http://localhost:8080', {
      incident_id: 'inc_123',
      status: 'success',
      repository: 'org/repo',
      workflow_run_id: 123456789,
    });

    const callArgs = mockFetch.mock.calls[0];
    const body = JSON.parse(callArgs[1].body);
    expect(body.workflow_run_id).toBe(123456789);
  });

  it('should handle all status types', async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      status: 200,
    });
    global.fetch = mockFetch as any;

    const statuses: Array<'success' | 'failed' | 'no_fix_needed'> = [
      'success',
      'failed',
      'no_fix_needed',
    ];

    for (const status of statuses) {
      await reportStatus('http://localhost:8080', {
        incident_id: 'inc_123',
        status,
        repository: 'org/repo',
      });
    }

    expect(mockFetch).toHaveBeenCalledTimes(statuses.length);
  });

  it('should include repository in payload', async () => {
    const mockFetch = jest.fn().mockResolvedValue({
      ok: true,
      status: 200,
    });
    global.fetch = mockFetch as any;

    await reportStatus('http://localhost:8080', {
      incident_id: 'inc_123',
      status: 'success',
      repository: 'myorg/myrepo',
    });

    const callArgs = mockFetch.mock.calls[0];
    const body = JSON.parse(callArgs[1].body);
    expect(body.repository).toBe('myorg/myrepo');
  });
});

describe('getWorkflowRunId', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    jest.resetModules();
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it('should return workflow run ID from environment', () => {
    process.env.GITHUB_RUN_ID = '123456789';
    expect(getWorkflowRunId()).toBe(123456789);
  });

  it('should return undefined if not set', () => {
    delete process.env.GITHUB_RUN_ID;
    expect(getWorkflowRunId()).toBeUndefined();
  });

  it('should parse numeric string correctly', () => {
    process.env.GITHUB_RUN_ID = '987654321';
    expect(getWorkflowRunId()).toBe(987654321);
  });
});
