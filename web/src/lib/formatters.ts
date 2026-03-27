import type {
  DomainMangaSummary,
  DomainMangaListResponse,
  DomainMangaDetailResponse,
  DomainChapterListResponse,
  DomainChapterReadResponse,
  DomainDictionaryResponse,
} from '../types'
import type {
  AppMangaItem,
  AppMangaPagedList,
  AppMangaDetail,
  AppChapterList,
  AppChapterRead,
  AppDictionaryEntry,
} from '../types/app'

export function formatMangaItem(raw: DomainMangaSummary): AppMangaItem {
  return {
    id: raw.id ?? '',
    title: raw.title ?? '',
    description: raw.description ?? '',
    status: raw.status ?? '',
    coverUrl: raw.cover_url ?? '',
    state: raw.state ?? '',
    authors: raw.authors ?? [],
    genres: raw.genres ?? [],
    languages: raw.languages ?? [],
    chaptersByLang: raw.chapters_by_lang ?? {},
    updatedAt: raw.updated_at ?? '',
  }
}

export function formatMangaPagedList(raw: DomainMangaListResponse): AppMangaPagedList {
  return {
    items: (raw.items ?? []).map(formatMangaItem),
    pageNumber: raw.pageNumber ?? 1,
    pageSize: raw.pageSize ?? 20,
    pageTotal: raw.pageTotal ?? 1,
    itemCount: raw.itemCount ?? 0,
  }
}

export function formatMangaDetail(raw: DomainMangaDetailResponse): AppMangaDetail {
  return {
    id: raw.id ?? '',
    title: raw.title ?? '',
    description: raw.description ?? '',
    status: raw.status ?? '',
    state: raw.state ?? '',
    coverUrl: raw.cover_url ?? '',
    authors: raw.authors ?? [],
    genres: raw.genres ?? [],
    sources: raw.sources ?? {},
    languages: (raw.languages ?? []).map(l => ({
      lang: l.lang ?? '',
      latestUpdate: l.latest_update ?? '',
      totalChapters: l.total_chapters ?? 0,
      fetchedChapters: l.fetched_chapters ?? 0,
      uploadedChapters: l.uploaded_chapters ?? 0,
    })),
    updatedAt: raw.updated_at ?? '',
  }
}

export function formatChapterList(raw: DomainChapterListResponse): AppChapterList {
  return {
    id: raw.id ?? '',
    lang: raw.lang ?? '',
    chapters: (raw.chapters ?? []).map(c => ({
      chapter: c.chapter ?? "0",
      pageCount: c.page_count ?? 0,
      uploadedAt: c.uploaded_at ?? '',
    })),
  }
}

function parseChapterFromUrl(url: string | null | undefined): number | null {
  if (!url) return null
  try {
    const ch = new URL(url, 'http://x').searchParams.get('chapter')
    return ch !== null ? parseFloat(ch) : null
  } catch {
    return null
  }
}

export function formatChapterRead(raw: DomainChapterReadResponse): AppChapterRead {
  return {
    id: raw.id ?? '',
    lang: raw.lang ?? '',
    chapter: raw.chapter ?? "0",
    pages: raw.pages ?? [],
    prevChapter: parseChapterFromUrl(raw.prev_chapter),
    nextChapter: parseChapterFromUrl(raw.next_chapter),
  }
}

export function formatDictionaryEntry(raw: DomainDictionaryResponse): AppDictionaryEntry {
  const chaptersByLang = raw.chapters_by_lang ?? {}
  return {
    id: raw.id ?? '',
    slug: raw.slug ?? '',
    title: raw.title ?? '',
    state: raw.state ?? '',
    coverUrl: raw.cover_url ?? '',
    sources: raw.sources ?? {},
    chaptersByLang,
    totalChapters: Object.values(chaptersByLang).reduce((a, b) => a + b, 0),
  }
}
