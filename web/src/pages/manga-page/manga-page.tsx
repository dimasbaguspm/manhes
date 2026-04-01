import { Link, useSearchParams, useParams } from 'react-router-dom'
import { useApiMangaDetail } from '@/hooks/use-api-manga-detail'
import { useApiChapterList } from '@/hooks/use-api-chapter-list'
import { useApiRefreshDictionary } from '@/hooks/use-api-refresh-dictionary'
import { useApiTrackers } from '@/hooks/use-api-tracker'
import { mangaApi } from '@/api/manga'
import { DEEP_LINKS } from '@/lib/deep-links'
import { CoverImage, MangaDetailHeader, StatusBadge, StateBadge, GenreBadge, LangTabs, LangProgressBar, ChaptersPanel, ActionButtons } from '@/pages/manga-page'
import { NoResults } from '@/components/ui'
import type { DomainTrackerResponse } from '@/types'

export default function MangaPage() {
  const { mangaId } = useParams<{ mangaId: string }>()
  const [searchParams, setSearchParams] = useSearchParams()

  const { data, loading, error } = useApiMangaDetail(mangaId)
  const resolvedLang = searchParams.get('tab') ?? (data?.languages?.[0]?.lang ?? null)
  const { data: chapterData, loading: chapterLoading, error: chapterError } = useApiChapterList(mangaId, resolvedLang ?? undefined)
  const { data: trackers } = useApiTrackers(mangaId)
  const { state: refreshState, refresh } = useApiRefreshDictionary()

  // Derive latest read chapter from trackers (most recent updated_at)
  const latestReadFromTracker = (() => {
    if (!trackers || trackers.length === 0) return undefined
    let latest: DomainTrackerResponse | null = null
    for (const t of trackers) {
      if (!latest || (t.updated_at && latest.updated_at && t.updated_at > latest.updated_at)) {
        latest = t
      }
    }
    return latest?.chapter_id
  })()

  async function upsertTracker(chapterId: string, isRead: boolean) {
    if (!mangaId) return
    try {
      await mangaApi.upsertTracker({
        manga_id: mangaId,
        chapter_id: chapterId,
        is_read: isRead,
      })
    } catch (err) {
      console.error('Failed to upsert tracker:', err)
    }
  }

  function selectLang(lang: string) {
    setSearchParams({ tab: lang }, { replace: true })
  }

  function handleRefresh() {
    if (data?.dictionary_id) {
      refresh(data.dictionary_id)
    }
  }

  if (loading) {
    return (
      <div className="animate-pulse">
        <div className="flex flex-col gap-4 sm:flex-row sm:gap-6">
          <div className="mx-auto h-56 w-36 flex-shrink-0 rounded-lg bg-gray-800 sm:mx-0 sm:h-64 sm:w-44" />
          <div className="flex-1 space-y-3">
            <div className="h-7 w-2/3 rounded bg-gray-800" />
            <div className="h-4 w-1/3 rounded bg-gray-800" />
            <div className="mt-4 h-24 rounded bg-gray-800" />
          </div>
        </div>
      </div>
    )
  }

  if (error) return <NoResults message={error} variant="error" />
  if (!data) return null

  const activeLangInfo = resolvedLang ? data.languages?.find(l => l.lang === resolvedLang) : null

  return (
    <div>
      <div className="mb-4 text-sm text-gray-500">
        <Link to={DEEP_LINKS.LIBRARY()} className="hover:text-gray-300">Library</Link>
        <span className="mx-2">/</span>
        <span className="text-gray-300">{data.title}</span>
      </div>

      <div className="flex flex-col items-center gap-4 sm:flex-row sm:items-start sm:gap-6">
        <CoverImage src={data.cover_url ?? ''} alt={data.title ?? ''} />

        <div className="w-full min-w-0 flex-1 text-center sm:text-left">
          <MangaDetailHeader
            title={data.title ?? ''}
            authors={data.authors ?? []}
          />

          <div className="mt-3 flex flex-wrap justify-center gap-2 sm:justify-start">
            <StatusBadge status={data.status ?? ''} />
            <StateBadge state={data.state ?? ''} />
            {data.genres?.slice(0, 5).map(g => (
              <GenreBadge key={g} genre={g} />
            ))}
          </div>

          {data.description && (
            <p className="mt-3 line-clamp-4 text-sm leading-relaxed text-gray-400">
              {data.description}
            </p>
          )}

          <ActionButtons
            state={data.state ?? ''}
            refreshState={refreshState}
            onRefresh={handleRefresh}
            updatedAt={data.updated_at ?? ''}
          />
        </div>
      </div>

      {/* Language tabs + chapters */}
      {data.state !== 'unavailable' && (
        <div className="mt-8">
          {(!data.languages || data.languages.length === 0) ? (
            <NoResults
              message={
                data.state === 'fetching'
                  ? 'Chapters are being fetched — please check back soon.'
                  : 'No chapters available yet.'
              }
            />
          ) : (
            <>
              <LangTabs
                langs={data.languages}
                activeLang={resolvedLang}
                onSelect={selectLang}
              />

              {activeLangInfo && (
                <LangProgressBar
                  totalChapters={activeLangInfo.total_chapters ?? 0}
                  availableChapters={activeLangInfo.available_chapters ?? 0}
                />
              )}

              {resolvedLang && (
                <ChaptersPanel
                  chapterData={chapterData}
                  chapterLoading={chapterLoading}
                  chapterError={chapterError}
                  latestRead={latestReadFromTracker}
                  trackers={trackers ?? undefined}
                  onUpsertTracker={upsertTracker}
                />
              )}
            </>
          )}
        </div>
      )}
    </div>
  )
}
