import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/cn'
import type { ButtonHTMLAttributes } from 'react'

const buttonIconVariants = cva(
  'inline-flex items-center justify-center rounded-lg font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 disabled:cursor-not-allowed disabled:opacity-50',
  {
    variants: {
      variant: {
        default: 'bg-gray-800 text-gray-300 hover:bg-gray-700 hover:text-white',
        outline: 'border border-gray-700 bg-transparent text-gray-300 hover:border-gray-500 hover:text-white',
        ghost: 'text-gray-300 hover:text-white hover:bg-gray-800',
      },
      size: {
        sm: 'h-8 w-8',
        md: 'h-10 w-10',
        lg: 'h-12 w-12',
      },
    },
    defaultVariants: {
      variant: 'ghost',
      size: 'md',
    },
  }
)

interface ButtonIconProps
  extends ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonIconVariants> {}

export function ButtonIcon({ variant, size, className, ...props }: ButtonIconProps) {
  return (
    <button
      className={buttonIconVariants({ variant, size, className })}
      data-button-icon
      {...props}
    />
  )
}
