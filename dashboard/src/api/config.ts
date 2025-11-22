import { apiClient } from './client'

export interface ServiceMapping {
  service_name: string
  repository: string
  branch: string
}

export interface ConfigResponse {
  service_mappings: ServiceMapping[]
}

export async function getConfig(): Promise<ConfigResponse> {
  const response = await apiClient.get<ConfigResponse>('/config')
  return response.data
}
