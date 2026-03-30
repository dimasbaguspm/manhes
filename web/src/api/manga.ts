import { api } from './client'
import type { DomainMangaListResponse, DomainMangaDetailResponse, DomainChapterListResponse, DomainChapterReadResponse } from '../types'

// Swagger params: id, q, genre, author, state, sortBy, sortOrder, page, pageSize
export interface ListMangaParams {
  id?: string | string[]
  q?: string
  genre?: string | string[]
  author?: string | string[]
  state?: string | string[]
  sortBy?: 'title' | 'updatedAt' | 'createdAt'
  sortOrder?: 'asc' | 'desc'
  page?: number
  pageSize?: number
}

function toQueryString(params: ListMangaParams): string {
  const q = new URLSearchParams()
  Object.entries(params).forEach(([k, v]) => {
    if (v === undefined || v === '') return
    if (Array.isArray(v)) v.forEach(item => q.append(k, String(item)))
    else q.set(k, String(v))
  })
  const s = q.toString()
  return s ? `?${s}` : ''
}

export const mangaApi = {
  list(params: ListMangaParams = {}) {
    return api.get<DomainMangaListResponse>(`/manga${toQueryString(params)}`)
  },

  get(mangaId: string) {
    return api.get<DomainMangaDetailResponse>(`/manga/${mangaId}`)
  },

  chapters(mangaId: string, lang: string) {
    return api.get<DomainChapterListResponse>(`/manga/${mangaId}/${lang}`)
  },

  // Swagger: GET /read/{chapterId}
  read(chapterId: string) {
    return api.get<DomainChapterReadResponse>(`/read/${encodeURIComponent(chapterId)}`)
  },
}
