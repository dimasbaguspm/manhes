import { useAsync } from './useAsync'
import { mangaApi, type ListMangaParams } from '../api/manga'
import { formatMangaList } from '../lib/formatData'

export function useApiMangaList(params: ListMangaParams = {}) {
  return useAsync(
    () => mangaApi.list(params).then(formatMangaList),
    [JSON.stringify(params)],
  )
}
