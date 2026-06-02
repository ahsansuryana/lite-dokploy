import { Router } from './router'
import { Layout } from './components/Layout'
import { Dashboard } from './pages/Dashboard'
import { AppDetail } from './pages/AppDetail'
import { CreateApp } from './pages/CreateApp'

function App() {
  return (
    <Layout>
      <Router
        routes={{
          '/': Dashboard,
          '/create': CreateApp,
          '*': () => {
            const path = window.location.hash.slice(1)
            if (path.startsWith('/app/')) return <AppDetail />
            return <div className="p-6 text-gray-400">404 — page not found</div>
          },
        }}
      />
    </Layout>
  )
}

export default App
