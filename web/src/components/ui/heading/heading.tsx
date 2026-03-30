import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/cn'

const headingBase = 'font-semibold text-gray-100'

const headingVariants = cva(headingBase, {
  variants: {
    level: {
      h1: 'text-4xl',
      h2: 'text-3xl',
      h3: 'text-2xl',
      h4: 'text-xl',
      h5: 'text-lg',
      h6: 'text-base',
    },
    size: {
      sm: 'text-lg',
      md: 'text-xl',
      lg: 'text-2xl',
      xl: 'text-3xl',
    },
  },
  defaultVariants: {
    level: 'h3',
  },
})

interface HeadingProps
  extends Omit<React.HTMLAttributes<HTMLHeadingElement>, 'level'>,
    VariantProps<typeof headingVariants> {}

export function Heading({ level = 'h3', className, ...props }: HeadingProps) {
  const Tag = level as 'h1'
  return (
    <Tag
      className={cn(headingVariants({ level }), className)}
      data-heading
      data-heading-level={level}
      {...props}
    />
  )
}
