import type { DomainMangaListResponse, DomainMangaSummary } from '../../types'

export function formatMangaList(raw: DomainMangaListResponse): DomainMangaListResponse {
  return {
    itemCount: raw.itemCount ?? 0,
    pageNumber: raw.pageNumber ?? 1,
    pageSize: raw.pageSize ?? 20,
    pageTotal: raw.pageTotal ?? 1,
    items: raw.items ?? [],
  }
}

export function formatMangaSummary(raw: DomainMangaSummary): DomainMangaSummary {
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
