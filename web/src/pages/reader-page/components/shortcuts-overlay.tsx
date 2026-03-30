import { X } from 'lucide-react'
import { Icon } from '@/components'
import { ButtonIcon } from '@/components/ui'

interface ShortcutsOverlayProps {
  open: boolean
  onClose: () => void
}

const KEYBOARD_SHORTCUTS = [
  { keys: ['f'], description: 'Toggle fullscreen' },
  { keys: ['s'], description: 'Open / close settings' },
  { keys: ['/'], description: 'Show keyboard shortcuts' },
  { keys: ['Esc'], description: 'Close any open panel' },
]

const GESTURE_SHORTCUTS = [
  { gesture: 'Double-tap', description: 'Show / hide header' },
  { gesture: 'Double-tap + hold', description: 'Open settings' },
]

export function ShortcutsOverlay({ open, onClose }: ShortcutsOverlayProps) {
  if (!open) return null

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 z-50 bg-black/60"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="fixed left-1/2 top-1/2 z-50 w-full max-w-sm -translate-x-1/2 -translate-y-1/2 rounded-xl border border-gray-800 bg-gray-950 p-5 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-gray-200">Keyboard shortcuts</h2>
          <ButtonIcon
            variant="ghost"
            size="sm"
            onClick={onClose}
            aria-label="Close shortcuts"
          >
            <Icon as={X} size="small" />
          </ButtonIcon>
        </div>

        <div className="space-y-2">
          {KEYBOARD_SHORTCUTS.map(s => (
            <div key={s.keys.join('+')} className="flex items-center justify-between gap-4">
              <span className="text-sm text-gray-400">{s.description}</span>
              <div className="flex shrink-0 gap-1">
                {s.keys.map(k => (
                  <kbd
                    key={k}
                    className="rounded border border-gray-700 bg-gray-800 px-2 py-0.5 font-mono text-xs text-gray-300"
                  >
                    {k}
                  </kbd>
                ))}
              </div>
            </div>
          ))}
        </div>

        <p className="mb-2 mt-5 text-xs font-semibold uppercase tracking-wider text-gray-600">
          Gestures
        </p>
        <div className="space-y-2">
          {GESTURE_SHORTCUTS.map(s => (
            <div key={s.gesture} className="flex items-center justify-between gap-4">
              <span className="text-sm text-gray-400">{s.description}</span>
              <span className="shrink-0 text-xs text-gray-500">{s.gesture}</span>
            </div>
          ))}
        </div>
      </div>
    </>
  )
}
