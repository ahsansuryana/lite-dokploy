import { useEffect, useState } from 'react'
import { api } from '../api'
import { usePath, Link } from '../router'
import type { Application, Deployment } from '../types'
import { StatusBadge, StatusDot } from '../components/StatusBadge'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Skeleton } from '../components/ui/skeleton'
import { Separator } from '../components/ui/separator'
import { ArrowLeft, Globe, Rocket, Terminal, Copy, RefreshCw, Play, Square, RotateCcw } from 'lucide-react'

export function AppDetail() {
  const path = usePath()
  const appId = path.replace('/app/', '')
  const [app, setApp] = useState<Application | null>(null)
  const [deps, setDeps] = useState<Deployment[]>([])
  const [log, setLog] = useState<string>('')
  const [logDepId, setLogDepId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  const load = () => {
    setLoading(true)
    Promise.all([
      api.getApp(appId),
      api.listDeployments(appId),
    ])
      .then(([app, deps]) => {
        setApp(app)
        setDeps(deps)
      })
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(load, [appId])

  const doAction = async (action: string, fn: () => Promise<unknown>) => {
    setActionLoading(action)
    try {
      await fn()
      load()
    } catch (e) {
      console.error(e)
    } finally {
      setActionLoading(null)
    }
  }

  const viewLog = async (depId: string) => {
    setLogDepId(depId)
    setLog('Loading...')
    try {
      const text = await api.getDeploymentLog(depId)
      setLog(text)
    } catch {
      setLog('Failed to load log')
    }
  }

  const webhookUrl = `${window.location.protocol}//${window.location.hostname}:3000/api/webhook/${appId}`

  if (loading) {
    return (
      <Card className="h-full bg-sidebar p-2.5 rounded-xl">
        <div className="rounded-xl bg-background shadow-md p-6">
          <Skeleton className="h-8 w-48 mb-4" />
          <Skeleton className="h-4 w-96 mb-8" />
          <div className="space-y-3">
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
            <Skeleton className="h-12 w-full" />
          </div>
        </div>
      </Card>
    )
  }

  if (!app) {
    return (
      <Card className="h-full bg-sidebar p-2.5 rounded-xl">
        <div className="rounded-xl bg-background shadow-md">
          <CardContent className="flex h-[60vh] items-center justify-center">
            <p className="text-muted-foreground">Application not found</p>
          </CardContent>
        </div>
      </Card>
    )
  }

  return (
    <Card className="h-full bg-sidebar p-2.5 rounded-xl">
      <div className="rounded-xl bg-background shadow-md">
        <CardHeader className="py-4 px-6">
          <Link href="/" className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground no-underline mb-2">
            <ArrowLeft size={14} />
            Back to Dashboard
          </Link>
          <div className="flex items-start justify-between">
            <div>
              <CardTitle className="text-xl flex items-center gap-3">
                <div className="size-10 rounded-lg bg-secondary flex items-center justify-center">
                  <Globe size={18} className="text-foreground" />
                </div>
                {app.name}
              </CardTitle>
              <CardDescription className="mt-1">
                {app.source === 'manual' ? 'Manual (pasted compose)' : `${app.repoUrl} · ${app.branch}`}
              </CardDescription>
            </div>
            <StatusBadge status={app.status} />
          </div>
        </CardHeader>
        <Separator />
        <CardContent className="pt-6">
          <div className="grid grid-cols-1 xl:grid-cols-3 gap-6 mb-8">
            <div className="xl:col-span-2 space-y-6">
              <Card>
                <CardHeader className="py-3 px-4">
                  <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                    Details
                  </CardTitle>
                </CardHeader>
                <Separator />
                <CardContent className="p-4">
                  <dl className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <dt className="text-muted-foreground text-xs">Source</dt>
                      <dd className="text-xs mt-0.5 capitalize">{app.source}</dd>
                    </div>
                    {app.source !== 'manual' && (
                    <>
                    <div>
                      <dt className="text-muted-foreground text-xs">Repository</dt>
                      <dd className="font-mono text-xs mt-0.5 text-foreground">{app.repoUrl}</dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-xs">Branch</dt>
                      <dd className="text-xs mt-0.5">{app.branch}</dd>
                    </div>
                    </>  
                    )}
                    <div>
                      <dt className="text-muted-foreground text-xs">Compose Path</dt>
                      <dd className="font-mono text-xs mt-0.5">{app.composePath}</dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-xs">Domain</dt>
                      <dd className="text-xs mt-0.5">{app.domain || <span className="text-muted-foreground/50">&mdash;</span>}</dd>
                    </div>
                  </dl>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="py-3 px-4">
                  <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                    Webhook
                  </CardTitle>
                </CardHeader>
                <Separator />
                <CardContent className="p-4">
                  <div className="flex items-center gap-2">
                    <code className="flex-1 rounded-lg bg-secondary px-3 py-2 text-xs font-mono text-muted-foreground truncate">
                      {webhookUrl}
                    </code>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => navigator.clipboard.writeText(webhookUrl)}
                    >
                      <Copy size={14} />
                    </Button>
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="py-3 px-4">
                  <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                    Environment Variables
                  </CardTitle>
                </CardHeader>
                <Separator />
                <CardContent className="p-4">
                  {app.envVars ? (
                    <pre className="text-xs font-mono text-muted-foreground bg-secondary rounded-lg p-3 overflow-x-auto">
                      {app.envVars}
                    </pre>
                  ) : (
                    <p className="text-sm text-muted-foreground/60">No environment variables configured.</p>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="py-3 px-4">
                  <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                    Deployments
                  </CardTitle>
                </CardHeader>
                <Separator />
                <CardContent className="p-4">
                  {deps.length === 0 ? (
                    <p className="text-sm text-muted-foreground/60">No deployments yet.</p>
                  ) : (
                    <div className="space-y-2 max-h-96 overflow-y-auto">
                      {deps.map((d) => (
                        <div
                          key={d.id}
                          className="flex items-center justify-between rounded-lg bg-secondary p-3 text-sm cursor-pointer hover:bg-border transition-colors"
                          onClick={() => viewLog(d.id)}
                        >
                          <div className="flex items-center gap-3 min-w-0">
                            <StatusDot status={d.status} />
                            {d.commitSha && (
                              <span className="font-mono text-xs text-muted-foreground">{d.commitSha.slice(0, 8)}</span>
                            )}
                            <span className="text-xs text-muted-foreground truncate">
                              {new Date(d.createdAt).toLocaleString()}
                            </span>
                          </div>
                          {d.commitMessage && (
                            <span className="text-xs text-muted-foreground truncate ml-2 max-w-[200px]">
                              {d.commitMessage}
                            </span>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>

            <div className="space-y-4">
              <Card>
                <CardHeader className="py-3 px-4">
                  <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                    Actions
                  </CardTitle>
                </CardHeader>
                <Separator />
                <CardContent className="p-4">
                  <div className="flex flex-col gap-2">
                    {([
                      ['Deploy', 'deploy', () => api.deploy(appId), Rocket],
                      ['Redeploy', 'redeploy', () => api.redeploy(appId), RefreshCw],
                      ['Restart', 'restart', () => api.restart(appId), RotateCcw],
                      ['Stop', 'stop', () => api.stop(appId), Square],
                      ['Start', 'start', () => api.start(appId), Play],
                    ] as const).map(([label, action, fn, Icon]) => (
                      <Button
                        key={action}
                        variant="secondary"
                        size="sm"
                        onClick={() => doAction(action, fn)}
                        disabled={actionLoading !== null}
                      >
                        {actionLoading === action ? (
                          <span className="flex items-center gap-2">
                            <RefreshCw size={14} className="animate-spin" />
                            Processing...
                          </span>
                        ) : (
                          <span className="flex items-center gap-2">
                            <Icon size={14} />
                            {label}
                          </span>
                        )}
                      </Button>
                    ))}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="py-3 px-4">
                  <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                    Delete
                  </CardTitle>
                </CardHeader>
                <Separator />
                <CardContent className="p-4">
                  <Button
                    variant="destructive"
                    size="sm"
                    className="w-full"
                    onClick={async () => {
                      if (confirm('Are you sure you want to delete this application?')) {
                        await api.deleteApp(appId)
                        window.location.hash = '#/'
                      }
                    }}
                  >
                    Delete Application
                  </Button>
                </CardContent>
              </Card>
            </div>
          </div>

          <Card>
            <CardHeader className="py-3 px-4">
              <div className="flex items-center gap-2">
                <Terminal size={16} className="text-muted-foreground" />
                <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                  Log{logDepId ? '' : 's'}
                </CardTitle>
              </div>
            </CardHeader>
            <Separator />
            <CardContent className="p-4">
              {log ? (
                <pre className="text-xs font-mono text-muted-foreground bg-secondary rounded-lg p-4 h-96 overflow-y-auto whitespace-pre-wrap custom-logs-scrollbar">
                  {log}
                </pre>
              ) : (
                <div className="flex items-center justify-center h-96 text-sm text-muted-foreground/60">
                  Click a deployment to view its log
                </div>
              )}
            </CardContent>
          </Card>
        </CardContent>
      </div>
    </Card>
  )
}
