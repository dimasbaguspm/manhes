import { Link } from 'react-router-dom'
import { DEEP_LINKS } from '@/lib/deep-links'
import { formatDate, DateFormat } from '@/lib/format-date'
import { Bookmark } from 'lucide-react'
import { Icon } from '@/components'
import { Badge, Button } from '@/components/ui'
import { NoResults } from '@/components/ui'
import type { DomainChapterListResponse } from '@/types'

interface ChaptersPanelProps {
  chapterData: DomainChapterListResponse | null
  chapterLoading: boolean
  chapterError: string | null
  bookmarkedSet: Set<string>
  latestRead: string | undefined
  getProgress: (chapter: string) => number | undefined
  onToggleBookmark: (chapter: string) => void
  onChapterClick: (chapterId: string, chapterName: string) => void
}

export function ChaptersPanel({
  chapterData,
  chapterLoading,
  chapterError,
  bookmarkedSet,
  latestRead,
  getProgress,
  onToggleBookmark,
  onChapterClick,
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

  return (
    <div className="space-y-1.5">
      {chapters.map(ch => {
        const chName = String(ch.name ?? ch.order ?? '0')
        const chId = ch.id ?? ''
        const isBookmarked = bookmarkedSet.has(chId)
        const isLatest = latestRead === chId
        const readPct = getProgress(chId)
        const hasProgress = readPct !== undefined && readPct > 0

        return (
          <div key={chId} className="flex items-center gap-2">
            <Link
              to={DEEP_LINKS.MANGA_READER({ chapterId: chId })}
              onClick={() => onChapterClick(chId, chName)}
              className={`relative flex min-w-0 flex-1 items-center justify-between overflow-hidden rounded-lg border bg-gray-900 px-4 py-3 transition hover:border-gray-600 hover:bg-gray-800 ${
                isLatest ? 'border-indigo-700' : 'border-gray-800'
              }`}
            >
              {hasProgress && (
                <div
                  className="pointer-events-none absolute bottom-0 left-0 h-0.5 bg-indigo-500/50"
                  style={{ width: `${readPct}%` }}
                />
              )}

              <div className="flex min-w-0 items-center gap-2">
                <span className={`font-medium ${isLatest ? 'text-indigo-300' : 'text-gray-100'}`}>
                  {chName}
                </span>
                {isLatest && (
                  <Badge variant="primary" size="sm">
                    Last read
                  </Badge>
                )}
              </div>

              <div className="ml-4 flex shrink-0 items-center gap-3 text-xs text-gray-500">
                {hasProgress && (
                  <span className={readPct! >= 100 ? 'text-emerald-500' : 'text-indigo-400'}>
                    {readPct! >= 100 ? 'Done' : `${readPct}%`}
                  </span>
                )}
                <span>{ch.page_count ?? 0} pgs</span>
                <span className="hidden sm:inline">
                  {ch.updated_at ? formatDate(ch.updated_at, DateFormat.ShortDate) : '—'}
                </span>
              </div>
            </Link>

            <Button
              variant={isBookmarked ? 'danger' : 'ghost'}
              color={isBookmarked ? 'danger' : 'muted'}
              size="sm"
              onClick={() => onToggleBookmark(chId)}
              aria-label={isBookmarked ? 'Remove bookmark' : 'Bookmark chapter'}
            >
              <Icon as={Bookmark} size="small" fill={isBookmarked ? 'currentColor' : 'none'} />
            </Button>
          </div>
        )
      })}
    </div>
  )
}
