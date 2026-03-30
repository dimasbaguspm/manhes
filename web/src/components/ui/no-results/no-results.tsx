import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/cn'

const noResultsVariants = cva(
  'rounded-lg px-4 py-6 text-center',
  {
    variants: {
      variant: {
        default: 'border border-gray-800 bg-gray-900 text-gray-500',
        error: 'border border-red-800 bg-red-950 text-red-300',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
)

interface NoResultsProps
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'message'>,
    VariantProps<typeof noResultsVariants> {
  message: string
}

export function NoResults({ variant, message, className, ...props }: NoResultsProps) {
  return (
    <div
      className={noResultsVariants({ variant, className })}
      data-no-results
      role="status"
      {...props}
    >
      {message}
    </div>
  )
}