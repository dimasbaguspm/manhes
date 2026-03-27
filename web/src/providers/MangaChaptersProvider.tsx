import { createContext, useContext, type ReactNode } from 'react'
import { mangaApi } from '../api/manga'
import { useAsync } from '../hooks/useAsync'
import { formatChapterList } from '../lib/formatters'
import type { AppChapterList } from '../types/app'

interface MangaChaptersContextValue {
  data: AppChapterList | null
  loading: boolean
  error: string | null
}

const MangaChaptersContext = createContext<MangaChaptersContextValue | null>(null)

interface MangaChaptersProviderProps {
  mangaId: string
  lang: string
  children: ReactNode
}

export function MangaChaptersProvider({ mangaId, lang, children }: MangaChaptersProviderProps) {
  const { data: raw, loading, error } = useAsync(
    () => mangaApi.chapters(mangaId, lang),
    [mangaId, lang],
  )

  const data = raw ? formatChapterList(raw) : null

  return (
    <MangaChaptersContext.Provider value={{ data, loading, error }}>
      {children}
    </MangaChaptersContext.Provider>
  )
}

export function useMangaChapters() {
  const ctx = useContext(MangaChaptersContext)
  if (!ctx) throw new Error('useMangaChapters must be used within MangaChaptersProvider')
  return ctx
}
