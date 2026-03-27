import { type ReactNode } from 'react'

interface BadgeProps {
  children: ReactNode
  className?: string
}

export function Badge({ children, className = '' }: BadgeProps) {
  return (
    <span className={`rounded-full px-2.5 py-1 text-xs font-medium ${className}`}>
      {children}
    </span>
  )
}
