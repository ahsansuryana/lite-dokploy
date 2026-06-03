import { useAuth } from './context/AuthContext'
import { Router } from './router'
import { AppShell } from './components/AppShell'
import { Dashboard } from './pages/Dashboard'
import { AppDetail } from './pages/AppDetail'
import { CreateApp } from './pages/CreateApp'
import { Login } from './pages/Login'

function Loading() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="h-8 w-8 animate-spin rounded-full border-b-2 border-primary" />
    </div>
  )
}

function App() {
  const { isAuthenticated, isLoading } = useAuth()

  if (isLoading) return <Loading />
  if (!isAuthenticated) return <Login />

  return (
    <AppShell>
      <Router
        routes={{
          '/': Dashboard,
          '/create': CreateApp,
          '*': () => {
            const path = window.location.hash.slice(1)
            if (path.startsWith('/app/')) return <AppDetail />
            return (
              <div className="flex h-[60vh] items-center justify-center text-muted-foreground">
                404 — page not found
              </div>
            )
          },
        }}
      />
    </AppShell>
  )
}

export default App
