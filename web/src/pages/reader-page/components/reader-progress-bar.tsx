import { Progress, Text } from '@/components/ui'
import { useBattery } from '@/hooks/use-battery'
import { useClock } from '@/hooks/use-clock'

interface ReaderProgressBarProps {
  /** Current scroll progress 0–100. */
  pct: number
  /** Whether to render the thin bar at the bottom of the viewport. */
  showBar: boolean
  /** Whether to render the floating percentage badge. */
  showIndicator: boolean
}

export function ReaderProgressBar({ pct, showBar, showIndicator }: ReaderProgressBarProps) {
  const { level, charging } = useBattery()
  const time = useClock()

  return (
    <>
      {showBar && (
        <div className="fixed bottom-0 left-0 right-0 z-10">
          <Progress size="sm" color="primary" value={pct} max={100} />
        </div>
      )}

      {/* Desktop indicator — right side */}
      {showIndicator && (
        <div className="fixed bottom-3 right-3 z-10 rounded bg-black/70 px-2 py-1 backdrop-blur md:hidden">
          <Text size="xs" color="white">{pct}%</Text>
        </div>
      )}

      {/* Mobile info row — centered, below progress bar */}
      {showIndicator && (
        <div className="fixed bottom-3 left-1/2 z-10 flex -translate-x-1/2 items-center gap-3 rounded bg-black/70 px-3 py-1 backdrop-blur md:hidden">
          <Text size="xs" color="white">{time}</Text>
          <div className="h-3 w-px bg-white" />
          <Text size="xs" color="white">{Math.round(level * 100)}%</Text>
          {charging && <div className="h-1.5 w-1.5 rounded-full bg-emerald-400" title="Charging" />}
        </div>
      )}
    </>
  )
}
