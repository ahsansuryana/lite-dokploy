import { cn } from '../../lib/utils'

const variants: Record<string, string> = {
  default: 'border-transparent bg-primary text-primary-foreground',
  secondary: 'border-transparent bg-secondary text-secondary-foreground',
  destructive: 'border-transparent bg-destructive text-destructive-foreground',
  outline: 'text-foreground',
  red: 'border-transparent bg-red-500/15 text-red-500',
  yellow: 'border-transparent bg-yellow-500/15 text-yellow-500',
  green: 'border-transparent bg-green-500/15 text-green-500',
  blue: 'border-transparent bg-blue-500/15 text-blue-500',
  gray: 'border-transparent bg-muted-foreground/15 text-muted-foreground',
}

export function Badge({ className, variant = 'default', children }: {
  className?: string
  variant?: string
  children?: React.ReactNode
}) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors',
        variants[variant] || variants.default,
        className,
      )}
    >
      {children}
    </span>
  )
}
