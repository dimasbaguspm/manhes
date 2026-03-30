import { useAsync } from './useAsync'
import { mangaApi } from '../api/manga'
import { formatChapterRead } from '../lib/formatData'

export function useApiChapterRead(chapterId: string | undefined) {
  return useAsync(
    () => chapterId ? mangaApi.read(chapterId).then(formatChapterRead) : Promise.reject('missing chapterId'),
    [chapterId],
  )
}
