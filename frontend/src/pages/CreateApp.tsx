import { useState } from 'react'
import { api } from '../api'
import { navigate, Link } from '../router'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Separator } from '../components/ui/separator'
import { ArrowLeft, Plus } from 'lucide-react'

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
    <Card className="h-full bg-sidebar p-2.5 rounded-xl">
      <div className="rounded-xl bg-background shadow-md">
        <CardHeader className="py-4 px-6">
          <Link href="/" className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground no-underline mb-2">
            <ArrowLeft size={14} />
            Back to Dashboard
          </Link>
          <CardTitle className="text-xl">New Application</CardTitle>
          <p className="text-sm text-muted-foreground mt-0.5">
            Configure a new Git-based application deployment.
          </p>
        </CardHeader>
        <Separator />
        <CardContent className="pt-6 max-w-2xl">
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Name <span className="text-destructive">*</span>
              </label>
              <input
                required
                value={form.name}
                onChange={(e) => update('name', e.target.value)}
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-colors"
                placeholder="my-app"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Git Repository URL <span className="text-destructive">*</span>
              </label>
              <input
                required
                value={form.repoUrl}
                onChange={(e) => update('repoUrl', e.target.value)}
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm font-mono ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-colors"
                placeholder="https://github.com/user/repo.git"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">Branch</label>
                <input
                  value={form.branch}
                  onChange={(e) => update('branch', e.target.value)}
                  className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-colors"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">Compose Path</label>
                <input
                  value={form.composePath}
                  onChange={(e) => update('composePath', e.target.value)}
                  className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm font-mono ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-colors"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">Domain</label>
              <input
                value={form.domain}
                onChange={(e) => update('domain', e.target.value)}
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-colors"
                placeholder="app.example.com"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Environment Variables
                <span className="text-muted-foreground font-normal ml-2">(KEY=value per line)</span>
              </label>
              <textarea
                value={form.envVars}
                onChange={(e) => update('envVars', e.target.value)}
                rows={6}
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm font-mono ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-colors"
                placeholder="DATABASE_URL=postgres://..."
              />
            </div>

            {error && (
              <div className="rounded-lg bg-destructive/10 border border-destructive/30 px-4 py-3 text-sm text-destructive">
                {error}
              </div>
            )}

            <div className="flex gap-3 pt-2">
              <Button type="submit" disabled={saving}>
                {saving ? (
                  <span className="flex items-center gap-2">
                    <span className="animate-spin inline-block size-4 border-2 border-current border-t-transparent rounded-full" />
                    Creating...
                  </span>
                ) : (
                  <span className="flex items-center gap-2">
                    <Plus size={16} />
                    Create Application
                  </span>
                )}
              </Button>
              <Link href="/">
                <Button type="button" variant="outline">Cancel</Button>
              </Link>
            </div>
          </form>
        </CardContent>
      </div>
    </Card>
  )
}
