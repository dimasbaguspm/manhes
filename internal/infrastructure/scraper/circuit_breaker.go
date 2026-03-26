package scraper

import (
	"context"

	"manga-engine/internal/domain"
	"manga-engine/pkg/circuitbreaker"
)

type cbScraper struct {
	inner domain.Scraper
	cb    *circuitbreaker.Breaker
}

// cbSearchScraper extends cbScraper to also implement domain.Searcher.
type cbSearchScraper struct{ *cbScraper }

// wrapCB wraps s with a circuit breaker. If s also implements domain.Searcher,
// the returned value implements it too, preserving optional capability checks.
func wrapCB(s domain.Scraper, cfg circuitbreaker.Config) domain.Scraper {
	cb := &cbScraper{inner: s, cb: circuitbreaker.New(cfg)}
	if _, ok := s.(domain.Searcher); ok {
		return &cbSearchScraper{cb}
	}
	return cb
}

func (c *cbScraper) Source() string { return c.inner.Source() }

func (c *cbScraper) FetchMangaDetail(ctx context.Context, id string) (*domain.Manga, error) {
	return circuitbreaker.Run(c.cb, func() (*domain.Manga, error) {
		return c.inner.FetchMangaDetail(ctx, id)
	})
}

func (c *cbScraper) FetchChapterList(ctx context.Context, mangaID string) ([]domain.Chapter, error) {
	return circuitbreaker.Run(c.cb, func() ([]domain.Chapter, error) {
		return c.inner.FetchChapterList(ctx, mangaID)
	})
}

func (c *cbScraper) FetchPageURLs(ctx context.Context, chapterID string) ([]string, error) {
	return circuitbreaker.Run(c.cb, func() ([]string, error) {
		return c.inner.FetchPageURLs(ctx, chapterID)
	})
}

func (c *cbSearchScraper) Search(ctx context.Context, query string) ([]domain.SearchResult, error) {
	return circuitbreaker.Run(c.cb, func() ([]domain.SearchResult, error) {
		return c.inner.(domain.Searcher).Search(ctx, query)
	})
}
