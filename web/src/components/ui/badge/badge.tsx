import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/cn'

const badgeVariants = cva(
  'inline-flex items-center rounded-full font-medium transition-colors',
  {
    variants: {
      variant: {
        default: 'bg-gray-800 text-gray-300 border border-gray-700',
        secondary: 'bg-gray-700 text-gray-300',
        success: 'bg-emerald-900/50 text-emerald-300 border border-emerald-700',
        warning: 'bg-amber-900/50 text-amber-300 border border-amber-700',
        error: 'bg-red-900/50 text-red-300 border border-red-700',
        primary: 'bg-indigo-900/50 text-indigo-300 border border-indigo-700',
      },
      size: {
        sm: 'px-2 py-0.5 text-[10px]',
        md: 'px-2.5 py-1 text-xs',
        lg: 'px-3 py-1.5 text-sm',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'md',
    },
  }
)

interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

export function Badge({ variant, size, className, ...props }: BadgeProps) {
  return (
    <span
      className={badgeVariants({ variant, size, className })}
      data-badge
      {...props}
    />
  )
}

export { badgeVariants }
