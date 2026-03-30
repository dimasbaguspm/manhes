import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { BookOpen, ChevronLeft, ChevronRight } from 'lucide-react'
import { Icon } from '@/components/icon'
import { Badge } from '@/components/ui/badge/badge'
import { Button } from '@/components/ui/button/button'
import { Text } from '@/components/ui/text/text'
import { NoResults } from '@/components/ui/no-results/no-results'
import { useApiMangaList } from '@/hooks/use-api-manga-list'
import type { ListMangaParams } from '@/api/manga'
import { DEEP_LINKS } from '@/lib/deep-links'
import { DateFormat, formatDate } from '@/lib/format-date'
import type { DomainMangaSummary } from '@/types'

const STATE_LABEL: Record<string, string> = {
  available: 'Available',
  fetching: 'Fetching',
  uploading: 'Uploading',
}

const STATE_BADGE_VARIANT: Record<string, 'success' | 'warning' | 'primary'> = {
  available: 'success',
  fetching: 'warning',
  uploading: 'primary',
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
            <Icon as={BookOpen} size="large" />
          </div>
        )}
      </div>

      <div className="min-w-0 flex-1">
        <div className="flex items-start justify-between gap-3">
          <h3 className="font-semibold text-gray-100">{manga.title}</h3>
          <Badge
            variant={STATE_BADGE_VARIANT[manga.state ?? ''] ?? 'default'}
            size="sm"
          >
            {STATE_LABEL[manga.state ?? ''] ?? manga.state}
          </Badge>
        </div>

        {manga.description && (
          <Text size="sm" color="muted" className="mt-1 line-clamp-2">
            {manga.description}
          </Text>
        )}

        {manga.languages && manga.languages.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1">
            {manga.languages.map(l => (
              <Text key={l.lang ?? l.lang} size="xs" color="muted">{l.lang?.toUpperCase() ?? l.lang}</Text>
            ))}
          </div>
        )}

        {manga.genres && manga.genres.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {manga.genres.slice(0, 5).map(g => (
              <Badge key={g} variant="default" size="sm">
                {g}
              </Badge>
            ))}
          </div>
        )}

        <div className="mt-3 flex items-center gap-3">
          <Link to={DEEP_LINKS.MANGA_DETAIL({ mangaId: manga.id ?? '' })}>
            <Button size="sm">View</Button>
          </Link>
          {manga.updated_at && (
            <Text size="xs" color="muted">Updated {formatDate(manga.updated_at, DateFormat.ShortDateTime)}</Text>
          )}
        </div>
      </div>
    </div>
  )
}

function Pagination({ page, total, onChange }: { page: number; total: number; onChange: (page: number) => void }) {
  if (total <= 1) return null

  return (
    <div className="flex items-center justify-center gap-3">
      <Button
        variant="outline"
        size="sm"
        onClick={() => onChange(page - 1)}
        disabled={page <= 1}
      >
        <Icon as={ChevronLeft} size="small" className="mr-1 inline" /> Prev
      </Button>
      <Text color="muted">
        {page} / {total}
      </Text>
      <Button
        variant="outline"
        size="sm"
        onClick={() => onChange(page + 1)}
        disabled={page >= total}
      >
        Next <Icon as={ChevronRight} size="small" className="ml-1 inline" />
      </Button>
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
        <Text color="muted" className="mb-4">{data.itemCount} manga</Text>
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
        <NoResults variant="error" message={error} />
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
              onChange={(p: number) => setFilter('page', p)}
            />
          </div>
        </>
      )}
    </div>
  )
}
