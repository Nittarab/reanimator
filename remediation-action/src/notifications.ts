/**
 * Notification handling for the remediation action
 */

import * as core from '@actions/core';

export interface NotificationPayload {
  incidentId: string;
  serviceName: string;
  severity: string;
  errorMessage: string;
  prUrl: string;
  timestamp: string;
}

/**
 * Send Slack notification
 * @param webhookUrl - Slack webhook URL
 * @param payload - Notification payload
 */
export async function sendSlackNotification(
  webhookUrl: string,
  payload: NotificationPayload
): Promise<void> {
  const message = {
    text: `🚨 Incident Remediation: ${payload.serviceName}`,
    blocks: [
      {
        type: 'header',
        text: {
          type: 'plain_text',
          text: `🚨 Incident Remediation: ${payload.serviceName}`,
        },
      },
      {
        type: 'section',
        fields: [
          {
            type: 'mrkdwn',
            text: `*Incident ID:*\n${payload.incidentId}`,
          },
          {
            type: 'mrkdwn',
            text: `*Severity:*\n${payload.severity}`,
          },
          {
            type: 'mrkdwn',
            text: `*Service:*\n${payload.serviceName}`,
          },
          {
            type: 'mrkdwn',
            text: `*Timestamp:*\n${payload.timestamp}`,
          },
        ],
      },
      {
        type: 'section',
        text: {
          type: 'mrkdwn',
          text: `*Error:*\n${payload.errorMessage.substring(0, 200)}${payload.errorMessage.length > 200 ? '...' : ''}`,
        },
      },
      {
        type: 'section',
        text: {
          type: 'mrkdwn',
          text: `*Pull Request:*\n<${payload.prUrl}|View PR>`,
        },
      },
    ],
  };

  const response = await fetch(webhookUrl, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(message),
  });

  if (!response.ok) {
    throw new Error(`Slack notification failed: HTTP ${response.status}`);
  }
}

/**
 * Send custom webhook notification
 * @param webhookUrl - Custom webhook URL
 * @param payload - Notification payload
 */
export async function sendCustomWebhook(
  webhookUrl: string,
  payload: NotificationPayload
): Promise<void> {
  const response = await fetch(webhookUrl, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    throw new Error(`Custom webhook notification failed: HTTP ${response.status}`);
  }
}

/**
 * Send all configured notifications
 * @param payload - Notification payload
 */
export async function sendNotifications(payload: NotificationPayload): Promise<void> {
  const slackWebhookUrl = process.env.SLACK_WEBHOOK_URL || '';
  const customWebhookUrl = process.env.CUSTOM_WEBHOOK_URL || '';

  const notifications: Promise<void>[] = [];

  // Send Slack notification if configured
  if (slackWebhookUrl) {
    core.info('Sending Slack notification...');
    notifications.push(
      sendSlackNotification(slackWebhookUrl, payload)
        .then(() => {
          core.info('Slack notification sent successfully');
        })
        .catch((error) => {
          const errorMessage = error instanceof Error ? error.message : String(error);
          core.warning(`Failed to send Slack notification: ${errorMessage}`);
          // Don't throw - we want to continue with other notifications
        })
    );
  }

  // Send custom webhook notification if configured
  if (customWebhookUrl) {
    core.info('Sending custom webhook notification...');
    notifications.push(
      sendCustomWebhook(customWebhookUrl, payload)
        .then(() => {
          core.info('Custom webhook notification sent successfully');
        })
        .catch((error) => {
          const errorMessage = error instanceof Error ? error.message : String(error);
          core.warning(`Failed to send custom webhook notification: ${errorMessage}`);
          // Don't throw - we want to continue with other notifications
        })
    );
  }

  // Wait for all notifications to complete (or fail gracefully)
  await Promise.all(notifications);

  if (notifications.length === 0) {
    core.info('No notification webhooks configured');
  }
}
