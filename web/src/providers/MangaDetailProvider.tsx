import { createContext, useContext, useState, useCallback, type ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { mangaApi } from '../api/manga'
import { watchlistApi } from '../api/watchlist'
import { dictionaryApi } from '../api/dictionary'
import { useAsync } from '../hooks/useAsync'
import { formatMangaDetail } from '../lib/formatters'
import type { AppMangaDetail } from '../types/app'

type AddState = 'idle' | 'loading' | 'done' | 'error'

interface MangaDetailContextValue {
  data: AppMangaDetail | null
  loading: boolean
  error: string | null
  addState: AddState
  addToWatchlist: () => Promise<void>
  refreshState: AddState
  refreshManga: () => Promise<void>
}

const MangaDetailContext = createContext<MangaDetailContextValue | null>(null)

export function MangaDetailProvider({ children }: { children: ReactNode }) {
  const { mangaId } = useParams<{ mangaId: string }>()
  const [addState, setAddState] = useState<AddState>('idle')
  const [refreshState, setRefreshState] = useState<AddState>('idle')

  const { data: raw, loading, error } = useAsync(
    () => mangaId ? mangaApi.get(mangaId) : Promise.reject('Not found'),
    [mangaId],
  )

  const data = raw ? formatMangaDetail(raw) : null

  async function addToWatchlist() {
    if (!data) return
    setAddState('loading')
    try {
      await watchlistApi.add(data.id)
      setAddState('done')
    } catch {
      setAddState('error')
    }
  }

  async function refreshManga() {
    if (!data) return
    setRefreshState('loading')
    try {
      await dictionaryApi.refresh(data.id)
      await watchlistApi.add(data.id)
      setRefreshState('done')
    } catch {
      setRefreshState('error')
    }
  }

  if (!mangaId) {
    return (
      <div className="rounded-lg border border-gray-800 bg-gray-900 px-4 py-8 text-center text-sm text-gray-500">
        Page not found.
      </div>
    )
  }

  return (
    <MangaDetailContext.Provider value={{ data, loading, error, addState, addToWatchlist, refreshState, refreshManga }}>
      {children}
    </MangaDetailContext.Provider>
  )
}

export function useMangaDetail() {
  const ctx = useContext(MangaDetailContext)
  if (!ctx) throw new Error('useMangaDetail must be used within MangaDetailProvider')
  return ctx
}
