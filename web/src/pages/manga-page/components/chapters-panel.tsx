import { Badge, NoResults } from '@/components/ui'
import { DEEP_LINKS } from '@/lib/deep-links'
import { DateFormat, formatDate } from '@/lib/format-date'
import type { DomainChapterListResponse, DomainTrackerResponse } from '@/types'
import { Check } from 'lucide-react'
import { Link } from 'react-router-dom'

interface ChaptersPanelProps {
  chapterData: DomainChapterListResponse | null
  chapterLoading: boolean
  chapterError: string | null
  latestRead: string | undefined
  trackers?: DomainTrackerResponse[]
  onUpsertTracker?: (chapterId: string, isRead: boolean) => void
}

export function ChaptersPanel({
  chapterData,
  chapterLoading,
  chapterError,
  latestRead,
  trackers,
  onUpsertTracker,
}: ChaptersPanelProps) {
  if (chapterLoading) {
    return (
      <div className="space-y-1.5">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="h-14 animate-pulse rounded-lg bg-gray-900" />
        ))}
      </div>
    )
  }

  if (chapterError) return <NoResults message={chapterError} variant="error" />

  const chapters = chapterData?.chapters ?? []
  if (chapters.length === 0) {
    return <NoResults message="No chapters available yet." />
  }

  // Build tracker lookup by chapterId
  const trackerByChapter = new Map<string, DomainTrackerResponse>()
  trackers?.forEach(t => {
    if (t.chapter_id) trackerByChapter.set(t.chapter_id, t)
  })

  return (
    <div className="space-y-1.5">
      {chapters.map(ch => {
        const chName = String(ch.name ?? ch.order ?? '0')
        const chId = ch.id ?? ''
        const isLatest = latestRead === chId
        const tracker = trackerByChapter.get(chId)
        const isRead = tracker?.is_read ?? false
        const lastReadAt = tracker?.updated_at

        return (
          <Link
            key={chId}
            to={DEEP_LINKS.MANGA_READER({ chapterId: chId })}
            onClick={() => {
              if (onUpsertTracker) onUpsertTracker(chId, true)
            }}
            className={`flex items-center justify-between overflow-hidden rounded-lg border bg-gray-900 px-4 py-3 transition hover:border-gray-600 hover:bg-gray-800 ${
              isLatest ? 'border-indigo-700' : 'border-gray-800'
            }`}
          >
            <div className="flex min-w-0 items-center gap-2">
              {isRead && (
                <span className="text-emerald-500 shrink-0">
                  <Check size={14} />
                </span>
              )}
              <span className={`font-medium ${isLatest ? 'text-indigo-300' : isRead ? 'text-gray-400' : 'text-gray-100'}`}>
                {chName}
              </span>
              {isLatest && (
                <Badge variant="primary" size="sm">
                  Last read
                </Badge>
              )}
            </div>

            <div className="flex shrink-0 items-center gap-3 text-xs text-gray-500">
                <span className="hidden sm:inline text-gray-600">
                  {ch.updated_at ? formatDate(ch.updated_at, DateFormat.ShortDate) : '—'}
                </span>
              <span>{ch.page_count ?? 0} pgs</span>
            </div>
          </Link>
        )
      })}
    </div>
  )
}
