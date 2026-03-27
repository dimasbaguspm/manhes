import { createContext, useContext, useEffect, useCallback, type ReactNode } from 'react'
import { useParams, useSearchParams } from 'react-router-dom'
import { mangaApi } from '../api/manga'
import { useAsync } from '../hooks/useAsync'
import { formatChapterRead } from '../lib/formatters'
import type { AppChapterRead } from '../types/app'

interface MangaReaderContextValue {
  data: AppChapterRead | null
  loading: boolean
  error: string | null
  chapter: string
  mangaId: string
  lang: string
  goNext: () => void
  goPrev: () => void
}

const MangaReaderContext = createContext<MangaReaderContextValue | null>(null)

export function MangaReaderDataProvider({ children }: { children: ReactNode }) {
  const { mangaId, lang } = useParams<{ mangaId: string; lang: string }>()
  const [searchParams, setSearchParams] = useSearchParams()
  const chapter = searchParams.get('chapter') ?? '1' 

  const { data: raw, loading, error } = useAsync(
    () => mangaId && lang ? mangaApi.read(mangaId, lang, chapter) : Promise.reject('Not found'),
    [mangaId, lang, chapter],
  )

  const data = raw ? formatChapterRead(raw) : null

  const goNext = useCallback(() => {
    if (data?.nextChapter != null) {
      setSearchParams({ chapter: String(data.nextChapter) })
      window.scrollTo(0, 0)
    }
  }, [data?.nextChapter, setSearchParams])

  const goPrev = useCallback(() => {
    if (data?.prevChapter != null) {
      setSearchParams({ chapter: String(data.prevChapter) })
      window.scrollTo(0, 0)
    }
  }, [data?.prevChapter, setSearchParams])

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'ArrowRight') goNext()
      if (e.key === 'ArrowLeft') goPrev()
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [goNext, goPrev])

  if (!mangaId || !lang) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-950 text-sm text-gray-500">
        Page not found.
      </div>
    )
  }

  return (
    <MangaReaderContext.Provider value={{ data, loading, error, chapter, mangaId, lang, goNext, goPrev }}>
      {children}
    </MangaReaderContext.Provider>
  )
}

export function useMangaReader() {
  const ctx = useContext(MangaReaderContext)
  if (!ctx) throw new Error('useMangaReader must be used within MangaReaderDataProvider')
  return ctx
}
