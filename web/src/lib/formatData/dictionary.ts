import type { DomainDictionaryResponse } from '../../types'

export function formatDictionary(raw: DomainDictionaryResponse): DomainDictionaryResponse {
  return {
    id: raw.id ?? '',
    slug: raw.slug ?? '',
    title: raw.title ?? '',
    cover_url: raw.cover_url ?? '',
    sources: raw.sources ?? {},
    chapters_by_lang: raw.chapters_by_lang ?? {},
    best_source: raw.best_source ?? {},
    source_stats: raw.source_stats ?? {},
    updated_at: raw.updated_at ?? '',
  }
}
