import { type ReactNode } from 'react'
import { cn } from '@/lib/cn'

interface BadgeGroupProps {
  children: ReactNode
  className?: string
  gap?: string
}

export function BadgeGroup({ children, className, gap = 'gap-1' }: BadgeGroupProps) {
  return (
    <div
      className={cn('inline-flex flex-wrap', gap, className)}
      data-badge-group
    >
      {children}
    </div>
  )
}
