import { useState } from 'react'
import { api } from '../api'
import { navigate, Link } from '../router'

export function CreateApp() {
  const [form, setForm] = useState({
    name: '',
    repoUrl: '',
    branch: 'main',
    composePath: 'docker-compose.yml',
    domain: '',
    envVars: '',
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSaving(true)
    try {
      const app = await api.createApp(form)
      navigate(`/app/${app.id}`)
    } catch (err) {
      setError(String(err))
    } finally {
      setSaving(false)
    }
  }

  const update = (field: string, value: string) =>
    setForm((f) => ({ ...f, [field]: value }))

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-6">
        <Link href="/" className="text-gray-500 hover:text-gray-300 text-sm">
          &larr; Back to Dashboard
        </Link>
      </div>

      <h1 className="text-2xl font-bold mb-6">New Application</h1>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-400 mb-1">Name *</label>
          <input
            required
            value={form.name}
            onChange={(e) => update('name', e.target.value)}
            className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-sm focus:outline-none focus:border-cyan-500 transition-colors"
            placeholder="my-app"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-400 mb-1">Git Repository URL *</label>
          <input
            required
            value={form.repoUrl}
            onChange={(e) => update('repoUrl', e.target.value)}
            className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-sm focus:outline-none focus:border-cyan-500 transition-colors font-mono"
            placeholder="https://github.com/user/repo.git"
          />
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-400 mb-1">Branch</label>
            <input
              value={form.branch}
              onChange={(e) => update('branch', e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-sm focus:outline-none focus:border-cyan-500 transition-colors"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-400 mb-1">Compose Path</label>
            <input
              value={form.composePath}
              onChange={(e) => update('composePath', e.target.value)}
              className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-sm focus:outline-none focus:border-cyan-500 transition-colors font-mono"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-400 mb-1">Domain</label>
          <input
            value={form.domain}
            onChange={(e) => update('domain', e.target.value)}
            className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-sm focus:outline-none focus:border-cyan-500 transition-colors"
            placeholder="app.example.com"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-400 mb-1">
            Environment Variables
            <span className="text-gray-600 font-normal ml-2">(KEY=value per line)</span>
          </label>
          <textarea
            value={form.envVars}
            onChange={(e) => update('envVars', e.target.value)}
            rows={6}
            className="w-full bg-gray-900 border border-gray-700 rounded px-3 py-2 text-sm focus:outline-none focus:border-cyan-500 transition-colors font-mono"
            placeholder="DATABASE_URL=postgres://..."
          />
        </div>

        {error && (
          <div className="bg-red-900/50 border border-red-700 rounded px-4 py-2 text-sm text-red-300">
            {error}
          </div>
        )}

        <button
          type="submit"
          disabled={saving}
          className="bg-cyan-600 hover:bg-cyan-500 disabled:opacity-50 disabled:cursor-not-allowed text-white px-6 py-2.5 rounded text-sm font-medium transition-colors"
        >
          {saving ? 'Creating...' : 'Create Application'}
        </button>
      </form>
    </div>
  )
}
