package domain

import (
	"context"
	"time"
)

// Repository is the single port for all DB operations.
type Repository interface {
	// Catalog
	UpsertManga(m Manga) error
	ListManga(filter MangaFilter) (MangaPage, error)
	GetMangaBySlug(slug string) (MangaDetail, bool, error)
	UpsertLang(slug, lang string, available, downloaded int) error
	UpsertChapter(slug, lang string, num float64, pageCount int) error
	MarkChapterUploaded(slug, lang string, num float64) error
	UpsertPage(slug, lang string, num float64, idx int, url string) error
	GetChapterPages(slug, lang string, num float64) ([]string, error)
	GetChaptersByLang(slug, lang string) ([]MangaChapter, error)
	GetPendingChapters() ([]ChapterRef, error)
	HasUploadedChapters(slug string) (bool, error)
	HasPendingChapters(slug string) (bool, error)

	// Dictionary
	UpsertDictionary(entry DictionaryEntry) error
	GetDictionary(id string) (DictionaryEntry, bool, error)
	GetDictionaryBySlug(slug string) (DictionaryEntry, bool, error)
	ListDictionary(filter DictionaryFilter) (DictionaryPage, error)
	SetDictionaryState(id string, state MangaState) error
	SetDictionaryStateBySlug(slug string, state MangaState) error

	// Watchlist
	ListWatchlist() ([]WatchlistEntry, error)
	AddWatchlist(entry WatchlistEntry) error
	RemoveWatchlist(slug string) error
	UpdateLastChecked(slug string, t time.Time) error

	// Ingest chapter state
	IsChapterDownloaded(slug, lang string, num float64) (bool, error)
	MarkChapterDownloaded(slug, lang string, num float64) error
	GetDownloadedByLang(slug string) (map[string]int, error)
	GetDownloadedChaptersByLang(slug, lang string) ([]float64, error)

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
	SavePage(slug, lang string, chapterNum float64, pageIdx int, data []byte, ext string) (string, error)
	SaveCover(slug string, data []byte, ext string) (string, error)
	WriteMetadata(slug string, m *Metadata) error
	ReadMetadata(slug string) (*Metadata, error)
	WriteLangMetadata(slug, lang string, m *LangMetadata) error
	ReadLangMetadata(slug, lang string) (*LangMetadata, error)
	WriteChapterManifest(slug, lang string, ch *Chapter) error
}
