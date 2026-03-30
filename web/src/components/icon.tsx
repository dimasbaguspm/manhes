import { type LucideIcon } from 'lucide-react'

type IconSize = 'small' | 'medium' | 'large'

const sizeClasses: Record<IconSize, string> = {
  small: 'h-4 w-4',
  medium: 'h-5 w-5',
  large: 'h-6 w-6',
}

interface IconProps {
  as: LucideIcon
  size?: IconSize
  className?: string
  fill?: string
}

export function Icon({ as: IconComponent, size = 'medium', className, ...props }: IconProps) {
  return (
    <IconComponent
      {...props}
      className={[sizeClasses[size], className].filter(Boolean).join(' ')}
    />
  )
}
