package application

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"manga-engine/internal/domain"
	"manga-engine/internal/infrastructure/storage"
)

var _ domain.DictionaryManager = (*DictionaryService)(nil)

type DictionaryConfig struct {
	RefreshInterval time.Duration
}

type DictionaryService struct {
	repo      domain.Repository
	registry  domain.SourceRegistry
	dl        domain.Downloader
	s3c       domain.ObjectStore
	publisher domain.EventPublisher
	interval  time.Duration
	log       *slog.Logger
}

func NewDictionaryService(repo domain.Repository, registry domain.SourceRegistry, dl domain.Downloader, s3c domain.ObjectStore, publisher domain.EventPublisher, cfg DictionaryConfig) *DictionaryService {
	return &DictionaryService{repo: repo, registry: registry, dl: dl, s3c: s3c, publisher: publisher, interval: cfg.RefreshInterval, log: slog.With("service", "dictionary")}
}

func (s *DictionaryService) Search(ctx context.Context, query string) ([]domain.DictionaryEntry, error) {
	type hit struct {
		title   string
		sources map[string]string
	}
	bySlug := make(map[string]*hit)

	for _, src := range s.registry.Ordered() {
		searcher, ok := src.(domain.Searcher)
		if !ok {
			continue
		}
		results, err := searcher.Search(ctx, query)
		if err != nil {
			s.log.Warn("dictionary search: source failed", "source", src.Source(), "query", query, "err", err)
			continue
		}
		s.log.Info("dictionary search: source results", "source", src.Source(), "query", query, "count", len(results))
		for _, r := range results {
			slug := storage.Slugify(r.Title)
			if h, ok := bySlug[slug]; ok {
				h.sources[src.Source()] = r.ID
			} else {
				bySlug[slug] = &hit{
					title:   r.Title,
					sources: map[string]string{src.Source(): r.ID},
				}
			}
		}
	}

	var entries []domain.DictionaryEntry
	for slug, h := range bySlug {
		entry, err := s.Upsert(ctx, slug, h.title, h.sources)
		if err != nil {
			s.log.Warn("dictionary search: upsert failed", "slug", slug, "err", err)
			continue
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// Refresh re-fetches source stats for all known sources and searches 3rd-party
// scrapers for any new sources matching the entry's title. It then persists the
// merged result and returns the updated entry.
func (s *DictionaryService) Refresh(ctx context.Context, id string) (domain.DictionaryEntry, error) {
	entry, found, err := s.repo.GetDictionary(id)
	if err != nil {
		return domain.DictionaryEntry{}, err
	}
	if !found {
		return domain.DictionaryEntry{}, domain.ErrNotFound
	}

	// Discover new sources from 3rd-party scrapers using the entry's title.
	newSources := false
	for _, src := range s.registry.Ordered() {
		searcher, ok := src.(domain.Searcher)
		if !ok {
			continue
		}
		results, err := searcher.Search(ctx, entry.Title)
		if err != nil {
			s.log.Warn("dictionary refresh: search failed", "source", src.Source(), "title", entry.Title, "err", err)
			continue
		}
		for _, r := range results {
			if storage.Slugify(r.Title) == entry.Slug {
				if _, exists := entry.Sources[src.Source()]; !exists {
					if entry.Sources == nil {
						entry.Sources = make(map[string]string)
					}
					entry.Sources[src.Source()] = r.ID
					newSources = true
				}
				break
			}
		}
	}

	if newSources {
		if err := s.repo.UpsertDictionary(entry); err != nil {
			s.log.Warn("dictionary refresh: upsert new sources", "id", id, "err", err)
		}
	}

	// Refresh source stats (chapter counts, cover, best-source selection).
	if err := s.refresh(ctx, id); err != nil {
		return domain.DictionaryEntry{}, err
	}

	// Stamp updated_at on the manga catalog entry so the detail page reflects
	// that we've fetched from the source, regardless of whether new chapters
	// were found.
	// UpsertManga auto-stamps synced_at = CURRENT_TIMESTAMP, recording that
	// we fetched from the source regardless of whether new chapters exist.
	if manga, found, err := s.repo.GetMangaBySlug(entry.Slug); err == nil && found {
		if err := s.repo.UpsertManga(manga.Manga); err != nil {
			s.log.Warn("dictionary refresh: update manga updated_at", "slug", entry.Slug, "err", err)
		}
	}

	updated, found, err := s.repo.GetDictionary(id)
	if err != nil {
		return domain.DictionaryEntry{}, err
	}
	if !found {
		return domain.DictionaryEntry{}, domain.ErrNotFound
	}

	if err := s.publisher.PublishIngestRequested(ctx, domain.IngestRequested{
		Slug:         updated.Slug,
		Sources:      updated.Sources,
		LangToSource: updated.BestSource,
	}); err != nil {
		s.log.Warn("dictionary refresh: publish ingest", "id", id, "err", err)
	}

	return updated, nil
}

func (s *DictionaryService) Upsert(ctx context.Context, slug, title string, sources map[string]string) (domain.DictionaryEntry, error) {
	entry := domain.DictionaryEntry{
		ID:          uuid.New().String(),
		Slug:        slug,
		Title:       title,
		Sources:     sources,
		SourceStats: map[string]domain.SourceStat{},
		BestSource:  map[string]string{},
		State:       domain.StateUnavailable,
		CreatedAt:   time.Now(),
	}
	if err := s.repo.UpsertDictionary(entry); err != nil {
		return domain.DictionaryEntry{}, err
	}

	// Read back to get canonical entry: actual ID, merged sources, preserved state/stats.
	result, found, err := s.repo.GetDictionaryBySlug(slug)
	if err != nil {
		return domain.DictionaryEntry{}, err
	}
	if !found {
		return domain.DictionaryEntry{}, nil
	}
	go s.refresh(context.Background(), result.ID)
	return result, nil
}
