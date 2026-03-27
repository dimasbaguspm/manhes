// App-level types: clean and non-optional, produced by formatters from domain types.

export interface AppMangaItem {
  id: string
  title: string
  description: string
  status: string
  coverUrl: string
  state: string
  authors: string[]
  genres: string[]
  languages: string[]
  chaptersByLang: Record<string, number>
  updatedAt: string
}

export interface AppMangaPagedList {
  items: AppMangaItem[]
  pageNumber: number
  pageSize: number
  pageTotal: number
  itemCount: number
}

export interface AppMangaDetail {
  id: string
  title: string
  description: string
  status: string
  state: string
  coverUrl: string
  authors: string[]
  genres: string[]
  sources: Record<string, string>
  languages: Array<{ lang: string; latestUpdate: string; totalChapters: number; fetchedChapters: number; uploadedChapters: number }>
  updatedAt: string
}

export interface AppChapterItem {
  chapter: string
  pageCount: number
  uploadedAt: string
}

export interface AppChapterList {
  id: string
  lang: string
  chapters: AppChapterItem[]
}

export interface AppChapterRead {
  id: string
  lang: string
  chapter: string
  pages: string[]
  prevChapter: string | null
  nextChapter: string | null
}

export interface AppDictionaryEntry {
  id: string
  slug: string
  title: string
  state: string
  coverUrl: string
  sources: Record<string, string>
  chaptersByLang: Record<string, number>
  totalChapters: number
}
