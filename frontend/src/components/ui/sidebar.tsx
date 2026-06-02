import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'
import { cn } from '../../lib/utils'
import { Separator } from './separator'

type SidebarContextType = {
  open: boolean
  toggle: () => void
  setOpen: (v: boolean) => void
}

const SidebarContext = createContext<SidebarContextType>({ open: true, toggle: () => {}, setOpen: () => {} })

export function SidebarProvider({ children, defaultOpen = true }: { children: ReactNode; defaultOpen?: boolean }) {
  const [open, setOpen] = useState(defaultOpen)
  const toggle = useCallback(() => setOpen((v) => !v), [])
  return (
    <SidebarContext.Provider value={{ open, toggle, setOpen }}>
      <div className="flex min-h-screen">{children}</div>
    </SidebarContext.Provider>
  )
}

export function Sidebar({ children }: { children: ReactNode }) {
  const { open } = useContext(SidebarContext)
  const width = open ? 'var(--sidebar-width)' : 'var(--sidebar-width-icon)'
  return (
    <div
      className="group/sidebar fixed top-0 left-0 z-50 flex h-full flex-col border-r border-sidebar-border bg-sidebar-background text-sidebar-foreground transition-[width] duration-200 ease-linear"
      style={{ width }}
    >
      {children}
    </div>
  )
}

export function SidebarHeader({ children }: { children: ReactNode }) {
  return (
    <div className="flex items-center gap-2 px-6 py-4">
      {children}
    </div>
  )
}

export function SidebarContent({ children }: { children: ReactNode }) {
  return (
    <div className="flex-1 overflow-auto px-3 py-2">
      {children}
    </div>
  )
}

export function SidebarFooter({ children }: { children: ReactNode }) {
  return (
    <div className="border-t border-sidebar-border p-3">
      {children}
    </div>
  )
}

export function SidebarGroup({ label, children }: { label?: string; children: ReactNode }) {
  return (
    <div className="mb-4">
      {label && (
        <div className="mb-2 px-2 text-xs font-medium text-muted-foreground uppercase tracking-wider">
          {label}
        </div>
      )}
      <div className="space-y-0.5">{children}</div>
    </div>
  )
}

export function SidebarItem({ icon, label, active, onClick }: {
  icon?: ReactNode
  label: string
  active?: boolean
  onClick?: () => void
}) {
  const { open } = useContext(SidebarContext)
  return (
    <button
      onClick={onClick}
      className={cn(
        'flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
        active
          ? 'bg-sidebar-accent text-sidebar-accent-foreground'
          : 'text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-accent-foreground',
      )}
    >
      <span className="shrink-0 size-4">{icon}</span>
      {open && <span className="truncate">{label}</span>}
    </button>
  )
}

export function SidebarInset({ children }: { children: ReactNode }) {
  const { open } = useContext(SidebarContext)
  const width = open ? 'var(--sidebar-width)' : 'var(--sidebar-width-icon)'
  return (
    <div
      className="flex flex-1 flex-col transition-[margin-left] duration-200 ease-linear"
      style={{ marginLeft: width }}
    >
      {children}
    </div>
  )
}

export function SidebarTrigger({ className }: { className?: string }) {
  const { toggle } = useContext(SidebarContext)
  return (
    <button
      onClick={toggle}
      className={cn(
        'inline-flex items-center justify-center rounded-lg text-muted-foreground hover:text-foreground transition-colors size-8',
        className,
      )}
      aria-label="Toggle sidebar"
    >
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="1" y="2" width="14" height="12" rx="2" />
        <line x1="6" y1="2" x2="6" y2="14" />
      </svg>
    </button>
  )
}

export function SidebarRail() {
  const { setOpen } = useContext(SidebarContext)
  return (
    <div
      className="fixed top-0 left-0 z-40 h-full w-1.5 cursor-col-resize transition-colors hover:bg-sidebar-ring/20"
      onMouseEnter={() => setOpen(true)}
    />
  )
}
