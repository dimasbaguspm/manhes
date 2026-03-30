import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/cn'
import type { LucideIcon } from 'lucide-react'

const iconBadgeContainerVariants = cva('relative inline-flex', {
  variants: {
    size: {
      sm: 'h-5 w-5',
      md: 'h-6 w-6',
      lg: 'h-8 w-8',
    },
  },
  defaultVariants: {
    size: 'md',
  },
})

const iconBadgeDotVariants = cva(
  'absolute rounded-full border border-gray-900',
  {
    variants: {
      variant: {
        default: 'bg-gray-400',
        success: 'bg-emerald-400',
        warning: 'bg-amber-400',
        error: 'bg-red-400',
        primary: 'bg-indigo-400',
      },
      size: {
        sm: 'h-2 w-2 -top-0.5 -right-0.5',
        md: 'h-2.5 w-2.5 -top-0.5 -right-0.5',
        lg: 'h-3 w-3 -top-1 -right-1',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'md',
    },
  }
)

interface IconBadgeProps
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'children'> {
  icon: LucideIcon
  children?: React.ReactNode
  variant?: 'default' | 'success' | 'warning' | 'error' | 'primary'
  badgeVariant?: 'default' | 'success' | 'warning' | 'error' | 'primary'
  size?: 'sm' | 'md' | 'lg'
  showBadge?: boolean
  iconClassName?: string
}

export function IconBadge({
  icon: IconComponent,
  children,
  variant = 'default',
  badgeVariant,
  size = 'md',
  showBadge = true,
  iconClassName,
  className,
  ...props
}: IconBadgeProps) {
  const resolvedBadgeVariant = badgeVariant ?? variant

  return (
    <div
      className={cn(iconBadgeContainerVariants({ size }), className)}
      data-icon-badge
      {...props}
    >
      <IconComponent className={cn('text-current', iconClassName)} />
      {showBadge && (
        <span
          className={iconBadgeDotVariants({ variant: resolvedBadgeVariant, size })}
          data-icon-badge-dot
        />
      )}
    </div>
  )
}