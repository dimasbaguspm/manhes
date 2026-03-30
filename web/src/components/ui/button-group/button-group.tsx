import { type ReactNode } from 'react'
import { cn } from '@/lib/cn'

interface ButtonGroupProps {
  children: ReactNode
  className?: string
  attached?: boolean
}

export function ButtonGroup({ children, className, attached = false }: ButtonGroupProps) {
  return (
    <div
      role="group"
      className={cn('inline-flex', attached ? 'rounded-md' : 'rounded-lg gap-1', className)}
      data-button-group
    >
      {children}
    </div>
  )
}