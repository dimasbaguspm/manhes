import { Icon } from '@/components'
import { ButtonIcon } from '@/components/ui'
import { Gauge } from 'lucide-react'

interface AutoScrollControlsProps {
  speed: number
  onCycle: () => void
}

const SPEED_BG = [
  'border border-gray-700 text-gray-500',
  'bg-indigo-900 text-indigo-300',
  'bg-indigo-800 text-indigo-200',
  'bg-indigo-600 text-white',
  'bg-indigo-500 text-white',
]

export function AutoScrollControls({ speed, onCycle }: AutoScrollControlsProps) {
  return (
    <div className="fixed bottom-3 left-3 z-10 md:hidden">
      <ButtonIcon
        size="sm"
        onClick={onCycle}
        aria-label={`Auto-scroll: ${speed === 0 ? 'Off' : speed}`}
        title={`Auto-scroll: ${speed === 0 ? 'Off' : speed} (tap to cycle)`}
        className={SPEED_BG[speed]}
      >
        <Icon as={Gauge} size="small" />
      </ButtonIcon>
    </div>
  )
}
