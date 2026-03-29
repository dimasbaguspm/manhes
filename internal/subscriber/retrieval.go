package subscriber

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"manga-engine/config"
	"manga-engine/internal/domain"
)

// RetrievalSubscriberConfig holds dependencies for RetrievalSubscriber.
type RetrievalSubscriberConfig struct {
	Repo     domain.Repository
	Registry domain.SourceRegistry
	Bus      domain.EventBus
	Cfg      *config.Config
}

// RetrievalSubscriber reacts to IngestRequested events, fetches manga detail and
// chapter lists from sources, diffs against stored chapters, and publishes
// ChaptersFound for chapters that need processing.
type RetrievalSubscriber struct {
	repo     domain.Repository
	registry domain.SourceRegistry
	bus      domain.EventBus
	cfg      *config.Config
	log      *slog.Logger
}

func NewRetrievalSubscriber(cfg RetrievalSubscriberConfig) *RetrievalSubscriber {
	return &RetrievalSubscriber{
		repo:     cfg.Repo,
		registry: cfg.Registry,
		bus:      cfg.Bus,
		cfg:      cfg.Cfg,
		log:      slog.With("component", "subscriber"),
	}
}

// HandleIngestRequested reacts to IngestRequested, fetches manga detail and
// chapter lists from sources, and publishes ChaptersFound for new chapters.
func (h *RetrievalSubscriber) HandleIngestRequested(ctx context.Context, e domain.IngestRequested) error {
	h.log.Info("[Retrieval Subscriber]: HandleIngestRequested: received event",
		"dictionaryID", e.DictionaryID,
		"mangaID", e.MangaID,
	)

	// Step 0: Generate or validate mangaID.
	mangaID := e.MangaID
	if mangaID == "" {
		mangaID = uuid.NewString()
	}

	// Step 1: Validate dictionary entry exists.
	entry, found, err := h.repo.GetDictionary(ctx, e.DictionaryID)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}

	// Step 2: Resolve source map and per-language best source.
	sourceMap := e.Sources
	if len(sourceMap) == 0 {
		sourceMap = entry.Sources
	}
	langToSrc := e.LangToSource
	if len(langToSrc) == 0 {
		langToSrc = entry.BestSource
	}
	if len(sourceMap) == 0 {
		h.log.Warn("[Retrieval Subscriber]: no sources available", "dictionaryID", e.DictionaryID)
		return nil
	}

	// Step 3: Fetch manga detail and chapter list concurrently per source.
	byLang := h.fetchFromSources(ctx, mangaID, e.DictionaryID, sourceMap, langToSrc)

	// Step 4: Diff chapters per language and collect only new chapters.
	newChaptersByLang := make(map[string][]domain.Chapter)
	for lang, chapters := range byLang {
		stored, err := h.repo.GetChaptersByLang(ctx, mangaID, lang)
		if err != nil {
			h.log.Warn("[Retrieval Subscriber]: get stored chapters", "mangaID", mangaID, "lang", lang, "err", err)
			continue
		}
		newChapters := diffChapters(stored, chapters)
		if len(newChapters) > 0 {
			newChaptersByLang[lang] = newChapters
		}
	}

	// Publish ChaptersFound if there are new chapters to process.
	if len(newChaptersByLang) > 0 {
		if err := h.bus.Publish(ctx, h.cfg.Bus.ChaptersFound, domain.ChaptersFound{
			DictionaryID: e.DictionaryID,
			MangaID:      mangaID,
			Chapters:     newChaptersByLang,
		}); err != nil {
			h.log.Warn("[Retrieval Subscriber]: publish ChaptersFound", "mangaID", mangaID, "err", err)
		}
	}

	return nil
}

// fetchFromSources queries all sources concurrently, groups chapters by language
// (respecting LangToSource), and returns the merged chapter list keyed by language.
func (h *RetrievalSubscriber) fetchFromSources(
	ctx context.Context,
	mangaID string,
	_ string, // dictionaryID (reserved for future use)
	sourceMap map[string]string,
	langToSrc map[string]string,
) map[string][]domain.Chapter {
	type result struct {
		srcName  string
		chapters []domain.Chapter
		err      error
	}

	resultCh := make(chan result, len(sourceMap))
	for srcName, srcID := range sourceMap {
		go func(srcName, srcID string) {
			src := h.scraperFor(srcName)
			if src == nil {
				resultCh <- result{srcName: srcName, err: fmt.Errorf("scraper not registered: %s", srcName)}
				return
			}

			chapters, err := src.FetchChapterList(ctx, srcID)
			resultCh <- result{srcName: srcName, chapters: chapters, err: err}
		}(srcName, srcID)
	}

	byLang := make(map[string][]domain.Chapter)
	for range sourceMap {
		res := <-resultCh
		if res.err != nil {
			h.log.Warn("[Retrieval Subscriber]: fetchFromSources: source error", "mangaID", mangaID, "source", res.srcName, "err", res.err)
			continue
		}
		for _, ch := range res.chapters {
			if len(langToSrc) > 0 {
				if bestSrc, ok := langToSrc[ch.Language]; ok && bestSrc != res.srcName {
					continue
				}
			}
			ch.MangaID = mangaID
			byLang[ch.Language] = append(byLang[ch.Language], ch)
		}
	}
	return byLang
}

// scraperFor returns the scraper registered under the given source name, or nil.
func (h *RetrievalSubscriber) scraperFor(name string) domain.Scraper {
	for _, candidate := range h.registry.Ordered() {
		if candidate.Source() == name {
			return candidate
		}
	}
	return nil
}

// diffChapters returns chapters from fetched that are not already in stored.
func diffChapters(stored, fetched []domain.Chapter) []domain.Chapter {
	storedSet := make(map[float64]bool)
	for _, sc := range stored {
		storedSet[sc.SortKey] = true
	}
	var new []domain.Chapter
	for _, ch := range fetched {
		if !storedSet[ch.SortKey] {
			new = append(new, ch)
		}
	}
	return new
}
