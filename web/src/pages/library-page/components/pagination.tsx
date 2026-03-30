import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Icon } from '@/components'

interface Props {
  page: number
  total: number
  onChange: (page: number) => void
}

export default function Pagination({ page, total, onChange }: Props) {
  if (total <= 1) return null

  return (
    <div className="flex items-center justify-center gap-3">
      <button
        onClick={() => onChange(page - 1)}
        disabled={page <= 1}
        className="rounded-lg border border-gray-700 bg-gray-900 px-4 py-1.5 text-sm text-gray-300 transition hover:border-gray-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-40"
      >
        <Icon as={ChevronLeft} size="small" className="mr-1 inline" /> Prev
      </button>
      <span className="text-sm text-gray-500">
        {page} / {total}
      </span>
      <button
        onClick={() => onChange(page + 1)}
        disabled={page >= total}
        className="rounded-lg border border-gray-700 bg-gray-900 px-4 py-1.5 text-sm text-gray-300 transition hover:border-gray-500 hover:text-white disabled:cursor-not-allowed disabled:opacity-40"
      >
        Next <Icon as={ChevronRight} size="small" className="ml-1 inline" />
      </button>
    </div>
  )
}
