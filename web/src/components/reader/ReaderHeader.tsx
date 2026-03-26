import { Link } from 'react-router-dom'

function fmtChapter(n: number) {
  return n % 1 === 0 ? n.toFixed(0) : String(n)
}

const navBtn =
  'rounded border border-gray-700 px-3 py-1 text-xs text-gray-300 transition hover:border-gray-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-30'

interface ReaderHeaderProps {
  visible: boolean
  lang: string
  chapter: number
  pageCount: number | null
  chaptersHref: string
  menuOpen: boolean
  onMenuToggle: () => void
  prevDisabled: boolean
  nextDisabled: boolean
  onPrev: () => void
  onNext: () => void
}

export function ReaderHeader({
  visible,
  lang,
  chapter,
  pageCount,
  chaptersHref,
  menuOpen,
  onMenuToggle,
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
        <Link to={chaptersHref} className="shrink-0 text-sm text-gray-400 transition hover:text-white">
          ← Chapters
        </Link>

        <div className="min-w-0 truncate text-center text-sm text-gray-300">
          <span className="uppercase">{lang}</span>
          {' — '}Ch.{fmtChapter(chapter)}
        </div>

        <div className="flex shrink-0 items-center gap-2">
          <div className="hidden items-center gap-2 md:flex">
            <button onClick={onPrev} disabled={prevDisabled} className={navBtn}>← Prev</button>
            <button onClick={onNext} disabled={nextDisabled} className={navBtn}>Next →</button>
          </div>

          <button
            onClick={onMenuToggle}
            aria-label="Reader settings"
            className={`${navBtn} ${menuOpen ? 'border-indigo-600 text-indigo-400' : ''}`}
          >
            <GearIcon />
          </button>
        </div>
      </div>
    </div>
  )
}

function GearIcon() {
  return (
    <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
      />
      <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
  )
}
