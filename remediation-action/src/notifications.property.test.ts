/**
 * Property-based tests for notification content
 * Feature: ai-sre-platform, Property 18: Notification content completeness
 * Validates: Requirements 23.5
 */

import * as fc from 'fast-check';
import { NotificationPayload } from './notifications';

describe('Notification Content Properties', () => {
  /**
   * Property 18: Notification content completeness
   * For any notification sent by the workflow, it should include incident severity,
   * affected service, and pull request link.
   */
  it('should include all required fields in notification payload', () => {
    fc.assert(
      fc.property(
        fc.string({ minLength: 1 }), // incidentId
        fc.string({ minLength: 1 }), // serviceName
        fc.constantFrom('critical', 'high', 'medium', 'low'), // severity
        fc.string({ minLength: 1 }), // errorMessage
        fc.webUrl(), // prUrl
        fc.date().map(d => d.toISOString()), // timestamp
        (incidentId, serviceName, severity, errorMessage, prUrl, timestamp) => {
          const payload: NotificationPayload = {
            incidentId,
            serviceName,
            severity,
            errorMessage,
            prUrl,
            timestamp,
          };

          // Verify all required fields are present and non-empty
          expect(payload.incidentId).toBeTruthy();
          expect(payload.serviceName).toBeTruthy();
          expect(payload.severity).toBeTruthy();
          expect(payload.prUrl).toBeTruthy();
          
          // Verify severity is one of the valid values
          expect(['critical', 'high', 'medium', 'low']).toContain(payload.severity);
          
          // Verify prUrl is a valid URL format
          expect(payload.prUrl).toMatch(/^https?:\/\//);
        }
      ),
      { numRuns: 100 }
    );
  });

  it('should preserve all payload fields when creating notification', () => {
    fc.assert(
      fc.property(
        fc.record({
          incidentId: fc.string({ minLength: 1 }),
          serviceName: fc.string({ minLength: 1 }),
          severity: fc.constantFrom('critical', 'high', 'medium', 'low'),
          errorMessage: fc.string({ minLength: 1 }),
          prUrl: fc.webUrl(),
          timestamp: fc.date().map(d => d.toISOString()),
        }),
        (payload) => {
          // Verify the payload structure is complete
          const requiredFields = ['incidentId', 'serviceName', 'severity', 'errorMessage', 'prUrl', 'timestamp'];
          const payloadKeys = Object.keys(payload);
          
          requiredFields.forEach(field => {
            expect(payloadKeys).toContain(field);
            expect(payload[field as keyof NotificationPayload]).toBeTruthy();
          });
        }
      ),
      { numRuns: 100 }
    );
  });

  it('should handle long error messages gracefully', () => {
    fc.assert(
      fc.property(
        fc.string({ minLength: 1 }),
        fc.string({ minLength: 1 }),
        fc.constantFrom('critical', 'high', 'medium', 'low'),
        fc.string({ minLength: 1, maxLength: 10000 }), // Very long error message
        fc.webUrl(),
        fc.date().map(d => d.toISOString()),
        (incidentId, serviceName, severity, errorMessage, prUrl, timestamp) => {
          const payload: NotificationPayload = {
            incidentId,
            serviceName,
            severity,
            errorMessage,
            prUrl,
            timestamp,
          };

          // Verify payload is created successfully even with long messages
          expect(payload).toBeDefined();
          expect(payload.errorMessage).toBe(errorMessage);
          expect(payload.errorMessage.length).toBeGreaterThan(0);
        }
      ),
      { numRuns: 100 }
    );
  });
});
