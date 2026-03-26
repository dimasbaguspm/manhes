import { Link } from 'react-router-dom'
import { useMangaDetail } from '../providers/MangaDetailProvider'
import { DEEP_LINKS } from '../lib/deepLinks'
import { formatDate, DateFormat } from '../lib/formatDate'

const STATUS_COLORS: Record<string, string> = {
  ongoing: 'bg-green-900 text-green-300',
  completed: 'bg-blue-900 text-blue-300',
  hiatus: 'bg-yellow-900 text-yellow-300',
}

export default function MangaPage() {
  const { data, loading, error, addState, addToWatchlist, refreshState, refreshManga } = useMangaDetail()

  if (loading) {
    return (
      <div className="animate-pulse">
        <div className="flex gap-6">
          <div className="h-64 w-44 flex-shrink-0 rounded-lg bg-gray-800" />
          <div className="flex-1 space-y-3">
            <div className="h-7 w-2/3 rounded bg-gray-800" />
            <div className="h-4 w-1/3 rounded bg-gray-800" />
            <div className="mt-4 h-24 rounded bg-gray-800" />
          </div>
        </div>
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
        <span className="text-gray-300">{data.title}</span>
      </div>

      <div className="flex gap-6">
        <div className="h-64 w-44 flex-shrink-0 overflow-hidden rounded-lg bg-gray-800">
          {data.coverUrl ? (
            <img src={data.coverUrl} alt={data.title} className="h-full w-full object-cover" />
          ) : (
            <div className="flex h-full flex-col items-center justify-center gap-2 px-3 text-center">
              <svg className="h-10 w-10 text-gray-700" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1}
                  d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
                />
              </svg>
              <p className="text-xs leading-snug text-gray-600">
                Cover on the way — please be patient
              </p>
            </div>
          )}
        </div>

        <div className="min-w-0 flex-1">
          <h1 className="text-2xl font-bold text-gray-100">{data.title}</h1>
          {data.authors.length > 0 && (
            <p className="mt-1 text-sm text-gray-400">{data.authors.join(', ')}</p>
          )}

          <div className="mt-3 flex flex-wrap gap-2">
            <span className={`rounded-full px-2.5 py-1 text-xs font-medium ${STATUS_COLORS[data.status] ?? 'bg-gray-800 text-gray-400'}`}>
              {data.status}
            </span>
            {data.genres.slice(0, 6).map(g => (
              <span key={g} className="rounded-full bg-gray-800 px-2.5 py-1 text-xs text-gray-400">
                {g}
              </span>
            ))}
          </div>

          {data.description && (
            <p className="mt-3 line-clamp-4 text-sm leading-relaxed text-gray-400">
              {data.description}
            </p>
          )}

          {data.state === 'unavailable' && (
            <button
              onClick={addToWatchlist}
              disabled={addState !== 'idle'}
              className={`mt-4 rounded-lg px-4 py-2 text-sm font-medium transition ${
                addState === 'error' ? 'cursor-default bg-red-900 text-red-300'
                : addState === 'loading' ? 'cursor-wait bg-gray-700 text-gray-400'
                : 'bg-indigo-600 text-white hover:bg-indigo-500'
              }`}
            >
              {addState === 'error' ? 'Failed to add'
                : addState === 'loading' ? 'Adding…'
                : '+ Add to Library'}
            </button>
          )}

          {data.state !== 'unavailable' && (
            <div className="mt-4 flex items-center gap-3">
              <button
                onClick={refreshManga}
                disabled={refreshState === 'loading' || refreshState === 'done'}
                className={`rounded-lg border px-4 py-2 text-sm font-medium transition disabled:cursor-not-allowed disabled:opacity-50 ${
                  refreshState === 'error' ? 'border-red-700 bg-red-900 text-red-300'
                  : refreshState === 'done' ? 'border-gray-700 bg-gray-800 text-gray-400'
                  : 'border-gray-700 bg-gray-800 text-gray-300 hover:border-gray-500 hover:text-white'
                }`}
              >
                {refreshState === 'loading' ? 'Refreshing…'
                  : refreshState === 'error' ? 'Refresh failed'
                  : refreshState === 'done' ? 'Queued'
                  : 'Refresh'}
              </button>
              {data.updatedAt && (
                <span className="text-xs text-gray-500">
                  Updated {formatDate(data.updatedAt, DateFormat.ShortDateTime)}
                </span>
              )}
            </div>
          )}
        </div>
      </div>

      {data.state !== 'unavailable' && (
        <div className="mt-8">
          <h2 className="mb-3 text-sm font-semibold uppercase tracking-wider text-gray-500">
            Available Languages
          </h2>
          {data.languages.length > 0 ? (
            <div className="flex flex-wrap gap-3">
              {data.languages.map(({ lang, latestUpdate, totalChapters, fetchedChapters, uploadedChapters }) => (
                <Link
                  key={lang}
                  to={DEEP_LINKS.MANGA_CHAPTERS({ mangaId: data.id, lang })}
                  className="flex min-w-48 flex-col gap-2 rounded-lg border border-gray-700 bg-gray-900 px-4 py-3 transition hover:border-gray-500"
                >
                  <div className="flex items-center justify-between gap-4">
                    <span className="font-medium uppercase text-gray-100">{lang}</span>
                    <span className="text-xs text-gray-500">
                      {latestUpdate ? new Date(latestUpdate).toLocaleDateString() : '—'}
                    </span>
                  </div>
                  {totalChapters > 0 && (
                    <div className="space-y-1">
                      <div className="h-1.5 w-full overflow-hidden rounded-full bg-gray-800">
                        <div className="flex h-full">
                          <div
                            className="h-full bg-indigo-500 transition-all"
                            style={{ width: `${Math.round((uploadedChapters / totalChapters) * 100)}%` }}
                          />
                          <div
                            className="h-full bg-yellow-700 transition-all"
                            style={{ width: `${Math.round(((fetchedChapters - uploadedChapters) / totalChapters) * 100)}%` }}
                          />
                        </div>
                      </div>
                      <div className="flex justify-between text-xs text-gray-500">
                        <span>
                          <span className="text-indigo-400">{uploadedChapters}</span>
                          {fetchedChapters > uploadedChapters && (
                            <span className="text-yellow-600"> +{fetchedChapters - uploadedChapters} fetching</span>
                          )}
                          <span> / {totalChapters}</span>
                        </span>
                      </div>
                    </div>
                  )}
                </Link>
              ))}
            </div>
          ) : (
            <div className="rounded-lg border border-gray-800 bg-gray-900 px-4 py-6 text-center text-gray-500">
              {data.state === 'fetching'
                ? 'Chapters are being fetched — please check back soon.'
                : 'No chapters available yet.'}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
