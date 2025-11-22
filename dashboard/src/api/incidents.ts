import { apiClient } from './client'
import type {
  Incident,
  IncidentEvent,
  IncidentFilters,
  IncidentListResponse,
} from './types'

export const getIncidents = async (
  filters?: IncidentFilters
): Promise<IncidentListResponse> => {
  const params = new URLSearchParams()
  
  if (filters?.status) params.append('status', filters.status)
  if (filters?.service) params.append('service', filters.service)
  if (filters?.repository) params.append('repository', filters.repository)
  if (filters?.start_time) params.append('start_time', filters.start_time)
  if (filters?.end_time) params.append('end_time', filters.end_time)

  const response = await apiClient.get<IncidentListResponse>(
    `/incidents?${params.toString()}`
  )
  return response.data
}

export const getIncident = async (id: string): Promise<Incident> => {
  const response = await apiClient.get<Incident>(`/incidents/${id}`)
  return response.data
}

export const getIncidentEvents = async (
  id: string
): Promise<IncidentEvent[]> => {
  const response = await apiClient.get<IncidentEvent[]>(
    `/incidents/${id}/events`
  )
  return response.data
}

export const triggerRemediation = async (id: string): Promise<void> => {
  await apiClient.post(`/incidents/${id}/trigger`)
}
