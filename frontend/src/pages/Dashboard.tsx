import { useEffect, useState } from 'react'
import { api } from '../api'
import type { DashboardItem } from '../types'
import { StatusBadge, StatusDot } from '../components/StatusBadge'
import { Link } from '../router'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Skeleton } from '../components/ui/skeleton'
import { Separator } from '../components/ui/separator'
import { Plus, Globe } from 'lucide-react'

export function Dashboard() {
  const [data, setData] = useState<DashboardItem[]>([])
  const [loading, setLoading] = useState(true)

  const load = () => {
    setLoading(true)
    api.dashboard()
      .then((res) => setData(res.applications || []))
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  return (
    <Card className="h-full bg-sidebar p-2.5 rounded-xl">
      <div className="rounded-xl bg-background shadow-md">
        <CardHeader className="flex flex-row items-center justify-between py-4 px-6">
          <div>
            <CardTitle className="text-xl">Applications</CardTitle>
            <p className="text-sm text-muted-foreground mt-0.5">
              Manage your deployed applications
            </p>
          </div>
          <Link href="/create">
            <Button size="sm">
              <Plus size={16} className="mr-1.5" />
              New Application
            </Button>
          </Link>
        </CardHeader>
        <Separator />
        <CardContent className="pt-6">
          {loading ? (
            <div className="space-y-3 min-h-[60vh]">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : data.length === 0 ? (
            <div className="flex h-[60vh] w-full flex-col items-center justify-center gap-3">
              <Globe size={32} className="text-muted-foreground" />
              <p className="font-medium text-muted-foreground">No applications yet</p>
              <p className="text-sm text-muted-foreground/60">
                Create your first application to get started.
              </p>
              <Link href="/create">
                <Button size="sm" className="mt-2">
                  <Plus size={16} className="mr-1.5" />
                  New Application
                </Button>
              </Link>
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-4">
              {data.map((item) => {
                const { latestDeployment, ...app } = item
                return (
                  <Link key={app.id} href={`/app/${app.id}`} className="no-underline">
                    <div className="group relative cursor-pointer rounded-xl border bg-card transition-colors hover:bg-border p-5 flex flex-col gap-3">
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex items-center gap-3 min-w-0">
                          <div className="size-10 rounded-lg bg-secondary flex items-center justify-center shrink-0">
                            <Globe size={18} className="text-foreground" />
                          </div>
                          <div className="min-w-0">
                            <p className="font-medium text-sm truncate">{app.name}</p>
                            <p className="text-xs text-muted-foreground truncate mt-0.5">
                              {app.source === 'manual' ? 'Manual (pasted compose)' : app.repoUrl}
                            </p>
                          </div>
                        </div>
                        <StatusBadge status={app.status} />
                      </div>
                      <Separator />
                      <div className="flex items-center justify-between text-xs text-muted-foreground">
                        <span>{app.branch}</span>
                        {latestDeployment ? (
                          <div className="flex items-center gap-2">
                            <StatusDot status={latestDeployment.status} />
                            <span className="font-mono">
                              {latestDeployment.commitSha?.slice(0, 7)}
                            </span>
                          </div>
                        ) : (
                          <span>No deployments</span>
                        )}
                      </div>
                    </div>
                  </Link>
                )
              })}
            </div>
          )}
        </CardContent>
      </div>
    </Card>
  )
}
