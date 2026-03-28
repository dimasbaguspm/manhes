package application

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"manga-engine/config"
	"manga-engine/internal/domain"
	"manga-engine/internal/infrastructure/storage"
)

var _ domain.DictionaryManager = (*DictionaryService)(nil)

// DictionaryServiceConfig holds dependencies and config for DictionaryService.
type DictionaryServiceConfig struct {
	Repo     domain.Repository
	Registry domain.SourceRegistry
	DL       domain.Downloader
	S3       domain.ObjectStore
	Bus      domain.EventBus
	Cfg      *config.Config
}

type DictionaryService struct {
	repo     domain.Repository
	registry domain.SourceRegistry
	dl       domain.Downloader
	s3c      domain.ObjectStore
	bus      domain.EventBus
	cfg      *config.Config
	log      *slog.Logger
}

func NewDictionaryService(cfg DictionaryServiceConfig) *DictionaryService {
	return &DictionaryService{
		repo:     cfg.Repo,
		registry: cfg.Registry,
		dl:       cfg.DL,
		s3c:      cfg.S3,
		bus:      cfg.Bus,
		cfg:      cfg.Cfg,
		log:      slog.With("service", "dictionary"),
	}
}

type searchHit struct {
	slug        string
	title       string
	sources     map[string]string
	sourceStats map[string]domain.SourceStat
}

func (s *DictionaryService) Search(ctx context.Context, query string) ([]domain.DictionaryEntry, error) {
	hits := make(map[string]*searchHit)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, src := range s.registry.Ordered() {
		searcher, ok := src.(domain.Searcher)
		if !ok {
			continue
		}
		wg.Add(1)
		go func(searcher domain.Searcher, scraperSrc domain.Scraper) {
			defer wg.Done()
			results, err := searcher.Search(ctx, query)
			if err != nil {
				s.log.Warn("search: source failed", "source", scraperSrc.Source(), "err", err)
				return
			}
			s.log.Info("search: source results", "source", scraperSrc.Source(), "query", query, "count", len(results))
			for _, r := range results {
				slug := storage.Slugify(r.Title)
				stat := s.fetchSourceStat(ctx, scraperSrc, r.ID)

				mu.Lock()
				if h, ok := hits[slug]; ok {
					h.sources[scraperSrc.Source()] = r.ID
					h.sourceStats[scraperSrc.Source()] = stat
				} else {
					hits[slug] = &searchHit{
						slug:        slug,
						title:       r.Title,
						sources:     map[string]string{scraperSrc.Source(): r.ID},
						sourceStats: map[string]domain.SourceStat{scraperSrc.Source(): stat},
					}
				}
				mu.Unlock()
			}
		}(searcher, src)
	}
	wg.Wait()

	// Build entries and batch upsert
	entries := make([]domain.DictionaryEntry, 0, len(hits))
	now := time.Now()
	for _, h := range hits {
		entries = append(entries, domain.DictionaryEntry{
			ID:          uuid.New().String(),
			Slug:        h.slug,
			Title:       h.title,
			Sources:     h.sources,
			SourceStats: h.sourceStats,
			BestSource:  map[string]string{},
			CoverURL:    "",
			UpdatedAt:   &now,
			CreatedAt:   now,
		})
	}

	if err := s.repo.UpsertDictionaryBatch(ctx, entries); err != nil {
		return nil, err
	}

	// Publish dictionary.updated for each upserted entry so MangaService can sync metadata.
	// TriggerIngest=false because Search does not trigger chapter ingestion.
	for _, entry := range entries {
		if err := s.bus.Publish(ctx, s.cfg.Bus.DictionaryUpdated, domain.DictionaryUpdated{
			DictionaryID:  entry.ID,
			TriggerIngest: false,
		}); err != nil {
			s.log.Warn("search: publish dictionary.updated", "id", entry.ID, "err", err)
		}
	}

	return entries, nil
}

// Refresh re-fetches source stats for all known sources and searches 3rd-party
// scrapers for any new sources matching the entry's title. It then persists the
// merged result and returns the updated entry.
func (s *DictionaryService) Refresh(ctx context.Context, id string) (domain.DictionaryEntry, error) {
	entry, found, err := s.repo.GetDictionary(ctx, id)
	if err != nil {
		return domain.DictionaryEntry{}, err
	}
	if !found {
		return domain.DictionaryEntry{}, domain.ErrNotFound
	}

	// Discover new sources concurrently from 3rd-party scrapers using the entry's title.
	type discoveredSource struct {
		sourceName string
		sourceID   string
	}
	var discovered []discoveredSource
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, src := range s.registry.Ordered() {
		searcher, ok := src.(domain.Searcher)
		if !ok {
			continue
		}
		wg.Add(1)
		go func(searcher domain.Searcher, scraperSrc domain.Scraper) {
			defer wg.Done()
			results, err := searcher.Search(ctx, entry.Title)
			if err != nil {
				s.log.Warn("dictionary refresh: search failed", "source", scraperSrc.Source(), "title", entry.Title, "err", err)
				return
			}
			for _, r := range results {
				if storage.Slugify(r.Title) == entry.Slug {
					mu.Lock()
					discovered = append(discovered, discoveredSource{
						sourceName: scraperSrc.Source(),
						sourceID:   r.ID,
					})
					mu.Unlock()
				}
			}
		}(searcher, src)
	}
	wg.Wait()

	// Capture state BEFORE merge for change detection.
	oldSources := entry.Sources

	// Merge newly discovered sources.
	for _, ds := range discovered {
		if _, exists := entry.Sources[ds.sourceName]; !exists {
			if entry.Sources == nil {
				entry.Sources = make(map[string]string)
			}
			entry.Sources[ds.sourceName] = ds.sourceID
		}
	}

	// Refresh source stats (chapter counts, cover, best-source selection).
	// Pass oldSources so entryChanged can compare pre- vs post-merge state.
	if _, err := s.refreshWithOldState(ctx, id, oldSources); err != nil {
		return domain.DictionaryEntry{}, err
	}

	updated, found, err := s.repo.GetDictionary(ctx, id)
	if err != nil {
		return domain.DictionaryEntry{}, err
	}
	if !found {
		return domain.DictionaryEntry{}, domain.ErrNotFound
	}

	// Publish dictionary.updated so manga service can sync metadata and (if TriggerIngest=true) trigger ingestion.
	if err := s.bus.Publish(ctx, s.cfg.Bus.DictionaryUpdated, domain.DictionaryUpdated{
		DictionaryID:  updated.ID,
		TriggerIngest: true, // Refresh always triggers ingestion so new chapters get synced
	}); err != nil {
		s.log.Warn("dictionary refresh: publish dictionary.updated", "id", id, "err", err)
	}

	return updated, nil
}
