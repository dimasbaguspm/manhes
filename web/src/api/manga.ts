import { api } from './client'
import type { DomainMangaListResponse, DomainMangaDetailResponse, DomainChapterListResponse, DomainChapterReadResponse } from '../types'

export interface ListMangaParams {
  title?: string
  status?: string
  state?: string
  sortBy?: string
  page?: number
  pageSize?: number
  hideUnavailable?: boolean
}

export const mangaApi = {
  list(params: ListMangaParams = {}) {
    const q = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') q.set(k, String(v))
    })
    const qs = q.size ? `?${q.toString()}` : ''
    return api.get<DomainMangaListResponse>(`/manga${qs}`)
  },

  get(mangaId: string) {
    return api.get<DomainMangaDetailResponse>(`/manga/${mangaId}`)
  },

  chapters(mangaId: string, lang: string) {
    return api.get<DomainChapterListResponse>(`/manga/${mangaId}/${lang}`)
  },

  read(mangaId: string, lang: string, chapter: number) {
    return api.get<DomainChapterReadResponse>(`/manga/${mangaId}/${lang}/read?chapter=${chapter}`)
  },
}
