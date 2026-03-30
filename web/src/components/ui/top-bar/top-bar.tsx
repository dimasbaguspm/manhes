import { cn } from '@/lib/cn'

interface TopBarProps {
  leading?: React.ReactNode
  center?: React.ReactNode
  trailing?: React.ReactNode
  className?: string
  sticky?: boolean
}

export function TopBar({ leading, center, trailing, className, sticky = false }: TopBarProps) {
  return (
    <header
      className={cn(
        'flex h-14 items-center border-b border-gray-800 bg-gray-900 px-4',
        sticky && 'sticky top-0 z-50',
        className
      )}
      data-top-bar
    >
      <div className="flex w-1/3 items-center" data-top-bar-leading>
        {leading}
      </div>
      <div className="flex flex-1 items-center justify-center" data-top-bar-center>
        {center}
      </div>
      <div className="flex w-1/3 items-center justify-end gap-2" data-top-bar-trailing>
        {trailing}
      </div>
    </header>
  )
}
