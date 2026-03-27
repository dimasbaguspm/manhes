import type { ListMangaParams } from '../api/manga'

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

  MANGA_CHAPTERS: ({ mangaId, lang }: { mangaId: string; lang: string }) =>
    `/manga/${mangaId}/${lang}`,

  MANGA_READER: ({ mangaId, lang, chapter }: { mangaId: string; lang: string; chapter: string }) =>
    `/manga/${mangaId}/${lang}/read?chapter=${chapter}`,
}
