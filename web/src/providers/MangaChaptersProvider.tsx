import { createContext, useContext, type ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { mangaApi } from '../api/manga'
import { useAsync } from '../hooks/useAsync'
import { formatChapterList, formatMangaDetail } from '../lib/formatters'
import type { AppChapterList } from '../types/app'

interface LangStats {
  totalChapters: number
  fetchedChapters: number
  uploadedChapters: number
}

interface MangaChaptersContextValue {
  data: AppChapterList | null
  mangaTitle: string
  langStats: LangStats | null
  loading: boolean
  error: string | null
}

const MangaChaptersContext = createContext<MangaChaptersContextValue | null>(null)

export function MangaChaptersProvider({ children }: { children: ReactNode }) {
  const { mangaId, lang } = useParams<{ mangaId: string; lang: string }>()

  const { data: chaptersRaw, loading, error } = useAsync(
    () => mangaId && lang ? mangaApi.chapters(mangaId, lang) : Promise.reject('Not found'),
    [mangaId, lang],
  )
  const { data: mangaRaw } = useAsync(
    () => mangaId ? mangaApi.get(mangaId) : Promise.reject('Not found'),
    [mangaId],
  )

  const data = chaptersRaw ? formatChapterList(chaptersRaw) : null
  const mangaDetail = mangaRaw ? formatMangaDetail(mangaRaw) : null
  const mangaTitle = mangaDetail?.title ?? ''
  const langStats = mangaDetail?.languages.find(l => l.lang === lang) ?? null

  if (!mangaId || !lang) {
    return (
      <div className="rounded-lg border border-gray-800 bg-gray-900 px-4 py-8 text-center text-sm text-gray-500">
        Page not found.
      </div>
    )
  }

  return (
    <MangaChaptersContext.Provider value={{ data, mangaTitle, langStats, loading, error }}>
      {children}
    </MangaChaptersContext.Provider>
  )
}

export function useMangaChapters() {
  const ctx = useContext(MangaChaptersContext)
  if (!ctx) throw new Error('useMangaChapters must be used within MangaChaptersProvider')
  return ctx
}
