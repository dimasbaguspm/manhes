package atsu

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"manga-engine/internal/domain"
)

func (a *Adapter) FetchMangaDetail(ctx context.Context, id string) (*domain.Manga, error) {
	u := fmt.Sprintf("%s/manga/page?id=%s", a.apiBase, id)
	var resp mangaPageResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("atsu fetch manga: %w", err)
	}

	p := resp.MangaPage
	var genres []string
	for _, g := range p.Genres {
		genres = append(genres, g.Name)
	}
	var authors []string
	for _, au := range p.Authors {
		authors = append(authors, au.Name)
	}

	coverURL := ""
	if p.Poster != nil && p.Poster.LargeImage != "" {
		coverURL = a.staticBase + "/" + p.Poster.LargeImage
	}

	return &domain.Manga{
		Title:       p.Title,
		Description: p.Synopsis,
		Status:      strings.ToLower(p.Status),
		Authors:     authors,
		Genres:      genres,
		CoverURL:    coverURL,
	}, nil
}

func (a *Adapter) FetchChapterList(ctx context.Context, mangaID string) ([]domain.Chapter, error) {
	raw, err := a.fetchRawChapters(ctx, mangaID)
	if err != nil {
		return nil, err
	}

	sort.Slice(raw, func(i, j int) bool {
		return int64(raw[i].Number) < int64(raw[j].Number)
	})

	var chapters []domain.Chapter
	for i, ch := range raw {
		sourceID := fmt.Sprintf("%s:%s", ch.ScanlationMangaID, ch.ID)
		chapters = append(chapters, domain.Chapter{
			Number:    ch.Title,
			SortKey:   float64(i),
			Title:     ch.Title,
			Language:  "en",
			Source:    "atsu",
			SourceID:  sourceID,
			ScrapedAt: time.Now(),
		})
	}

	return chapters, nil
}

type rawChapterEntry struct {
	ID                string
	ScanlationMangaID string
	Title             string
	Number            float64
	CreatedAt         int64
	PageCount         int
}

func (a *Adapter) fetchRawChapters(ctx context.Context, mangaID string) ([]rawChapterEntry, error) {
	u := fmt.Sprintf("%s/manga/allChapters?mangaId=%s", a.apiBase, mangaID)
	var resp allChaptersResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("atsu fetch chapters: %w", err)
	}

	// use the first scanlation type, as atsu could return multiple scanlation type
	fixedScanlationID := resp.Chapters[0].ScanlationMangaID

	uniqueChapters := slices.DeleteFunc(resp.Chapters, func(i chapterItem) bool {
		return i.ScanlationMangaID != fixedScanlationID
	})

	out := make([]rawChapterEntry, len(uniqueChapters))

	for i, ch := range uniqueChapters {
		out[i] = rawChapterEntry{
			ID:                ch.ID,
			ScanlationMangaID: ch.ScanlationMangaID,
			Title:             ch.Title,
			Number:            ch.Number,
			CreatedAt:         ch.CreatedAt,
			PageCount:         ch.PageCount,
		}
	}

	return out, nil
}

func (a *Adapter) FetchPageURLs(ctx context.Context, sourceID string) ([]string, error) {
	parts := strings.SplitN(sourceID, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("atsu: invalid sourceID %q", sourceID)
	}
	mangaID, chapterID := parts[0], parts[1]

	u := fmt.Sprintf("%s/read/chapter?mangaId=%s&chapterId=%s", a.apiBase, mangaID, chapterID)
	var resp readChapterResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("atsu fetch pages: %w", err)
	}

	pages := resp.ReadChapter.Pages
	sort.Slice(pages, func(i, j int) bool {
		return pages[i].Number < pages[j].Number
	})

	urls := make([]string, len(pages))
	for i, p := range pages {
		if strings.HasPrefix(p.Image, "http") {
			urls[i] = p.Image
		} else {
			urls[i] = a.baseURL + p.Image
		}
	}
	return urls, nil
}
