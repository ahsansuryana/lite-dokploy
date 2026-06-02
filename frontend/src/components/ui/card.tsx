import { forwardRef } from 'react'
import { cn } from '../../lib/utils'

export const Card = forwardRef<HTMLDivElement, { className?: string; children?: React.ReactNode }>(
  ({ className, children }, ref) => (
    <div ref={ref} className={cn('rounded-xl border bg-card text-card-foreground shadow-sm', className)}>
      {children}
    </div>
  ),
)

export const CardHeader = forwardRef<HTMLDivElement, { className?: string; children?: React.ReactNode }>(
  ({ className, children }, ref) => (
    <div ref={ref} className={cn('flex flex-col space-y-1.5 p-6', className)}>
      {children}
    </div>
  ),
)

export const CardTitle = forwardRef<HTMLHeadingElement, { className?: string; children?: React.ReactNode }>(
  ({ className, children }, ref) => (
    <h3 ref={ref} className={cn('font-semibold leading-none tracking-tight', className)}>
      {children}
    </h3>
  ),
)

export const CardDescription = forwardRef<HTMLParagraphElement, { className?: string; children?: React.ReactNode }>(
  ({ className, children }, ref) => (
    <p ref={ref} className={cn('text-sm text-muted-foreground', className)}>
      {children}
    </p>
  ),
)

export const CardContent = forwardRef<HTMLDivElement, { className?: string; children?: React.ReactNode }>(
  ({ className, children }, ref) => (
    <div ref={ref} className={cn('p-6 pt-0', className)}>
      {children}
    </div>
  ),
)

export const CardFooter = forwardRef<HTMLDivElement, { className?: string; children?: React.ReactNode }>(
  ({ className, children }, ref) => (
    <div ref={ref} className={cn('flex items-center p-6 pt-0', className)}>
      {children}
    </div>
  ),
)
