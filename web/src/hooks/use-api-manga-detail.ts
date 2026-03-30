import { useAsync } from '@/hooks/use-async'
import { mangaApi } from '@/api/manga'
import { formatMangaDetail } from '@/lib/format-data'

export function useApiMangaDetail(mangaId: string | undefined) {
  return useAsync(
    () => mangaId ? mangaApi.get(mangaId).then(formatMangaDetail) : Promise.reject('missing mangaId'),
    [mangaId],
  )
}
