import { SidebarProvider, Sidebar, SidebarHeader, SidebarContent, SidebarFooter, SidebarGroup, SidebarItem, SidebarInset, SidebarTrigger, SidebarRail } from './ui/sidebar'
import { Separator } from './ui/separator'
import { ModeToggle } from './ui/mode-toggle'
import { usePath, navigate, Link } from '../router'
import { useAuth } from '../context/AuthContext'
import { House, Plus, LayoutDashboard, LogOut } from 'lucide-react'

export function AppShell({ children }: { children: React.ReactNode }) {
  const path = usePath()
  const { logout, username } = useAuth()

  const nav = [
    { href: '/', label: 'Applications', icon: <House size={16} /> },
    { href: '/create', label: 'New Application', icon: <Plus size={16} /> },
  ]

  return (
    <SidebarProvider>
      <Sidebar>
        <SidebarHeader>
          <div className="flex items-center gap-2">
            <LayoutDashboard size={20} className="text-sidebar-primary" />
            <span className="text-base font-semibold">lite-dokploy</span>
          </div>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup label="Home">
            {nav.map((item) => (
              <SidebarItem
                key={item.href}
                icon={item.icon}
                label={item.label}
                active={path === item.href}
                onClick={() => navigate(item.href)}
              />
            ))}
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          <div className="flex items-center justify-between px-2">
            <span className="text-xs text-muted-foreground">{username}</span>
            <button
              onClick={logout}
              className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground"
              title="Sign out"
            >
              <LogOut size={14} />
            </button>
          </div>
          <div className="px-2 text-xs text-muted-foreground">
            v0.1.0
          </div>
        </SidebarFooter>
        <SidebarRail />
      </Sidebar>
      <SidebarInset>
        <header className="flex h-16 items-center gap-3 border-b border-border px-4">
          <SidebarTrigger />
          <Separator orientation="vertical" className="h-6" />
          <Link href="/" className="text-sm font-medium text-muted-foreground hover:text-foreground no-underline">
            Lite Dokploy
          </Link>
          {path.startsWith('/app/') && (
            <>
              <span className="text-muted-foreground">/</span>
              <span className="text-sm font-medium text-foreground">Application</span>
            </>
          )}
          {path === '/create' && (
            <>
              <span className="text-muted-foreground">/</span>
              <span className="text-sm font-medium text-foreground">New Application</span>
            </>
          )}
          <div className="ml-auto">
            <ModeToggle />
          </div>
        </header>
        <main className="flex-1 p-4">
          {children}
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
