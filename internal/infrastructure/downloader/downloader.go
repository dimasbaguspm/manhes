package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"manga-engine/internal/domain"
)

var _ domain.Downloader = (*Downloader)(nil)

type Downloader struct {
	client *http.Client
}

func New(client *http.Client) *Downloader {
	return &Downloader{client: client}
}

func (d *Downloader) Download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
