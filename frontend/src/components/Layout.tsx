import { Link } from '../router'

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex flex-col">
      <header className="border-b border-gray-800 px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-6">
          <Link href="/" className="text-lg font-bold text-cyan-400 no-underline">
            lite-dokploy
          </Link>
          <nav className="flex gap-4 text-sm">
            <Link href="/" className="text-gray-400 hover:text-white transition-colors">
              Dashboard
            </Link>
          </nav>
        </div>
      </header>
      <main className="flex-1 p-6">{children}</main>
    </div>
  )
}
