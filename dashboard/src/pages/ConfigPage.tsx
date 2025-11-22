import { useQuery } from '@tanstack/react-query'
import { getConfig } from '@/api/config'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

export function ConfigPage() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['config'],
    queryFn: getConfig,
  })

  if (isLoading) {
    return (
      <div>
        <h2 className="text-3xl font-bold mb-6">Configuration</h2>
        <p className="text-muted-foreground">Loading configuration...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div>
        <h2 className="text-3xl font-bold mb-6">Configuration</h2>
        <p className="text-red-500">Failed to load configuration: {(error as Error).message}</p>
      </div>
    )
  }

  return (
    <div>
      <h2 className="text-3xl font-bold mb-6">Configuration</h2>
      
      <Card>
        <CardHeader>
          <CardTitle>Service-to-Repository Mappings</CardTitle>
          <CardDescription>
            Configured mappings between service names and GitHub repositories
          </CardDescription>
        </CardHeader>
        <CardContent>
          {data?.service_mappings && data.service_mappings.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-3 px-4 font-semibold">Service Name</th>
                    <th className="text-left py-3 px-4 font-semibold">Repository</th>
                    <th className="text-left py-3 px-4 font-semibold">Branch</th>
                  </tr>
                </thead>
                <tbody>
                  {data.service_mappings.map((mapping, index) => (
                    <tr key={index} className="border-b hover:bg-muted/50">
                      <td className="py-3 px-4 font-mono text-sm">{mapping.service_name}</td>
                      <td className="py-3 px-4 font-mono text-sm">{mapping.repository}</td>
                      <td className="py-3 px-4 font-mono text-sm">{mapping.branch}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-muted-foreground">No service mappings configured.</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
