package domain

import (
	"regexp"
	"strconv"
	"time"
)

// Chapter is a downloaded chapter entity belonging to a Manga.
type Chapter struct {
	MangaSlug string
	Number    string  // human-readable: "78", "78.5", "S1 - Chapter 78"
	SortKey   float64 // internal numeric sort key for ordering; not serialized
	Title     string
	Language  string
	Source    string
	SourceID  string
	PageURLs  []string
	ScrapedAt time.Time
}

var seasonChapterRe = regexp.MustCompile(`(?i)^S(\d+)\s*-\s*Chapter\s+([\d.]+)$`)

// ParseChapterSortKey derives a float64 sort key from a human-readable chapter number.
// "S1 - Chapter 78" → 10078.0, "78.5" → 78.5, "78" → 78.0
func ParseChapterSortKey(num string) float64 {
	if m := seasonChapterRe.FindStringSubmatch(num); len(m) == 3 {
		season, _ := strconv.ParseFloat(m[1], 64)
		chapter, _ := strconv.ParseFloat(m[2], 64)
		return season*10000 + chapter
	}
	f, _ := strconv.ParseFloat(num, 64)
	return f
}
