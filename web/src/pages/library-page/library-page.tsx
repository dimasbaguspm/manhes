import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { Icon } from '@/components/icon'
import { Badge } from '@/components/ui/badge/badge'
import { Button } from '@/components/ui/button/button'
import { Text } from '@/components/ui/text/text'
import { NoResults } from '@/components/ui/no-results/no-results'
import { useApiMangaList } from '@/hooks/use-api-manga-list'
import type { ListMangaParams } from '@/api/manga'
import { DEEP_LINKS } from '@/lib/deep-links'
import { MangaCard } from './components/manga-card'

const DEFAULT_FILTERS: ListMangaParams = {
  state: ['fetching', 'uploading', 'available'],
  sortBy: 'updatedAt',
  sortOrder: 'desc',
  page: 1,
  pageSize: 20,
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
        <div className="grid grid-cols-2 gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 12 }).map((_, i) => (
            <div key={i} className="overflow-hidden rounded-lg border border-gray-800 bg-gray-900">
              <div className="aspect-[2/3] w-full animate-pulse bg-gray-800" />
              <div className="p-3">
                <div className="h-4 w-full animate-pulse rounded bg-gray-800" />
                <div className="mt-2 h-3 w-2/3 animate-pulse rounded bg-gray-800" />
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
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-5 lg:grid-cols-6">
            {data.items.map(manga => (
              <MangaCard key={manga.id} manga={manga} />
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
