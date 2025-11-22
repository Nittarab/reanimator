import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getIncidents, getIncident, triggerRemediation } from './incidents'
import { apiClient } from './client'
import type { Incident, IncidentListResponse } from './types'

vi.mock('./client')

describe('Incidents API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('getIncidents', () => {
    it('fetches incidents without filters', async () => {
      const mockResponse: IncidentListResponse = {
        incidents: [],
        total: 0,
      }

      vi.mocked(apiClient.get).mockResolvedValue({ data: mockResponse })

      const result = await getIncidents()

      expect(apiClient.get).toHaveBeenCalledWith('/incidents?')
      expect(result).toEqual(mockResponse)
    })

    it('fetches incidents with filters', async () => {
      const mockResponse: IncidentListResponse = {
        incidents: [],
        total: 0,
      }

      vi.mocked(apiClient.get).mockResolvedValue({ data: mockResponse })

      await getIncidents({
        status: 'pending',
        service: 'api-gateway',
      })

      expect(apiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('status=pending')
      )
      expect(apiClient.get).toHaveBeenCalledWith(
        expect.stringContaining('service=api-gateway')
      )
    })
  })

  describe('getIncident', () => {
    it('fetches a single incident by ID', async () => {
      const mockIncident: Incident = {
        id: 'inc_123',
        service_name: 'api-gateway',
        repository: 'org/api-gateway',
        error_message: 'Test error',
        severity: 'high',
        status: 'pending',
        provider: 'datadog',
        provider_data: {},
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      vi.mocked(apiClient.get).mockResolvedValue({ data: mockIncident })

      const result = await getIncident('inc_123')

      expect(apiClient.get).toHaveBeenCalledWith('/incidents/inc_123')
      expect(result).toEqual(mockIncident)
    })
  })

  describe('triggerRemediation', () => {
    it('triggers remediation for an incident', async () => {
      vi.mocked(apiClient.post).mockResolvedValue({ data: {} })

      await triggerRemediation('inc_123')

      expect(apiClient.post).toHaveBeenCalledWith('/incidents/inc_123/trigger')
    })
  })
})
