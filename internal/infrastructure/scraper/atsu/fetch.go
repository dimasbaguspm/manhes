package atsu

import (
	"context"
	"fmt"
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
		Sources:     map[string]string{"atsu": id},
		ScrapedAt:   time.Now(),
	}, nil
}

func (a *Adapter) FetchChapterList(ctx context.Context, mangaID string) ([]domain.Chapter, error) {
	u := fmt.Sprintf("%s/manga/allChapters?mangaId=%s", a.apiBase, mangaID)
	var resp allChaptersResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("atsu fetch chapters: %w", err)
	}

	sort.Slice(resp.Chapters, func(i, j int) bool {
		if resp.Chapters[i].Number != resp.Chapters[j].Number {
			return resp.Chapters[i].Number < resp.Chapters[j].Number
		}
		return resp.Chapters[i].CreatedAt < resp.Chapters[j].CreatedAt
	})

	seen := make(map[float64]struct{})
	var chapters []domain.Chapter
	for _, ch := range resp.Chapters {
		if _, ok := seen[ch.Number]; ok {
			continue
		}
		seen[ch.Number] = struct{}{}
		chapters = append(chapters, domain.Chapter{
			Number:    ch.Number,
			Title:     ch.Title,
			Language:  "en",
			Source:    "atsu",
			SourceID:  mangaID + ":" + ch.ID,
			ScrapedAt: time.Now(),
		})
	}
	return chapters, nil
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
