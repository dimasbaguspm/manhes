package domain

import (
	"context"
	"time"
)

// Repository is the single port for all DB operations.
type Repository interface {
	// Catalog
	UpsertManga(ctx context.Context, m Manga) error
	ListManga(ctx context.Context, filter MangaFilter) (MangaPage, error)
	GetMangaBySlug(ctx context.Context, slug string) (MangaDetail, bool, error)
	UpsertLang(ctx context.Context, slug, lang string, available, downloaded int) error
	UpsertChapter(ctx context.Context, slug, lang, num string, sortKey float64, pageCount int) error
	MarkChapterUploaded(ctx context.Context, slug, lang, num string) error
	UpsertPage(ctx context.Context, slug, lang, num string, idx int, url string) error
	GetChapterPages(ctx context.Context, slug, lang, num string) ([]string, error)
	GetChaptersByLang(ctx context.Context, slug, lang string) ([]MangaChapter, error)
	GetPendingChapters(ctx context.Context) ([]ChapterRef, error)
	HasUploadedChapters(ctx context.Context, slug string) (bool, error)
	HasPendingChapters(ctx context.Context, slug string) (bool, error)

	// Dictionary
	UpsertDictionary(ctx context.Context, entry DictionaryEntry) error
	GetDictionary(ctx context.Context, id string) (DictionaryEntry, bool, error)
	GetDictionaryBySlug(ctx context.Context, slug string) (DictionaryEntry, bool, error)
	ListDictionary(ctx context.Context, filter DictionaryFilter) (DictionaryPage, error)
	SetDictionaryState(ctx context.Context, id string, state MangaState) error
	SetDictionaryStateBySlug(ctx context.Context, slug string, state MangaState) error

	// Watchlist
	ListWatchlist(ctx context.Context) ([]WatchlistEntry, error)
	AddWatchlist(ctx context.Context, entry WatchlistEntry) error
	RemoveWatchlist(ctx context.Context, slug string) error
	UpdateLastChecked(ctx context.Context, slug string, t time.Time) error

	// Ingest chapter state
	IsChapterDownloaded(ctx context.Context, slug, lang, num string) (bool, error)
	MarkChapterDownloaded(ctx context.Context, slug, lang, num string, sortKey float64) error
	GetDownloadedByLang(ctx context.Context, slug string) (map[string]int, error)
	GetDownloadedChaptersByLang(ctx context.Context, slug, lang string) ([]string, error)

	Close() error
}

// EventPublisher publishes domain events to the event bus.
type EventPublisher interface {
	PublishIngestRequested(ctx context.Context, e IngestRequested) error
	PublishChapterDownloaded(ctx context.Context, e ChapterDownloaded) error
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
