import type { DomainChapterListResponse, DomainChapterItem } from '../../types'

export function formatChapterList(raw: DomainChapterListResponse): DomainChapterListResponse {
  return {
    id: raw.id ?? '',
    lang: raw.lang ?? '',
    chapters: (raw.chapters ?? []).map(formatChapterItem),
  }
}

export function formatChapterItem(raw: DomainChapterItem): DomainChapterItem {
  return {
    id: raw.id ?? '',
    name: raw.name ?? '',
    order: raw.order ?? 0,
    page_count: raw.page_count ?? 0,
    updated_at: raw.updated_at ?? '',
  }
}
