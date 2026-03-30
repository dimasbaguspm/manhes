import type { ReactNode } from 'react'

interface ReaderStripProps {
  /** Tailwind max-width class controlling how wide the strip is. */
  maxWidthClass: string
  /** Called on pointerdown — used for double-tap and hold detection. */
  onPointerDown: () => void
  /** Called on pointerup — completes the tap or hold gesture. */
  onPointerUp: () => void
  /** Called on pointercancel — cancels any in-flight hold timer. */
  onPointerCancel: () => void
  /** Render prop for the loading state. Omit when not loading. */
  renderLoading?: () => ReactNode
  /** Render prop for an error banner. Omit when there's no error. */
  renderError?: () => ReactNode
  /** Render prop for the actual page content inside the constrained container. */
  renderPages?: () => ReactNode
  /**
   * Render prop for the bottom chapter-navigation footer.
   * Automatically isolated from pointer handlers so nav clicks
   * don't accidentally trigger tap/hold gestures.
   */
  renderFooter?: () => ReactNode
}

export function ReaderStrip({
  maxWidthClass,
  onPointerDown,
  onPointerUp,
  onPointerCancel,
  renderLoading,
  renderError,
  renderPages,
  renderFooter,
}: ReaderStripProps) {
  return (
    <div
      className="flex flex-col items-center py-4"
      onPointerDown={onPointerDown}
      onPointerUp={onPointerUp}
      onPointerCancel={onPointerCancel}
    >

      {renderLoading?.()}
      {renderError?.()}

      {renderPages && (
        <div className={`flex w-full flex-col ${maxWidthClass} mx-auto`}>
          {renderPages()}
        </div>
      )}

      {/* Footer pointer events must not bubble up to the gesture handler. */}
      {renderFooter && (
        <div
          onPointerDown={e => e.stopPropagation()}
          onPointerUp={e => e.stopPropagation()}
          onPointerCancel={e => e.stopPropagation()}
        >
          {renderFooter()}
        </div>
      )}

    </div>
  )
}
