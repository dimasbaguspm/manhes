import { Link } from 'react-router-dom'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Icon } from '@/components'
import { Button } from '@/components/ui'

interface ChapterNavFooterProps {
  chaptersHref: string
  prevDisabled: boolean
  nextDisabled: boolean
  onPrev: () => void
  onNext: () => void
}

export function ChapterNavFooter({
  chaptersHref,
  prevDisabled,
  nextDisabled,
  onPrev,
  onNext,
}: ChapterNavFooterProps) {
  return (
    <div className="my-10 flex items-center justify-center gap-2 px-4 sm:gap-4">
      <Button
        variant="outline"
        color="muted"
        onClick={onPrev}
        disabled={prevDisabled}
        className="px-3 sm:px-5"
      >
        <Icon as={ChevronLeft} size="small" className="shrink-0" />
        <span className="hidden sm:inline">Previous</span>
        <span className="sm:hidden">Prev</span>
      </Button>

      <Link to={chaptersHref} className="inline-flex items-center gap-1.5 rounded-lg border border-gray-700 bg-gray-900 py-2.5 px-3 text-sm text-gray-300 transition hover:border-gray-500 hover:text-white sm:px-5">
        <span className="hidden sm:inline">Chapter List</span>
        <span className="sm:hidden">List</span>
      </Link>

      <Button
        variant="outline"
        color="muted"
        onClick={onNext}
        disabled={nextDisabled}
        className="px-3 sm:px-5"
      >
        <span className="hidden sm:inline">Next</span>
        <span className="sm:hidden">Next</span>
        <Icon as={ChevronRight} size="small" className="shrink-0" />
      </Button>
    </div>
  )
}
