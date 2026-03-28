package application

import (
	"context"
	"fmt"
	"log/slog"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

var _ domain.WatchlistManager = (*WatchlistService)(nil)

// WatchlistServiceConfig holds dependencies for WatchlistService.
type WatchlistServiceConfig struct {
	Repo     domain.Repository
	Dict     *DictionaryService
	Registry domain.SourceRegistry
	Bus      domain.EventBus
	Cfg      *config.Config
}

type WatchlistService struct {
	repo     domain.Repository
	dict     *DictionaryService
	registry domain.SourceRegistry
	bus      domain.EventBus
	cfg      *config.Config
	log      *slog.Logger
}

func NewWatchlistService(cfg WatchlistServiceConfig) *WatchlistService {
	return &WatchlistService{
		repo:     cfg.Repo,
		dict:     cfg.Dict,
		registry: cfg.Registry,
		bus:      cfg.Bus,
		cfg:      cfg.Cfg,
		log:      slog.With("service", "watchlist"),
	}
}

func (s *WatchlistService) AddByDictionaryID(ctx context.Context, dictionaryID string) (string, error) {
	entry, found, err := s.repo.GetDictionary(ctx, dictionaryID)
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("dictionary entry not found: %s", dictionaryID)
	}

	wl := domain.WatchlistEntry{
		Slug:         entry.Slug,
		Title:        entry.Title,
		DictionaryID: dictionaryID,
		Sources:      entry.Sources,
	}
	if err := s.repo.AddWatchlist(ctx, wl); err != nil {
		return "", err
	}
	if err := s.repo.SetDictionaryState(ctx, dictionaryID, domain.StateFetching); err != nil {
		s.log.Warn("watchlist add: set dictionary state fetching", "id", dictionaryID, "err", err)
	}
	go s.dict.refresh(context.Background(), dictionaryID)

	sources := s.sourcesForIngest(entry)
	return entry.Slug, s.bus.Publish(ctx, s.cfg.Bus.IngestRequested, domain.IngestRequested{
		DictionaryID: entry.ID,
		Slug:        entry.Slug,
		Sources:     sources,
		LangToSource: entry.BestSource,
	})
}

func (s *WatchlistService) resolveSourcesForEntry(_ context.Context, entry *domain.WatchlistEntry) (sources, langToSource map[string]string, err error) {
	if entry.DictionaryID != "" {
		dictEntry, found, err := s.repo.GetDictionary(context.Background(), entry.DictionaryID)
		if err != nil {
			return nil, nil, err
		}
		if found {
			return s.sourcesForIngest(dictEntry), dictEntry.BestSource, nil
		}
	}
	return entry.Sources, nil, nil
}

// sourcesForIngest returns only sources that are best for at least one language,
// avoiding redundant downloads when BestSource is populated.
// Falls back to all sources before the first dictionary refresh has run.
func (s *WatchlistService) sourcesForIngest(entry domain.DictionaryEntry) map[string]string {
	if len(entry.Sources) == 0 {
		return map[string]string{}
	}
	if len(entry.BestSource) > 0 {
		result := make(map[string]string)
		for _, sourceName := range entry.BestSource {
			if id, ok := entry.Sources[sourceName]; ok {
				result[sourceName] = id
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return entry.Sources
}
