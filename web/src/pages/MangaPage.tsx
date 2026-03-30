import { Link, useSearchParams, useParams } from 'react-router-dom'
import { useApiMangaDetail } from '../hooks/useApiMangaDetail'
import { useApiChapterList } from '../hooks/useApiChapterList'
import { useApiRefreshDictionary } from '../hooks/useApiRefreshDictionary'
import { DEEP_LINKS } from '../lib/deepLinks'
import { usePersistedState } from '../hooks/usePersistedState'
import { CoverImage } from '../components/manga-detail/CoverImage'
import { MangaDetailHeader } from '../components/manga-detail/Header'
import { StatusBadge, StateBadge, GenreBadge } from '../components/manga-detail/Badge'
import { LangTabs } from '../components/manga-detail/LangTabs'
import { LangProgressBar } from '../components/manga-detail/ProgressBar'
import { ChaptersPanel } from '../components/manga-detail/ChaptersPanel'
import { NoResults } from '../components/manga-detail/NoResults'
import { ActionButtons } from '../components/manga-detail/ActionButtons'

export default function MangaPage() {
  const { mangaId } = useParams<{ mangaId: string }>()
  const [searchParams, setSearchParams] = useSearchParams()
  const [favorites, setFavorites] = usePersistedState<Record<string, true>>({
    key: 'manhes_favorites',
    fallback: {},
  })
  const [bookmarks, setBookmarks] = usePersistedState<Record<string, true>>({
    key: 'manhes_bookmarks',
    fallback: {},
  })
  const [latestRead, setLatestRead] = usePersistedState<Record<string, string>>({
    key: 'manhes_latest_read',
    fallback: {},
  })
  const [readProgress] = usePersistedState<Record<string, number>>({
    key: 'manhes_read_progress',
    fallback: {},
  })

  const { data, loading, error } = useApiMangaDetail(mangaId)
  const resolvedLang = searchParams.get('tab') ?? (data?.languages?.[0]?.lang ?? null)
  const { data: chapterData, loading: chapterLoading, error: chapterError } = useApiChapterList(mangaId, resolvedLang ?? undefined)
  const { state: refreshState, refresh } = useApiRefreshDictionary()

  const isFavorite = mangaId ? !!favorites[mangaId] : false

  function toggleFavorite() {
    if (!mangaId) return
    setFavorites(prev => {
      const { [mangaId]: _, ...rest } = prev
      return _ !== undefined ? (rest as Record<string, true>) : { ...prev, [mangaId]: true }
    })
  }

  function toggleBookmark(chapterId: string) {
    if (!mangaId || !resolvedLang) return
    const key = `${mangaId}/${resolvedLang}/${chapterId}`
    setBookmarks(prev => {
      const { [key]: _, ...rest } = prev
      return _ !== undefined ? (rest as Record<string, true>) : { ...prev, [key]: true }
    })
  }

  function markRead(chapterId: string) {
    if (!mangaId || !resolvedLang) return
    setLatestRead(prev => ({ ...prev, [`${mangaId}/${resolvedLang}`]: chapterId }))
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

  if (error) return <NoResults message={error} error />
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
            isFavorite={isFavorite}
            onToggleFavorite={toggleFavorite}
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
                  bookmarkedSet={new Set(
                    Object.keys(bookmarks)
                      .filter(k => k.startsWith(`${mangaId}/${resolvedLang}/`))
                      .map(k => k.split('/')[2]),
                  )}
                  latestRead={latestRead[`${mangaId}/${resolvedLang}`]}
                  getProgress={chapter => readProgress[`${mangaId}/${resolvedLang}/${chapter}`]}
                  onToggleBookmark={toggleBookmark}
                  onChapterClick={(chapterId) => markRead(chapterId)}
                />
              )}
            </>
          )}
        </div>
      )}
    </div>
  )
}
