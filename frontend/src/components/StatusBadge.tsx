import { cn } from '../lib/utils'

const statusColors: Record<string, string> = {
  running: 'bg-green-500',
  done: 'bg-green-500',
  deploying: 'bg-yellow-500',
  deploying_started: 'bg-yellow-500',
  failed: 'bg-red-500',
  error: 'bg-red-500',
  stopped: 'bg-gray-500',
  idle: 'bg-muted-foreground/40',
  cancelled: 'bg-muted-foreground',
}

const badgeVariants: Record<string, string> = {
  running: 'bg-green-500/15 text-green-500',
  done: 'bg-green-500/15 text-green-500',
  deploying: 'bg-yellow-500/15 text-yellow-500',
  deploying_started: 'bg-yellow-500/15 text-yellow-500',
  failed: 'bg-red-500/15 text-red-500',
  error: 'bg-red-500/15 text-red-500',
  stopped: 'bg-muted-foreground/15 text-muted-foreground',
  idle: 'bg-muted-foreground/15 text-muted-foreground',
  cancelled: 'bg-muted-foreground/15 text-muted-foreground',
}

export function StatusDot({ status }: { status: string }) {
  return (
    <span
      className={cn('inline-block size-2 rounded-full', statusColors[status] || statusColors.idle)}
      aria-hidden
    />
  )
}

export function StatusBadge({ status, className }: { status: string; className?: string }) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-md px-2.5 py-0.5 text-xs font-semibold transition-colors',
        badgeVariants[status] || 'bg-muted-foreground/15 text-muted-foreground',
        className,
      )}
    >
      <StatusDot status={status} />
      {status}
    </span>
  )
}
