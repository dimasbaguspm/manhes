interface ReaderProgressBarProps {
  /** Current scroll progress 0–100. */
  pct: number
  /** Whether to render the thin bar at the bottom of the viewport. */
  showBar: boolean
  /** Whether to render the floating percentage badge. */
  showIndicator: boolean
}

export function ReaderProgressBar({ pct, showBar, showIndicator }: ReaderProgressBarProps) {
  return (
    <>
      {showBar && (
        <div className="fixed bottom-0 left-0 right-0 z-10 h-1 bg-gray-800/50">
          <div
            className="h-full bg-indigo-500 transition-[width] duration-75"
            style={{ width: `${pct}%` }}
          />
        </div>
      )}

      {showIndicator && (
        <div className="fixed bottom-3 right-3 z-10 rounded bg-black/70 px-2 py-1 text-xs font-medium text-white backdrop-blur">
          {pct}%
        </div>
      )}
    </>
  )
}
