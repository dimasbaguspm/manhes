package atsu

import (
	"context"
	"fmt"
	"net/url"

	"manga-engine/internal/domain"
)

func (a *Adapter) Search(ctx context.Context, query string) ([]domain.SearchResult, error) {
	u := fmt.Sprintf("%s/search/page?query=%s", a.apiBase, url.QueryEscape(query))
	var resp searchResp
	if err := a.get(ctx, u, &resp); err != nil {
		return nil, fmt.Errorf("atsu search: %w", err)
	}

	out := make([]domain.SearchResult, 0, len(resp.Hits))
	for _, h := range resp.Hits {
		out = append(out, domain.SearchResult{
			ID:    h.ID,
			Title: h.Title,
			Type:  h.Type,
		})
	}
	return out, nil
}
