import type { DomainChapterReadResponse } from '@/types'

// Parses next/prev chapter IDs from full URLs like "/read/abc-123"
export function parseChapterIdFromUrl(url: string | null | undefined): string | null {
  if (!url) return null
  try {
    const parsed = new URL(url, window.location.origin)
    return parsed.pathname.split('/').pop() ?? null
  } catch {
    return null
  }
}

export function formatChapterRead(raw: DomainChapterReadResponse): DomainChapterReadResponse {
  return {
    chapter_id: raw.chapter_id ?? '',
    manga_id: raw.manga_id ?? '',
    pages: raw.pages ?? [],
    prev_chapter: raw.prev_chapter ?? undefined,
    next_chapter: raw.next_chapter ?? undefined,
  }
}
