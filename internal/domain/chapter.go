package domain

import (
	"time"
)

// Chapter is a downloaded chapter entity belonging to a Manga.
type Chapter struct {
	ID        string  // optional; auto-generated if empty before upsert
	MangaID   string  // references manga.id
	Number    string  // human-readable: "78", "78.5", "S1 - Chapter 78"
	SortKey   float64 // internal numeric sort key for ordering; not serialized
	Title     string
	Language  string
	Source    string
	SourceID  string
	PageURLs  []string
	ScrapedAt time.Time
}
