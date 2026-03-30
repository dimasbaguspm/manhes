import { Link } from 'react-router-dom'
import { ChevronLeft, ChevronRight, Keyboard, Settings } from 'lucide-react'
import { Icon } from '@/components'

const navBtn =
  'inline-flex items-center gap-1 rounded border border-gray-700 px-3 py-1 text-xs text-gray-300 transition hover:border-gray-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-30'

interface ReaderHeaderProps {
  visible: boolean
  chapter: string
  pageCount: number | null
  chaptersHref: string
  menuOpen: boolean
  onMenuToggle: () => void
  onShortcutsToggle: () => void
  prevDisabled: boolean
  nextDisabled: boolean
  onPrev: () => void
  onNext: () => void
}

export function ReaderHeader({
  visible,
  chapter,
  pageCount,
  chaptersHref,
  menuOpen,
  onMenuToggle,
  onShortcutsToggle,
  prevDisabled,
  nextDisabled,
  onPrev,
  onNext,
}: ReaderHeaderProps) {
  return (
    <div
      className={`sticky top-0 z-20 overflow-hidden transition-[max-height,opacity] duration-300 ease-in-out ${
        visible ? 'max-h-20 opacity-100' : 'max-h-0 opacity-0 pointer-events-none'
      }`}
    >
      <div className="flex items-center justify-between gap-2 border-b border-gray-800 bg-gray-950/95 px-4 py-3 backdrop-blur">
        <Link to={chaptersHref} className="inline-flex shrink-0 items-center gap-1 text-sm text-gray-400 transition hover:text-white">
          <Icon as={ChevronLeft} size="small" />
          <span>Chapters</span>
        </Link>

        <div className="min-w-0 truncate text-center text-sm text-gray-300">
          {'Ch. '}{chapter}
          {pageCount !== null && (
            <span className="text-gray-500"> · {pageCount} pgs</span>
          )}
        </div>

        <div className="flex shrink-0 items-center gap-2">
          <div className="hidden items-center gap-2 md:flex">
            <button onClick={onPrev} disabled={prevDisabled} className={navBtn}>
              <Icon as={ChevronLeft} size="small" /> Prev
            </button>
            <button onClick={onNext} disabled={nextDisabled} className={navBtn}>
              Next <Icon as={ChevronRight} size="small" />
            </button>
          </div>

          <button
            onClick={onShortcutsToggle}
            aria-label="Keyboard shortcuts"
            title="Keyboard shortcuts (/)"
            className={navBtn}
          >
            <Icon as={Keyboard} />
          </button>

          <button
            onClick={onMenuToggle}
            aria-label="Reader settings"
            className={`${navBtn} ${menuOpen ? 'border-indigo-600 text-indigo-400' : ''}`}
          >
            <Icon as={Settings} />
          </button>
        </div>
      </div>
    </div>
  )
}
