package domain

import "context"

// Scraper fetches manga metadata and chapter content from a single source.
type Scraper interface {
	Source() string
	FetchMangaDetail(ctx context.Context, id string) (*Manga, error)
	FetchChapterList(ctx context.Context, mangaID string) ([]Chapter, error)
	FetchPageURLs(ctx context.Context, chapterID string) ([]string, error)
}

// Searcher is an optional capability scrapers may implement.
type Searcher interface {
	Search(ctx context.Context, query string) ([]SearchResult, error)
}

// SourceRegistry provides ordered scrapers for use by application services.
type SourceRegistry interface {
	Ordered() []Scraper
}

// SearchResult is a single hit returned by a source's search.
type SearchResult struct {
	ID     string
	Title  string
	Status string
	Genres []string
	Type   string
	Source string
}
