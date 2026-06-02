import { useEffect, useState } from 'react'
import { api } from '../api'
import type { DashboardItem } from '../types'
import { StatusBadge } from '../components/StatusBadge'
import { Link } from '../router'

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
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Applications</h1>
        <Link
          href="/create"
          className="bg-cyan-600 hover:bg-cyan-500 text-white px-4 py-2 rounded text-sm font-medium transition-colors"
        >
          + New Application
        </Link>
      </div>

      {loading ? (
        <div className="text-gray-500 text-center py-12">Loading...</div>
      ) : data.length === 0 ? (
        <div className="text-gray-500 text-center py-12 border border-dashed border-gray-800 rounded-lg">
          <p className="mb-2 text-lg">No applications yet</p>
          <p className="text-sm">Create your first application to get started.</p>
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-800 text-gray-400 uppercase text-xs tracking-wider">
                <th className="text-left py-3 px-2">Name</th>
                <th className="text-left py-3 px-2">Status</th>
                <th className="text-left py-3 px-2">Domain</th>
                <th className="text-left py-3 px-2">Branch</th>
                <th className="text-left py-3 px-2">Last Deploy</th>
                <th className="text-left py-3 px-2">Commit</th>
              </tr>
            </thead>
            <tbody>
              {data.map(({ application: app, latestDeployment }) => (
                <tr key={app.id} className="border-b border-gray-800/50 hover:bg-gray-900/50 transition-colors">
                  <td className="py-3 px-2">
                    <Link href={`/app/${app.id}`} className="text-cyan-400 hover:text-cyan-300 font-medium">
                      {app.name}
                    </Link>
                  </td>
                  <td className="py-3 px-2">
                    <StatusBadge status={app.status} />
                  </td>
                  <td className="py-3 px-2 text-gray-400">
                    {app.domain || <span className="text-gray-600">—</span>}
                  </td>
                  <td className="py-3 px-2 text-gray-400">{app.branch}</td>
                  <td className="py-3 px-2 text-gray-400">
                    {latestDeployment
                      ? new Date(latestDeployment.createdAt).toLocaleString()
                      : <span className="text-gray-600">—</span>}
                  </td>
                  <td className="py-3 px-2 text-gray-400 font-mono text-xs">
                    {latestDeployment?.commitSha
                      ? latestDeployment.commitSha.slice(0, 8)
                      : <span className="text-gray-600">—</span>}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
