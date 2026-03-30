import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Toggle } from '@/pages/reader-page/components/toggle'
import { Icon } from '@/components/icon'
import type { ReaderSettings } from '@/pages/reader-page/components/use-reader-settings'

const STRIP_WIDTH_OPTIONS = [
  { label: 'Narrow', value: 'narrow' },
  { label: 'Normal', value: 'normal' },
  { label: 'Wide',   value: 'wide'   },
  { label: 'Full',   value: 'full'   },
]

const BG_COLOR_OPTIONS = [
  { label: 'Black', value: 'black', bg: 'bg-black',    text: 'text-gray-200' },
  { label: 'Dark',  value: 'dark',  bg: 'bg-gray-950', text: 'text-gray-200' },
  { label: 'White', value: 'white', bg: 'bg-white',    text: 'text-gray-900' },
]

const AUTO_SCROLL_LABELS = ['', 'Slow', 'Medium', 'Fast', 'Faster', 'Turbo']

const navBtn =
  'rounded border border-gray-700 px-3 py-1 text-xs text-gray-300 transition hover:border-gray-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-30'

interface ReaderSettingsPanelProps {
  settings: ReaderSettings
  set: <K extends keyof ReaderSettings>(key: K, value: ReaderSettings[K]) => void
  headerVisible: boolean
  onHeaderToggle: () => void
  prevDisabled: boolean
  nextDisabled: boolean
  onPrev: () => void
  onNext: () => void
}

export function ReaderSettingsPanel({
  settings,
  set,
  headerVisible,
  onHeaderToggle,
  prevDisabled,
  nextDisabled,
  onPrev,
  onNext,
}: ReaderSettingsPanelProps) {
  return (
    <div className="flex flex-col gap-5 p-4">

      <section>
        <SectionLabel>Strip width</SectionLabel>
        <div className="flex gap-1">
          {STRIP_WIDTH_OPTIONS.map(opt => (
            <button
              key={opt.value}
              onClick={() => set('stripWidth', opt.value)}
              className={`flex-1 rounded px-2 py-1.5 text-xs transition ${
                settings.stripWidth === opt.value
                  ? 'bg-indigo-600 text-white'
                  : 'bg-gray-800 text-gray-400 hover:bg-gray-700 hover:text-white'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </section>

      <section>
        <SectionLabel>Background</SectionLabel>
        <div className="flex gap-2">
          {BG_COLOR_OPTIONS.map(opt => (
            <button
              key={opt.value}
              onClick={() => set('bgColor', opt.value)}
              className={`flex-1 rounded border px-2 py-1.5 text-xs transition ${opt.bg} ${opt.text} ${
                settings.bgColor === opt.value ? 'border-indigo-500' : 'border-gray-700'
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </section>

      <section className="space-y-3">
        <SectionLabel>Display</SectionLabel>
        <ToggleRow label="Show header"    on={headerVisible}              onToggle={onHeaderToggle} />
        <ToggleRow label="Progress bar"   on={settings.showProgress}      onToggle={() => set('showProgress', !settings.showProgress)} />
        <ToggleRow label="Page indicator" on={settings.showPageIndicator} onToggle={() => set('showPageIndicator', !settings.showPageIndicator)} />
        <ToggleRow label="Auto scroll"    on={settings.autoScroll}        onToggle={() => set('autoScroll', !settings.autoScroll)} />
        {settings.autoScroll && (
          <div>
            <p className="mb-1 text-xs text-gray-500">
              Speed —{' '}
              <span className="font-normal text-gray-400">{AUTO_SCROLL_LABELS[settings.autoScrollSpeed]}</span>
            </p>
            <input
              type="range"
              min={1}
              max={5}
              value={settings.autoScrollSpeed}
              onChange={e => set('autoScrollSpeed', parseInt(e.target.value))}
              className="w-full accent-indigo-500"
            />
          </div>
        )}
      </section>

      <section>
        <button
          onClick={() => document.documentElement.requestFullscreen?.()}
          className="w-full rounded border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-300 transition hover:border-gray-500 hover:text-white"
        >
          Enter fullscreen
        </button>
      </section>

      {/* Chapter navigation shown only inside the mobile sheet */}
      <section className="flex gap-2 pb-2 md:hidden">
        <button onClick={onPrev} disabled={prevDisabled} className={`flex-1 py-2 ${navBtn}`}>
          <Icon as={ChevronLeft} size="small" className="mr-1 inline" /> Prev chapter
        </button>
        <button onClick={onNext} disabled={nextDisabled} className={`flex-1 py-2 ${navBtn}`}>
          Next chapter <Icon as={ChevronRight} size="small" className="ml-1 inline" />
        </button>
      </section>

    </div>
  )
}

// ── Shared sub-components ─────────────────────────────────────────────────────

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-500">{children}</p>
  )
}

function ToggleRow({ label, on, onToggle }: { label: string; on: boolean; onToggle: () => void }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-gray-300">{label}</span>
      <Toggle on={on} onToggle={onToggle} />
    </div>
  )
}
