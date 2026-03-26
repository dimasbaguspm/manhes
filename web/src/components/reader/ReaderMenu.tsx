import type { ReactNode } from 'react'

interface ReaderMenuProps {
  open: boolean
  onClose: () => void
  children: ReactNode
}

/**
 * Adaptive menu container:
 * - Mobile  → animated bottom sheet (always in DOM, slides in/out via transform).
 * - Desktop → fixed side panel (conditionally mounted).
 *
 * Renders `children` once in the active container. The inactive container is
 * hidden via `md:hidden` / `hidden md:block` so only one copy is interactive.
 */
export function ReaderMenu({ open, onClose, children }: ReaderMenuProps) {
  return (
    <>
      {/* Mobile dim backdrop */}
      <div
        className={`fixed inset-0 z-30 bg-black/60 transition-opacity duration-300 md:hidden ${
          open ? 'opacity-100' : 'opacity-0 pointer-events-none'
        }`}
        onClick={onClose}
      />

      {/* Mobile bottom sheet — always mounted for smooth slide animation */}
      <div
        className={`fixed bottom-0 left-0 right-0 z-40 max-h-[82vh] overflow-y-auto rounded-t-2xl border-t border-gray-800 bg-gray-950 transition-transform duration-300 ease-out md:hidden ${
          open ? 'translate-y-0' : 'translate-y-full'
        }`}
      >
        <DragHandle />
        {children}
      </div>

      {/* Desktop side panel — conditionally mounted */}
      {open && (
        <aside className="fixed right-0 top-[3.25rem] z-40 hidden h-[calc(100vh-3.25rem)] w-72 overflow-y-auto border-l border-gray-800 bg-gray-950 md:block">
          <div className="flex items-center justify-between border-b border-gray-800 px-4 py-3">
            <span className="text-sm font-semibold text-gray-200">Reader settings</span>
            <button
              onClick={onClose}
              aria-label="Close settings"
              className="rounded p-0.5 text-gray-600 transition hover:text-gray-300"
            >
              <CloseIcon />
            </button>
          </div>
          {children}
        </aside>
      )}
    </>
  )
}

function DragHandle() {
  return (
    <div className="flex justify-center pt-3 pb-1">
      <div className="h-1 w-10 rounded-full bg-gray-700" />
    </div>
  )
}

function CloseIcon() {
  return (
    <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
    </svg>
  )
}
