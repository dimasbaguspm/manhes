package domain

import (
	"context"
	"time"
)

// MangaState represents the availability state of a manga in the dictionary.
type MangaState string

const (
	StateUnavailable MangaState = "unavailable"
	StateFetching    MangaState = "fetching"
	StateUploading   MangaState = "uploading"
	StateAvailable   MangaState = "available"
)

// SourceStat holds chapter availability data for a single source.
type SourceStat struct {
	ChaptersByLang map[string]int `json:"chapters_by_lang"`
	FetchedAt      time.Time      `json:"fetched_at"`
	Err            string         `json:"err"`
}

// DictionaryFilter holds optional filters and pagination for ListDictionary.
type DictionaryFilter struct {
	Q        string
	Page     int // 1-based
	PageSize int
}

// DictionaryPage is a paginated slice of DictionaryEntry.
type DictionaryPage struct {
	Items      []DictionaryEntry
	TotalItems int
	TotalPages int
	PageSize   int
	PageNumber int
}

// DictionaryEntry is the cross-source index record for a manga title.
// It is the source of truth for the /manga endpoints.
type DictionaryEntry struct {
	ID          string
	Slug        string
	Title       string
	CoverURL    string
	Sources     map[string]string     // source_name → source_id
	SourceStats map[string]SourceStat // source_name → latest stats
	BestSource  map[string]string     // lang → source_name with most chapters
	UpdatedAt   *time.Time
	CreatedAt   time.Time
}

// DictionaryManager is the dictionary management port used by the HTTP handler.
type DictionaryManager interface {
	Search(ctx context.Context, query string) ([]DictionaryEntry, error)
	Refresh(ctx context.Context, id string) (DictionaryEntry, error)
}

// DictionaryResponse is the API representation of a dictionary entry.
type DictionaryResponse struct {
	ID             string                `json:"id"`
	Slug           string                `json:"slug"`
	Title          string                `json:"title"`
	CoverURL       string                `json:"cover_url"`
	Sources        map[string]string     `json:"sources"`
	BestSource     map[string]string     `json:"best_source"`
	SourceStats    map[string]SourceStat `json:"source_stats"`
	ChaptersByLang map[string]int        `json:"chapters_by_lang"`
	UpdatedAt      *time.Time            `json:"updated_at"`
}
