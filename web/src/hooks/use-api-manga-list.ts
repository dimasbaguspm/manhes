import { useAsync } from '@/hooks/use-async'
import { mangaApi, type ListMangaParams } from '@/api/manga'
import { formatMangaList } from '@/lib/format-data'

export function useApiMangaList(params: ListMangaParams = {}) {
  return useAsync(
    () => mangaApi.list(params).then(formatMangaList),
    [JSON.stringify(params)],
  )
}
