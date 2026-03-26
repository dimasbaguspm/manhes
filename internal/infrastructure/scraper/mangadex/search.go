package mangadex

import (
	"context"
	"fmt"
	"net/url"

	"manga-engine/internal/domain"
)

func (a *Adapter) Search(ctx context.Context, query string) ([]domain.SearchResult, error) {
	u := fmt.Sprintf("%s/manga?title=%s&limit=10&order[relevance]=desc", a.baseURL, url.QueryEscape(query))
	var resp searchResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("mangadex search: %w", err)
	}

	out := make([]domain.SearchResult, 0, len(resp.Data))
	for _, m := range resp.Data {
		attr := m.Attributes
		title := firstOf(attr.Title, "en", "ja-ro", "ja")
		var genres []string
		for _, tag := range attr.Tags {
			if name := firstOf(tag.Attributes.Name, "en"); name != "" {
				genres = append(genres, name)
			}
		}
		out = append(out, domain.SearchResult{
			ID:     m.ID,
			Title:  title,
			Status: attr.Status,
			Genres: genres,
			Type:   "Manga",
		})
	}
	return out, nil
}
