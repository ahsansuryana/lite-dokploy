import { useState } from 'react'
import { api } from '../api'
import { navigate, Link } from '../router'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Button } from '../components/ui/button'
import { Separator } from '../components/ui/separator'
import { ArrowLeft, GitBranch, FileText, Plus } from 'lucide-react'

type SourceTab = 'git' | 'manual'

export function CreateApp() {
  const [sourceTab, setSourceTab] = useState<SourceTab>('git')
  const [form, setForm] = useState({
    name: '',
    repoUrl: '',
    branch: 'main',
    composePath: 'docker-compose.yml',
    composeContent: '',
    envVars: '',
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSaving(true)
    try {
      const payload = {
        name: form.name,
        repoUrl: form.repoUrl,
        branch: form.branch,
        composePath: form.composePath,
        composeContent: sourceTab === 'manual' ? form.composeContent : undefined,
        envVars: form.envVars,
        source: sourceTab === 'manual' ? 'manual' : 'git',
      }
      const app = await api.createApp(payload)
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
            Configure a new application deployment.
          </p>
        </CardHeader>

        <div className="px-6 pb-0">
          <div className="flex border-b border-border">
            <button
              type="button"
              onClick={() => setSourceTab('git')}
              className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
                sourceTab === 'git'
                  ? 'border-primary text-foreground'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              <GitBranch size={16} />
              Git Repository
            </button>
            <button
              type="button"
              onClick={() => setSourceTab('manual')}
              className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors ${
                sourceTab === 'manual'
                  ? 'border-primary text-foreground'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              <FileText size={16} />
              Docker Compose
            </button>
          </div>
        </div>

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

            {sourceTab === 'git' ? (
              <>
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
              </>
            ) : (
              <div>
                <label className="block text-sm font-medium text-foreground mb-1.5">
                  Docker Compose Content <span className="text-destructive">*</span>
                </label>
                <textarea
                  required
                  value={form.composeContent}
                  onChange={(e) => update('composeContent', e.target.value)}
                  rows={14}
                  className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm font-mono ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-colors"
                  placeholder="services:&#10;  app:&#10;    image: nginx:latest&#10;    ports:&#10;      - '8080:80'"
                />
                <div className="flex items-center gap-2 mt-1.5">
                  <input
                    value={form.composePath}
                    onChange={(e) => update('composePath', e.target.value)}
                    className="flex-1 rounded-lg border border-input bg-background px-3 py-2 text-sm font-mono ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 transition-colors"
                    placeholder="docker-compose.yml"
                  />
                  <span className="text-xs text-muted-foreground whitespace-nowrap">Filename</span>
                </div>
              </div>
            )}

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
