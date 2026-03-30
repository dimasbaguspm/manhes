import { Play, Pause, RotateCw } from 'lucide-react'
import { Icon } from '@/components'
import { ButtonIcon } from '@/components/ui'

interface AutoScrollControlsProps {
  isActive: boolean
  speed: number
  onToggle: () => void
  onCycleSpeed: () => void
}

const SPEED_LABELS = ['', 'Slow', 'Medium', 'Fast', 'Faster', 'Turbo']

export function AutoScrollControls({ isActive, speed, onToggle, onCycleSpeed }: AutoScrollControlsProps) {
  return (
    <div className="fixed bottom-12 left-4 z-20 flex flex-col gap-2 md:hidden">
      <ButtonIcon
        variant={isActive ? 'default' : 'outline'}
        size="md"
        onClick={onToggle}
        aria-label={isActive ? 'Pause auto-scroll' : 'Start auto-scroll'}
        title={isActive ? 'Pause auto-scroll (s)' : 'Start auto-scroll (s)'}
        className={isActive ? 'bg-indigo-600 text-white' : 'border-gray-700 text-gray-300'}
      >
        <Icon as={isActive ? Pause : Play} size="small" />
      </ButtonIcon>

      {isActive && (
        <ButtonIcon
          variant="ghost"
          size="sm"
          onClick={onCycleSpeed}
          aria-label="Cycle scroll speed"
          title={`Speed: ${SPEED_LABELS[speed]}`}
          className="text-gray-400 hover:text-white"
        >
          <Icon as={RotateCw} size="small" />
        </ButtonIcon>
      )}

      {isActive && (
        <div className="mt-1 rounded bg-black/70 px-1.5 py-0.5 text-center">
          <span className="text-xs text-gray-300">{SPEED_LABELS[speed]}</span>
        </div>
      )}
    </div>
  )
}
