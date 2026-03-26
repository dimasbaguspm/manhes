package scraper

import (
	"sort"

	"manga-engine/internal/domain"
	"manga-engine/pkg/circuitbreaker"
)

var _ domain.SourceRegistry = (*Registry)(nil)

type entry struct {
	priority int
	scraper  domain.Scraper
}

// Registry holds scrapers ordered by priority (lower = higher priority).
type Registry struct {
	sources []entry
}

func (r *Registry) Register(priority int, s domain.Scraper) {
	r.sources = append(r.sources, entry{priority: priority, scraper: wrapCB(s, circuitbreaker.Default())})
	sort.Slice(r.sources, func(i, j int) bool {
		return r.sources[i].priority < r.sources[j].priority
	})
}

func (r *Registry) Ordered() []domain.Scraper {
	out := make([]domain.Scraper, len(r.sources))
	for i, e := range r.sources {
		out[i] = e.scraper
	}
	return out
}
