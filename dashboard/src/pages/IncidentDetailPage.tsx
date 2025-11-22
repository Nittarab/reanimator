import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getIncident, getIncidentEvents, triggerRemediation } from '@/api/incidents'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { IncidentStatus } from '@/api/types'

const statusColors: Record<IncidentStatus, string> = {
  pending: 'bg-yellow-500',
  workflow_triggered: 'bg-blue-500',
  in_progress: 'bg-blue-600',
  pr_created: 'bg-green-500',
  resolved: 'bg-green-700',
  failed: 'bg-red-500',
  no_fix_needed: 'bg-gray-500',
}

const severityColors: Record<string, string> = {
  critical: 'bg-red-600',
  high: 'bg-orange-500',
  medium: 'bg-yellow-500',
  low: 'bg-blue-500',
}

export function IncidentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  const { data: incident, isLoading: incidentLoading, error: incidentError } = useQuery({
    queryKey: ['incident', id],
    queryFn: () => getIncident(id!),
    enabled: !!id,
    refetchInterval: 10000, // Refresh every 10 seconds
  })

  const { data: events = [], isLoading: eventsLoading } = useQuery({
    queryKey: ['incident-events', id],
    queryFn: () => getIncidentEvents(id!),
    enabled: !!id,
    refetchInterval: 10000,
  })

  const triggerMutation = useMutation({
    mutationFn: () => triggerRemediation(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['incident', id] })
      queryClient.invalidateQueries({ queryKey: ['incident-events', id] })
    },
  })

  if (incidentLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-muted-foreground">Loading incident details...</p>
      </div>
    )
  }

  if (incidentError || !incident) {
    return (
      <div className="flex flex-col items-center justify-center h-64 space-y-4">
        <p className="text-destructive">Failed to load incident details</p>
        <Button onClick={() => navigate('/')}>Back to Incidents</Button>
      </div>
    )
  }

  const isWorkflowActive = incident.status === 'workflow_triggered' || incident.status === 'in_progress'
  const canTrigger = !isWorkflowActive && incident.status !== 'resolved'

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <Button variant="outline" onClick={() => navigate('/')} className="mb-4">
            ← Back to Incidents
          </Button>
          <h2 className="text-3xl font-bold">Incident Details</h2>
          <p className="text-muted-foreground mt-1">ID: {incident.id}</p>
        </div>
        <div className="flex items-center gap-4">
          <Badge className={statusColors[incident.status]}>
            {incident.status.replace('_', ' ').toUpperCase()}
          </Badge>
          <Badge className={severityColors[incident.severity] || 'bg-gray-500'}>
            {incident.severity.toUpperCase()}
          </Badge>
        </div>
      </div>

      {/* Main Incident Information */}
      <Card>
        <CardHeader>
          <CardTitle>Incident Information</CardTitle>
          <CardDescription>
            Reported by {incident.provider} on {new Date(incident.created_at).toLocaleString()}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <h4 className="font-semibold mb-1">Service</h4>
            <p className="text-sm">{incident.service_name}</p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Repository</h4>
            <p className="text-sm">{incident.repository || 'Not mapped'}</p>
          </div>
          <div>
            <h4 className="font-semibold mb-1">Error Message</h4>
            <p className="text-sm text-destructive">{incident.error_message}</p>
          </div>
          {incident.stack_trace && (
            <div>
              <h4 className="font-semibold mb-1">Stack Trace</h4>
              <pre className="text-xs bg-muted p-4 rounded-md overflow-x-auto whitespace-pre-wrap">
                {incident.stack_trace}
              </pre>
            </div>
          )}
          {incident.diagnosis && (
            <div>
              <h4 className="font-semibold mb-1">Diagnosis</h4>
              <p className="text-sm">{incident.diagnosis}</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Links and Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Actions & Links</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-wrap gap-4">
            {incident.workflow_run_id && (
              <a
                href={`https://github.com/${incident.repository}/actions/runs/${incident.workflow_run_id}`}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-blue-600 hover:underline"
              >
                View GitHub Workflow →
              </a>
            )}
            {incident.pull_request_url && (
              <a
                href={incident.pull_request_url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-blue-600 hover:underline"
              >
                View Pull Request →
              </a>
            )}
          </div>
          <div>
            <Button
              onClick={() => triggerMutation.mutate()}
              disabled={!canTrigger || triggerMutation.isPending}
            >
              {triggerMutation.isPending ? 'Triggering...' : 'Trigger Remediation'}
            </Button>
            {isWorkflowActive && (
              <p className="text-sm text-muted-foreground mt-2">
                Workflow is currently active. Manual trigger is disabled.
              </p>
            )}
            {triggerMutation.isError && (
              <p className="text-sm text-destructive mt-2">
                Failed to trigger remediation. Please try again.
              </p>
            )}
            {triggerMutation.isSuccess && (
              <p className="text-sm text-green-600 mt-2">
                Remediation triggered successfully!
              </p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Timeline */}
      <Card>
        <CardHeader>
          <CardTitle>Timeline</CardTitle>
          <CardDescription>History of events for this incident</CardDescription>
        </CardHeader>
        <CardContent>
          {eventsLoading ? (
            <p className="text-sm text-muted-foreground">Loading timeline...</p>
          ) : events.length === 0 ? (
            <p className="text-sm text-muted-foreground">No events recorded yet</p>
          ) : (
            <div className="space-y-4">
              {events.map((event) => (
                <div key={event.id} className="flex gap-4 border-l-2 border-muted pl-4">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-sm">
                        {event.event_type.replace('_', ' ').toUpperCase()}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {new Date(event.created_at).toLocaleString()}
                      </span>
                    </div>
                    {Object.keys(event.details).length > 0 && (
                      <pre className="text-xs text-muted-foreground mt-1 overflow-x-auto">
                        {JSON.stringify(event.details, null, 2)}
                      </pre>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Timestamps */}
      <Card>
        <CardHeader>
          <CardTitle>Timestamps</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-2 gap-4">
          <div>
            <h4 className="font-semibold text-sm mb-1">Created</h4>
            <p className="text-sm text-muted-foreground">
              {new Date(incident.created_at).toLocaleString()}
            </p>
          </div>
          <div>
            <h4 className="font-semibold text-sm mb-1">Updated</h4>
            <p className="text-sm text-muted-foreground">
              {new Date(incident.updated_at).toLocaleString()}
            </p>
          </div>
          {incident.triggered_at && (
            <div>
              <h4 className="font-semibold text-sm mb-1">Triggered</h4>
              <p className="text-sm text-muted-foreground">
                {new Date(incident.triggered_at).toLocaleString()}
              </p>
            </div>
          )}
          {incident.completed_at && (
            <div>
              <h4 className="font-semibold text-sm mb-1">Completed</h4>
              <p className="text-sm text-muted-foreground">
                {new Date(incident.completed_at).toLocaleString()}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
