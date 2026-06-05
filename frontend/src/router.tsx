import { useSyncExternalStore } from 'react'

function getPath() {
  return window.location.pathname || '/'
}

const listeners = new Set<() => void>()

function subscribe(cb: () => void) {
  listeners.add(cb)
  const onPop = () => {
    for (const fn of listeners) fn()
  }
  window.addEventListener('popstate', onPop)
  return () => {
    listeners.delete(cb)
    window.removeEventListener('popstate', onPop)
  }
}

export function usePath() {
  return useSyncExternalStore(subscribe, getPath, () => '/')
}

export function navigate(path: string) {
  window.history.pushState(null, '', path)
  window.dispatchEvent(new PopStateEvent('popstate'))
}

export function useNavigate() {
  return navigate
}

export function Router({ routes }: { routes: Record<string, React.ComponentType> }) {
  const path = usePath()
  const Route = routes[path] || routes['*']
  if (!Route) return <div className="p-6 text-gray-400">404 — page not found</div>
  return <Route />
}

export function Link({ href, children, className }: {
  href: string
  children: React.ReactNode
  className?: string
}) {
  return (
    <a
      href={href}
      className={className}
      onClick={(e) => {
        e.preventDefault()
        navigate(href)
      }}
    >
      {children}
    </a>
  )
}
