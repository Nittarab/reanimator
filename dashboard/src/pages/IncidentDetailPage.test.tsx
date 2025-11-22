import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { IncidentDetailPage } from './IncidentDetailPage'
import * as incidentsApi from '@/api/incidents'
import type { Incident, IncidentEvent } from '@/api/types'

const mockIncident: Incident = {
  id: 'inc_test_123',
  service_name: 'api-gateway',
  repository: 'org/api-gateway',
  error_message: 'NullPointerException in UserService',
  stack_trace: 'at UserService.getUser(UserService.java:42)',
  severity: 'high',
  status: 'pr_created',
  provider: 'datadog',
  provider_data: {},
  workflow_run_id: 123456,
  pull_request_url: 'https://github.com/org/api-gateway/pull/1',
  diagnosis: 'Null check missing',
  created_at: '2024-01-15T10:00:00Z',
  updated_at: '2024-01-15T10:05:00Z',
  triggered_at: '2024-01-15T10:01:00Z',
  completed_at: '2024-01-15T10:05:00Z',
}

const mockEvents: IncidentEvent[] = [
  {
    id: 'evt_1',
    incident_id: 'inc_test_123',
    event_type: 'incident_created',
    details: {},
    created_at: '2024-01-15T10:00:00Z',
  },
  {
    id: 'evt_2',
    incident_id: 'inc_test_123',
    event_type: 'workflow_triggered',
    details: { workflow_run_id: 123456 },
    created_at: '2024-01-15T10:01:00Z',
  },
]

describe('IncidentDetailPage', () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })

  const renderWithRouter = () => {
    return render(
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Routes>
            <Route path="/incidents/:id" element={<IncidentDetailPage />} />
          </Routes>
        </BrowserRouter>
      </QueryClientProvider>,
      { wrapper: ({ children }) => <div>{children}</div> }
    )
  }

  it('should display incident details', async () => {
    vi.spyOn(incidentsApi, 'getIncident').mockResolvedValue(mockIncident)
    vi.spyOn(incidentsApi, 'getIncidentEvents').mockResolvedValue(mockEvents)

    window.history.pushState({}, '', '/incidents/inc_test_123')
    renderWithRouter()

    await waitFor(() => {
      expect(screen.getByText('Incident Details')).toBeInTheDocument()
    })

    expect(screen.getByText('api-gateway')).toBeInTheDocument()
    expect(screen.getByText('org/api-gateway')).toBeInTheDocument()
    expect(screen.getByText('NullPointerException in UserService')).toBeInTheDocument()
  })

  it('should display stack trace when available', async () => {
    vi.spyOn(incidentsApi, 'getIncident').mockResolvedValue(mockIncident)
    vi.spyOn(incidentsApi, 'getIncidentEvents').mockResolvedValue(mockEvents)

    window.history.pushState({}, '', '/incidents/inc_test_123')
    renderWithRouter()

    await waitFor(() => {
      expect(screen.getByText('Stack Trace')).toBeInTheDocument()
    })

    expect(screen.getByText(/at UserService\.getUser/)).toBeInTheDocument()
  })

  it('should display timeline events', async () => {
    vi.spyOn(incidentsApi, 'getIncident').mockResolvedValue(mockIncident)
    vi.spyOn(incidentsApi, 'getIncidentEvents').mockResolvedValue(mockEvents)

    window.history.pushState({}, '', '/incidents/inc_test_123')
    renderWithRouter()

    await waitFor(() => {
      expect(screen.getByText('Timeline')).toBeInTheDocument()
    })

    expect(screen.getByText(/INCIDENT CREATED/)).toBeInTheDocument()
    expect(screen.getByText(/WORKFLOW TRIGGERED/)).toBeInTheDocument()
  })

  it('should display links to GitHub workflow and PR', async () => {
    vi.spyOn(incidentsApi, 'getIncident').mockResolvedValue(mockIncident)
    vi.spyOn(incidentsApi, 'getIncidentEvents').mockResolvedValue(mockEvents)

    window.history.pushState({}, '', '/incidents/inc_test_123')
    renderWithRouter()

    await waitFor(() => {
      expect(screen.getByText('Actions & Links')).toBeInTheDocument()
    })

    const workflowLink = screen.getByText('View GitHub Workflow →')
    expect(workflowLink).toHaveAttribute(
      'href',
      'https://github.com/org/api-gateway/actions/runs/123456'
    )

    const prLink = screen.getByText('View Pull Request →')
    expect(prLink).toHaveAttribute('href', 'https://github.com/org/api-gateway/pull/1')
  })

  it('should disable trigger button when workflow is active', async () => {
    const activeIncident = { ...mockIncident, status: 'in_progress' as const }
    vi.spyOn(incidentsApi, 'getIncident').mockResolvedValue(activeIncident)
    vi.spyOn(incidentsApi, 'getIncidentEvents').mockResolvedValue(mockEvents)

    window.history.pushState({}, '', '/incidents/inc_test_123')
    renderWithRouter()

    await waitFor(() => {
      const button = screen.getByRole('button', { name: /Trigger Remediation/ })
      expect(button).toBeDisabled()
    })

    expect(screen.getByText(/Workflow is currently active/)).toBeInTheDocument()
  })

  it('should enable trigger button when workflow is not active', async () => {
    const failedIncident = { ...mockIncident, status: 'failed' as const }
    vi.spyOn(incidentsApi, 'getIncident').mockResolvedValue(failedIncident)
    vi.spyOn(incidentsApi, 'getIncidentEvents').mockResolvedValue(mockEvents)

    window.history.pushState({}, '', '/incidents/inc_test_123')
    renderWithRouter()

    await waitFor(() => {
      const button = screen.getByRole('button', { name: /Trigger Remediation/ })
      expect(button).not.toBeDisabled()
    })
  })
})
