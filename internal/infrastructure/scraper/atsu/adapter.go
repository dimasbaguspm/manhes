package atsu

import (
	"net/http"
	"strings"

	"golang.org/x/time/rate"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

var _ domain.Scraper = (*Adapter)(nil)
var _ domain.Searcher = (*Adapter)(nil)

type Adapter struct {
	baseURL    string
	apiBase    string
	staticBase string
	client     *http.Client
	limiter    *rate.Limiter
}

func New(cfg config.AtsuConfig, client *http.Client) *Adapter {
	base := strings.TrimRight(cfg.BaseURL, "/")
	return &Adapter{
		baseURL:    base,
		apiBase:    base + "/api",
		staticBase: base + "/static",
		client:     client,
		limiter:    rate.NewLimiter(rate.Limit(cfg.RateLimit), int(cfg.RateLimit)+1),
	}
}

func (a *Adapter) Source() string { return "atsu" }
