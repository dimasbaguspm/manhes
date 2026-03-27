import { useEffect, useRef, useCallback } from 'react'

/**
 * Persists the current reading progress percentage to localStorage whenever
 * the chapter changes or the component unmounts. Uses a ref so the cleanup
 * closure always captures the latest scrollPct without re-registering the
 * effect on every scroll event.
 */
export function useProgressSave(
  mangaId: string | undefined,
  lang: string | undefined,
  chapter: string | undefined,
  scrollPct: number,
) {
  const scrollPctRef = useRef(scrollPct)
  useEffect(() => { scrollPctRef.current = scrollPct })

  const saveProgress = useCallback(() => {
    if (!mangaId || !lang || !chapter) return
    try {
      const raw = localStorage.getItem('manhes_read_progress')
      const prev = (raw ? JSON.parse(raw) : {}) as Record<string, number>
      localStorage.setItem('manhes_read_progress', JSON.stringify({
        ...prev,
        [`${mangaId}/${lang}/${chapter}`]: scrollPctRef.current,
      }))
    } catch {}
  }, [mangaId, lang, chapter])

  useEffect(() => {
    return () => { saveProgress() }
  }, [saveProgress])
}
