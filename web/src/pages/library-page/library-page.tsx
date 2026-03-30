import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useApiMangaList } from '../../hooks/useApiMangaList'
import type { ListMangaParams } from '../../api/manga'
import Pagination from './components/Pagination'
import { DEEP_LINKS } from '../../lib/deepLinks'
import { DateFormat, formatDate } from '../../lib/formatDate'
import type { DomainMangaSummary } from '../../types'

const STATE_LABEL: Record<string, string> = {
  available: 'Available',
  fetching: 'Fetching',
  uploading: 'Uploading',
}

const STATE_COLOR: Record<string, string> = {
  available: 'bg-emerald-900 text-emerald-300',
  fetching: 'bg-amber-900 text-amber-300',
  uploading: 'bg-blue-900 text-blue-300',
}

function LibraryItem({ manga }: { manga: DomainMangaSummary }) {
  return (
    <div className="flex gap-4 rounded-lg border border-gray-800 bg-gray-900 p-4 transition hover:border-gray-700">
      <div className="aspect-[2/3] w-20 flex-shrink-0 overflow-hidden rounded-md bg-gray-800">
        {manga.cover_url ? (
          <img
            src={manga.cover_url}
            alt={manga.title}
            className="h-full w-full object-cover"
            loading="lazy"
          />
        ) : (
          <div className="flex h-full items-center justify-center text-gray-600">
            <svg className="h-7 w-7" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1}
                d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
              />
            </svg>
          </div>
        )}
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-3">
          <h3 className="font-semibold text-gray-100">{manga.title}</h3>
          <span className={`flex-shrink-0 rounded px-2 py-0.5 text-xs font-medium ${STATE_COLOR[manga.state ?? ''] ?? 'bg-gray-800 text-gray-400'}`}>
            {STATE_LABEL[manga.state ?? ''] ?? manga.state}
          </span>
        </div>

        {manga.description && (
          <p className="mt-1 line-clamp-2 text-sm leading-relaxed text-gray-400">
            {manga.description}
          </p>
        )}

        {manga.languages && manga.languages.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-xs text-gray-500">
            {manga.languages.map(l => (
              <span key={l.lang ?? l.lang}>{l.lang?.toUpperCase() ?? l.lang}</span>
            ))}
          </div>
        )}

        {manga.genres && manga.genres.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {manga.genres.slice(0, 5).map(g => (
              <span key={g} className="rounded bg-gray-800 px-1.5 py-0.5 text-xs text-gray-500">
                {g}
              </span>
            ))}
          </div>
        )}

        <div className="mt-3 flex items-center gap-3">
          <Link
            to={DEEP_LINKS.MANGA_DETAIL({ mangaId: manga.id ?? '' })}
            className="rounded-lg bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-indigo-500"
          >
            View
          </Link>
          {manga.updated_at && (
            <span className="text-xs text-gray-600">Updated {formatDate(manga.updated_at, DateFormat.ShortDateTime)}</span>
          )}
        </div>
      </div>
    </div>
  )
}

const DEFAULT_FILTERS: ListMangaParams = {
  sortBy: 'title',
  page: 1,
  pageSize: 20,
}

export default function LibraryPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [filters, setFilters] = useState<ListMangaParams>(() => {
    const q = searchParams.get('q') ?? ''
    const page = parseInt(searchParams.get('page') ?? '1', 10)
    return { ...DEFAULT_FILTERS, q, page: isNaN(page) ? 1 : page }
  })

  const { data, loading, error } = useApiMangaList(filters)

  function setFilter(key: keyof ListMangaParams, value: string | number) {
    setFilters(f => {
      const next = { ...f, [key]: value }
      if (key !== 'page') next.page = 1
      return next
    })
    setSearchParams(prev => {
      const next = new URLSearchParams(prev)
      if (key === 'page') {
        next.set('page', String(value))
      } else {
        if (value) next.set(key, String(value))
        else next.delete(key)
        next.set('page', '1')
      }
      return next
    }, { replace: true })
  }

  return (
    <div>
      <div className="mb-6 flex flex-wrap gap-3">
        <input
          type="text"
          placeholder="Search title..."
          value={filters.q ?? ''}
          onChange={e => setFilter('q', e.target.value)}
          className="min-w-[200px] flex-1 rounded-lg border border-gray-700 bg-gray-900 px-3 py-2 text-sm text-gray-100 placeholder-gray-500 focus:border-indigo-500 focus:outline-none"
        />
      </div>

      {data && (
        <p className="mb-4 text-sm text-gray-500">{data.itemCount} manga</p>
      )}

      {loading && (
        <div className="space-y-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="flex gap-4 rounded-lg border border-gray-800 bg-gray-900 p-4">
              <div className="aspect-[2/3] w-20 flex-shrink-0 animate-pulse rounded-md bg-gray-800" />
              <div className="flex-1 space-y-2 py-1">
                <div className="h-4 w-2/3 animate-pulse rounded bg-gray-800" />
                <div className="h-3 animate-pulse rounded bg-gray-800" />
                <div className="h-3 w-5/6 animate-pulse rounded bg-gray-800" />
              </div>
            </div>
          ))}
        </div>
      )}

      {error && (
        <div className="rounded-lg border border-red-800 bg-red-950 px-4 py-3 text-sm text-red-300">
          {error}
        </div>
      )}

      {data && data.items && data.items.length === 0 && (
        <div className="py-20 text-center text-gray-500">
          No manga found. Try adjusting your filters or{' '}
          <Link to={DEEP_LINKS.DISCOVER()} className="text-indigo-400 hover:underline">discover new ones</Link>.
        </div>
      )}

      {data && data.items && data.items.length > 0 && (
        <>
          <div className="space-y-3">
            {data.items.map(manga => (
              <LibraryItem key={manga.id} manga={manga} />
            ))}
          </div>
          <div className="mt-8">
            <Pagination
              page={data.pageNumber ?? 1}
              total={data.pageTotal ?? 1}
              onChange={p => setFilter('page', p)}
            />
          </div>
        </>
      )}
    </div>
  )
}
