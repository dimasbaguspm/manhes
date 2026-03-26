package mangadex

import (
	"net/http"

	"golang.org/x/time/rate"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

var _ domain.Scraper = (*Adapter)(nil)
var _ domain.Searcher = (*Adapter)(nil)

const coverBaseURL = "https://uploads.mangadex.org/covers"

type Adapter struct {
	baseURL string
	client  *http.Client
	limiter *rate.Limiter
}

func New(cfg config.MangadexConfig, client *http.Client) *Adapter {
	return &Adapter{
		baseURL: cfg.BaseURL,
		client:  client,
		limiter: rate.NewLimiter(rate.Limit(cfg.RateLimit), int(cfg.RateLimit)+1),
	}
}

func (a *Adapter) Source() string { return "mangadex" }
