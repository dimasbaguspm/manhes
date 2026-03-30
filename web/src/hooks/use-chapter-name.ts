import { useApiChapterList } from '@/hooks/use-api-chapter-list'

/**
 * Looks up a chapter's display name from the manga's chapter list.
 * mangaId comes from the chapter read response (data.manga_id).
 * lang defaults to 'en' as a sensible fallback when not provided.
 */
export function useChapterName(mangaId: string | undefined, chapterId: string | undefined, lang = 'en') {
  const { data, loading } = useApiChapterList(mangaId, lang)

  const name = (() => {
    if (!data?.chapters || !chapterId) return null
    return data.chapters.find(ch => ch.id === chapterId)?.name ?? null
  })()

  return { name, loading }
}
