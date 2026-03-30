import type { DomainMangaDetailResponse } from '@/types'

export function formatMangaDetail(raw: DomainMangaDetailResponse): DomainMangaDetailResponse {
  return {
    id: raw.id ?? '',
    title: raw.title ?? '',
    description: raw.description ?? '',
    status: raw.status ?? '',
    state: raw.state ?? '',
    cover_url: raw.cover_url ?? '',
    authors: raw.authors ?? [],
    genres: raw.genres ?? [],
    languages: raw.languages ?? [],
    updated_at: raw.updated_at ?? '',
    created_at: raw.created_at ?? '',
    dictionary_id: raw.dictionary_id ?? '',
  }
}
