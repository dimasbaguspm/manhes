import { useParams, Link } from 'react-router-dom'
import { useMangaChapters } from '../providers/MangaChaptersProvider'
import { DEEP_LINKS } from '../lib/deepLinks'

function fmtChapter(n: number) {
  return n % 1 === 0 ? n.toFixed(0) : String(n)
}

export default function ChaptersPage() {
  const { mangaId, lang } = useParams<{ mangaId: string; lang: string }>()
  const { data, mangaTitle, langStats, loading, error } = useMangaChapters()

  if (loading) {
    return (
      <div className="space-y-2">
        {Array.from({ length: 8 }).map((_, i) => (
          <div key={i} className="h-14 animate-pulse rounded-lg bg-gray-900" />
        ))}
      </div>
    )
  }

  if (error) {
    return (
      <div className="rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
        {error}
      </div>
    )
  }

  if (!data) return null

  return (
    <div>
      <div className="mb-4 text-sm text-gray-500">
        <Link to={DEEP_LINKS.LIBRARY()} className="hover:text-gray-300">Library</Link>
        <span className="mx-2">/</span>
        <Link to={DEEP_LINKS.MANGA_DETAIL({ mangaId: mangaId! })} className="hover:text-gray-300">
          {mangaTitle || '…'}
        </Link>
        <span className="mx-2">/</span>
        <span className="uppercase text-gray-300">{lang}</span>
      </div>

      <div className="mb-4 space-y-2">
        <h1 className="text-lg font-semibold uppercase text-gray-100">
          {lang} — {data.chapters.length} chapters
        </h1>
        {langStats && langStats.totalChapters > 0 && (
          <div className="space-y-1">
            <div className="h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
              <div className="flex h-full">
                <div
                  className="h-full bg-indigo-500 transition-all"
                  style={{ width: `${Math.round((langStats.uploadedChapters / langStats.totalChapters) * 100)}%` }}
                />
                <div
                  className="h-full bg-yellow-700 transition-all"
                  style={{ width: `${Math.round(((langStats.fetchedChapters - langStats.uploadedChapters) / langStats.totalChapters) * 100)}%` }}
                />
              </div>
            </div>
            <p className="text-xs text-gray-500">
              <span className="text-indigo-400">{langStats.uploadedChapters}</span>
              {langStats.fetchedChapters > langStats.uploadedChapters && (
                <span className="text-yellow-600"> +{langStats.fetchedChapters - langStats.uploadedChapters} fetching</span>
              )}
              <span> / {langStats.totalChapters} chapters from source</span>
            </p>
          </div>
        )}
      </div>

      <div className="space-y-1.5">
        {[...data.chapters].map(ch => (
          <Link
            key={ch.chapter}
            to={DEEP_LINKS.MANGA_READER({ mangaId: mangaId!, lang: lang!, chapter: ch.chapter })}
            className="flex items-center justify-between rounded-lg border border-gray-800 bg-gray-900 px-4 py-3 transition hover:border-gray-600 hover:bg-gray-800"
          >
            <span className="font-medium text-gray-100">
              Chapter {ch.chapter}
            </span>
            <div className="flex items-center gap-4 text-xs text-gray-500">
              <span>{ch.pageCount} pages</span>
              <span>{ch.uploadedAt ? new Date(ch.uploadedAt).toLocaleDateString() : '—'}</span>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}
