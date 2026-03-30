import { Link } from 'react-router-dom'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Icon } from '@/components'

interface ChapterNavFooterProps {
  chaptersHref: string
  prevDisabled: boolean
  nextDisabled: boolean
  onPrev: () => void
  onNext: () => void
}

const navBtn =
  'inline-flex items-center gap-1.5 rounded-lg border border-gray-700 bg-gray-900 py-2.5 text-sm text-gray-300 transition hover:border-gray-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-30'

export function ChapterNavFooter({
  chaptersHref,
  prevDisabled,
  nextDisabled,
  onPrev,
  onNext,
}: ChapterNavFooterProps) {
  return (
    <div className="my-10 flex items-center justify-center gap-2 px-4 sm:gap-4">
      <button
        onClick={onPrev}
        disabled={prevDisabled}
        className={`${navBtn} px-3 sm:px-5`}
      >
        <Icon as={ChevronLeft} size="small" className="shrink-0" />
        <span className="hidden sm:inline">Previous</span>
        <span className="sm:hidden">Prev</span>
      </button>

      <Link to={chaptersHref} className={`${navBtn} px-3 sm:px-5`}>
        <span className="hidden sm:inline">Chapter List</span>
        <span className="sm:hidden">List</span>
      </Link>

      <button
        onClick={onNext}
        disabled={nextDisabled}
        className={`${navBtn} px-3 sm:px-5`}
      >
        <span className="hidden sm:inline">Next</span>
        <span className="sm:hidden">Next</span>
        <Icon as={ChevronRight} size="small" className="shrink-0" />
      </button>
    </div>
  )
}
