package domain

import "context"

// MangaHandler is the read-only manga catalog interface exposed by the handler layer.
type MangaHandler interface {
	ListManga(ctx context.Context, filter MangaFilter) (MangaPage, []MangaSummary, error)
	GetManga(ctx context.Context, mangaID string) (MangaDetail, bool, error)
}

// ChapterHandler is the chapter read interface exposed by the handler layer.
type ChapterHandler interface {
	GetChaptersByLang(ctx context.Context, mangaID, lang string) ([]ChapterItem, bool, error)
	ReadChapter(ctx context.Context, chapterID string) (ChapterRead, bool, error)
}

// DictionaryHandler is the dictionary management interface exposed by the handler layer.
type DictionaryHandler interface {
	Search(ctx context.Context, query string) ([]DictionaryEntry, error)
	Refresh(ctx context.Context, id string) (DictionaryEntry, error)
}
