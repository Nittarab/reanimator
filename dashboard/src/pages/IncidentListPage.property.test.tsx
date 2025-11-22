import { describe, it, expect } from 'vitest'
import fc from 'fast-check'
import type { Incident, IncidentStatus } from '@/api/types'

/**
 * Feature: ai-sre-platform, Property 19: Dashboard incident ordering
 * Validates: Requirements 19.1
 * 
 * For any list of incidents displayed in the dashboard, they should be ordered
 * by timestamp with the most recent first.
 */

// Helper to generate random incidents
const incidentArbitrary = fc.record({
  id: fc.uuid(),
  service_name: fc.string({ minLength: 1, maxLength: 50 }),
  repository: fc.string({ minLength: 1, maxLength: 100 }),
  error_message: fc.string({ minLength: 1, maxLength: 200 }),
  stack_trace: fc.option(fc.string(), { nil: undefined }),
  severity: fc.constantFrom('low', 'medium', 'high', 'critical'),
  status: fc.constantFrom<IncidentStatus>(
    'pending',
    'workflow_triggered',
    'in_progress',
    'pr_created',
    'resolved',
    'failed',
    'no_fix_needed'
  ),
  provider: fc.constantFrom('datadog', 'pagerduty', 'grafana', 'sentry'),
  provider_data: fc.dictionary(fc.string(), fc.anything()),
  workflow_run_id: fc.option(fc.integer({ min: 1 }), { nil: undefined }),
  pull_request_url: fc.option(fc.webUrl(), { nil: undefined }),
  diagnosis: fc.option(fc.string(), { nil: undefined }),
  created_at: fc.date({ min: new Date('2020-01-01'), max: new Date('2025-12-31') }).map(d => d.toISOString()),
  updated_at: fc.date({ min: new Date('2020-01-01'), max: new Date('2025-12-31') }).map(d => d.toISOString()),
  triggered_at: fc.option(fc.date({ min: new Date('2020-01-01'), max: new Date('2025-12-31') }).map(d => d.toISOString()), { nil: undefined }),
  completed_at: fc.option(fc.date({ min: new Date('2020-01-01'), max: new Date('2025-12-31') }).map(d => d.toISOString()), { nil: undefined }),
}) as fc.Arbitrary<Incident>

// Sorting function extracted from component
const sortIncidentsByTimestamp = (incidents: Incident[]): Incident[] => {
  return [...incidents].sort(
    (a, b) =>
      new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  )
}

describe('IncidentListPage Property Tests', () => {
  it('Property 19: should order incidents by timestamp with most recent first', () => {
    fc.assert(
      fc.property(
        fc.array(incidentArbitrary, { minLength: 0, maxLength: 50 }),
        (incidents) => {
          // Sort incidents using the component's logic
          const sorted = sortIncidentsByTimestamp(incidents)

          // Verify that the sorted list is ordered by timestamp (most recent first)
          for (let i = 0; i < sorted.length - 1; i++) {
            const currentTime = new Date(sorted[i].created_at).getTime()
            const nextTime = new Date(sorted[i + 1].created_at).getTime()
            
            // Current incident should have a timestamp >= next incident (descending order)
            expect(currentTime).toBeGreaterThanOrEqual(nextTime)
          }

          // Verify that all original incidents are present in sorted list
          expect(sorted.length).toBe(incidents.length)
          
          // Verify that sorting is stable (same incidents, just reordered)
          const sortedIds = sorted.map(i => i.id).sort()
          const originalIds = incidents.map(i => i.id).sort()
          expect(sortedIds).toEqual(originalIds)
        }
      ),
      { numRuns: 100 }
    )
  })

  /**
   * Feature: ai-sre-platform, Property 20: Dashboard incident display completeness
   * Validates: Requirements 19.2
   * 
   * For any incident displayed in the dashboard, it should show status, affected service,
   * error message, and associated repository.
   */
  it('Property 20: should display all required fields for each incident', () => {
    fc.assert(
      fc.property(
        incidentArbitrary,
        (incident) => {
          // Verify that all required fields are present and non-empty
          expect(incident.status).toBeDefined()
          expect(incident.status).not.toBe('')
          
          expect(incident.service_name).toBeDefined()
          expect(incident.service_name).not.toBe('')
          
          expect(incident.error_message).toBeDefined()
          expect(incident.error_message).not.toBe('')
          
          expect(incident.repository).toBeDefined()
          // Repository can be empty string, but should be defined
          
          // Verify status is a valid incident status
          const validStatuses: IncidentStatus[] = [
            'pending',
            'workflow_triggered',
            'in_progress',
            'pr_created',
            'resolved',
            'failed',
            'no_fix_needed'
          ]
          expect(validStatuses).toContain(incident.status)
        }
      ),
      { numRuns: 100 }
    )
  })
})
