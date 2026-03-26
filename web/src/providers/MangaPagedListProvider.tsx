import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'
import { mangaApi, type ListMangaParams } from '../api/manga'
import { useAsync } from '../hooks/useAsync'
import { formatMangaPagedList } from '../lib/formatters'
import type { AppMangaPagedList } from '../types/app'

interface Filters extends Required<ListMangaParams> {}

interface MangaPagedListContextValue {
  data: AppMangaPagedList | null
  loading: boolean
  error: string | null
  filters: Filters
  setFilter: (key: keyof Filters, value: string | number) => void
  refresh: () => void
}

const MangaPagedListContext = createContext<MangaPagedListContextValue | null>(null)

const DEFAULT_FILTERS: Filters = {
  title: '',
  status: '',
  state: '',
  sortBy: 'title',
  page: 1,
  pageSize: 20,
  hideUnavailable: true,
}

export function MangaPagedListProvider({ children }: { children: ReactNode }) {
  const [filters, setFilters] = useState<Filters>(DEFAULT_FILTERS)
  const [refreshKey, setRefreshKey] = useState(0)

  const { data: raw, loading, error } = useAsync(
    () => mangaApi.list(filters),
    [filters, refreshKey],
  )

  const data = raw ? formatMangaPagedList(raw) : null

  function setFilter(key: keyof Filters, value: string | number) {
    setFilters(f => ({
      ...f,
      [key]: value,
      ...(key !== 'page' && { page: 1 }),
    }))
  }

  const refresh = useCallback(() => setRefreshKey(k => k + 1), [])

  return (
    <MangaPagedListContext.Provider value={{ data, loading, error, filters, setFilter, refresh }}>
      {children}
    </MangaPagedListContext.Provider>
  )
}

export function useMangaPagedList() {
  const ctx = useContext(MangaPagedListContext)
  if (!ctx) throw new Error('useMangaPagedList must be used within MangaPagedListProvider')
  return ctx
}
