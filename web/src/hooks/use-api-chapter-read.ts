import { useAsync } from '@/hooks/use-async'
import { mangaApi } from '@/api/manga'
import { formatChapterRead } from '@/lib/format-data'

export function useApiChapterRead(chapterId: string | undefined) {
  return useAsync(
    () => chapterId ? mangaApi.read(chapterId).then(formatChapterRead) : Promise.reject('missing chapterId'),
    [chapterId],
  )
}
