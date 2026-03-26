import type { ReactNode } from 'react'

interface ReaderStripProps {
  /** Tailwind max-width class controlling how wide the strip is. */
  maxWidthClass: string
  /** Gap between images in Tailwind scale units (value × 4 = px). 0 = no gap. */
  gap: number
  /** Called on every click/tap — double-tap detection lives in the caller. */
  onClick: () => void
  /** Render prop for the loading state. Omit when not loading. */
  renderLoading?: () => ReactNode
  /** Render prop for an error banner. Omit when there's no error. */
  renderError?: () => ReactNode
  /** Render prop for the actual page images inside the constrained container. */
  renderPages?: () => ReactNode
  /**
   * Render prop for the bottom chapter-navigation footer.
   * Automatically isolated from the double-tap handler so nav clicks
   * don't accidentally toggle the header.
   */
  renderFooter?: () => ReactNode
}

export function ReaderStrip({
  maxWidthClass,
  gap,
  onClick,
  renderLoading,
  renderError,
  renderPages,
  renderFooter,
}: ReaderStripProps) {
  return (
    <div className="flex flex-col items-center py-4" onClick={onClick}>

      {renderLoading?.()}
      {renderError?.()}

      {renderPages && (
        <div
          className={`flex w-full flex-col ${maxWidthClass} mx-auto`}
          style={gap > 0 ? { gap: `${gap * 4}px` } : undefined}
        >
          {renderPages()}
        </div>
      )}

      {/* Footer clicks must not bubble up to the double-tap handler. */}
      {renderFooter && (
        <div onClick={e => e.stopPropagation()}>
          {renderFooter()}
        </div>
      )}

    </div>
  )
}
