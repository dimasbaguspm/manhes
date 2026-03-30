import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/cn'

const progressTrackVariants = cva(
  'w-full overflow-hidden rounded-full bg-gray-800',
  {
    variants: {
      size: {
        sm: 'h-1',
        md: 'h-2',
        lg: 'h-4',
      },
    },
    defaultVariants: {
      size: 'md',
    },
  }
)

const progressBarVariants = cva(
  'h-full transition-all duration-300 ease-in-out',
  {
    variants: {
      color: {
        default: 'bg-indigo-500',
        primary: 'bg-indigo-500',
        success: 'bg-emerald-500',
        warning: 'bg-amber-500',
        error: 'bg-red-500',
      },
    },
    defaultVariants: {
      color: 'default',
    },
  }
)

interface ProgressProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'value'>, VariantProps<typeof progressTrackVariants> {
  value?: number
  max?: number
  color?: 'default' | 'primary' | 'success' | 'warning' | 'error'
  showBar?: boolean
}

export function Progress({
  value = 0,
  max = 100,
  size,
  color = 'default',
  showBar = true,
  className,
  ...props
}: ProgressProps) {
  const pct = Math.min(100, Math.max(0, (value / max) * 100))

  return (
    <div
      className={cn(progressTrackVariants({ size }), className)}
      role="progressbar"
      aria-valuenow={value}
      aria-valuemin={0}
      aria-valuemax={max}
      data-progress
      {...props}
    >
      {showBar && (
        <div
          className={progressBarVariants({ color })}
          style={{ width: `${pct}%` }}
          data-progress-bar
        />
      )}
    </div>
  )
}