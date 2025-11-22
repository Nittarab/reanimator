export type IncidentStatus =
  | 'pending'
  | 'workflow_triggered'
  | 'in_progress'
  | 'pr_created'
  | 'resolved'
  | 'failed'
  | 'no_fix_needed'

export interface Incident {
  id: string
  service_name: string
  repository: string
  error_message: string
  stack_trace?: string
  severity: string
  status: IncidentStatus
  provider: string
  provider_data: Record<string, unknown>
  workflow_run_id?: number
  pull_request_url?: string
  diagnosis?: string
  created_at: string
  updated_at: string
  triggered_at?: string
  completed_at?: string
}

export interface IncidentEvent {
  id: string
  incident_id: string
  event_type: string
  details: Record<string, unknown>
  created_at: string
}

export interface IncidentFilters {
  status?: IncidentStatus
  service?: string
  repository?: string
  start_time?: string
  end_time?: string
}

export interface IncidentListResponse {
  incidents: Incident[]
  total: number
}

export interface IncidentStats {
  total: number
  success_rate: number
  mean_time_to_resolution: number
  by_status: Record<IncidentStatus, number>
}
