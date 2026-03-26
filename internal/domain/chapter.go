package domain

import "time"

// Chapter is a downloaded chapter entity belonging to a Manga.
type Chapter struct {
	MangaSlug string
	Number    float64 // supports ch 10.5 etc.
	Title     string
	Language  string
	Source    string
	SourceID  string
	PageURLs  []string
	ScrapedAt time.Time
}
