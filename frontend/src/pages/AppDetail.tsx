import { useEffect, useState } from 'react'
import { api } from '../api'
import { usePath, navigate, Link } from '../router'
import type { Application, Deployment, AppDomain } from '../types'
import { StatusBadge, StatusDot } from '../components/StatusBadge'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../components/ui/card'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Skeleton } from '../components/ui/skeleton'
import { Separator } from '../components/ui/separator'
import {
  ArrowLeft, Globe, Rocket, Terminal, Copy, RefreshCw, Play, Square, RotateCcw,
  Plus, X, Pencil, Trash2, ShieldCheck, Save,
} from 'lucide-react'

export function AppDetail() {
  const path = usePath()
  const appId = path.replace('/app/', '')
  const [app, setApp] = useState<Application | null>(null)
  const [deps, setDeps] = useState<Deployment[]>([])
  const [domains, setDomains] = useState<AppDomain[]>([])
  const [log, setLog] = useState<string>('')
  const [logDepId, setLogDepId] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [isEditing, setIsEditing] = useState(false)
  const [saving, setSaving] = useState(false)
  const [editForm, setEditForm] = useState({
    name: '',
    source: 'manual',
    repoUrl: '',
    branch: '',
    composePath: '',
    composeContent: '',
    envVars: '',
  })

  const [showDomainForm, setShowDomainForm] = useState(false)
  const [editingDomain, setEditingDomain] = useState<AppDomain | null>(null)
  const [domainForm, setDomainForm] = useState({
    host: '',
    port: 80,
    path: '/',
    internalPath: '/',
    stripPath: false,
    https: false,
    serviceName: '',
  })
  const [domainSaving, setDomainSaving] = useState(false)

  const resetDomainForm = () => {
    setDomainForm({ host: '', port: 80, path: '/', internalPath: '/', stripPath: false, https: false, serviceName: '' })
    setEditingDomain(null)
    setShowDomainForm(false)
  }

  const editDomain = (d: AppDomain) => {
    setDomainForm({
      host: d.host,
      port: d.port,
      path: d.path,
      internalPath: d.internalPath,
      stripPath: d.stripPath,
      https: d.https,
      serviceName: d.serviceName,
    })
    setEditingDomain(d)
    setShowDomainForm(true)
  }

  const saveDomain = async () => {
    setDomainSaving(true)
    try {
      if (editingDomain) {
        await api.updateDomain(appId, editingDomain.id, domainForm)
      } else {
        await api.createDomain(appId, domainForm)
      }
      resetDomainForm()
      const d = await api.listDomains(appId)
      setDomains(d)
    } catch (e) {
      console.error(e)
    } finally {
      setDomainSaving(false)
    }
  }

  const deleteDomain = async (domainId: string) => {
    if (!confirm('Delete this domain?')) return
    try {
      await api.deleteDomain(appId, domainId)
      setDomains(domains.filter((d) => d.id !== domainId))
    } catch (e) {
      console.error(e)
    }
  }

  const load = () => {
    setLoading(true)
    Promise.all([
      api.getApp(appId),
      api.listDeployments(appId),
      api.listDomains(appId),
    ])
      .then(([app, deps, domains]) => {
        setApp(app)
        setDeps(deps)
        setDomains(domains)
      })
      .catch(console.error)
      .finally(() => setLoading(false))
  }

  useEffect(load, [appId])

  const startEditing = () => {
    if (!app) return
    setEditForm({
      name: app.name,
      source: app.source,
      repoUrl: app.repoUrl,
      branch: app.branch,
      composePath: app.composePath,
      composeContent: app.composeContent || '',
      envVars: app.envVars,
    })
    setIsEditing(true)
  }

  const cancelEditing = () => {
    setIsEditing(false)
  }

  const saveEdit = async () => {
    if (!app) return
    setSaving(true)
    try {
      await api.updateApp(appId, {
        name: editForm.name,
        source: editForm.source,
        repoUrl: editForm.repoUrl,
        branch: editForm.branch,
        composePath: editForm.composePath,
        composeContent: editForm.composeContent,
        envVars: editForm.envVars,
      })
      setIsEditing(false)
      load()
    } catch (e) {
      console.error(e)
    } finally {
      setSaving(false)
    }
  }

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
                {isEditing ? (
                  <input
                    value={editForm.name}
                    onChange={(e) => setEditForm((f) => ({ ...f, name: e.target.value }))}
                    className="rounded-md border border-input bg-background px-2.5 py-1.5 text-lg font-semibold ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  />
                ) : (
                  app.name
                )}
              </CardTitle>
              {!isEditing && (
                <CardDescription className="mt-1">
                  {app.source === 'manual' ? 'Manual (pasted compose)' : `${app.repoUrl} · ${app.branch}`}
                </CardDescription>
              )}
            </div>
            <div className="flex items-center gap-2">
              {isEditing ? (
                <>
                  <Button size="sm" variant="outline" onClick={cancelEditing} disabled={saving}>
                    Cancel
                  </Button>
                  <Button size="sm" onClick={saveEdit} disabled={saving}>
                    {saving ? (
                      <span className="flex items-center gap-2">
                        <RefreshCw size={14} className="animate-spin" />
                        Saving...
                      </span>
                    ) : (
                      <span className="flex items-center gap-2">
                        <Save size={14} />
                        Save
                      </span>
                    )}
                  </Button>
                </>
              ) : (
                <Button size="sm" variant="outline" onClick={startEditing}>
                  <Pencil size={14} className="mr-1" />
                  Edit
                </Button>
              )}
              <StatusBadge status={app.status} />
            </div>
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
                  {isEditing ? (
                    <div className="space-y-4">
                      <div>
                        <label className="block text-xs text-muted-foreground mb-1">Source</label>
                        <select
                          value={editForm.source}
                          onChange={(e) => setEditForm((f) => ({ ...f, source: e.target.value }))}
                          className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                        >
                          <option value="manual">Manual (pasted compose)</option>
                          <option value="git">Git Repository</option>
                        </select>
                      </div>
                      {editForm.source === 'git' && (
                        <>
                          <div>
                            <label className="block text-xs text-muted-foreground mb-1">Repository URL</label>
                            <input
                              value={editForm.repoUrl}
                              onChange={(e) => setEditForm((f) => ({ ...f, repoUrl: e.target.value }))}
                              className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm font-mono ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                              placeholder="https://github.com/user/repo.git"
                            />
                          </div>
                          <div>
                            <label className="block text-xs text-muted-foreground mb-1">Branch</label>
                            <input
                              value={editForm.branch}
                              onChange={(e) => setEditForm((f) => ({ ...f, branch: e.target.value }))}
                              className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                              placeholder="main"
                            />
                          </div>
                        </>
                      )}
                      <div>
                        <label className="block text-xs text-muted-foreground mb-1">Compose Path</label>
                        <input
                          value={editForm.composePath}
                          onChange={(e) => setEditForm((f) => ({ ...f, composePath: e.target.value }))}
                          className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm font-mono ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                          placeholder="docker-compose.yml"
                        />
                      </div>
                    </div>
                  ) : (
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
                    </dl>
                  )}
                </CardContent>
              </Card>

              {(isEditing || app.composeContent) && (
                <Card>
                  <CardHeader className="py-3 px-4">
                    <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                      Compose File
                    </CardTitle>
                  </CardHeader>
                  <Separator />
                  <CardContent className="p-4">
                    {isEditing ? (
                      <textarea
                        value={editForm.composeContent}
                        onChange={(e) => setEditForm((f) => ({ ...f, composeContent: e.target.value }))}
                        className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-xs font-mono ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                        rows={20}
                        placeholder="version: '3'
services:
  app:
    image: nginx"
                      />
                    ) : (
                      <pre className="text-xs font-mono text-muted-foreground bg-secondary rounded-lg p-3 overflow-x-auto max-h-96 overflow-y-auto">
                        {app.composeContent}
                      </pre>
                    )}
                  </CardContent>
                </Card>
              )}

              <Card>
                <CardHeader className="py-3 px-4">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                      Domains
                    </CardTitle>
                    <Button size="sm" variant="ghost" onClick={() => { resetDomainForm(); setShowDomainForm(true) }}>
                      <Plus size={14} className="mr-1" />
                      Add Domain
                    </Button>
                  </div>
                </CardHeader>
                <Separator />
                <CardContent className="p-4 space-y-3">
                  {domains.length === 0 && !showDomainForm && (
                    <p className="text-sm text-muted-foreground/60">No domains configured.</p>
                  )}
                  {domains.map((d) => (
                    <div key={d.id} className="flex items-center justify-between rounded-lg bg-secondary p-3 text-sm">
                      <div className="flex items-center gap-3 min-w-0">
                        <Globe size={14} className="text-muted-foreground shrink-0" />
                        <div className="min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-medium">{d.host}</span>
                            {d.path !== '/' && <Badge variant="outline" className="text-[10px]">{d.path}</Badge>}
                            {d.https && <ShieldCheck size={12} className="text-green-500" />}
                          </div>
                          <div className="flex items-center gap-2 mt-0.5 text-xs text-muted-foreground">
                            <span>port {d.port}</span>
                            {d.serviceName && <><span>·</span><span>service: {d.serviceName}</span></>}
                            {d.path !== d.internalPath && <><span>·</span><span>→ {d.internalPath}</span></>}
                            {d.stripPath && <><span>·</span><Badge variant="secondary" className="text-[10px]">strip</Badge></>}
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-1 shrink-0">
                        <Button size="icon" variant="ghost" className="size-7" onClick={() => editDomain(d)}>
                          <Pencil size={12} />
                        </Button>
                        <Button size="icon" variant="ghost" className="size-7 text-destructive" onClick={() => deleteDomain(d.id)}>
                          <Trash2 size={12} />
                        </Button>
                      </div>
                    </div>
                  ))}
                  {showDomainForm && (
                    <div className="rounded-lg border border-input bg-background p-4 space-y-3">
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium">{editingDomain ? 'Edit Domain' : 'New Domain'}</span>
                        <Button size="icon" variant="ghost" className="size-6" onClick={resetDomainForm}>
                          <X size={14} />
                        </Button>
                      </div>
                      <div className="grid grid-cols-2 gap-3">
                        <div className="col-span-2">
                          <label className="block text-xs text-muted-foreground mb-1">Host *</label>
                          <input
                            required
                            value={domainForm.host}
                            onChange={(e) => setDomainForm((f) => ({ ...f, host: e.target.value }))}
                            className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            placeholder="example.com"
                          />
                        </div>
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1">Port</label>
                          <input
                            type="number"
                            value={domainForm.port}
                            onChange={(e) => setDomainForm((f) => ({ ...f, port: Number(e.target.value) }))}
                            className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                          />
                        </div>
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1">Service Name</label>
                          <input
                            value={domainForm.serviceName}
                            onChange={(e) => setDomainForm((f) => ({ ...f, serviceName: e.target.value }))}
                            className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            placeholder="(auto)"
                          />
                        </div>
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1">Path (outer)</label>
                          <input
                            value={domainForm.path}
                            onChange={(e) => setDomainForm((f) => ({ ...f, path: e.target.value }))}
                            className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                          />
                        </div>
                        <div>
                          <label className="block text-xs text-muted-foreground mb-1">Internal Path</label>
                          <input
                            value={domainForm.internalPath}
                            onChange={(e) => setDomainForm((f) => ({ ...f, internalPath: e.target.value }))}
                            className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                          />
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <label className="flex items-center gap-2 text-sm">
                          <input
                            type="checkbox"
                            checked={domainForm.stripPath}
                            onChange={(e) => setDomainForm((f) => ({ ...f, stripPath: e.target.checked }))}
                            className="rounded border-input"
                          />
                          Strip path
                        </label>
                        <label className="flex items-center gap-2 text-sm">
                          <input
                            type="checkbox"
                            checked={domainForm.https}
                            onChange={(e) => setDomainForm((f) => ({ ...f, https: e.target.checked }))}
                            className="rounded border-input"
                          />
                          HTTPS
                        </label>
                      </div>
                      <div className="flex justify-end gap-2 pt-1">
                        <Button size="sm" variant="outline" onClick={resetDomainForm}>Cancel</Button>
                        <Button size="sm" onClick={saveDomain} disabled={domainSaving || !domainForm.host}>
                          {domainSaving ? 'Saving...' : editingDomain ? 'Update' : 'Create'}
                        </Button>
                      </div>
                    </div>
                  )}
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
                  {isEditing ? (
                    <textarea
                      value={editForm.envVars}
                      onChange={(e) => setEditForm((f) => ({ ...f, envVars: e.target.value }))}
                      className="w-full rounded-md border border-input bg-background px-2.5 py-1.5 text-xs font-mono ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                      rows={8}
                      placeholder="NODE_ENV=production
DATABASE_URL=postgres://..."
                    />
                  ) : app.envVars ? (
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
                        navigate('/')
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
