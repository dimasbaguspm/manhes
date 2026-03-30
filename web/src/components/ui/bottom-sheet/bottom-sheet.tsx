import { createContext, useContext, useState, useCallback, useEffect, type ReactNode, type ReactElement } from 'react'
import { cn } from '@/lib/cn'
import { X } from 'lucide-react'

interface BottomSheetContextValue {
  open: boolean
  onOpenChange: (open: boolean) => void
}

const BottomSheetContext = createContext<BottomSheetContextValue | null>(null)

function useBottomSheetContext() {
  const ctx = useContext(BottomSheetContext)
  if (!ctx) throw new Error('BottomSheet sub-components must be used within <BottomSheet>')
  return ctx
}

interface BottomSheetProps {
  open?: boolean
  onOpenChange?: (open: boolean) => void
  children: ReactNode
}

export function BottomSheet({ open: controlledOpen, onOpenChange, children }: BottomSheetProps) {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(false)
  const isControlled = controlledOpen !== undefined
  const open = isControlled ? controlledOpen : uncontrolledOpen

  const handleOpenChange = useCallback((newOpen: boolean) => {
    if (isControlled) {
      onOpenChange?.(newOpen)
    } else {
      setUncontrolledOpen(newOpen)
    }
  }, [isControlled, onOpenChange])

  return (
    <BottomSheetContext.Provider value={{ open, onOpenChange: handleOpenChange }}>
      {children}
    </BottomSheetContext.Provider>
  )
}

interface BottomSheetTriggerProps {
  children: ReactNode
  asChild?: boolean
}

export function BottomSheetTrigger({ children, asChild }: BottomSheetTriggerProps) {
  const { onOpenChange } = useBottomSheetContext()

  if (asChild && typeof children !== 'string') {
    // Clone and attach onClick
    const child = children as ReactElement
    return (
      <div onClick={() => onOpenChange(true)} data-bottom-sheet-trigger>
        {child}
      </div>
    )
  }

  return (
    <button onClick={() => onOpenChange(true)} data-bottom-sheet-trigger>
      {children}
    </button>
  )
}

interface BottomSheetContentProps {
  children: ReactNode
  className?: string
}

export function BottomSheetContent({ children, className }: BottomSheetContentProps) {
  const { open, onOpenChange } = useBottomSheetContext()

  useEffect(() => {
    if (!open) return
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onOpenChange(false)
    }
    document.addEventListener('keydown', handleEsc)
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', handleEsc)
      document.body.style.overflow = ''
    }
  }, [open, onOpenChange])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50" data-bottom-sheet-overlay onClick={() => onOpenChange(false)}>
      <div
        className="fixed bottom-0 left-0 right-0 z-50 max-h-[85vh] overflow-y-auto rounded-t-xl bg-gray-900 p-4 pb-8 shadow-xl transition-all"
        onClick={e => e.stopPropagation()}
        data-bottom-sheet-content
      >
        <div className="mb-4 flex justify-center">
          <div className="h-1 w-12 rounded-full bg-gray-700" />
        </div>
        <div className={cn(className)} data-bottom-sheet-body>
          {children}
        </div>
      </div>
      {/* Backdrop */}
      <div className="fixed inset-0 bg-black/60" aria-hidden="true" />
    </div>
  )
}

interface BottomSheetHeaderProps {
  children: ReactNode
  className?: string
}

export function BottomSheetHeader({ children, className }: BottomSheetHeaderProps) {
  return (
    <div className={cn('flex items-center justify-between mb-4', className)} data-bottom-sheet-header>
      {children}
    </div>
  )
}

interface BottomSheetTitleProps {
  children: ReactNode
  className?: string
}

export function BottomSheetTitle({ children, className }: BottomSheetTitleProps) {
  return (
    <h2 className={cn('text-lg font-semibold text-gray-100', className)} data-bottom-sheet-title>
      {children}
    </h2>
  )
}

interface BottomSheetCloseProps {
  asChild?: boolean
  children?: ReactNode
  className?: string
}

export function BottomSheetClose({ asChild, children, className }: BottomSheetCloseProps) {
  const { onOpenChange } = useBottomSheetContext()

  if (asChild && typeof children !== 'string') {
    const child = children as React.ReactElement
    return (
      <div onClick={() => onOpenChange(false)} className={className} data-bottom-sheet-close>
        {child}
      </div>
    )
  }

  return (
    <button
      onClick={() => onOpenChange(false)}
      className={cn('rounded-lg p-1.5 text-gray-400 hover:bg-gray-800 hover:text-white transition', className)}
      aria-label="Close"
      data-bottom-sheet-close
    >
      {children ?? <X className="h-5 w-5" />}
    </button>
  )
}
