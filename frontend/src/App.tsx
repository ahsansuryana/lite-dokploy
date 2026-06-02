import { Router } from './router'
import { AppShell } from './components/AppShell'
import { Dashboard } from './pages/Dashboard'
import { AppDetail } from './pages/AppDetail'
import { CreateApp } from './pages/CreateApp'

function App() {
  return (
    <AppShell>
      <Router
        routes={{
          '/': Dashboard,
          '/create': CreateApp,
          '*': () => {
            const path = window.location.hash.slice(1)
            if (path.startsWith('/app/')) return <AppDetail />
            return <div className="flex h-[60vh] items-center justify-center text-muted-foreground">404 — page not found</div>
          },
        }}
      />
    </AppShell>
  )
}

export default App
