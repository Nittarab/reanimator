/**
 * Unit tests for notification functions
 */

import { sendSlackNotification, sendCustomWebhook, NotificationPayload } from './notifications';

// Mock fetch globally
global.fetch = jest.fn();

describe('Notification Functions', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  const mockPayload: NotificationPayload = {
    incidentId: 'inc_123',
    serviceName: 'api-gateway',
    severity: 'high',
    errorMessage: 'NullPointerException in UserService',
    prUrl: 'https://github.com/org/repo/pull/42',
    timestamp: '2024-01-15T10:00:00Z',
  };

  describe('sendSlackNotification', () => {
    it('should send Slack notification with correct format', async () => {
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
      });

      await sendSlackNotification('https://hooks.slack.com/test', mockPayload);

      expect(global.fetch).toHaveBeenCalledTimes(1);
      const call = (global.fetch as jest.Mock).mock.calls[0];
      expect(call[0]).toBe('https://hooks.slack.com/test');
      
      const body = JSON.parse(call[1].body);
      expect(body.text).toContain(mockPayload.serviceName);
      expect(body.blocks).toBeDefined();
      expect(body.blocks.length).toBeGreaterThan(0);
    });

    it('should include all required fields in Slack message', async () => {
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
      });

      await sendSlackNotification('https://hooks.slack.com/test', mockPayload);

      const call = (global.fetch as jest.Mock).mock.calls[0];
      const body = JSON.parse(call[1].body);
      const bodyString = JSON.stringify(body);
      
      expect(bodyString).toContain(mockPayload.incidentId);
      expect(bodyString).toContain(mockPayload.serviceName);
      expect(bodyString).toContain(mockPayload.severity);
      expect(bodyString).toContain(mockPayload.prUrl);
    });

    it('should throw error on failed Slack notification', async () => {
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 500,
      });

      await expect(
        sendSlackNotification('https://hooks.slack.com/test', mockPayload)
      ).rejects.toThrow('Slack notification failed');
    });

    it('should truncate long error messages', async () => {
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
      });

      const longPayload = {
        ...mockPayload,
        errorMessage: 'A'.repeat(500),
      };

      await sendSlackNotification('https://hooks.slack.com/test', longPayload);

      const call = (global.fetch as jest.Mock).mock.calls[0];
      const body = JSON.parse(call[1].body);
      const bodyString = JSON.stringify(body);
      
      // Should be truncated to 200 chars + "..."
      expect(bodyString).toContain('...');
    });
  });

  describe('sendCustomWebhook', () => {
    it('should send custom webhook with payload', async () => {
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: true,
        status: 200,
      });

      await sendCustomWebhook('https://example.com/webhook', mockPayload);

      expect(global.fetch).toHaveBeenCalledTimes(1);
      const call = (global.fetch as jest.Mock).mock.calls[0];
      expect(call[0]).toBe('https://example.com/webhook');
      
      const body = JSON.parse(call[1].body);
      expect(body).toEqual(mockPayload);
    });

    it('should throw error on failed custom webhook', async () => {
      (global.fetch as jest.Mock).mockResolvedValueOnce({
        ok: false,
        status: 404,
      });

      await expect(
        sendCustomWebhook('https://example.com/webhook', mockPayload)
      ).rejects.toThrow('Custom webhook notification failed');
    });
  });
});
