import { type ButtonHTMLAttributes } from 'react'

export type ButtonVariant = 'primary' | 'default' | 'danger' | 'outline-danger' | 'muted'

const variantClasses: Record<ButtonVariant, string> = {
  primary:         'bg-indigo-600 text-white hover:bg-indigo-500',
  default:         'border border-gray-700 bg-gray-800 text-gray-300 hover:border-gray-500 hover:text-white',
  danger:          'bg-red-900 text-red-300',
  'outline-danger':'border border-red-700 bg-red-900 text-red-300',
  muted:           'border border-gray-700 bg-gray-800 text-gray-400',
}

const base = 'rounded-lg px-4 py-2 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-50'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
}

export function Button({ variant = 'default', className, ...props }: ButtonProps) {
  return (
    <button
      {...props}
      className={[base, variantClasses[variant], className].filter(Boolean).join(' ')}
    />
  )
}
