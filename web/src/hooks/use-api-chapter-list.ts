import { useAsync } from '@/hooks/use-async'
import { mangaApi } from '@/api/manga'
import { formatChapterList } from '@/lib/format-data'

export function useApiChapterList(mangaId: string | undefined, lang: string | undefined) {
  return useAsync(
    () => mangaId && lang ? mangaApi.chapters(mangaId, lang).then(formatChapterList) : Promise.reject('missing params'),
    [mangaId, lang],
  )
}
