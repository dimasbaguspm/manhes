import { useAsync } from './useAsync'
import { mangaApi } from '../api/manga'
import { formatMangaDetail } from '../lib/formatData'

export function useApiMangaDetail(mangaId: string | undefined) {
  return useAsync(
    () => mangaId ? mangaApi.get(mangaId).then(formatMangaDetail) : Promise.reject('missing mangaId'),
    [mangaId],
  )
}
