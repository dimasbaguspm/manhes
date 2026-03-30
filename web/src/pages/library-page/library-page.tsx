import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight, RefreshCw } from 'lucide-react'
import { Icon } from '@/components/icon'
import { Badge } from '@/components/ui/badge/badge'
import { Button } from '@/components/ui/button/button'
import { Text } from '@/components/ui/text/text'
import { ButtonIcon } from '@/components/ui/button-icon'
import { NoResults } from '@/components/ui/no-results/no-results'
import { Image } from '@/components/ui/image'
import { useApiMangaList } from '@/hooks/use-api-manga-list'
import { useApiRefreshDictionary } from '@/hooks/use-api-refresh-dictionary'
import type { ListMangaParams } from '@/api/manga'
import { DEEP_LINKS } from '@/lib/deep-links'
import { formatDate, DateFormat } from '@/lib/format-date'
import { cn } from '@/lib/cn'
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

const NEEDS_REFRESH_THRESHOLD_MS = 24 * 60 * 60 * 1000

function needsRefresh(updatedAt: string | undefined): boolean {
  if (!updatedAt) return false
  return Date.now() - new Date(updatedAt).getTime() > NEEDS_REFRESH_THRESHOLD_MS
}

function LibraryItem({ manga }: { manga: DomainMangaSummary }) {
  const { state: refreshState, refresh } = useApiRefreshDictionary()
  const isRefreshing = refreshState === 'loading'
  const isStale = needsRefresh(manga.updated_at)
  const languages = manga.languages?.map(l => l.lang?.toUpperCase()).filter(Boolean).join(', ')

  return (
    <div className="group flex flex-col gap-4 rounded-lg border border-gray-800 bg-gray-900 p-4 transition hover:border-gray-700 sm:flex-row">
      {/* Image - wider on desktop, full width on mobile */}
      <div className="relative w-full sm:w-32 md:w-40">
        <Image
          src={manga.cover_url}
          alt={manga.title}
          size="lg"
          aspect="landscape"
          className="w-full sm:w-32 md:w-40"
        />
        {/* State badge overlay on mobile */}
        <div className="absolute top-2 right-2 sm:hidden">
          <Badge
            variant={STATE_BADGE_VARIANT[manga.state ?? ''] ?? 'default'}
            size="sm"
          >
            {STATE_LABEL[manga.state ?? ''] ?? manga.state}
          </Badge>
        </div>
      </div>

      {/* Metadata */}
      <div className="min-w-0 flex-1">
        {/* Desktop: title and badge side by side. Mobile: title only (badge is on image) */}
        <div className="hidden items-start justify-between gap-3 sm:flex">
          <h3 className="font-semibold text-gray-100">{manga.title}</h3>
          <Badge
            variant={STATE_BADGE_VARIANT[manga.state ?? ''] ?? 'default'}
            size="sm"
          >
            {STATE_LABEL[manga.state ?? ''] ?? manga.state}
          </Badge>
        </div>
        {/* Mobile title */}
        <h3 className="font-semibold text-gray-100 sm:hidden">{manga.title}</h3>

        {/* Mobile: languages and relative time below title */}
        <div className="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 sm:hidden">
          {languages && <Text size="xs" color="muted">{languages}</Text>}
          {manga.updated_at && (
            <Text size="xs" color="muted">{formatDate(manga.updated_at, DateFormat.Relative)}</Text>
          )}
        </div>

        {/* Desktop: description and genres */}
        <div className="hidden sm:block">
          {manga.description && (
            <Text size="sm" color="muted" className="mt-1 line-clamp-2">
              {manga.description}
            </Text>
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

          {languages && (
            <div className="mt-2 flex flex-wrap gap-x-3 gap-y-1">
              <Text size="xs" color="muted">{languages}</Text>
            </div>
          )}
        </div>

        {/* Action row */}
        <div className="mt-3 flex items-center gap-3">
          <Link to={DEEP_LINKS.MANGA_DETAIL({ mangaId: manga.id ?? '' })}>
            <Button size="sm">View</Button>
          </Link>
          {/* Desktop timestamp */}
          <div className="hidden items-center gap-2 sm:flex">
            {manga.updated_at && (
              <Text size="xs" color="muted">{formatDate(manga.updated_at, DateFormat.Relative)}</Text>
            )}
            {isStale && manga.dictionary_id && (
              <ButtonIcon
                variant="ghost"
                size="sm"
                onClick={() => refresh(manga.dictionary_id!)}
                disabled={isRefreshing}
                aria-label="Refresh manga"
                className={isRefreshing ? 'animate-spin' : 'text-gray-400 hover:text-indigo-400'}
              >
                <RefreshCw size={14} />
              </ButtonIcon>
            )}
          </div>
          {/* Mobile refresh button */}
          {isStale && manga.dictionary_id && (
            <ButtonIcon
              variant="ghost"
              size="sm"
              onClick={() => refresh(manga.dictionary_id!)}
              disabled={isRefreshing}
              aria-label="Refresh manga"
              className={cn('ml-auto sm:hidden', isRefreshing ? 'animate-spin' : 'text-gray-400 hover:text-indigo-400')}
            >
              <RefreshCw size={14} />
            </ButtonIcon>
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
  state: ['fetching', 'uploading', 'available'],
  sortBy: 'updatedAt',
  sortOrder: 'desc',
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
              <div className="aspect-[3/4] w-32 animate-pulse rounded-md bg-gray-800" />
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
