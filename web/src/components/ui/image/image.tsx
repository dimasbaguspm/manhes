import { BookOpen } from 'lucide-react'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/cn'
import { Icon } from '@/components/icon'

const imageVariants = cva('overflow-hidden rounded-md bg-gray-800 flex-shrink-0', {
  variants: {
    size: {
      xs: 'w-12',
      sm: 'w-16',
      md: 'w-20',
      lg: 'w-28',
      xl: 'w-36',
    },
    aspect: {
      portrait: 'aspect-[2/3]',
      landscape: 'aspect-[3/4]',
      wide: 'aspect-[4/3]',
      square: 'aspect-square',
    },
  },
  defaultVariants: {
    size: 'md',
    aspect: 'portrait',
  },
})

interface ImageProps extends Omit<React.ImgHTMLAttributes<HTMLImageElement>, 'src'>, VariantProps<typeof imageVariants> {
  src?: string | null
  alt?: string
  fallback?: React.ReactNode
}

export function Image({ src, alt, size, aspect, className, fallback, ...props }: ImageProps) {
  if (!src) {
    return (
      <div className={cn(imageVariants({ size, aspect }), 'flex items-center justify-center', className)}>
        {fallback ?? <Icon as={BookOpen} size="medium" className="text-gray-600" />}
      </div>
    )
  }

  return (
    <div className={cn(imageVariants({ size, aspect }), 'relative', className)}>
      <img
        src={src}
        alt={alt ?? ''}
        className="h-full w-full object-cover"
        loading="lazy"
        {...props}
      />
    </div>
  )
}

export { imageVariants }
