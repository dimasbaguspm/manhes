import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/cn'

const textVariants = cva('text-gray-300', {
  variants: {
    size: {
      xs: 'text-xs',
      sm: 'text-sm',
      md: 'text-base',
      lg: 'text-lg',
    },
    color: {
      default: 'text-gray-300',
      muted: 'text-gray-500',
      primary: 'text-indigo-400',
      error: 'text-red-400',
      success: 'text-emerald-400',
      white: 'text-white',
    },
  },
  defaultVariants: {
    size: 'sm',
    color: 'default',
  },
})

interface TextProps
  extends Omit<React.HTMLAttributes<HTMLParagraphElement>, 'color'>,
    VariantProps<typeof textVariants> {}

export function Text({ size, color, className, ...props }: TextProps) {
  return (
    <p
      className={textVariants({ size, color, className })}
      data-text
      {...props}
    />
  )
}
