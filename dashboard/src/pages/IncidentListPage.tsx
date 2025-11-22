import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { getIncidents } from '@/api/incidents'
import type { IncidentFilters, IncidentStatus } from '@/api/types'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

const statusColors: Record<IncidentStatus, string> = {
  pending: 'bg-yellow-500',
  workflow_triggered: 'bg-blue-500',
  in_progress: 'bg-blue-600',
  pr_created: 'bg-green-500',
  resolved: 'bg-green-700',
  failed: 'bg-red-500',
  no_fix_needed: 'bg-gray-500',
}

const statusLabels: Record<IncidentStatus, string> = {
  pending: 'Pending',
  workflow_triggered: 'Workflow Triggered',
  in_progress: 'In Progress',
  pr_created: 'PR Created',
  resolved: 'Resolved',
  failed: 'Failed',
  no_fix_needed: 'No Fix Needed',
}

export function IncidentListPage() {
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [serviceFilter, setServiceFilter] = useState<string>('')
  const [repositoryFilter, setRepositoryFilter] = useState<string>('')

  // Build filters object
  const activeFilters: IncidentFilters = {
    ...(statusFilter !== 'all' && { status: statusFilter as IncidentStatus }),
    ...(serviceFilter && { service: serviceFilter }),
    ...(repositoryFilter && { repository: repositoryFilter }),
  }

  // Fetch incidents with polling every 10 seconds
  const { data, isLoading, error } = useQuery({
    queryKey: ['incidents', activeFilters],
    queryFn: () => getIncidents(activeFilters),
    refetchInterval: 10000, // Poll every 10 seconds
  })

  // Sort incidents by timestamp (most recent first)
  const sortedIncidents = data?.incidents
    ? [...data.incidents].sort(
        (a, b) =>
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      )
    : []

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp)
    return date.toLocaleString()
  }

  const truncateMessage = (message: string, maxLength: number = 100) => {
    if (message.length <= maxLength) return message
    return message.substring(0, maxLength) + '...'
  }

  return (
    <div>
      <h2 className="text-3xl font-bold mb-6">Incidents</h2>

      {/* Filters */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle className="text-lg">Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">Status</label>
              <select
                className="w-full h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
              >
                <option value="all">All</option>
                <option value="pending">Pending</option>
                <option value="workflow_triggered">Workflow Triggered</option>
                <option value="in_progress">In Progress</option>
                <option value="pr_created">PR Created</option>
                <option value="resolved">Resolved</option>
                <option value="failed">Failed</option>
                <option value="no_fix_needed">No Fix Needed</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">Service</label>
              <Input
                placeholder="Filter by service..."
                value={serviceFilter}
                onChange={(e) => setServiceFilter(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">
                Repository
              </label>
              <Input
                placeholder="Filter by repository..."
                value={repositoryFilter}
                onChange={(e) => setRepositoryFilter(e.target.value)}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Loading state */}
      {isLoading && (
        <div className="text-center py-8 text-muted-foreground">
          Loading incidents...
        </div>
      )}

      {/* Error state */}
      {error && (
        <div className="text-center py-8 text-red-500">
          Error loading incidents: {error.message}
        </div>
      )}

      {/* Incidents table */}
      {!isLoading && !error && (
        <Card>
          <CardContent className="p-0">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="border-b bg-muted/50">
                  <tr>
                    <th className="px-4 py-3 text-left text-sm font-medium">
                      Status
                    </th>
                    <th className="px-4 py-3 text-left text-sm font-medium">
                      Service
                    </th>
                    <th className="px-4 py-3 text-left text-sm font-medium">
                      Repository
                    </th>
                    <th className="px-4 py-3 text-left text-sm font-medium">
                      Error Message
                    </th>
                    <th className="px-4 py-3 text-left text-sm font-medium">
                      Time
                    </th>
                    <th className="px-4 py-3 text-left text-sm font-medium">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {sortedIncidents.length === 0 ? (
                    <tr>
                      <td
                        colSpan={6}
                        className="px-4 py-8 text-center text-muted-foreground"
                      >
                        No incidents found
                      </td>
                    </tr>
                  ) : (
                    sortedIncidents.map((incident) => (
                      <tr
                        key={incident.id}
                        className="border-b hover:bg-muted/50 transition-colors"
                      >
                        <td className="px-4 py-3">
                          <Badge
                            className={`${statusColors[incident.status]} text-white`}
                          >
                            {statusLabels[incident.status]}
                          </Badge>
                        </td>
                        <td className="px-4 py-3 text-sm">
                          {incident.service_name}
                        </td>
                        <td className="px-4 py-3 text-sm">
                          {incident.repository || '-'}
                        </td>
                        <td className="px-4 py-3 text-sm max-w-md">
                          {truncateMessage(incident.error_message)}
                        </td>
                        <td className="px-4 py-3 text-sm text-muted-foreground">
                          {formatTimestamp(incident.created_at)}
                        </td>
                        <td className="px-4 py-3">
                          <Link
                            to={`/incidents/${incident.id}`}
                            className="text-sm text-blue-600 hover:text-blue-800 hover:underline"
                          >
                            View Details
                          </Link>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Total count */}
      {!isLoading && !error && data && (
        <div className="mt-4 text-sm text-muted-foreground">
          Showing {sortedIncidents.length} incident
          {sortedIncidents.length !== 1 ? 's' : ''}
        </div>
      )}
    </div>
  )
}
