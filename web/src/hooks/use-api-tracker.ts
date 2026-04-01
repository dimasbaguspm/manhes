import { useAsync } from '@/hooks/use-async'
import { mangaApi } from '@/api/manga'

export function useApiTrackers(mangaId: string | undefined) {
  return useAsync(
    () => (mangaId ? mangaApi.getTrackers(mangaId) : Promise.reject('missing mangaId')),
    [mangaId],
  )
}
