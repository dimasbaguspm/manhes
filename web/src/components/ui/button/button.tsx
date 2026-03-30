import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '@/lib/cn'
import type { ButtonHTMLAttributes } from 'react'

const buttonVariants = cva(
  'inline-flex items-center justify-center rounded-lg font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500 disabled:cursor-not-allowed disabled:opacity-50',
  {
    variants: {
      variant: {
        default: 'bg-indigo-600 text-white hover:bg-indigo-500',
        outline: 'border border-gray-700 bg-transparent text-gray-300 hover:border-gray-500 hover:text-white',
        ghost: 'text-gray-300 hover:text-white hover:bg-gray-800',
        danger: 'bg-red-900 text-red-300 hover:bg-red-800',
        'outline-danger': 'border border-red-700 bg-red-900 text-red-300 hover:bg-red-800',
        muted: 'border border-gray-700 bg-gray-800 text-gray-400 hover:bg-gray-700 hover:text-gray-200',
      },
      color: {
        white: 'bg-white text-gray-900 hover:bg-gray-100',
        transparent: 'bg-transparent text-gray-300 hover:text-white',
        primary: 'bg-indigo-600 text-white hover:bg-indigo-500',
        muted: 'bg-gray-800 text-gray-400 hover:bg-gray-700 hover:text-gray-200',
        danger: 'bg-red-900 text-red-300 hover:bg-red-800',
      },
      size: {
        sm: 'h-8 px-3 text-xs',
        md: 'h-10 px-4 text-sm',
        lg: 'h-12 px-6 text-base',
      },
    },
    compoundVariants: [
      { variant: 'outline', color: 'white', class: 'border-white text-white hover:bg-white/10' },
      { variant: 'ghost', color: 'white', class: 'text-white hover:bg-white/10' },
      { variant: 'default', color: 'white', class: 'bg-white text-gray-900 hover:bg-gray-100' },
      { variant: 'outline', color: 'primary', class: 'border-indigo-600 text-indigo-400 hover:bg-indigo-600/20' },
      { variant: 'ghost', color: 'primary', class: 'text-indigo-400 hover:bg-indigo-600/20' },
    ],
    defaultVariants: {
      variant: 'default',
      color: 'muted',
      size: 'md',
    },
  }
)

interface ButtonProps
  extends Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'color'>,
    VariantProps<typeof buttonVariants> {}

export function Button({ variant, color, size, className, ...props }: ButtonProps) {
  return (
    <button
      className={buttonVariants({ variant, color, size, className })}
      data-button
      {...props}
    />
  )
}

export { buttonVariants }