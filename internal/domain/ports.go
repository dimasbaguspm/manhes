package domain

import (
	"context"

	"manga-engine/pkg/eventbus"
)

// Event is re-exported from pkg/eventbus for domain consumers.
type Event = eventbus.Event

// Handler is re-exported from pkg/eventbus for domain consumers.
type Handler = eventbus.Handler

// EventBus is re-exported from pkg/eventbus for domain consumers.
type EventBus = eventbus.Bus

// Repository is the single port for all DB operations.
type Repository interface {
	// Manga
	UpsertManga(ctx context.Context, m Manga) error
	ListManga(ctx context.Context, filter MangaFilter) (MangaPage, error)
	GetMangaByDictionaryID(ctx context.Context, dictionaryID string) (MangaDetail, bool, error)
	GetMangaByID(ctx context.Context, id string) (MangaDetail, bool, error)

	// Chapters
	UpsertChapter(ctx context.Context, id, mangaID, name string, chapterOrder int, lang, imageSrc string) error
	UpsertChapterBatch(ctx context.Context, chapters []Chapter) error
	GetChapterCountByLang(ctx context.Context, mangaID, lang string) (int, error)
	GetChaptersByLang(ctx context.Context, mangaID, lang string) ([]Chapter, error)
	GetChaptersByManga(ctx context.Context, mangaID string) ([]Chapter, error)
	IsChapterIngested(ctx context.Context, mangaID, lang string, chapterOrder int) (bool, error)
	GetChapterUploaded(ctx context.Context, mangaID, lang, chapterNum string) (bool, error)

	// Dictionary
	UpsertDictionary(ctx context.Context, entry DictionaryEntry) error
	UpsertDictionaryBatch(ctx context.Context, entries []DictionaryEntry) error
	GetDictionary(ctx context.Context, id string) (DictionaryEntry, bool, error)
	GetDictionaryBySlug(ctx context.Context, slug string) (DictionaryEntry, bool, error)
	ListDictionary(ctx context.Context, filter DictionaryFilter) (DictionaryPage, error)

	// Ingest
	GetDownloadedByLang(ctx context.Context, mangaID string) (map[string]int, error)
	GetDownloadedChaptersByLang(ctx context.Context, mangaID, lang string) ([]string, error)

	// Manga (cover)
	UpdateMangaCover(ctx context.Context, mangaID, coverURL string) error

	Close() error
}

// Downloader fetches remote resources over HTTP.
type Downloader interface {
	Download(ctx context.Context, url string) ([]byte, error)
}

// ObjectStore uploads and deletes objects in an S3-compatible store.
type ObjectStore interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)
	Delete(ctx context.Context, key string) error
}

// Storer persists manga pages and metadata to disk.
type Storer interface {
	SavePage(slug, lang string, chapterNum string, pageIdx int, data []byte, ext string) (string, error)
	SaveCover(slug string, data []byte, ext string) (string, error)
	WriteMetadata(slug string, m *Metadata) error
	ReadMetadata(slug string) (*Metadata, error)
	WriteLangMetadata(slug, lang string, m *LangMetadata) error
	ReadLangMetadata(slug, lang string) (*LangMetadata, error)
	WriteChapterManifest(slug, lang string, ch *Chapter) error
}

// MangaManager is the manga management port, subscribed to dictionary.updated events.
type MangaManager interface {
	HandleDictionaryUpdated(ctx context.Context, e DictionaryUpdated) error
}
