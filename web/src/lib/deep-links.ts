import type { ListMangaParams } from '@/api/manga'

function buildQuery(params: Record<string, string | number | undefined>): string {
  const q = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== '' && v !== 0) q.set(k, String(v))
  })
  return q.size ? `?${q.toString()}` : ''
}

export const DEEP_LINKS = {
  LIBRARY: (filters?: Partial<ListMangaParams>) =>
    `/${filters ? buildQuery(filters as Record<string, string | number | undefined>) : ''}`,

  DISCOVER: () => '/discover',

  MANGA_DETAIL: ({ mangaId }: { mangaId: string }) =>
    `/manga/${mangaId}`,

  MANGA_READER: ({ chapterId }: { chapterId: string }) =>
    `/read/${encodeURIComponent(chapterId)}`,
}
