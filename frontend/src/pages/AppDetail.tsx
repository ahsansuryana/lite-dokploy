import { useEffect, useState } from 'react'
import { api } from '../api'
import { usePath, Link } from '../router'
import type { Application, Deployment } from '../types'
import { StatusBadge } from '../components/StatusBadge'

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
      alert(`Action failed: ${e}`)
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

  if (loading) return <div className="text-gray-500 text-center py-12">Loading...</div>
  if (!app) return <div className="text-gray-500 text-center py-12">Application not found</div>

  return (
    <div>
      <div className="mb-6">
        <Link href="/" className="text-gray-500 hover:text-gray-300 text-sm">
          &larr; Back to Dashboard
        </Link>
      </div>

      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold">{app.name}</h1>
          <p className="text-gray-500 text-sm mt-1">
            {app.repoUrl} &middot; {app.branch}
          </p>
        </div>
        <StatusBadge status={app.status} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        <div className="lg:col-span-2 space-y-4">
          <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
            <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">Details</h2>
            <dl className="grid grid-cols-2 gap-3 text-sm">
              <div>
                <dt className="text-gray-500">Repository</dt>
                <dd className="font-mono text-xs mt-0.5">{app.repoUrl}</dd>
              </div>
              <div>
                <dt className="text-gray-500">Branch</dt>
                <dd>{app.branch}</dd>
              </div>
              <div>
                <dt className="text-gray-500">Compose Path</dt>
                <dd className="font-mono text-xs mt-0.5">{app.composePath}</dd>
              </div>
              <div>
                <dt className="text-gray-500">Domain</dt>
                <dd>{app.domain || <span className="text-gray-600">—</span>}</dd>
              </div>
            </dl>
          </div>

          <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
            <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">Webhook</h2>
            <div className="flex items-center gap-2">
              <code className="flex-1 bg-gray-950 border border-gray-700 rounded px-3 py-2 text-xs font-mono text-gray-300 truncate">
                {webhookUrl}
              </code>
              <button
                onClick={() => navigator.clipboard.writeText(webhookUrl)}
                className="bg-gray-800 hover:bg-gray-700 text-xs px-3 py-2 rounded transition-colors"
              >
                Copy
              </button>
            </div>
          </div>

          <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
            <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">
              Environment Variables
            </h2>
            {app.envVars ? (
              <pre className="text-xs font-mono text-gray-300 bg-gray-950 rounded p-3 overflow-x-auto">
                {app.envVars}
              </pre>
            ) : (
              <p className="text-gray-600 text-sm">No environment variables configured.</p>
            )}
          </div>
        </div>

        <div className="space-y-3">
          <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Actions</h2>
          <div className="flex flex-col gap-2">
            {[
              ['Deploy', 'deploy', () => api.deploy(appId)],
              ['Redeploy', 'redeploy', () => api.redeploy(appId)],
              ['Restart', 'restart', () => api.restart(appId)],
              ['Stop', 'stop', () => api.stop(appId)],
              ['Start', 'start', () => api.start(appId)],
            ].map(([label, action, fn]) => (
              <button
                key={action as string}
                onClick={() => doAction(action as string, fn as () => Promise<unknown>)}
                disabled={actionLoading !== null}
                className="bg-gray-800 hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed px-4 py-2 rounded text-sm font-medium transition-colors text-left"
              >
                {actionLoading === action ? 'Processing...' : label as string}
              </button>
            ))}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
          <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">
            Deployments
          </h2>
          {deps.length === 0 ? (
            <p className="text-gray-600 text-sm">No deployments yet.</p>
          ) : (
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {deps.map((d) => (
                <div
                  key={d.id}
                  className="flex items-center justify-between bg-gray-950 rounded px-3 py-2 text-sm cursor-pointer hover:bg-gray-800/50 transition-colors"
                  onClick={() => viewLog(d.id)}
                >
                  <div className="flex items-center gap-2 min-w-0">
                    <StatusBadge status={d.status} />
                    {d.commitSha && (
                      <span className="font-mono text-xs text-gray-400">{d.commitSha.slice(0, 8)}</span>
                    )}
                    <span className="text-xs text-gray-500 truncate">
                      {new Date(d.createdAt).toLocaleString()}
                    </span>
                  </div>
                  {d.commitMessage && (
                    <span className="text-xs text-gray-500 truncate ml-2 max-w-[200px]">
                      {d.commitMessage}
                    </span>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="bg-gray-900 border border-gray-800 rounded-lg p-4">
          <h2 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">
            Log{logDepId ? '' : 's'}
          </h2>
          {log ? (
            <pre className="text-xs font-mono text-gray-300 bg-gray-950 rounded p-3 h-96 overflow-y-auto whitespace-pre-wrap">
              {log}
            </pre>
          ) : (
            <div className="flex items-center justify-center h-96 text-gray-600 text-sm">
              Click a deployment to view its log
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
