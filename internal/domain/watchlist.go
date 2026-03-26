package domain

import (
	"context"
	"time"
)

// WatchlistEntry holds a registered manga pending ingestion.
type WatchlistEntry struct {
	ID           string
	Slug         string
	Title        string
	DictionaryID string
	Sources      map[string]string // source_name → source_id
	LastChecked  *time.Time
}

// WatchlistManager is the watchlist management port used by the HTTP handler.
type WatchlistManager interface {
	AddByDictionaryID(ctx context.Context, dictionaryID string) (string, error)
}

// WatchlistResponse is the API representation of a watchlist entry.
type WatchlistResponse struct {
	Slug         string `json:"slug"`
	DictionaryID string `json:"dictionaryId"`
}
