import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/cn'

const loaderVariants = cva(
  'animate-spin rounded-full border-2 border-current',
  {
    variants: {
      size: {
        sm: 'h-4 w-4',
        md: 'h-6 w-6',
        lg: 'h-8 w-8',
      },
      color: {
        white: 'border-white/30 border-t-white',
        gray: 'border-gray-600 border-t-gray-300',
        primary: 'border-indigo-600/30 border-t-indigo-400',
      },
    },
    defaultVariants: {
      size: 'md',
      color: 'primary',
    },
  }
)

interface LoaderProps extends Omit<React.HTMLAttributes<HTMLDivElement>, 'color'>, VariantProps<typeof loaderVariants> {}

export function Loader({ size, color, className, ...props }: LoaderProps) {
  return (
    <div
      className={cn(loaderVariants({ size, color }), className)}
      role="status"
      aria-label="Loading"
      data-loader
      {...props}
    >
      <span className="sr-only">Loading...</span>
    </div>
  )
}